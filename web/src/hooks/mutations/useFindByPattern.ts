// Mutation hook for finding playlists by regex pattern

import { useMutation } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { PlaylistService } from '@/lib/playlist-service'

export function useFindByPattern() {
  const client = useSpotifyClient()

  return useMutation({
    mutationFn: (pattern: string) => {
      const service = new PlaylistService(client)
      return service.findPlaylistsByPattern(pattern)
    },
  })
}
