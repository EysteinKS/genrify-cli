// Status bar component at bottom of layout

import { useStatusBar } from '@/contexts/StatusBarContext'
import styles from './StatusBar.module.css'

export function StatusBar() {
  const { message, isError, isLoading, entries, openHistory } = useStatusBar()

  const latestEntry = entries[0]

  return (
    <div className={styles.statusBar}>
      {/* Current status message (loading/temporary) */}
      {message && (
        <div className={`${styles.statusMessage} ${isError ? styles.error : ''}`}>
          {isLoading && <span className={styles.spinner} />}
          <span className={styles.message}>{message}</span>
        </div>
      )}

      {/* Latest completed action (clickable) */}
      {latestEntry && !message && (
        <button
          className={`${styles.actionIndicator} ${styles[latestEntry.type]}`}
          onClick={openHistory}
          title="View action history"
        >
          <span className={styles.actionIcon}>
            {latestEntry.type === 'success' ? '✓' : '✕'}
          </span>
          <span className={styles.actionText}>
            {latestEntry.message || latestEntry.title}
          </span>
        </button>
      )}
    </div>
  )
}
