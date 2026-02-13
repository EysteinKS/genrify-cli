// Add tracks page - mirrors internal/gui/add_tracks_view.go

import { useState } from 'react'
import { useAddTracks } from '@/hooks/mutations/useAddTracks'
import { useStatusBar } from '@/contexts/StatusBarContext'
import { isCancelledError } from '@/lib/cancelled'
import { normalizePlaylistID, normalizeTrackURI } from '@/lib/helpers'
import styles from './AddTracksPage.module.css'

export function AddTracksPage() {
  const [playlistId, setPlaylistId] = useState('')
  const [tracksInput, setTracksInput] = useState('')
  const [warnings, setWarnings] = useState<string[]>([])
  const [success, setSuccess] = useState<string | null>(null)

  const addTracks = useAddTracks()
  const { setError } = useStatusBar()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setWarnings([])
    setSuccess(null)

    // Normalize playlist ID
    let normalizedPlaylistId: string
    try {
      normalizedPlaylistId = normalizePlaylistID(playlistId)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Invalid playlist ID')
      return
    }

    // Parse and normalize track URIs
    const lines = tracksInput
      .split(/[\n,]/)
      .map((s) => s.trim())
      .filter((s) => s !== '')

    if (lines.length === 0) {
      setError('Please enter at least one track URI')
      return
    }

    const uris: string[] = []
    const warns: string[] = []

    for (const line of lines) {
      try {
        const uri = normalizeTrackURI(line)
        uris.push(uri)
      } catch (err) {
        warns.push(`Skipped invalid track: ${line} (${err instanceof Error ? err.message : 'error'})`)
      }
    }

    if (uris.length === 0) {
      setError('No valid track URIs found')
      setWarnings(warns)
      return
    }

    try {
      await addTracks.mutateAsync({ playlistId: normalizedPlaylistId, uris })
      setSuccess(`Added ${uris.length} track${uris.length !== 1 ? 's' : ''} to playlist`)
      setWarnings(warns)
      setTracksInput('')
    } catch (err) {
      if (isCancelledError(err)) return
    }
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Add Tracks to Playlist</h1>

      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.field}>
          <label htmlFor="playlistId">Playlist ID / URI / URL *</label>
          <input
            id="playlistId"
            type="text"
            value={playlistId}
            onChange={(e) => setPlaylistId(e.target.value)}
            placeholder="Enter playlist ID, URI, or URL"
            required
          />
        </div>

        <div className={styles.field}>
          <label htmlFor="tracks">Track URIs / URLs *</label>
          <textarea
            id="tracks"
            value={tracksInput}
            onChange={(e) => setTracksInput(e.target.value)}
            placeholder="Enter track URIs or URLs (one per line or comma-separated)"
            rows={10}
            required
          />
          <p className={styles.hint}>
            Accepts: Track IDs, spotify:track:ID URIs, or open.spotify.com/track/ID URLs
          </p>
        </div>

        <button type="submit" disabled={addTracks.isPending} className={styles.submitButton}>
          {addTracks.isPending ? 'Adding...' : 'Add Tracks'}
        </button>

        {success && <div className={styles.success}>{success}</div>}

        {warnings.length > 0 && (
          <div className={styles.warnings}>
            <h3>Warnings:</h3>
            <ul>
              {warnings.map((w, i) => (
                <li key={i}>{w}</li>
              ))}
            </ul>
          </div>
        )}
      </form>
    </div>
  )
}
