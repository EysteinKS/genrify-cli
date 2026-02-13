// Query hook for playlist tracks

import { useQuery } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useAuth } from '@/contexts/AuthContext'

export function usePlaylistTracks(playlistId: string | null, max = 0) {
  const client = useSpotifyClient()
  const { isLoggedIn } = useAuth()

  return useQuery({
    queryKey: ['playlist-tracks', playlistId, max],
    queryFn: () => client.listPlaylistTracks(playlistId!, max),
    enabled: isLoggedIn && !!playlistId,
  })
}
