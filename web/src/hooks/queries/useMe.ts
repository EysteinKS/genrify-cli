// Query hook for current user profile

import { useQuery } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useAuth } from '@/contexts/AuthContext'

export function useMe() {
  const client = useSpotifyClient()
  const { isLoggedIn } = useAuth()

  return useQuery({
    queryKey: ['me'],
    queryFn: () => client.getMe(),
    enabled: isLoggedIn,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}
