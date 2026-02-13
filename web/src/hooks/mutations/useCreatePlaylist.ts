// Mutation hook for creating a playlist

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useConfirmedWrite } from '@/hooks/useConfirmedWrite'
import { useStatusBar } from '@/contexts/StatusBarContext'

interface CreatePlaylistParams {
  name: string
  description: string
  isPublic: boolean
}

export function useCreatePlaylist() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()
  const { confirmAndRun } = useConfirmedWrite()
  const { logSuccess, logError } = useStatusBar()

  return useMutation({
    mutationFn: ({ name, description, isPublic }: CreatePlaylistParams) => {
      const trimmedName = name.trim()
      const body = { name: trimmedName, public: isPublic, description }

      return confirmAndRun({
        plan: {
          title: 'Create playlist',
          summary: [
            `Playlist name: ${trimmedName ? `"${trimmedName}"` : '(empty)'}`,
            `Public: ${isPublic ? 'Yes' : 'No'}`,
            `Description: ${description?.trim() ? `"${description.trim()}"` : '(empty)'}`,
          ],
          requests: [
            {
              method: 'POST',
              path: '/me/playlists',
              description: 'Create a new playlist in your Spotify account.',
              body,
            },
          ],
          confirmLabel: 'Create',
        },
        startingMessage: 'Creating playlist...',
        successMessage: (pl) => `Created playlist: ${pl.name}`,
        errorPrefix: 'Failed to create playlist',
        action: () => client.createPlaylist(trimmedName, description, isPublic),
      })
    },
    onSuccess: (data, variables) => {
      // Invalidate playlists cache
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      
      logSuccess('Create Playlist', {
        message: `Created playlist "${data.name}"`,
        variables,
        data,
      })
    },
    onError: (error, variables) => {
      logError('Create Playlist Failed', {
        message: error instanceof Error ? error.message : 'Unknown error',
        variables,
        error,
      })
    },
  })
}
