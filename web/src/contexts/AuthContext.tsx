// Auth context - manages token lifecycle with automatic refresh
// Port of TokenManager from internal/spotify/token_manager.go

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'
import type { Token } from '@/types/auth'
import { storage } from '@/lib/storage'
import { isTokenExpired, refreshToken as refreshTokenAPI, initiateLogin } from '@/lib/auth'
import { useConfig } from './ConfigContext'

interface AuthContextValue {
  token: Token | null
  isLoggedIn: boolean
  login: () => Promise<void>
  logout: () => void
  getAccessToken: () => Promise<string>
  forceRefresh: () => Promise<string>
  setToken: (token: Token) => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const { config } = useConfig()
  const [token, setTokenState] = useState<Token | null>(() => {
    return storage.getToken()
  })

  const isLoggedIn = token !== null && !isTokenExpired(token)

  const setToken = useCallback((newToken: Token) => {
    setTokenState(newToken)
    storage.setToken(newToken)
  }, [])

  const logout = useCallback(() => {
    setTokenState(null)
    storage.clearToken()
  }, [])

  const login = useCallback(async () => {
    await initiateLogin(config)
  }, [config])

  const forceRefresh = useCallback(async (): Promise<string> => {
    if (!token?.refresh_token) {
      throw new Error('No refresh token available')
    }

    const newToken = await refreshTokenAPI(config, token.refresh_token)
    setToken(newToken)
    return newToken.access_token
  }, [config, token, setToken])

  const getAccessToken = useCallback(async (): Promise<string> => {
    if (!token) {
      throw new Error('Not logged in')
    }

    // Check if token needs refresh (with 60s leeway)
    if (isTokenExpired(token, 60)) {
      if (!token.refresh_token) {
        throw new Error('Token expired and no refresh token available')
      }
      return forceRefresh()
    }

    return token.access_token
  }, [token, forceRefresh])

  // Load token from storage on mount
  useEffect(() => {
    const stored = storage.getToken()
    if (stored) {
      setTokenState(stored)
    }
  }, [])

  return (
    <AuthContext.Provider
      value={{
        token,
        isLoggedIn,
        login,
        logout,
        getAccessToken,
        forceRefresh,
        setToken,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}
