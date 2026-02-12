# genrify

[![CI](https://github.com/EysteinKS/genrify-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/EysteinKS/genrify-cli/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/EysteinKS/genrify-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/EysteinKS/genrify-cli)

CLI and GUI for interacting with Spotify (login + playlists).

## Features

- 🖥️ **GTK3 GUI** - Graphical interface with all playlist features
- 💻 **Interactive TUI** - Terminal-based menu interface
- ⌨️ **CLI commands** - Direct command-line operations
- 🔐 **OAuth PKCE** - Secure authentication with auto-generated certificates
- 📋 **Playlist management** - Create, merge, add tracks, and more

## Prereqs

- A Spotify app (Client ID)

Optional (only if you run from source):

- Go 1.22+

## Install

### GUI Version (with GTK3)

The GUI version provides a graphical interface for all features.

#### macOS

1. Install GTK3 using Homebrew:
   ```sh
   brew install gtk+3 pkg-config
   ```

2. Download the GUI build for macOS from the [GUI Releases](https://github.com/EysteinKS/genrify-cli/releases) (look for `genrify-gui_*_darwin_*.tar.gz`)

3. Extract and run:
   ```sh
   tar xzf genrify-gui_*_darwin_*.tar.gz
   ./genrify-gui-darwin-*
   ```

#### Linux

1. Install GTK3:
   ```sh
   # Ubuntu/Debian
   sudo apt-get install libgtk-3-0

   # Fedora/RHEL
   sudo dnf install gtk3

   # Arch
   sudo pacman -S gtk3
   ```

2. Download the GUI build for Linux from the [GUI Releases](https://github.com/EysteinKS/genrify-cli/releases) (look for `genrify-gui_*_linux_*.tar.gz`)

3. Extract and run:
   ```sh
   tar xzf genrify-gui_*_linux_*.tar.gz
   ./genrify-gui-linux-*
   ```

### CLI-Only Version (no GTK3 required)

Download the latest release asset for your OS from GitHub Releases (not the "Source code" zip), unzip it, and run `genrify`.

macOS note: if you get "Apple could not verify \"genrify\" is free of malware", this is Gatekeeper blocking an unsigned binary.
If you trust the download, you can either:

- Finder → right-click `genrify` → **Open** → **Open**
- Or remove the quarantine attribute:

```sh
xattr -d com.apple.quarantine /path/to/genrify
```

On first run, `genrify` will ask for the required Spotify settings and save them to a config file in your user config directory.

### Build from Source

#### CLI-only (no GTK3 required)
```sh
make build-cli
./genrify version
```

Or directly:
```sh
CGO_ENABLED=0 go build -tags nogui -o genrify ./cmd/genrify
```

#### GUI version (requires GTK3 installed)
```sh
# First install GTK3 (see instructions above)
make build
./genrify version
```

Or directly:
```sh
CGO_ENABLED=1 go build -o genrify ./cmd/genrify
```

## Config

On first run, the CLI will prompt you for config and save it.

Advanced: you can still configure via environment variables (they override the saved config):

- `SPOTIFY_CLIENT_ID` (required)
- `SPOTIFY_REDIRECT_URI` (optional, default: `http://localhost:8888/callback`)
- `SPOTIFY_SCOPES` (optional, default is playlist read/write scopes)
- `SPOTIFY_TLS_CERT_FILE` / `SPOTIFY_TLS_KEY_FILE` (required only if `SPOTIFY_REDIRECT_URI` is `https://...`)

## HTTPS redirect

If your Spotify app uses an `https://localhost:...` redirect URI, Genrify will automatically generate self-signed certificates on first use. The certificates are stored in `~/.config/genrify/.certs/`.

### Manual certificate generation (optional)

If you prefer to use mkcert for browser-trusted certificates:

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

### GUI Mode (if GTK3 is installed)

Simply run `genrify` without arguments to launch the GUI:

```sh
genrify
# or explicitly
genrify gui
```

The GUI provides:
- 🔐 Login/logout interface
- 📋 Browse and search playlists
- 🎵 View playlist tracks
- ➕ Create new playlists
- 🎶 Add tracks to playlists
- 🔀 Merge multiple playlists with deduplication

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

## Development

### Build

```sh
# Build CLI-only version (no GTK3 required)
make build-cli

# Build GUI version (requires GTK3)
make build
```

### Test

```sh
# Run tests with nogui tag
go test -tags nogui ./...

# Run tests with coverage
make test

# View coverage
go tool cover -func=coverage.out
```

### Lint

```sh
make lint
```

### Build Tags

- Default build includes GUI (requires GTK3 and `CGO_ENABLED=1`)
- Build with `-tags nogui` for CLI-only version (no GTK3 required, allows `CGO_ENABLED=0`)
- GoReleaser uses `nogui` tag for cross-platform releases
