// Mutation hook for deleting a playlist

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'

export function useDeletePlaylist() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (playlistId: string) => client.deletePlaylist(playlistId),
    onSuccess: () => {
      // Invalidate playlists cache
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
    },
  })
}
