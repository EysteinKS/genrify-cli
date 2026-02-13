// Typed localStorage wrapper - replaces internal/auth/store.go filesystem storage

import type { Token } from '@/types/auth'
import type { AppConfig } from '@/types/config'

const KEYS = {
  TOKEN: 'genrify:token',
  CONFIG: 'genrify:config',
  PKCE_VERIFIER: 'genrify:pkce_verifier',
  AUTH_STATE: 'genrify:auth_state',
} as const

export const storage = {
  // Token management
  getToken(): Token | null {
    return getJSON<Token>(KEYS.TOKEN)
  },

  setToken(token: Token): void {
    setJSON(KEYS.TOKEN, token)
  },

  clearToken(): void {
    remove(KEYS.TOKEN)
  },

  // Config management
  getConfig(): AppConfig | null {
    return getJSON<AppConfig>(KEYS.CONFIG)
  },

  setConfig(config: AppConfig): void {
    setJSON(KEYS.CONFIG, config)
  },

  clearConfig(): void {
    remove(KEYS.CONFIG)
  },

  // PKCE verifier (temporary, for OAuth flow)
  getPKCEVerifier(): string | null {
    return localStorage.getItem(KEYS.PKCE_VERIFIER)
  },

  setPKCEVerifier(verifier: string): void {
    localStorage.setItem(KEYS.PKCE_VERIFIER, verifier)
  },

  clearPKCEVerifier(): void {
    remove(KEYS.PKCE_VERIFIER)
  },

  // Auth state (temporary, for OAuth flow)
  getAuthState(): string | null {
    return sessionStorage.getItem(KEYS.AUTH_STATE)
  },

  setAuthState(state: string): void {
    sessionStorage.setItem(KEYS.AUTH_STATE, state)
  },

  clearAuthState(): void {
    sessionStorage.removeItem(KEYS.AUTH_STATE)
  },

  // Clear all
  clearAll(): void {
    remove(KEYS.TOKEN)
    remove(KEYS.CONFIG)
    remove(KEYS.PKCE_VERIFIER)
    sessionStorage.removeItem(KEYS.AUTH_STATE)
  },
}

function getJSON<T>(key: string): T | null {
  try {
    const item = localStorage.getItem(key)
    if (!item) return null
    return JSON.parse(item) as T
  } catch (err) {
    console.error(`Failed to parse ${key} from localStorage:`, err)
    return null
  }
}

function setJSON(key: string, value: unknown): void {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch (err) {
    console.error(`Failed to save ${key} to localStorage:`, err)
  }
}

function remove(key: string): void {
  localStorage.removeItem(key)
}
