# CI and Windows Support Fixes

This document explains the fixes for the CI linting errors and the addition of Windows GUI support.

## Issue 1: CI Linting Error ✅ FIXED

### Problem
```
Error: internal/gui/async.go:6:2: could not import github.com/gotk3/gotk3/glib
```

The golangci-lint step in the test job was trying to lint GUI code, but GTK3 headers weren't available in the CI environment.

### Root Cause
The test job had a duplicate `golangci-lint` step that wasn't using the `nogui` build tag, causing it to try to compile GUI code without GTK3 installed.

### Solution
Updated `.github/workflows/ci.yml` to add `--build-tags=nogui` to the golangci-lint args in the test job:

```yaml
- name: golangci-lint (required)
  uses: golangci/golangci-lint-action@v6
  with:
    version: v1.63.4
    args: --timeout=5m --build-tags=nogui
```

Now the linter only checks code that doesn't require GTK3.

## Issue 2: Windows GUI Support ✅ ADDED

### Requirements
Windows GUI builds require:
- MSYS2 (provides Unix-like environment)
- GTK3 via MSYS2 package manager
- MinGW-w64 GCC compiler

### Implementation

#### 1. Added Windows Build Job to `.github/workflows/release-gui.yml`

```yaml
build-gui-windows:
  name: Build GUI for Windows
  runs-on: windows-latest

  steps:
    - name: Install MSYS2 and GTK3
      uses: msys2/setup-msys2@v2
      with:
        msystem: MINGW64
        update: true
        install: >-
          mingw-w64-x86_64-gtk3
          mingw-w64-x86_64-pkg-config
          mingw-w64-x86_64-gcc

    - name: Build GUI for Windows amd64
      shell: msys2 {0}
      run: |
        export PATH="/mingw64/bin:$PATH"
        export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig"
        CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build \
          -ldflags="-s -w -X genrify/internal/buildinfo.Version=... -H windowsgui" \
          -o genrify-gui-windows-amd64.exe ./cmd/genrify
```

#### 2. Package Structure
The Windows build creates a ZIP file containing:
- `genrify.exe` - The GUI application
- Required GTK3 DLLs (copied from MSYS2)

#### 3. Updated Release Job
Modified the release job to include Windows artifacts:

```yaml
needs: [build-gui-linux, build-gui-windows, build-gui-macos]
```

## What's Different About Windows

### GTK3 on Windows
Unlike Linux and macOS, Windows doesn't have a native GTK3 package. We use MSYS2 which provides:
- MinGW-w64 toolchain (GCC compiler for Windows)
- GTK3 libraries compiled for Windows
- pkg-config for build configuration

### Build Flags
The Windows build uses `-H windowsgui` linker flag to create a proper Windows GUI application (no console window).

### Runtime Requirements
Users must install GTK3 runtime on their Windows system:
- **Option A:** MSYS2 (recommended) - Full Unix-like environment
- **Option B:** GTK3 Runtime Bundle - Minimal GTK3 only

## Testing Locally

### Verify CI Fix
```bash
go test -tags nogui ./...
```

### Test Windows Build (if you have MSYS2)
In MSYS2 MINGW64 terminal:
```bash
export PATH="/mingw64/bin:$PATH"
export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig"
CGO_ENABLED=1 go build -o genrify.exe ./cmd/genrify
./genrify.exe version
```

## Documentation Updates

### New Files
1. **`INSTALL_WINDOWS.md`** - Comprehensive Windows installation guide
   - MSYS2 installation steps
   - GTK3 setup
   - Building from source
   - Troubleshooting

### Updated Files
1. **`README.md`** - Added Windows section to installation instructions
2. **`GTK3_SETUP.md`** - Updated platform support matrix and dependencies
3. **`.github/workflows/ci.yml`** - Fixed linting to use nogui tag
4. **`.github/workflows/release-gui.yml`** - Added Windows build job

## Platform Support Summary

| Platform | Status | Package Format | GTK3 Installation |
|----------|--------|----------------|-------------------|
| Linux (amd64) | ✅ | .tar.gz | `apt/dnf/pacman install gtk3` |
| macOS (Intel) | ✅ | .tar.gz | `brew install gtk+3` |
| macOS (Apple Silicon) | ✅ | .tar.gz | `brew install gtk+3` |
| Windows (amd64) | ✅ | .zip | MSYS2 or GTK Runtime |

## CI Workflow

All workflows now properly handle build tags:

1. **Lint Job** - Uses `--build-tags=nogui`
2. **Test Job** - Uses `-tags nogui`
3. **Build CLI Job** - Uses `-tags nogui`
4. **Build GUI Job** - Installs GTK3 (Linux only in CI)
5. **Release GUI Job** - Builds for Linux, Windows, macOS

## Release Process

When you push a version tag:

1. **Main Release Workflow** runs:
   - Creates cross-platform CLI-only binaries
   - No GTK3 required to run

2. **GUI Release Workflow** runs:
   - Builds Linux GUI (Ubuntu + GTK3)
   - Builds Windows GUI (MSYS2 + GTK3)
   - Builds macOS GUI (Homebrew + GTK3)
   - Creates draft release with all platforms

## User Experience

### CLI-Only Users (Main Release)
- Download from main releases
- No GTK3 installation needed
- Works everywhere Go works
- Uses terminal UI (`genrify start`)

### GUI Users (GUI Release)
- Download platform-specific GUI release
- Install GTK3 runtime first
- Get full graphical interface
- Can still use terminal UI if desired

## Future Improvements

Potential enhancements:
- [ ] Bundle GTK3 DLLs with Windows build (larger but standalone)
- [ ] Create Windows installer (.msi) with GTK3 runtime
- [ ] Create macOS .app bundle
- [ ] Linux AppImage or Flatpak (bundles GTK3)
- [ ] Add Windows ARM64 support
