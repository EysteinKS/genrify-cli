// OAuth callback page - handles code exchange

import { useEffect, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuth } from '@/contexts/AuthContext'
import { useConfig } from '@/contexts/ConfigContext'
import { handleCallback } from '@/lib/auth'
import styles from './CallbackPage.module.css'

export function CallbackPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const { setToken } = useAuth()
  const { config } = useConfig()
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const exchangeCode = async () => {
      const code = searchParams.get('code')
      const state = searchParams.get('state')
      const spotifyError = searchParams.get('error')

      if (spotifyError) {
        setError(`Spotify authorization failed: ${spotifyError}`)
        setTimeout(() => navigate('/login'), 3000)
        return
      }

      if (!code || !state) {
        setError('Missing code or state parameter')
        setTimeout(() => navigate('/login'), 3000)
        return
      }

      try {
        const token = await handleCallback(config, code, state)
        setToken(token)
        navigate('/login')
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error')
        setTimeout(() => navigate('/login'), 3000)
      }
    }

    exchangeCode()
  }, [searchParams, navigate, setToken, config])

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        {error ? (
          <>
            <div className={styles.error}>✕</div>
            <h2 className={styles.title}>Authentication Error</h2>
            <p className={styles.message}>{error}</p>
            <p className={styles.hint}>Redirecting to login...</p>
          </>
        ) : (
          <>
            <div className={styles.spinner} />
            <h2 className={styles.title}>Completing Login</h2>
            <p className={styles.message}>Exchanging authorization code for access token...</p>
          </>
        )}
      </div>
    </div>
  )
}
