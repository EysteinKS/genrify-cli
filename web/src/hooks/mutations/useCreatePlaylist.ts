// Mutation hook for creating a playlist

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'

interface CreatePlaylistParams {
  name: string
  description: string
  isPublic: boolean
}

export function useCreatePlaylist() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ name, description, isPublic }: CreatePlaylistParams) =>
      client.createPlaylist(name, description, isPublic),
    onSuccess: () => {
      // Invalidate playlists cache
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
    },
  })
}
