// Merge playlists page - mirrors internal/gui/merge_view.go (3-step flow)

import { useState } from 'react'
import { useFindByPattern } from '@/hooks/mutations/useFindByPattern'
import { useMergePlaylists } from '@/hooks/mutations/useMergePlaylists'
import { useDeletePlaylists } from '@/hooks/mutations/useDeletePlaylists'
import { useStatusBar } from '@/contexts/StatusBarContext'
import { DataTable, type Column } from '@/components/DataTable/DataTable'
import type { SimplifiedPlaylist } from '@/types/spotify'
import type { MergeResult } from '@/lib/playlist-service'
import { isCancelledError } from '@/lib/cancelled'
import styles from './MergePage.module.css'

type Step = 'find' | 'merge' | 'results'

export function MergePage() {
  const [step, setStep] = useState<Step>('find')
  const [pattern, setPattern] = useState('')
  const [matchedPlaylists, setMatchedPlaylists] = useState<SimplifiedPlaylist[]>([])
  const [targetName, setTargetName] = useState('')
  const [targetDescription, setTargetDescription] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [deduplicate, setDeduplicate] = useState(true)
  const [mergeResult, setMergeResult] = useState<MergeResult | null>(null)

  const findByPattern = useFindByPattern()
  const mergePlaylists = useMergePlaylists()
  const deletePlaylists = useDeletePlaylists()
  const { setStatus, setError } = useStatusBar()

  const handleFind = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      setStatus('Finding playlists...')
      const result = await findByPattern.mutateAsync(pattern)
      setMatchedPlaylists(result)
      setStatus(`Found ${result.length} matching playlist${result.length !== 1 ? 's' : ''}`)
      setStep('merge')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to find playlists')
    }
  }

  const handleMerge = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const sourceIds = matchedPlaylists.map((p) => p.id)
      const result = await mergePlaylists.mutateAsync({
        sourceIds,
        targetName,
        options: { deduplicate, public: isPublic, description: targetDescription },
      })
      setMergeResult(result)
      setStep('results')
    } catch (err) {
      if (isCancelledError(err)) return
    }
  }

  const handleDeleteSources = async () => {
    if (!mergeResult?.verified) {
      setError('Cannot delete sources - merge was not verified')
      return
    }
    try {
      await deletePlaylists.mutateAsync(matchedPlaylists.map((p) => ({ id: p.id, name: p.name })))
      handleReset()
    } catch (err) {
      if (isCancelledError(err)) return
    }
  }

  const handleReset = () => {
    setStep('find')
    setPattern('')
    setMatchedPlaylists([])
    setTargetName('')
    setTargetDescription('')
    setMergeResult(null)
  }

  const columns: Column<SimplifiedPlaylist>[] = [
    { key: 'name', header: 'Name', render: (p) => p.name, sortable: true },
    { key: 'tracks', header: 'Tracks', render: (p) => p.tracks.total, sortable: true },
    { key: 'owner', header: 'Owner', render: (p) => p.owner.display_name || p.owner.id },
  ]

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Merge Playlists</h1>

      {step === 'find' && (
        <form onSubmit={handleFind} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="pattern">Pattern (Regex) *</label>
            <input
              id="pattern"
              type="text"
              value={pattern}
              onChange={(e) => setPattern(e.target.value)}
              placeholder="Enter regex pattern to match playlist names"
              required
            />
            <p className={styles.hint}>Example: "^Genre -" matches all playlists starting with "Genre -"</p>
          </div>

          <button type="submit" disabled={findByPattern.isPending} className={styles.button}>
            {findByPattern.isPending ? 'Finding...' : 'Find Matches'}
          </button>
        </form>
      )}

      {step === 'merge' && (
        <>
          <div className={styles.section}>
            <h2>Matched Playlists ({matchedPlaylists.length})</h2>
            <DataTable
              columns={columns}
              data={matchedPlaylists}
              keyExtractor={(p) => p.id}
              emptyMessage="No playlists matched"
            />
          </div>

          <form onSubmit={handleMerge} className={styles.form}>
            <div className={styles.field}>
              <label htmlFor="targetName">Target Playlist Name *</label>
              <input
                id="targetName"
                type="text"
                value={targetName}
                onChange={(e) => setTargetName(e.target.value)}
                placeholder="Name for merged playlist"
                required
              />
            </div>

            <div className={styles.field}>
              <label htmlFor="targetDescription">Description</label>
              <textarea
                id="targetDescription"
                value={targetDescription}
                onChange={(e) => setTargetDescription(e.target.value)}
                placeholder="Description for merged playlist"
                rows={2}
              />
            </div>

            <div className={styles.checkboxes}>
              <div className={styles.checkbox}>
                <input
                  id="public"
                  type="checkbox"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                />
                <label htmlFor="public">Make playlist public</label>
              </div>

              <div className={styles.checkbox}>
                <input
                  id="deduplicate"
                  type="checkbox"
                  checked={deduplicate}
                  onChange={(e) => setDeduplicate(e.target.checked)}
                />
                <label htmlFor="deduplicate">Remove duplicates</label>
              </div>
            </div>

            <div className={styles.actions}>
              <button type="button" onClick={handleReset} className={styles.secondaryButton}>
                Start Over
              </button>
              <button type="submit" disabled={mergePlaylists.isPending} className={styles.button}>
                {mergePlaylists.isPending ? 'Merging...' : 'Merge Playlists'}
              </button>
            </div>
          </form>
        </>
      )}

      {step === 'results' && mergeResult && (
        <div className={styles.results}>
          <div className={styles.resultCard}>
            <h2>Merge Complete</h2>
            <dl className={styles.resultList}>
              <dt>New Playlist ID:</dt>
              <dd>{mergeResult.newPlaylistId}</dd>

              <dt>Total Tracks:</dt>
              <dd>{mergeResult.trackCount}</dd>

              <dt>Duplicates Removed:</dt>
              <dd>{mergeResult.duplicatesRemoved}</dd>

              <dt>Verified:</dt>
              <dd className={mergeResult.verified ? styles.success : styles.error}>
                {mergeResult.verified ? 'Yes' : 'No'}
              </dd>

              {mergeResult.missingURIs.length > 0 && (
                <>
                  <dt>Missing URIs:</dt>
                  <dd>{mergeResult.missingURIs.length}</dd>
                </>
              )}
            </dl>

            <div className={styles.actions}>
              {mergeResult.verified && (
                <button onClick={handleDeleteSources} className={styles.deleteButton}>
                  Delete Source Playlists
                </button>
              )}
              <button onClick={handleReset} className={styles.button}>
                Start New Merge
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
