// Playlist and track helpers - port of internal/helpers/playlist.go

import type { SimplifiedPlaylist, Artist } from '@/types/spotify'

const OPEN_TRACK_URL_RE = /^https?:\/\/open\.spotify\.com\/track\/([A-Za-z0-9]+)(\?.*)?$/i
const OPEN_PLAYLIST_URL_RE = /^https?:\/\/open\.spotify\.com\/playlist\/([A-Za-z0-9]+)(\?.*)?$/i
const SPOTIFY_PLAYLIST_URI_RE = /^spotify:playlist:([A-Za-z0-9]+)$/i

/**
 * Join artist names into a comma-separated string.
 * Port of JoinArtistNames from playlist.go:18-27
 */
export function joinArtistNames(artists: Artist[]): string {
  const names = artists.map((a) => a.name).filter((n) => n !== '')
  return names.join(', ')
}

/**
 * Convert a track ID, URI, or URL to a Spotify track URI.
 * Port of NormalizeTrackURI from playlist.go:29-46
 * @throws Error if input is empty or unsupported URL format
 */
export function normalizeTrackURI(s: string): string {
  s = s.trim()
  if (s === '') {
    throw new Error('empty track value')
  }

  if (s.toLowerCase().startsWith('spotify:track:')) {
    return s
  }

  const match = OPEN_TRACK_URL_RE.exec(s)
  if (match) {
    return `spotify:track:${match[1]}`
  }

  // Check if it's a URL we don't support
  if (isURL(s)) {
    throw new Error(`unsupported track url: ${s}`)
  }

  // Treat as raw track ID
  return `spotify:track:${s}`
}

/**
 * Convert a playlist ID, URI, or URL to a Spotify playlist ID.
 * Port of NormalizePlaylistID from playlist.go:48-65
 * @throws Error if input is empty or unsupported URL format
 */
export function normalizePlaylistID(s: string): string {
  s = s.trim()
  if (s === '') {
    throw new Error('empty playlist id')
  }

  // spotify:playlist:ID
  let match = SPOTIFY_PLAYLIST_URI_RE.exec(s)
  if (match) {
    return match[1]
  }

  // https://open.spotify.com/playlist/ID
  match = OPEN_PLAYLIST_URL_RE.exec(s)
  if (match) {
    return match[1]
  }

  // Check if it's a URL we don't support
  if (isURL(s)) {
    throw new Error(`unsupported playlist url: ${s}`)
  }

  // Treat as raw playlist ID
  return s
}

/**
 * Filter playlists by name (case-insensitive substring match).
 * Port of FilterPlaylistsByName from playlist.go:67-80
 */
export function filterPlaylistsByName(
  playlists: SimplifiedPlaylist[],
  filter: string
): SimplifiedPlaylist[] {
  const want = filter.toLowerCase().trim()
  if (want === '') {
    return playlists
  }
  return playlists.filter((p) => p.name.toLowerCase().includes(want))
}

/**
 * Check if a string looks like a URL.
 */
function isURL(s: string): boolean {
  try {
    const u = new URL(s)
    return u.protocol !== ''
  } catch {
    return false
  }
}
