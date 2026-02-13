// Header component with app title and settings icon

import { useState } from 'react'
import { SettingsDialog } from '../SettingsDialog/SettingsDialog'
import styles from './Header.module.css'

export function Header() {
  const [showSettings, setShowSettings] = useState(false)

  return (
    <>
      <header className={styles.header}>
        <h1 className={styles.title}>Genrify</h1>
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
      </header>
      {showSettings && <SettingsDialog onClose={() => setShowSettings(false)} />}
    </>
  )
}
