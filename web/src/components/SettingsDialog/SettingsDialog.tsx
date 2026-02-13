// Settings dialog for configuring Spotify client ID and redirect URI

import { useState, useEffect } from 'react'
import { useConfig } from '@/contexts/ConfigContext'
import type { AppConfig } from '@/types/config'
import styles from './SettingsDialog.module.css'

interface SettingsDialogProps {
  onClose: () => void
}

export function SettingsDialog({ onClose }: SettingsDialogProps) {
  const { config, setConfig } = useConfig()
  const [formData, setFormData] = useState<AppConfig>(config)

  useEffect(() => {
    setFormData(config)
  }, [config])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setConfig(formData)
    onClose()
  }

  const handleCancel = () => {
    onClose()
  }

  return (
    <div className={styles.overlay} onClick={handleCancel}>
      <div className={styles.dialog} onClick={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          <h2 className={styles.title}>Settings</h2>
          <button className={styles.closeButton} onClick={handleCancel} title="Close">
            ×
          </button>
        </div>

        <form onSubmit={handleSubmit} className={styles.form}>
          <div className={styles.field}>
            <label htmlFor="clientId" className={styles.label}>
              Spotify Client ID *
            </label>
            <input
              id="clientId"
              type="text"
              className={styles.input}
              value={formData.clientId}
              onChange={(e) => setFormData({ ...formData, clientId: e.target.value })}
              placeholder="Enter your Spotify Client ID"
              required
            />
            <p className={styles.hint}>
              Get your Client ID from{' '}
              <a
                href="https://developer.spotify.com/dashboard"
                target="_blank"
                rel="noopener noreferrer"
              >
                Spotify Developer Dashboard
              </a>
            </p>
          </div>

          <div className={styles.field}>
            <label htmlFor="redirectUri" className={styles.label}>
              Redirect URI *
            </label>
            <input
              id="redirectUri"
              type="text"
              className={styles.input}
              value={formData.redirectUri}
              onChange={(e) => setFormData({ ...formData, redirectUri: e.target.value })}
              placeholder="http://localhost:5173/callback"
              required
            />
            <p className={styles.hint}>
              Must match a redirect URI registered in your Spotify app settings
            </p>
          </div>

          <div className={styles.field}>
            <label htmlFor="scopes" className={styles.label}>
              Scopes
            </label>
            <textarea
              id="scopes"
              className={styles.textarea}
              value={formData.scopes.join('\n')}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  scopes: e.target.value.split('\n').filter((s) => s.trim() !== ''),
                })
              }
              rows={4}
              placeholder="One scope per line"
            />
          </div>

          <div className={styles.actions}>
            <button type="button" className={styles.cancelButton} onClick={handleCancel}>
              Cancel
            </button>
            <button type="submit" className={styles.saveButton}>
              Save
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
