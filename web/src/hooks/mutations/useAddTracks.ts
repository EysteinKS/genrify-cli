// Mutation hook for adding tracks to a playlist

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'

interface AddTracksParams {
  playlistId: string
  uris: string[]
}

export function useAddTracks() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ playlistId, uris }: AddTracksParams) =>
      client.addTracksToPlaylist(playlistId, uris),
    onSuccess: (_, { playlistId }) => {
      // Invalidate tracks cache for this playlist
      queryClient.invalidateQueries({ queryKey: ['playlist-tracks', playlistId] })
      queryClient.invalidateQueries({ queryKey: ['playlist', playlistId] })
    },
  })
}
