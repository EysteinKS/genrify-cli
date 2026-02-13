// Mutation hook for deleting multiple playlists

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { PlaylistService } from '@/lib/playlist-service'
import { useConfirmedWrite } from '@/hooks/useConfirmedWrite'
import { useStatusBar } from '@/contexts/StatusBarContext'

type PlaylistRef = { id: string; name?: string }

export function useDeletePlaylists() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()
  const { confirmAndRun } = useConfirmedWrite()
  const { logSuccess, logError } = useStatusBar()

  return useMutation({
    mutationFn: async (playlists: Array<string | PlaylistRef>) => {
      const items: PlaylistRef[] = playlists
        .map((p) => (typeof p === 'string' ? { id: p.trim() } : { id: p.id.trim(), name: p.name?.trim() }))
        .filter((p) => p.id !== '')

      const ids = items.map((p) => p.id)

      const displayList = items
        .slice(0, 10)
        .map((p) => (p.name ? `"${p.name}" (${p.id})` : p.id))

      const requests = items.map((p) => ({
        method: 'DELETE' as const,
        path: `/playlists/${encodeURIComponent(p.id)}/followers`,
        description: p.name
          ? `Unfollow/delete "${p.name}" (ID: ${p.id}).`
          : `Unfollow/delete playlist ${p.id}.`,
      }))

      const service = new PlaylistService(client)
      await confirmAndRun({
        plan: {
          title: 'Delete source playlists',
          intro: 'This will remove your follow for each playlist. If you own them, Spotify treats this as deleting them.',
          summary: [
            `Playlists: ${ids.length}`,
            ...displayList,
            ...(items.length > 10 ? [`...and ${items.length - 10} more`] : []),
          ],
          requests,
          confirmLabel: 'Delete',
        },
        startingMessage: `Deleting ${ids.length} playlist${ids.length !== 1 ? 's' : ''}...`,
        successMessage: `Deleted ${ids.length} playlist${ids.length !== 1 ? 's' : ''}`,
        errorPrefix: 'Failed to delete playlists',
        action: () => service.deletePlaylists(ids),
      })
      
      return { deletedCount: ids.length, items }
    },
    onSuccess: (data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      
      logSuccess('Delete Playlists', {
        message: `Deleted ${data.deletedCount} playlist${data.deletedCount !== 1 ? 's' : ''}`,
        variables,
        data,
      })
    },
    onError: (error, variables) => {
      logError('Delete Playlists Failed', {
        message: error instanceof Error ? error.message : 'Unknown error',
        variables,
        error,
      })
    },
  })
}
