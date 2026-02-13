// App config types - ported from internal/config/config.go

export interface AppConfig {
  clientId: string
  redirectUri: string
  scopes: string[]
}

export const DEFAULT_CONFIG: AppConfig = {
  clientId: '',
  // Default is resolved at runtime from the current app origin + BASE_URL + /callback.
  redirectUri: '',
  scopes: [
    'playlist-read-private',
    'playlist-read-collaborative',
    'playlist-modify-private',
    'playlist-modify-public',
  ],
}
