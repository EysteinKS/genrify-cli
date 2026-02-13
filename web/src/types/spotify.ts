// Spotify API types - ported from internal/spotify/types.go

export interface User {
  id: string
  display_name: string
}

export interface SimplifiedPlaylist {
  id: string
  name: string
  description: string
  public: boolean
  collaborative: boolean
  owner: User
  tracks: {
    total: number
  }
}

export interface Paging<T> {
  href: string
  items: T[]
  limit: number
  next: string | null
  offset: number
  previous: string | null
  total: number
}

export interface Artist {
  id: string
  name: string
}

export interface Album {
  id: string
  name: string
}

export interface FullTrack {
  id: string
  name: string
  uri: string
  artists: Artist[]
  album: Album
}

export interface PlaylistTrackItem {
  track: FullTrack
}

export interface SnapshotResponse {
  snapshot_id: string
}

export interface SpotifyApiError {
  error: {
    status: number
    message: string
  }
}
