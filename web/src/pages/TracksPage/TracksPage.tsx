// Tracks page - mirrors internal/gui/tracks_view.go

import { useState, useEffect } from 'react'
import { useSearchParams } from 'react-router-dom'
import { usePlaylistTracks } from '@/hooks/queries/usePlaylistTracks'
import { DataTable, type Column } from '@/components/DataTable/DataTable'
import { normalizePlaylistID, joinArtistNames } from '@/lib/helpers'
import type { FullTrack } from '@/types/spotify'
import styles from './TracksPage.module.css'

export function TracksPage() {
  const [searchParams] = useSearchParams()
  const [input, setInput] = useState('')
  const [playlistId, setPlaylistId] = useState<string | null>(null)
  const [limit, setLimit] = useState(0)

  const { data: tracks, isLoading, error } = usePlaylistTracks(playlistId, limit)

  // Auto-load if playlistId is in URL
  useEffect(() => {
    const id = searchParams.get('playlistId')
    if (id) {
      setInput(id)
      setPlaylistId(id)
    }
  }, [searchParams])

  const handleLoad = () => {
    try {
      const id = normalizePlaylistID(input)
      setPlaylistId(id)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Invalid playlist ID')
    }
  }

  const columns: Column<FullTrack>[] = [
    {
      key: 'uri',
      header: 'URI',
      render: (t) => t.uri,
    },
    {
      key: 'name',
      header: 'Name',
      render: (t) => t.name,
      sortable: true,
    },
    {
      key: 'artists',
      header: 'Artists',
      render: (t) => joinArtistNames(t.artists),
      sortable: true,
    },
  ]

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Playlist Tracks</h1>

      <div className={styles.controls}>
        <div className={styles.field}>
          <label htmlFor="playlistId">Playlist ID / URI / URL</label>
          <input
            id="playlistId"
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Enter playlist ID, URI, or URL"
            onKeyDown={(e) => e.key === 'Enter' && handleLoad()}
          />
        </div>

        <div className={styles.field}>
          <label htmlFor="limit">Limit (0 = all)</label>
          <input
            id="limit"
            type="number"
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            min={0}
          />
        </div>

        <button onClick={handleLoad} className={styles.loadButton}>
          Load Tracks
        </button>
      </div>

      {error && <p className={styles.error}>Error: {error.message}</p>}

      {isLoading ? (
        <p className={styles.loading}>Loading tracks...</p>
      ) : tracks ? (
        <>
          <p className={styles.count}>
            {tracks.length} track{tracks.length !== 1 ? 's' : ''}
          </p>
          <DataTable
            columns={columns}
            data={tracks}
            keyExtractor={(t) => t.uri}
            emptyMessage="No tracks found"
          />
        </>
      ) : (
        <p className={styles.hint}>Enter a playlist ID above and click Load Tracks</p>
      )}
    </div>
  )
}
