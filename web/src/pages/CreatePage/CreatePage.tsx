// Create playlist page - mirrors internal/gui/create_view.go

import { useState } from 'react'
import { useCreatePlaylist } from '@/hooks/mutations/useCreatePlaylist'
import { useStatusBar } from '@/contexts/StatusBarContext'
import styles from './CreatePage.module.css'

export function CreatePage() {
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isPublic, setIsPublic] = useState(false)
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(
    null
  )

  const createPlaylist = useCreatePlaylist()
  const { setStatus, setError } = useStatusBar()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFeedback(null)

    if (!name.trim()) {
      setFeedback({ type: 'error', message: 'Name is required' })
      return
    }

    try {
      setStatus('Creating playlist...')
      const result = await createPlaylist.mutateAsync({ name, description, isPublic })
      setStatus(`Created playlist: ${result.name}`)
      setFeedback({
        type: 'success',
        message: `Playlist "${result.name}" created successfully! ID: ${result.id}`,
      })
      // Clear form
      setName('')
      setDescription('')
      setIsPublic(false)
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error'
      setError(`Failed to create playlist: ${message}`)
      setFeedback({ type: 'error', message })
    }
  }

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Create Playlist</h1>

      <form onSubmit={handleSubmit} className={styles.form}>
        <div className={styles.field}>
          <label htmlFor="name">Name *</label>
          <input
            id="name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter playlist name"
            required
          />
        </div>

        <div className={styles.field}>
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Enter playlist description (optional)"
            rows={3}
          />
        </div>

        <div className={styles.checkbox}>
          <input
            id="public"
            type="checkbox"
            checked={isPublic}
            onChange={(e) => setIsPublic(e.target.checked)}
          />
          <label htmlFor="public">Make playlist public</label>
        </div>

        <button type="submit" disabled={createPlaylist.isPending} className={styles.submitButton}>
          {createPlaylist.isPending ? 'Creating...' : 'Create Playlist'}
        </button>

        {feedback && (
          <div
            className={
              feedback.type === 'success' ? styles.successFeedback : styles.errorFeedback
            }
          >
            {feedback.message}
          </div>
        )}
      </form>
    </div>
  )
}
