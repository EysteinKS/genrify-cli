# Genrify Web - Quick Start Guide

## Prerequisites

- Node.js 18+ and npm
- Spotify Developer account

## Setup (5 minutes)

### 1. Register Spotify Application

```bash
# 1. Visit: https://developer.spotify.com/dashboard
# 2. Click "Create app"
# 3. Fill in:
#    - App name: "Genrify Local"
#    - App description: "Local Spotify playlist manager"
#    - Redirect URI: http://localhost:5173/callback
#    - Check "Web API"
# 4. Save and copy your Client ID
```

### 2. Install and Run

```bash
cd web
npm install
npm run dev
```

Open http://localhost:5173

### 3. Configure

On first visit:
1. Settings dialog will open automatically
2. Paste your Spotify Client ID
3. Verify redirect URI is `http://localhost:5173/callback`
4. Click Save

### 4. Login

1. Click "Login with Spotify"
2. Authorize the app on Spotify's page
3. You'll be redirected back to the app
4. Your display name will appear

## Features Overview

### Browse Playlists
- Navigate to **Playlists** in sidebar
- Filter by name (case-insensitive)
- Adjust limit (default 50)
- Click any row to view tracks

### View Tracks
- Navigate to **Tracks** in sidebar
- Enter playlist ID, URI, or URL
- Click "Load Tracks"
- Tracks display with name and artists

### Create Playlist
- Navigate to **Create Playlist** in sidebar
- Enter name (required) and description
- Toggle public/private
- Click "Create Playlist"

### Add Tracks
- Navigate to **Add Tracks** in sidebar
- Enter playlist ID
- Paste track URIs/URLs (one per line or comma-separated)
- Click "Add Tracks"
- Invalid tracks shown as warnings

### Merge Playlists
- Navigate to **Merge Playlists** in sidebar

**Step 1: Find**
- Enter regex pattern (e.g., `^Genre -` for all playlists starting with "Genre -")
- Click "Find Matches"
- Matched playlists display in table

**Step 2: Merge**
- Enter target playlist name
- Optionally add description
- Toggle public/private and deduplicate
- Click "Merge Playlists"

**Step 3: Results**
- View merge results (track count, duplicates removed, verification status)
- Optionally delete source playlists (only if verified)

## Troubleshooting

### "Invalid state parameter"
- Clear browser storage and try logging in again
- Make sure you're using the same browser/tab for login flow

### "Token expired"
- Logout and login again
- Check that your Client ID is correct in settings

### "Failed to find playlists"
- Check your regex pattern syntax
- Make sure you're logged in

### Build errors
- Delete `node_modules` and `package-lock.json`
- Run `npm install` again
- Make sure you're using Node.js 18+

## Development

### Run dev server
```bash
npm run dev
```

### Build for production
```bash
npm run build
```

### Preview production build
```bash
npm run preview
```

### Run tests
```bash
npm test
```

### Lint
```bash
npm run lint
```

### Format
```bash
npm run format
```

## Architecture

```
Browser (http://localhost:5173)
    ↓
Genrify Web (React SPA)
    ↓
Spotify Web API (https://api.spotify.com/v1)
```

No backend required - everything runs in your browser!

## Security Notes

- Client ID is not secret (it's designed to be public)
- Tokens are stored in browser localStorage (cleared on logout)
- OAuth PKCE ensures secure authentication without client secret
- All API calls go directly to Spotify (no proxy)

## Next Steps

- Add your own features
- Customize the theme in `src/globals.css`
- Add unit tests in `src/__tests__/`
- Deploy to static hosting (Netlify, Vercel, GitHub Pages)

## Support

- [Spotify Web API Docs](https://developer.spotify.com/documentation/web-api)
- [React Query Docs](https://tanstack.com/query/latest)
- [React Router Docs](https://reactrouter.com/)
