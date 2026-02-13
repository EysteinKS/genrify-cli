// Login page - mirrors internal/gui/login_view.go

import { useAuth } from '@/contexts/AuthContext'
import { useConfig } from '@/contexts/ConfigContext'
import { useMe } from '@/hooks/queries/useMe'
import { SettingsDialog } from '@/components/SettingsDialog/SettingsDialog'
import { useState } from 'react'
import styles from './LoginPage.module.css'

export function LoginPage() {
  const { isLoggedIn, login, logout } = useAuth()
  const { isConfigured } = useConfig()
  const { data: user } = useMe()
  const [showSettings, setShowSettings] = useState(!isConfigured)

  const handleLogin = async () => {
    if (!isConfigured) {
      setShowSettings(true)
      return
    }
    await login()
  }

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2 className={styles.title}>Spotify Authentication</h2>

        {isLoggedIn && user ? (
          <div className={styles.status}>
            <p className={styles.welcome}>
              Logged in as <strong>{user.display_name || user.id}</strong>
            </p>
            <button onClick={logout} className={styles.logoutButton}>
              Logout
            </button>
          </div>
        ) : (
          <div className={styles.status}>
            <p className={styles.message}>You need to log in to use Genrify.</p>
            <button onClick={handleLogin} className={styles.loginButton}>
              Login with Spotify
            </button>
            {!isConfigured && (
              <p className={styles.hint}>
                Configure your Spotify Client ID in settings before logging in.
              </p>
            )}
          </div>
        )}
      </div>

      {showSettings && <SettingsDialog onClose={() => setShowSettings(false)} />}
    </div>
  )
}
