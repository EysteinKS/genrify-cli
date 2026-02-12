# genrify

[![CI](https://github.com/EysteinKS/genrify-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/EysteinKS/genrify-cli/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/EysteinKS/genrify-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/EysteinKS/genrify-cli)

CLI for interacting with Spotify (login + playlists).

## Prereqs

- A Spotify app (Client ID)

Optional (only if you run from source):

- Go 1.22+

## Install

### Prebuilt executable (recommended)

Download the latest release for your OS from GitHub Releases, unzip it, and run `genrify`.

On first run, `genrify` will ask for the required Spotify settings and save them to a config file in your user config directory.

### Run from source

```sh
go run ./cmd/genrify version
```

## Config

On first run, the CLI will prompt you for config and save it.

Advanced: you can still configure via environment variables (they override the saved config):

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

If you installed the prebuilt executable, run:

```sh
genrify login
```

### Interactive mode

```sh
go run ./cmd/genrify start
```

With the prebuilt executable:

```sh
genrify start
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

# Race + coverage (writes ./coverage.out)
go test ./... -race -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out
```
