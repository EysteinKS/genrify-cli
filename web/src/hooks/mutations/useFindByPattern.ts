// Mutation hook for finding playlists by regex pattern

import { useMutation } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { PlaylistService } from '@/lib/playlist-service'
import { useStatusBar } from '@/contexts/StatusBarContext'

export function useFindByPattern() {
  const client = useSpotifyClient()
  const { logSuccess, logError } = useStatusBar()

  return useMutation({
    mutationFn: (pattern: string) => {
      const service = new PlaylistService(client)
      return service.findPlaylistsByPattern(pattern)
    },
    onSuccess: (data, variables) => {
      logSuccess('Find Playlists', {
        message: `Found ${data.length} playlist${data.length !== 1 ? 's' : ''} matching pattern`,
        variables: { pattern: variables },
        data,
      })
    },
    onError: (error, variables) => {
      logError('Find Playlists Failed', {
        message: error instanceof Error ? error.message : 'Unknown error',
        variables: { pattern: variables },
        error,
      })
    },
  })
}
