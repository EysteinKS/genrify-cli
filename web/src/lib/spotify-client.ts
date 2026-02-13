// Spotify API client - port of internal/spotify/client.go

import type {
  User,
  SimplifiedPlaylist,
  FullTrack,
  Paging,
  PlaylistTrackItem,
  SnapshotResponse,
  SpotifyApiError,
} from '@/types/spotify'

const BASE_URL = 'https://api.spotify.com/v1'

/**
 * Spotify API client with automatic token refresh and retry logic.
 * Port of Client from client.go
 */
export class SpotifyClient {
  private getAccessToken: () => Promise<string>
  private forceRefresh: () => Promise<string>

  constructor(getAccessToken: () => Promise<string>, forceRefresh: () => Promise<string>) {
    this.getAccessToken = getAccessToken
    this.forceRefresh = forceRefresh
  }

  /**
   * Get current user profile.
   * Port of GetMe from client.go:80-86
   */
  async getMe(): Promise<User> {
    return this.doJSON<User>('GET', '/me')
  }

  /**
   * List current user's playlists (all pages up to max).
   * Port of ListCurrentUserPlaylists from client.go:88-102
   */
  async listCurrentUserPlaylists(max = 0): Promise<SimplifiedPlaylist[]> {
    const pageSize = 50
    return this.collectPaged<SimplifiedPlaylist>(
      pageSize,
      max,
      async (limit, offset) => {
        const params = new URLSearchParams({
          limit: limit.toString(),
          offset: offset.toString(),
        })
        return this.doJSON<Paging<SimplifiedPlaylist>>('GET', `/me/playlists?${params}`)
      }
    )
  }

  /**
   * List tracks in a playlist (all pages up to max).
   * Port of ListPlaylistTracks from client.go:104-132
   */
  async listPlaylistTracks(playlistId: string, max = 0): Promise<FullTrack[]> {
    playlistId = playlistId.trim()
    if (!playlistId) {
      throw new Error('playlist id is required')
    }

    const pageSize = 100
    const items = await this.collectPaged<PlaylistTrackItem>(
      pageSize,
      max,
      async (limit, offset) => {
        const params = new URLSearchParams({
          limit: limit.toString(),
          offset: offset.toString(),
        })
        return this.doJSON<Paging<PlaylistTrackItem>>(
          'GET',
          `/playlists/${encodeURIComponent(playlistId)}/tracks?${params}`
        )
      }
    )

    // Filter out null tracks
    const tracks: FullTrack[] = []
    for (const item of items) {
      if (item.track && item.track.uri) {
        tracks.push(item.track)
      }
    }
    return tracks
  }

  /**
   * Create a new playlist for the current user.
   * Port of CreatePlaylist from client.go:134-154
   */
  async createPlaylist(
    name: string,
    description: string,
    isPublic: boolean
  ): Promise<SimplifiedPlaylist> {
    name = name.trim()
    if (!name) {
      throw new Error('name is required')
    }

    const body = {
      name,
      public: isPublic,
      description,
    }

    // Spotify supports creating playlists via /me
    return this.doJSON<SimplifiedPlaylist>('POST', '/me/playlists', body)
  }

  /**
   * Add tracks to a playlist (batches in groups of 100).
   * Port of AddTracksToPlaylist from client.go:156-187
   */
  async addTracksToPlaylist(playlistId: string, uris: string[]): Promise<string> {
    playlistId = playlistId.trim()
    if (!playlistId) {
      throw new Error('playlist id is required')
    }

    const clean = uris.map((u) => u.trim()).filter((u) => u !== '')
    if (clean.length === 0) {
      throw new Error('at least one track uri is required')
    }

    const endpoint = `/playlists/${encodeURIComponent(playlistId)}/tracks`
    let lastSnapshot = ''

    // Batch in groups of 100
    for (let i = 0; i < clean.length; i += 100) {
      const batch = clean.slice(i, i + 100)
      const body = { uris: batch }
      const resp = await this.doJSON<SnapshotResponse>('POST', endpoint, body)
      lastSnapshot = resp.snapshot_id
    }

    return lastSnapshot
  }

  /**
   * Get a single playlist by ID.
   * Port of GetPlaylist from client.go:189-200
   */
  async getPlaylist(playlistId: string): Promise<SimplifiedPlaylist> {
    playlistId = playlistId.trim()
    if (!playlistId) {
      throw new Error('playlist id is required')
    }
    return this.doJSON<SimplifiedPlaylist>('GET', `/playlists/${encodeURIComponent(playlistId)}`)
  }

  /**
   * Delete (unfollow) a playlist.
   * Port of DeletePlaylist from client.go:202-212
   */
  async deletePlaylist(playlistId: string): Promise<void> {
    playlistId = playlistId.trim()
    if (!playlistId) {
      throw new Error('playlist id is required')
    }
    await this.doJSON('DELETE', `/playlists/${encodeURIComponent(playlistId)}/followers`)
  }

  /**
   * Remove tracks from a playlist.
   * Port of RemoveTracksFromPlaylist from client.go:214-239
   */
  async removeTracksFromPlaylist(playlistId: string, uris: string[]): Promise<string> {
    playlistId = playlistId.trim()
    if (!playlistId) {
      throw new Error('playlist id is required')
    }

    const tracks = uris
      .map((u) => u.trim())
      .filter((u) => u !== '')
      .map((u) => ({ uri: u }))

    if (tracks.length === 0) {
      throw new Error('at least one track uri is required')
    }

    const body = { tracks }
    const resp = await this.doJSON<SnapshotResponse>(
      'DELETE',
      `/playlists/${encodeURIComponent(playlistId)}/tracks`,
      body
    )
    return resp.snapshot_id
  }

  /**
   * Collect all pages from a paginated endpoint.
   * Port of collectPaged from paging.go
   */
  private async collectPaged<T>(
    pageSize: number,
    max: number,
    fetch: (limit: number, offset: number) => Promise<Paging<T>>
  ): Promise<T[]> {
    if (max < 0) {
      throw new Error('max must be >= 0')
    }

    let limit = pageSize
    if (max > 0 && max < limit) {
      limit = max
    }

    const out: T[] = []
    let offset = 0

    while (true) {
      const page = await fetch(limit, offset)
      out.push(...page.items)

      if (max > 0 && out.length >= max) {
        return out.slice(0, max)
      }
      if (!page.next || page.items.length === 0) {
        return out
      }

      offset += limit

      if (max > 0) {
        const remaining = max - out.length
        if (remaining < limit) {
          limit = remaining
        }
      }
    }
  }

  /**
   * Execute a JSON API request with retry logic.
   * Port of doJSONWithRetry from client.go:245-323
   */
  private async doJSON<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = `${BASE_URL}${path}`
    let bodyBytes: string | undefined
    if (body !== undefined) {
      bodyBytes = JSON.stringify(body)
    }

    let refreshed = false
    let rateRetries = 0

    while (true) {
      const headers: Record<string, string> = {
        Accept: 'application/json',
      }

      const accessToken = await this.getAccessToken()
      headers['Authorization'] = `Bearer ${accessToken}`

      if (bodyBytes) {
        headers['Content-Type'] = 'application/json'
      }

      const resp = await fetch(url, {
        method,
        headers,
        body: bodyBytes,
      })

      const respBody = await resp.text()

      // 401: try refresh once and retry
      if (resp.status === 401 && !refreshed) {
        try {
          await this.forceRefresh()
          refreshed = true
          continue
        } catch {
          // Refresh failed, fall through to error
        }
      }

      // 429: rate limit with exponential backoff (up to 5 retries)
      if (resp.status === 429 && rateRetries < 5) {
        const wait = this.retryAfterDuration(resp.headers.get('Retry-After'), rateRetries)
        rateRetries++
        await this.sleep(wait)
        continue
      }

      // Error response
      if (!resp.ok) {
        throw this.decodeAPIError(respBody, resp.status)
      }

      // Success - no response body expected
      if (!respBody) {
        return undefined as T
      }

      return JSON.parse(respBody) as T
    }
  }

  /**
   * Decode Spotify API error from response body.
   * Port of decodeAPIError from errors.go:29-41
   */
  private decodeAPIError(body: string, fallbackStatus: number): Error {
    try {
      const parsed = JSON.parse(body) as SpotifyApiError
      if (parsed.error) {
        const status = parsed.error.status || fallbackStatus
        const message = parsed.error.message || ''
        if (message) {
          return new Error(`Spotify API error: HTTP ${status}: ${message}`)
        }
        return new Error(`Spotify API error: HTTP ${status}`)
      }
    } catch {
      // Not valid JSON or unexpected structure
    }
    return new Error(`Spotify API error: HTTP ${fallbackStatus}`)
  }

  /**
   * Calculate retry delay for rate limiting.
   * Port of retryAfterDuration from client.go:326-342
   */
  private retryAfterDuration(headerVal: string | null, attempt: number): number {
    if (headerVal) {
      const secs = parseInt(headerVal.trim(), 10)
      if (!isNaN(secs) && secs >= 0) {
        return secs === 0 ? 0 : secs * 1000
      }
    }

    // Exponential backoff with cap
    const base = 250
    const d = base * Math.pow(2, attempt)
    return Math.min(d, 5000)
  }

  /**
   * Sleep for specified milliseconds.
   */
  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }
}
