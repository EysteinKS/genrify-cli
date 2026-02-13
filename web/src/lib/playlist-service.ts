// Playlist service - port of internal/playlist/service.go

import type { SimplifiedPlaylist } from '@/types/spotify'
import type { SpotifyClient } from './spotify-client'

export interface MergeOptions {
  deduplicate: boolean
  public: boolean
  description: string
}

export interface MergeResult {
  newPlaylistId: string
  trackCount: number
  duplicatesRemoved: number
  verified: boolean
  missingURIs: string[]
}

export const ErrNoPlaylistsMatched = new Error('no playlists matched pattern')

/**
 * Playlist service for high-level playlist operations.
 * Port of Service from service.go
 */
export class PlaylistService {
  constructor(private client: SpotifyClient) {}

  /**
   * Find playlists matching a regex pattern.
   * Port of FindPlaylistsByPattern from service.go:38-63
   */
  async findPlaylistsByPattern(pattern: string): Promise<SimplifiedPlaylist[]> {
    pattern = pattern.trim()
    if (!pattern) {
      throw new Error('pattern is required')
    }

    let re: RegExp
    try {
      re = new RegExp(pattern)
    } catch (err) {
      throw new Error(`invalid pattern: ${err}`)
    }

    const playlists = await this.client.listCurrentUserPlaylists(0)

    const matched = playlists.filter((p) => re.test(p.name))
    if (matched.length === 0) {
      throw ErrNoPlaylistsMatched
    }

    return matched
  }

  /**
   * Merge multiple playlists into a new playlist.
   * Port of MergePlaylists from service.go:65-129
   */
  async mergePlaylists(
    sourceIds: string[],
    targetName: string,
    opts: MergeOptions,
    onProgress?: (message: string) => void
  ): Promise<MergeResult> {
    targetName = targetName.trim()
    if (!targetName) {
      throw new Error('target name is required')
    }
    if (sourceIds.length === 0) {
      throw new Error('at least one source playlist is required')
    }

    // Collect tracks first so we can fail early without creating anything
    onProgress?.('Collecting tracks from source playlists...')
    let uris: string[] = []
    for (const id of sourceIds) {
      const trimmedId = id.trim()
      if (!trimmedId) continue

      const tracks = await this.client.listPlaylistTracks(trimmedId, 0)
      for (const t of tracks) {
        if (t.uri) {
          uris.push(t.uri)
        }
      }
    }

    let duplicatesRemoved = 0
    if (opts.deduplicate) {
      onProgress?.('Deduplicating tracks...')
      const result = deduplicate(uris)
      uris = result.kept
      duplicatesRemoved = result.duplicates
    }

    onProgress?.('Creating target playlist...')
    const pl = await this.client.createPlaylist(targetName, opts.description, opts.public)

    // Best-effort rollback if we fail after creating the playlist
    const rollback = async () => {
      try {
        await this.client.deletePlaylist(pl.id)
      } catch {
        // Ignore rollback errors
      }
    }

    if (uris.length > 0) {
      onProgress?.(`Adding ${uris.length} tracks...`)
      try {
        await this.client.addTracksToPlaylist(pl.id, uris)
      } catch (err) {
        await rollback()
        throw new Error(`add tracks: ${err}`)
      }
    }

    onProgress?.('Verifying playlist contents...')
    const [verified, missingURIs] = await this.verifyPlaylistContents(pl.id, uris)
    if (!verified) {
      await rollback()
      throw new Error('verification failed')
    }

    return {
      newPlaylistId: pl.id,
      trackCount: uris.length,
      duplicatesRemoved,
      verified,
      missingURIs,
    }
  }

  /**
   * Verify playlist contents match expected URIs.
   * Port of VerifyPlaylistContents from service.go:131-180
   */
  async verifyPlaylistContents(
    playlistId: string,
    expectedURIs: string[]
  ): Promise<[boolean, string[]]> {
    playlistId = playlistId.trim()
    if (!playlistId) {
      throw new Error('playlist id is required')
    }

    const expected = new Set<string>()
    for (const u of expectedURIs) {
      const trimmed = u.trim()
      if (trimmed) {
        expected.add(trimmed)
      }
    }

    if (expected.size === 0) {
      return [true, []]
    }

    // Spotify can be eventually consistent; retry a couple times
    for (let attempt = 0; attempt < 3; attempt++) {
      const tracks = await this.client.listPlaylistTracks(playlistId, 0)
      const seen = new Set<string>()
      for (const t of tracks) {
        if (t.uri) {
          seen.add(t.uri)
        }
      }

      const missing: string[] = []
      for (const u of expected) {
        if (!seen.has(u)) {
          missing.push(u)
        }
      }

      if (missing.length === 0) {
        return [true, []]
      }

      if (attempt < 2) {
        await this.sleep(200)
        continue
      }

      return [false, missing]
    }

    return [false, []]
  }

  /**
   * Delete multiple playlists.
   * Port of DeletePlaylists from service.go:182-197
   */
  async deletePlaylists(playlistIds: string[]): Promise<void> {
    for (const id of playlistIds) {
      const trimmedId = id.trim()
      if (!trimmedId) continue

      try {
        await this.client.deletePlaylist(trimmedId)
      } catch (err) {
        // Check for 403 (permission denied)
        if (err instanceof Error && err.message.includes('403')) {
          throw new Error(`delete playlist ${trimmedId}: permission denied`)
        }
        throw new Error(`delete playlist ${trimmedId}: ${err}`)
      }
    }
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }
}

/**
 * Remove duplicate URIs while preserving order.
 * Port of deduplicate from service.go:199-216
 */
function deduplicate(uris: string[]): { kept: string[]; duplicates: number } {
  const seen = new Set<string>()
  const out: string[] = []
  let dupes = 0

  for (const u of uris) {
    const trimmed = u.trim()
    if (!trimmed) continue

    if (seen.has(trimmed)) {
      dupes++
      continue
    }

    seen.add(trimmed)
    out.push(trimmed)
  }

  return { kept: out, duplicates: dupes }
}
