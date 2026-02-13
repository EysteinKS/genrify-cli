// Query hook for single playlist details

import { useQuery } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useAuth } from '@/contexts/AuthContext'

export function usePlaylist(playlistId: string | null) {
  const client = useSpotifyClient()
  const { isLoggedIn } = useAuth()

  return useQuery({
    queryKey: ['playlist', playlistId],
    queryFn: () => client.getPlaylist(playlistId!),
    enabled: isLoggedIn && !!playlistId,
  })
}
