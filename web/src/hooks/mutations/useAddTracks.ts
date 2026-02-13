// Mutation hook for adding tracks to a playlist

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useConfirmedWrite } from '@/hooks/useConfirmedWrite'
import { useStatusBar } from '@/contexts/StatusBarContext'

interface AddTracksParams {
  playlistId: string
  uris: string[]
}

export function useAddTracks() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()
  const { confirmAndRun } = useConfirmedWrite()
  const { logSuccess, logError } = useStatusBar()

  return useMutation({
    mutationFn: ({ playlistId, uris }: AddTracksParams) => {
      const trimmedPlaylistId = playlistId.trim()
      const clean = uris.map((u) => u.trim()).filter((u) => u !== '')

      const endpoint = `/playlists/${encodeURIComponent(trimmedPlaylistId)}/tracks`
      const batches: string[][] = []
      for (let i = 0; i < clean.length; i += 100) {
        batches.push(clean.slice(i, i + 100))
      }

      return confirmAndRun({
        plan: {
          title: 'Add tracks to playlist',
          summary: [
            `Playlist ID: ${trimmedPlaylistId || '(empty)'}`,
            `Tracks to add: ${clean.length}`,
            `Requests: ${batches.length} (Spotify limits 100 tracks per request)`,
          ],
          requests: batches.map((batch, idx) => ({
            method: 'POST',
            path: endpoint,
            description: `Add ${batch.length} track URI(s) (batch ${idx + 1}/${batches.length}).`,
            body: { uris: batch },
          })),
          confirmLabel: 'Add tracks',
        },
        startingMessage: `Adding ${clean.length} track${clean.length !== 1 ? 's' : ''}...`,
        successMessage: () => `Successfully added ${clean.length} track${clean.length !== 1 ? 's' : ''}`,
        errorPrefix: 'Failed to add tracks',
        action: () => client.addTracksToPlaylist(trimmedPlaylistId, clean),
      })
    },
    onSuccess: (data, variables) => {
      // Invalidate tracks cache for this playlist
      queryClient.invalidateQueries({ queryKey: ['playlist-tracks', variables.playlistId] })
      queryClient.invalidateQueries({ queryKey: ['playlist', variables.playlistId] })
      
      logSuccess('Add Tracks', {
        message: `Added ${variables.uris.length} track${variables.uris.length !== 1 ? 's' : ''} to playlist`,
        variables,
        data,
      })
    },
    onError: (error, variables) => {
      logError('Add Tracks Failed', {
        message: error instanceof Error ? error.message : 'Unknown error',
        variables,
        error,
      })
    },
  })
}
