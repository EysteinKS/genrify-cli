# Installing Genrify with GUI on macOS

This guide will help you install and run Genrify with the GTK3 GUI on your MacBook Pro.

## Prerequisites

- macOS 11 (Big Sur) or later
- Homebrew package manager

## Step 1: Install Homebrew (if not already installed)

If you don't have Homebrew installed, open Terminal and run:

```sh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Follow the on-screen instructions to complete the installation.

## Step 2: Install GTK3 and Dependencies

GTK3 is required for the GUI. Install it using Homebrew:

```sh
brew install gtk+3 pkg-config
```

This will install:
- GTK+ 3.x libraries
- pkg-config (required for building with CGO)
- All necessary dependencies (GLib, Cairo, Pango, etc.)

The installation may take a few minutes.

## Step 3: Verify GTK3 Installation

Check that GTK3 is properly installed:

```sh
pkg-config --modversion gtk+-3.0
```

You should see a version number like `3.24.x`.

## Step 4: Install Genrify

### Option A: Download Pre-built Binary (Recommended)

1. Go to the [GUI Releases](https://github.com/EysteinKS/genrify-cli/releases) page

2. Download the appropriate file for your Mac:
   - For Apple Silicon (M1/M2/M3): `genrify-gui_*_darwin_arm64.tar.gz`
   - For Intel Macs: `genrify-gui_*_darwin_amd64.tar.gz`

3. Extract the archive:
   ```sh
   cd ~/Downloads
   tar xzf genrify-gui_*_darwin_*.tar.gz
   ```

4. Move the binary to a location in your PATH:
   ```sh
   sudo mv genrify-gui-darwin-* /usr/local/bin/genrify
   sudo chmod +x /usr/local/bin/genrify
   ```

5. If you get a Gatekeeper warning when first running, either:
   - Right-click the binary in Finder → Open → Open
   - Or remove the quarantine attribute:
     ```sh
     xattr -d com.apple.quarantine /usr/local/bin/genrify
     ```

### Option B: Build from Source

If you prefer to build from source:

1. Install Go 1.22 or later:
   ```sh
   brew install go
   ```

2. Clone the repository:
   ```sh
   git clone https://github.com/EysteinKS/genrify-cli.git
   cd genrify-cli
   ```

3. Build the GUI version:
   ```sh
   make build
   ```

4. Optionally, move to PATH:
   ```sh
   sudo mv genrify /usr/local/bin/
   ```

## Step 5: Run Genrify

Simply run:

```sh
genrify
```

This will launch the GUI by default (since GTK3 is installed).

### First Run Configuration

On first run, Genrify will show a configuration dialog where you can enter:
- Spotify Client ID (required - get this from your [Spotify Developer Dashboard](https://developer.spotify.com/dashboard))
- Redirect URI (default: `http://localhost:8888/callback`)
- Scopes (default includes playlist read/write)
- Auto-generate certificates option (for HTTPS redirects)

## Alternative: CLI-Only Version

If you prefer not to install GTK3, you can download the CLI-only version from the main [Releases](https://github.com/EysteinKS/genrify-cli/releases) page. This version works without GTK3 and provides the same functionality through a terminal interface.

## Troubleshooting

### "dyld: Library not loaded" error

If you see an error about missing GTK libraries:

```sh
# Make sure GTK3 is installed
brew install gtk+3

# Check that libraries are linked
brew link gtk+3
```

### "command not found: genrify"

Make sure the binary is in your PATH:
```sh
echo $PATH
```

If `/usr/local/bin` is not in your PATH, add it to your shell profile:
```sh
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### GUI doesn't launch

Try running explicitly:
```sh
genrify gui
```

If it still doesn't work, check that GTK3 is installed:
```sh
brew list gtk+3
```

### "Apple could not verify..."

This is macOS Gatekeeper blocking unsigned binaries. To allow:
```sh
xattr -d com.apple.quarantine /path/to/genrify
```

Or right-click → Open → Open in Finder.

## Uninstalling

To remove Genrify:

```sh
# Remove the binary
sudo rm /usr/local/bin/genrify

# Remove config (optional)
rm -rf ~/.config/genrify

# Remove GTK3 (optional)
brew uninstall gtk+3
```

## Getting Help

- [GitHub Issues](https://github.com/EysteinKS/genrify-cli/issues)
- [Main README](README.md)
