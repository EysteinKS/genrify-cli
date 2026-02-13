# Genrify Web

Browser-based SPA for Spotify playlist management. Pure TypeScript React frontend with OAuth PKCE authentication.

## Features

- **Login**: OAuth PKCE authentication (no backend required)
- **Browse Playlists**: View and filter your Spotify playlists
- **View Tracks**: Load tracks from any playlist
- **Create Playlist**: Create new playlists with custom settings
- **Add Tracks**: Bulk add tracks to playlists
- **Merge Playlists**: Find playlists by regex pattern, merge, deduplicate, and delete sources

## Tech Stack

- **React 18** + **TypeScript** (strict mode)
- **Vite** (dev server + build)
- **TanStack Query v5** (server state)
- **React Router v6** (routing)
- **CSS Modules** (styling)

## Setup

### 1. Install Dependencies

```bash
cd web
npm install
```

### 2. Register Spotify App

1. Go to [Spotify Developer Dashboard](https://developer.spotify.com/dashboard)
2. Create a new app
3. Add redirect URI: `http://localhost:5173/callback`
4. Copy your **Client ID**

### 3. Configure

On first launch, you'll be prompted to enter your Spotify Client ID. The settings are stored in `localStorage`.

**Default scopes:**
- `playlist-read-private`
- `playlist-read-collaborative`
- `playlist-modify-private`
- `playlist-modify-public`

## Development

```bash
npm run dev
```

Opens dev server at http://localhost:5173

## Build

```bash
npm run build
```

Outputs to `dist/`

## Preview Production Build

```bash
npm run preview
```

## Testing

```bash
npm test
```

Runs Vitest unit tests.

## Architecture

```
src/
├── types/          # TypeScript type definitions (ported from Go)
├── lib/            # Core logic layer (no React dependencies)
│   ├── storage.ts       # localStorage wrapper
│   ├── pkce.ts          # OAuth PKCE generation
│   ├── auth.ts          # OAuth flow functions
│   ├── helpers.ts       # URI normalization
│   ├── spotify-client.ts  # Spotify API client
│   └── playlist-service.ts  # Merge/verify logic
├── contexts/       # React contexts (auth, config, status)
├── hooks/          # TanStack Query hooks
│   ├── queries/         # GET operations
│   └── mutations/       # POST/DELETE operations
├── components/     # Shared UI components
└── pages/          # Route pages
```

All TypeScript library code in `lib/` is a direct port of the corresponding Go packages.

## OAuth Flow

1. **Login**: Generate PKCE challenge, store verifier in `localStorage`, redirect to Spotify
2. **Callback**: Validate state, exchange code for token, store in `localStorage`
3. **Refresh**: Transparent token refresh on API calls (60s leeway)

No backend required - all OAuth is handled in the browser using PKCE.

## Comparison with Go Version

| Feature | Go CLI/GUI | Web SPA |
|---------|-----------|---------|
| Auth | Local HTTP server | OAuth redirect (PKCE) |
| Storage | Filesystem | localStorage |
| UI | GTK3 | React + CSS Modules |
| API Client | `http.Client` | `fetch()` |
| State | In-memory | TanStack Query |

All core business logic (URI normalization, merge, verification) is identical between Go and TypeScript implementations.
