// Header component with app title and settings icon

import { useState } from 'react'
import { SettingsDialog } from '../SettingsDialog/SettingsDialog'
import { useStatusBar } from '@/contexts/StatusBarContext'
import styles from './Header.module.css'

export function Header() {
  const [showSettings, setShowSettings] = useState(false)
  const { entries, openHistory } = useStatusBar()

  const latestEntry = entries[0]

  return (
    <>
      <header className={styles.header}>
        <h1 className={styles.title}>Genrify</h1>
        
        <div className={styles.actions}>
          {latestEntry && (
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
          
          <button
            className={styles.settingsButton}
            onClick={() => setShowSettings(true)}
            title="Settings"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="12" r="3" />
              <path d="M12 1v6m0 6v6m9-9h-6m-6 0H3" />
            </svg>
          </button>
        </div>
      </header>
      {showSettings && <SettingsDialog onClose={() => setShowSettings(false)} />}
    </>
  )
}
