// Mutation hook for merging playlists

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useSpotifyClient } from '../useSpotifyClient'
import { PlaylistService, type MergeOptions } from '@/lib/playlist-service'
import { useConfirmedWrite } from '@/hooks/useConfirmedWrite'
import { useStatusBar } from '@/contexts/StatusBarContext'

interface MergePlaylistsParams {
  sourceIds: string[]
  targetName: string
  options: MergeOptions
  onProgress?: (message: string) => void
}

export function useMergePlaylists() {
  const client = useSpotifyClient()
  const queryClient = useQueryClient()
  const { confirmAndRun } = useConfirmedWrite()
  const { setLoading, logSuccess, logError } = useStatusBar()

  return useMutation({
    mutationFn: async ({ sourceIds, targetName, options, onProgress }: MergePlaylistsParams) => {
      const service = new PlaylistService(client)

      const report = (msg: string) => {
        setLoading(msg)
        onProgress?.(msg)
      }

      // Read-only: compute the exact URIs we intend to send before asking for confirmation.
      const prepared = await service.prepareMergeTracks(sourceIds, options.deduplicate, report)

      const createBody = {
        name: targetName.trim(),
        public: options.public,
        description: options.description,
      }

      const addBatches: string[][] = []
      for (let i = 0; i < prepared.uris.length; i += 100) {
        addBatches.push(prepared.uris.slice(i, i + 100))
      }

      const addRequests = addBatches.map((batch, idx) => ({
        method: 'POST' as const,
        path: '/playlists/<new-playlist-id>/tracks',
        description: `Add ${batch.length} track URI(s) to the newly created playlist (batch ${idx + 1}/${addBatches.length}).`,
        body: { uris: batch },
      }))

      const rollbackRequest = {
        method: 'DELETE' as const,
        path: '/playlists/<new-playlist-id>/followers',
        description:
          'Rollback (only if a later step fails): unfollow/delete the newly created playlist so no partial playlist remains.',
      }

      return confirmAndRun({
        plan: {
          title: 'Merge playlists',
          intro:
            'This action will create a new playlist and add tracks collected from the selected source playlists.',
          summary: [
            `Source playlists: ${sourceIds.length}`,
            `Target name: ${createBody.name ? `"${createBody.name}"` : '(empty)'}`,
            `Public: ${options.public ? 'Yes' : 'No'}`,
            `Deduplicate: ${options.deduplicate ? 'Yes' : 'No'} (duplicates removed: ${prepared.duplicatesRemoved})`,
            `Tracks to add: ${prepared.uris.length}`,
            `Add requests: ${addRequests.length} (Spotify limits 100 tracks per request)`,
          ],
          requests: [
            {
              method: 'POST',
              path: '/me/playlists',
              description: 'Create the target playlist in your Spotify account.',
              body: createBody,
            },
            ...addRequests,
            rollbackRequest,
          ],
          confirmLabel: 'Merge',
        },
        startingMessage: 'Merging playlists...',
        successMessage: 'Merge completed successfully',
        errorPrefix: 'Failed to merge playlists',
        action: () => service.mergePreparedTracks(targetName, options, prepared, report),
      })
    },
    onSuccess: (data, variables) => {
      // Invalidate playlists cache
      queryClient.invalidateQueries({ queryKey: ['playlists'] })
      
      logSuccess('Merge Playlists', {
        message: `Merged ${variables.sourceIds.length} playlists into "${variables.targetName}" (${data.trackCount} tracks)`,
        variables,
        data,
      })
    },
    onError: (error, variables) => {
      logError('Merge Playlists Failed', {
        message: error instanceof Error ? error.message : 'Unknown error',
        variables,
        error,
      })
    },
  })
}
