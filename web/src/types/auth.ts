// Auth types - ported from internal/auth/token.go

export interface Token {
  access_token: string
  token_type: string
  scope?: string
  expires_at: string // ISO 8601 string (time.Time in Go)
  refresh_token?: string
}

export interface PKCEChallenge {
  verifier: string
  challenge: string
}
