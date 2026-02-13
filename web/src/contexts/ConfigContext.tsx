// Config context - manages app configuration (Spotify client ID, redirect URI, scopes)

import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import type { AppConfig } from '@/types/config'
import { DEFAULT_CONFIG } from '@/types/config'
import { storage } from '@/lib/storage'
import { getAppCallbackRedirectUri, isLoopbackUrl } from '@/lib/redirect-uri'

interface ConfigContextValue {
  config: AppConfig
  setConfig: (config: AppConfig) => void
  isConfigured: boolean
}

const ConfigContext = createContext<ConfigContextValue | null>(null)

export function ConfigProvider({ children }: { children: ReactNode }) {
  const [config, setConfigState] = useState<AppConfig>(() => {
    const stored = storage.getConfig()
    const appRedirectUri = getAppCallbackRedirectUri()

    if (stored) {
      if (!stored.redirectUri || isLoopbackUrl(stored.redirectUri)) {
        return { ...stored, redirectUri: appRedirectUri || stored.redirectUri }
      }
      return stored
    }

    return { ...DEFAULT_CONFIG, redirectUri: appRedirectUri || DEFAULT_CONFIG.redirectUri }
  })

  const isConfigured = config.clientId.trim() !== ''

  const setConfig = (newConfig: AppConfig) => {
    setConfigState(newConfig)
    storage.setConfig(newConfig)
  }

  // Load config from storage on mount
  useEffect(() => {
    const stored = storage.getConfig()
    if (stored) {
      const appRedirectUri = getAppCallbackRedirectUri()
      if (!stored.redirectUri || isLoopbackUrl(stored.redirectUri)) {
        setConfigState({ ...stored, redirectUri: appRedirectUri || stored.redirectUri })
        return
      }
      setConfigState(stored)
    }
  }, [])

  return (
    <ConfigContext.Provider value={{ config, setConfig, isConfigured }}>
      {children}
    </ConfigContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useConfig() {
  const context = useContext(ConfigContext)
  if (!context) {
    throw new Error('useConfig must be used within ConfigProvider')
  }
  return context
}
