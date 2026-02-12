# Installing Genrify with GUI on Windows

This guide will help you install and run Genrify with the GTK3 GUI on Windows.

## Prerequisites

- Windows 10 or later
- Administrator access for installation

## Step 1: Install GTK3 Runtime

You have two options for installing GTK3 on Windows:

### Option A: MSYS2 (Recommended)

MSYS2 provides a Unix-like environment with a package manager for easy GTK3 installation.

1. **Download MSYS2:**
   - Visit https://www.msys2.org/
   - Download the installer (e.g., `msys2-x86_64-*.exe`)
   - Run the installer and follow the prompts
   - Default installation path: `C:\msys64`

2. **Update MSYS2:**
   Open MSYS2 MINGW64 terminal (from Start Menu) and run:
   ```bash
   pacman -Syu
   ```
   Close the terminal when prompted, then reopen and run:
   ```bash
   pacman -Su
   ```

3. **Install GTK3:**
   In the MSYS2 MINGW64 terminal:
   ```bash
   pacman -S mingw-w64-x86_64-gtk3 mingw-w64-x86_64-pkg-config
   ```

4. **Add to PATH:**
   - Open Windows Settings → System → About → Advanced system settings
   - Click "Environment Variables"
   - Under "System variables", find "Path" and click "Edit"
   - Click "New" and add: `C:\msys64\mingw64\bin`
   - Click OK to save

### Option B: GTK Runtime Bundle

1. Download the GTK3 Runtime from the [GTK website](https://www.gtk.org/docs/installations/windows)
2. Run the installer
3. Add the GTK bin directory to your PATH

## Step 2: Verify GTK3 Installation

Open Command Prompt or PowerShell and run:
```cmd
pkg-config --modversion gtk+-3.0
```

You should see a version number like `3.24.x`.

## Step 3: Install Genrify

### Option A: Download Pre-built Binary (Recommended)

1. Go to the [GUI Releases](https://github.com/EysteinKS/genrify-cli/releases) page

2. Download the Windows GUI build: `genrify-gui_*_windows_amd64.zip`

3. Extract the ZIP file to a folder (e.g., `C:\Program Files\Genrify`)

4. **Optional:** Add to PATH:
   - Add the Genrify folder to your PATH (same process as Step 1.4)
   - This allows you to run `genrify` from any directory

### Option B: Build from Source

If you prefer to build from source:

1. **Install Go:**
   - Download from https://go.dev/dl/
   - Install Go 1.22 or later

2. **Install Git:**
   - Download from https://git-scm.com/download/win

3. **Clone and Build:**
   Open Command Prompt or MSYS2 MINGW64 terminal:
   ```bash
   git clone https://github.com/EysteinKS/genrify-cli.git
   cd genrify-cli

   # Set environment for MSYS2
   export PATH="/mingw64/bin:$PATH"
   export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig"

   # Build
   CGO_ENABLED=1 go build -o genrify.exe ./cmd/genrify
   ```

## Step 4: Run Genrify

### Using Command Prompt
```cmd
cd "C:\Program Files\Genrify"
genrify.exe
```

### Using PowerShell
```powershell
cd "C:\Program Files\Genrify"
.\genrify.exe
```

### If Added to PATH
Simply run from any directory:
```cmd
genrify
```

This will launch the GUI by default.

### First Run Configuration

On first run, Genrify will show a configuration dialog where you can enter:
- Spotify Client ID (required - get from [Spotify Developer Dashboard](https://developer.spotify.com/dashboard))
- Redirect URI (default: `http://localhost:8888/callback`)
- Scopes (default includes playlist read/write)
- Auto-generate certificates option (for HTTPS redirects)

## Alternative: CLI-Only Version

If you prefer not to install GTK3, download the CLI-only version from the main [Releases](https://github.com/EysteinKS/genrify-cli/releases) page. This version works without GTK3 and provides the same functionality through a terminal interface.

## Troubleshooting

### "The code execution cannot proceed because gtk-3.dll was not found"

GTK3 is not properly installed or not in your PATH. Solutions:
1. Verify MSYS2 installation: `where gtk-3.dll` should show `C:\msys64\mingw64\bin\gtk-3.dll`
2. Ensure `C:\msys64\mingw64\bin` is in your PATH
3. Restart Command Prompt/PowerShell after modifying PATH

### "pkg-config: command not found"

pkg-config is not in your PATH:
```bash
# In MSYS2 MINGW64 terminal
pacman -S mingw-w64-x86_64-pkg-config
```

### GUI doesn't launch

Try running from MSYS2 MINGW64 terminal:
```bash
cd /c/Program\ Files/Genrify
./genrify.exe
```

Check for error messages that might indicate missing DLLs.

### "Could not load a pixbuf" warning

This warning is usually harmless but if icons don't show:
```bash
# In MSYS2 MINGW64 terminal
gdk-pixbuf-query-loaders --update-cache
```

### Firewall Blocking OAuth

When logging in, Windows Firewall may block the OAuth callback server:
1. Click "Allow access" when prompted
2. Or manually add an exception for genrify.exe

## Building from Source (Advanced)

### Requirements
- Go 1.22+
- MSYS2 with GTK3
- Git

### Build Commands

Using MSYS2 MINGW64 terminal:
```bash
# Set up environment
export PATH="/mingw64/bin:$PATH"
export PKG_CONFIG_PATH="/mingw64/lib/pkgconfig"

# Clone repository
git clone https://github.com/EysteinKS/genrify-cli.git
cd genrify-cli

# Build GUI version
CGO_ENABLED=1 go build -o genrify.exe ./cmd/genrify

# Build CLI-only version (no GTK3 required)
CGO_ENABLED=0 go build -tags nogui -o genrify-cli.exe ./cmd/genrify
```

## Uninstalling

1. Delete the Genrify folder
2. Remove from PATH if added
3. **Optional:** Uninstall MSYS2 or GTK3 Runtime if no longer needed
4. **Optional:** Delete config: `%USERPROFILE%\.config\genrify`

## Getting Help

- [GitHub Issues](https://github.com/EysteinKS/genrify-cli/issues)
- [Main README](README.md)
- [macOS Installation Guide](INSTALL_MACOS.md)
