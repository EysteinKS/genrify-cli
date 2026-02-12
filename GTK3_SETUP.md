# GTK3 GitHub Actions Setup Summary

This document summarizes the GitHub Actions setup for building Genrify with GTK3 support.

## Changes Made

### 1. CI Workflow Updates (`.github/workflows/ci.yml`)

#### Updated Test Job
- Now runs tests with `-tags nogui` to avoid GTK3 dependency
- Ensures tests pass in CI environment without GTK3

#### New: Build CLI-Only Job
- Tests building the CLI-only version with `-tags nogui`
- Verifies `CGO_ENABLED=0` builds work
- Runs on Ubuntu without GTK3 dependencies

#### New: Build GUI Job
- Installs GTK3 dependencies (`libgtk-3-dev pkg-config`)
- Builds the GUI version with `CGO_ENABLED=1`
- Uploads Linux GUI binary as artifact
- Verifies GUI build compiles successfully

### 2. New GUI Release Workflow (`.github/workflows/release-gui.yml`)

Creates separate GUI releases with GTK3 support:

#### Linux GUI Build
- Runs on `ubuntu-latest`
- Installs GTK3 dependencies
- Builds for Linux amd64
- Creates tarball with version in filename

#### Windows GUI Build
- Runs on `windows-latest`
- Installs GTK3 via MSYS2
- Builds for Windows amd64
- Packages with essential DLLs as ZIP archive

#### macOS GUI Build
- Runs on `macos-latest`
- Installs GTK3 via Homebrew
- Builds for both macOS architectures:
  - arm64 (Apple Silicon)
  - amd64 (Intel)
- Creates separate tarballs for each architecture

#### Release Creation
- Combines artifacts from all platforms
- Creates draft release with:
  - All GUI binaries (Linux + macOS)
  - Installation instructions in release notes
  - Requirements for each platform

### 3. Documentation Updates

#### README.md
- Added features section highlighting GUI
- Added GUI installation instructions for macOS and Linux
- Updated build instructions for both CLI and GUI versions
- Added GUI usage section
- Updated development section with build tags info
- Mentioned auto-certificate generation

#### New: INSTALL_MACOS.md
Comprehensive macOS installation guide with:
- Step-by-step Homebrew installation
- GTK3 installation and verification
- Both pre-built and source build options
- Troubleshooting common issues
- Uninstallation instructions

## Workflow Behavior

### On Every Push/PR
The CI workflow runs:
1. **Lint** - Checks code with nogui tag
2. **Test** - Runs all tests with nogui tag
3. **Build CLI** - Verifies CLI-only build works
4. **Build GUI** - Verifies GUI build compiles (Ubuntu + GTK3)

### On Version Tag Push
Both workflows run:
1. **Main Release** (`.github/workflows/release.yml`):
   - Creates cross-platform CLI-only releases via GoReleaser
   - Uses `-tags nogui` for maximum compatibility
   - No GTK3 required to run

2. **GUI Release** (`.github/workflows/release-gui.yml`):
   - Creates platform-specific GUI builds
   - Linux amd64 with GTK3
   - macOS arm64 and amd64 with GTK3
   - Requires users to install GTK3 to run

## Platform Support Matrix

| Platform | CLI-Only | GUI | Notes |
|----------|----------|-----|-------|
| Linux (amd64) | ✅ | ✅ | GUI requires `libgtk-3-0` |
| Linux (arm64) | ✅ | ⚠️ | Possible but not in CI |
| macOS (Intel) | ✅ | ✅ | GUI requires `brew install gtk+3` |
| macOS (Apple Silicon) | ✅ | ✅ | GUI requires `brew install gtk+3` |
| Windows (amd64) | ✅ | ✅ | GUI requires MSYS2/GTK3 Runtime |

## GTK3 Dependencies by Platform

### Ubuntu/Debian
```bash
sudo apt-get install libgtk-3-dev pkg-config  # Build time
sudo apt-get install libgtk-3-0                # Runtime
```

### macOS
```bash
brew install gtk+3 pkg-config  # Build and runtime
```

### Windows
```bash
# Using MSYS2 MINGW64 terminal
pacman -S mingw-w64-x86_64-gtk3 mingw-w64-x86_64-pkg-config mingw-w64-x86_64-gcc

# Add to PATH: C:\msys64\mingw64\bin
```

### Fedora/RHEL
```bash
sudo dnf install gtk3-devel pkg-config  # Build time
sudo dnf install gtk3                   # Runtime
```

## Testing Locally

### Test CLI Build (no GTK3)
```bash
CGO_ENABLED=0 go build -tags nogui -o genrify ./cmd/genrify
./genrify version
```

### Test GUI Build (requires GTK3)
```bash
# macOS: brew install gtk+3 pkg-config
# Linux: sudo apt-get install libgtk-3-dev pkg-config

CGO_ENABLED=1 go build -o genrify ./cmd/genrify
./genrify version
./genrify gui
```

### Test with Make
```bash
make build-cli  # CLI-only
make build      # GUI (requires GTK3)
```

## Release Process

1. **Tag a version**: `git tag v0.2.0 && git push origin v0.2.0`
2. **Main release workflow** creates CLI-only cross-platform binaries
3. **GUI release workflow** creates platform-specific GUI binaries
4. **Two releases appear**:
   - Main release: CLI-only, works everywhere
   - GUI release: Platform-specific, requires GTK3

## Future Enhancements

Potential improvements:
- [ ] Add Windows GUI build (requires GTK3 setup on Windows runners)
- [ ] Add Linux arm64 GUI build
- [ ] Create macOS .app bundle with bundled GTK3
- [ ] Add AppImage for Linux (bundles GTK3)
- [ ] Add Flatpak for Linux
