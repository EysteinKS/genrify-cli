// Playlists page - mirrors internal/gui/playlists_view.go

import { useState, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { usePlaylists } from '@/hooks/queries/usePlaylists'
import { DataTable, type Column } from '@/components/DataTable/DataTable'
import { filterPlaylistsByName } from '@/lib/helpers'
import type { SimplifiedPlaylist } from '@/types/spotify'
import styles from './PlaylistsPage.module.css'

export function PlaylistsPage() {
  const navigate = useNavigate()
  const [limit, setLimit] = useState(50)
  const [filter, setFilter] = useState('')
  const { data: playlists, isLoading, error, refetch } = usePlaylists(limit)

  const filteredPlaylists = useMemo(() => {
    if (!playlists) return []
    return filterPlaylistsByName(playlists, filter)
  }, [playlists, filter])

  const columns: Column<SimplifiedPlaylist>[] = [
    {
      key: 'id',
      header: 'ID',
      render: (p) => p.id,
      sortable: true,
    },
    {
      key: 'name',
      header: 'Name',
      render: (p) => p.name,
      sortable: true,
    },
    {
      key: 'tracks',
      header: 'Tracks',
      render: (p) => p.tracks.total,
      sortable: true,
    },
    {
      key: 'owner',
      header: 'Owner',
      render: (p) => p.owner.display_name || p.owner.id,
      sortable: true,
    },
  ]

  const handleRowClick = (playlist: SimplifiedPlaylist) => {
    navigate(`/tracks?playlistId=${playlist.id}`)
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Your Playlists</h1>

      <div className={styles.controls}>
        <div className={styles.field}>
          <label htmlFor="filter">Filter by name</label>
          <input
            id="filter"
            type="text"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Search playlists..."
          />
        </div>

        <div className={styles.field}>
          <label htmlFor="limit">Limit</label>
          <input
            id="limit"
            type="number"
            value={limit}
            onChange={(e) => setLimit(Number(e.target.value))}
            min={1}
            max={500}
          />
        </div>

        <button onClick={() => refetch()} className={styles.refreshButton}>
          Refresh
        </button>
      </div>

      {error && <p className={styles.error}>Error: {error.message}</p>}

      {isLoading ? (
        <p className={styles.loading}>Loading playlists...</p>
      ) : (
        <DataTable
          columns={columns}
          data={filteredPlaylists}
          keyExtractor={(p) => p.id}
          onRowClick={handleRowClick}
          emptyMessage="No playlists found"
        />
      )}
    </div>
  )
}
