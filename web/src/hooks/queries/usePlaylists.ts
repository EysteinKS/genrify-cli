// Query hook for user's playlists

import { useQuery } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useAuth } from '@/contexts/AuthContext'

export function usePlaylists(max = 50) {
  const client = useSpotifyClient()
  const { isLoggedIn } = useAuth()

  return useQuery({
    queryKey: ['playlists', max],
    queryFn: () => client.listCurrentUserPlaylists(max),
    enabled: isLoggedIn,
    staleTime: 30 * 1000, // 30 seconds
  })
}
