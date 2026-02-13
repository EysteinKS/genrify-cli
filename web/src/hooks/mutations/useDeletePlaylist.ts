// Mutation hook for deleting a playlist

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { useConfirmedWrite } from '@/hooks/useConfirmedWrite'
import { useStatusBar } from '@/contexts/StatusBarContext'

type PlaylistRef = { id: string; name?: string }

export function useDeletePlaylist() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()
  const { confirmAndRun } = useConfirmedWrite()
  const { logSuccess, logError } = useStatusBar()

  return useMutation({
    mutationFn: (playlist: string | PlaylistRef) => {
      const trimmed = typeof playlist === 'string' ? playlist.trim() : playlist.id.trim()
      const name = typeof playlist === 'string' ? undefined : playlist.name?.trim()
      const endpoint = `/playlists/${encodeURIComponent(trimmed)}/followers`

      return confirmAndRun({
        plan: {
          title: 'Delete playlist',
          intro:
            'This will remove your follow for the playlist. If you own it, Spotify treats this as deleting the playlist.',
          summary: [
            `Playlist name: ${name ? `"${name}"` : '(unknown)'}`,
            `Playlist ID: ${trimmed || '(empty)'}`,
          ],
          requests: [
            {
              method: 'DELETE',
              path: endpoint,
              description: name
                ? `Unfollow/delete "${name}" (ID: ${trimmed}).`
                : 'Unfollow/delete playlist.',
            },
          ],
          confirmLabel: 'Delete',
        },
        startingMessage: 'Deleting playlist...',
        successMessage: 'Playlist deleted',
        errorPrefix: 'Failed to delete playlist',
        action: () => client.deletePlaylist(trimmed),
      })
    },
    onSuccess: (data, variables) => {
      // Invalidate playlists cache
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      
      const name = typeof variables === 'string' ? undefined : variables.name
      const id = typeof variables === 'string' ? variables : variables.id
      
      logSuccess('Delete Playlist', {
        message: name ? `Deleted playlist "${name}"` : `Deleted playlist ${id}`,
        variables,
        data,
      })
    },
    onError: (error, variables) => {
      logError('Delete Playlist Failed', {
        message: error instanceof Error ? error.message : 'Unknown error',
        variables,
        error,
      })
    },
  })
}
