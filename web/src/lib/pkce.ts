// PKCE (Proof Key for Code Exchange) implementation - port of internal/auth/pkce.go

import type { PKCEChallenge } from '@/types/auth'

/**
 * Generate a PKCE challenge for OAuth authorization.
 * Uses Web Crypto API (crypto.getRandomValues + crypto.subtle.digest).
 * @returns PKCEChallenge with verifier and challenge (both base64url encoded)
 */
export async function generatePKCE(): Promise<PKCEChallenge> {
  const verifier = randomURLSafe(64)
  const challenge = await sha256Base64URL(verifier)
  return { verifier, challenge }
}

/**
 * Generate a random base64url-encoded string from n random bytes.
 * Port of randomURLSafe from pkce.go:27-33
 */
function randomURLSafe(n: number): string {
  const bytes = new Uint8Array(n)
  crypto.getRandomValues(bytes)
  return base64URLEncode(bytes)
}

/**
 * Compute SHA-256 hash of input and return base64url-encoded result.
 * Port of pkce.go:21-22
 */
async function sha256Base64URL(input: string): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(input)
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  return base64URLEncode(new Uint8Array(hashBuffer))
}

/**
 * Base64URL encoding (no padding) per RFC 7636.
 * Port of base64.RawURLEncoding from Go.
 */
function base64URLEncode(bytes: Uint8Array): string {
  let binary = ''
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i])
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
}
