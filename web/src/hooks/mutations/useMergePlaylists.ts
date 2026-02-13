// Mutation hook for merging playlists

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { PlaylistService, type MergeOptions } from '@/lib/playlist-service'

interface MergePlaylistsParams {
  sourceIds: string[]
  targetName: string
  options: MergeOptions
  onProgress?: (message: string) => void
}

export function useMergePlaylists() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ sourceIds, targetName, options, onProgress }: MergePlaylistsParams) => {
      const service = new PlaylistService(client)
      return service.mergePlaylists(sourceIds, targetName, options, onProgress)
    },
    onSuccess: () => {
      // Invalidate playlists cache
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
    },
  })
}
