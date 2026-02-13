// Hook to get a memoized SpotifyClient instance

import { useMemo } from 'react'
import { SpotifyClient } from '@/lib/spotify-client'
import { useAuth } from '@/contexts/AuthContext'

export function useSpotifyClient(): SpotifyClient {
  const { getAccessToken, forceRefresh } = useAuth()

  return useMemo(
    () => new SpotifyClient(getAccessToken, forceRefresh),
    [getAccessToken, forceRefresh]
  )
}
