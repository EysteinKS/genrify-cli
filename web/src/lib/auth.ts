// OAuth PKCE browser flow - adapted from internal/auth/oauth.go

import type { Token } from '@/types/auth'
import type { AppConfig } from '@/types/config'
import { generatePKCE } from './pkce'
import { storage } from './storage'
import { getEffectiveRedirectUri } from './redirect-uri'

const AUTH_URL = 'https://accounts.spotify.com/authorize'
const TOKEN_URL = 'https://accounts.spotify.com/api/token'

interface TokenResponse {
  access_token: string
  token_type: string
  scope?: string
  expires_in: number
  refresh_token?: string
}

interface TokenError {
  error: string
  error_description?: string
}

/**
 * Build Spotify authorization URL with PKCE parameters.
 * Port of buildAuthorizeURL from oauth.go:179-194
 */
export function buildAuthorizeURL(
  config: AppConfig,
  state: string,
  codeChallenge: string
): string {
  const redirectUri = getEffectiveRedirectUri(config.redirectUri)
  const params = new URLSearchParams({
    client_id: config.clientId,
    response_type: 'code',
    redirect_uri: redirectUri,
    state,
    code_challenge_method: 'S256',
    code_challenge: codeChallenge,
  })

  if (config.scopes.length > 0) {
    params.set('scope', config.scopes.join(' '))
  }

  return `${AUTH_URL}?${params.toString()}`
}

/**
 * Exchange authorization code for access token.
 * Port of exchangeCode from oauth.go:209-257
 */
export async function exchangeCode(
  config: AppConfig,
  code: string,
  codeVerifier: string
): Promise<Token> {
  const redirectUri = getEffectiveRedirectUri(config.redirectUri)
  const form = new URLSearchParams({
    client_id: config.clientId,
    grant_type: 'authorization_code',
    code,
    redirect_uri: redirectUri,
    code_verifier: codeVerifier,
  })

  const resp = await fetch(TOKEN_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: form.toString(),
  })

  const body = await resp.text()

  if (!resp.ok) {
    let te: TokenError | undefined
    try {
      te = JSON.parse(body) as TokenError
    } catch {
      // ignore
    }
    if (te?.error) {
      throw new Error(
        `Token exchange failed: ${te.error}${te.error_description ? ` (${te.error_description})` : ''}`
      )
    }
    throw new Error(`Token exchange failed: HTTP ${resp.status}`)
  }

  const tr = JSON.parse(body) as TokenResponse
  if (!tr.access_token) {
    throw new Error('Missing access_token in response')
  }

  return {
    access_token: tr.access_token,
    token_type: tr.token_type,
    scope: tr.scope,
    refresh_token: tr.refresh_token,
    expires_at: new Date(Date.now() + tr.expires_in * 1000).toISOString(),
  }
}

/**
 * Refresh an access token using a refresh token.
 * Port of Refresh from oauth.go:259-314
 */
export async function refreshToken(config: AppConfig, refreshToken: string): Promise<Token> {
  if (!refreshToken) {
    throw new Error('Missing refresh token')
  }

  const form = new URLSearchParams({
    client_id: config.clientId,
    grant_type: 'refresh_token',
    refresh_token: refreshToken,
  })

  const resp = await fetch(TOKEN_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded',
    },
    body: form.toString(),
  })

  const body = await resp.text()

  if (!resp.ok) {
    let te: TokenError | undefined
    try {
      te = JSON.parse(body) as TokenError
    } catch {
      // ignore
    }
    if (te?.error) {
      throw new Error(
        `Token refresh failed: ${te.error}${te.error_description ? ` (${te.error_description})` : ''}`
      )
    }
    throw new Error(`Token refresh failed: HTTP ${resp.status}`)
  }

  const tr = JSON.parse(body) as TokenResponse
  if (!tr.access_token) {
    throw new Error('Missing access_token in refresh response')
  }

  // Spotify may return a new refresh token; if not, keep the old one
  const newRefreshToken = tr.refresh_token || refreshToken

  return {
    access_token: tr.access_token,
    token_type: tr.token_type,
    scope: tr.scope,
    refresh_token: newRefreshToken,
    expires_at: new Date(Date.now() + tr.expires_in * 1000).toISOString(),
  }
}

/**
 * Check if a token is expired (with leeway in seconds).
 * Port of token.Expired from token.go:17-22
 */
export function isTokenExpired(token: Token | null, leewaySeconds = 60): boolean {
  if (!token || !token.access_token) {
    return true
  }
  const expiresAt = new Date(token.expires_at).getTime()
  const now = Date.now()
  return expiresAt - now <= leewaySeconds * 1000
}

/**
 * Initiate OAuth login flow.
 * Generates PKCE challenge, stores verifier and state, redirects browser.
 * Adapted from LoginPKCE in oauth.go (browser redirect instead of local server)
 */
export async function initiateLogin(config: AppConfig): Promise<void> {
  // Generate PKCE
  const pkce = await generatePKCE()
  storage.setPKCEVerifier(pkce.verifier)

  // Generate random state
  const stateBytes = new Uint8Array(24)
  crypto.getRandomValues(stateBytes)
  const state = Array.from(stateBytes, (b) => b.toString(16).padStart(2, '0')).join('')
  storage.setAuthState(state)

  // Build authorization URL and redirect
  const authURL = buildAuthorizeURL(config, state, pkce.challenge)
  window.location.href = authURL
}

/**
 * Handle OAuth callback.
 * Validates state, exchanges code for token, cleans up PKCE artifacts.
 * Called from CallbackPage component.
 */
export async function handleCallback(
  config: AppConfig,
  code: string,
  state: string
): Promise<Token> {
  // Validate state
  const storedState = storage.getAuthState()
  if (!storedState || storedState !== state) {
    throw new Error('Invalid state parameter')
  }

  // Retrieve PKCE verifier
  const verifier = storage.getPKCEVerifier()
  if (!verifier) {
    throw new Error('Missing PKCE verifier')
  }

  // Exchange code for token
  const token = await exchangeCode(config, code, verifier)

  // Clean up
  storage.clearPKCEVerifier()
  storage.clearAuthState()

  return token
}
