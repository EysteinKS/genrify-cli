// App config types - ported from internal/config/config.go

export interface AppConfig {
  clientId: string
  redirectUri: string
  scopes: string[]
}

export const DEFAULT_CONFIG: AppConfig = {
  clientId: '',
  redirectUri: 'http://localhost:5173/callback',
  scopes: [
    'playlist-read-private',
    'playlist-read-collaborative',
    'playlist-modify-private',
    'playlist-modify-public',
  ],
}
