# genrify

CLI for interacting with Spotify (login + playlists).

## Prereqs

- Go 1.22+
- A Spotify app (Client ID)

## Config

Environment variables:

- `SPOTIFY_CLIENT_ID` (required)
- `SPOTIFY_REDIRECT_URI` (optional, default: `http://localhost:8888/callback`)
- `SPOTIFY_SCOPES` (optional, default is playlist read/write scopes)
- `SPOTIFY_TLS_CERT_FILE` / `SPOTIFY_TLS_KEY_FILE` (required only if `SPOTIFY_REDIRECT_URI` is `https://...`)

## HTTPS redirect (mkcert)

If your Spotify app uses an `https://localhost:...` redirect URI, the CLI must serve HTTPS locally.

```sh
brew install mkcert
mkcert -install

mkdir -p .certs
mkcert -cert-file .certs/localhost.pem -key-file .certs/localhost-key.pem localhost 127.0.0.1 ::1

export SPOTIFY_REDIRECT_URI='https://localhost:8888/callback'
export SPOTIFY_TLS_CERT_FILE="$PWD/.certs/localhost.pem"
export SPOTIFY_TLS_KEY_FILE="$PWD/.certs/localhost-key.pem"
```

## Usage

### Login

```sh
go run ./cmd/genrify login
```

### Interactive mode

```sh
go run ./cmd/genrify start
```

The `start` command launches an interactive menu where you can:
- List playlists (with filtering)
- Show tracks from a playlist
- Create new playlists
- Add tracks to playlists

### Command-line mode

```sh
# List playlists
go run ./cmd/genrify playlists list
go run ./cmd/genrify playlists list --filter "workout" --limit 10

# Show tracks
go run ./cmd/genrify playlists tracks <playlist-id> --limit 50

# Create playlist
go run ./cmd/genrify playlists create --name "My Playlist" --description "Made by genrify"

# Add tracks
go run ./cmd/genrify playlists add <playlist-id> spotify:track:<id> https://open.spotify.com/track/<id>
```

## Dev

```sh
go test ./...
```
