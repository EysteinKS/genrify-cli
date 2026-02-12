# macOS GTK3 GUI Fixes

This document explains the fixes applied to make the GTK3 GUI work on macOS.

## Issues Encountered

### 1. Missing Pixbuf Loaders
**Error:**
```
Gtk-WARNING **: Could not load a pixbuf from /org/gtk/libgtk/theme/Adwaita/assets/check-symbolic.svg.
This may indicate that pixbuf loaders or the mime database could not be found.
```

**Cause:** GTK3 couldn't find the pixbuf loaders needed to render icons and SVG images.

**Fix:**
- Updated the pixbuf loaders cache
- Set environment variables to point to the correct loader directory
- Created `run-gui.sh` launch script to set these automatically

### 2. NSWindow Main Thread Error
**Error:**
```
Terminating app due to uncaught exception 'NSInternalInconsistencyException',
reason: 'NSWindow should only be instantiated on the main thread!'
```

**Cause:** macOS requires all GUI operations (including GTK/NSWindow) to happen on the main OS thread. By default, Go's runtime can move goroutines between threads.

**Fix:**
- Added `runtime.LockOSThread()` in the `init()` function of `cmd/genrify/main.go`
- This ensures the main goroutine stays locked to the main OS thread
- Required for macOS Quartz backend compatibility

## Code Changes

### 1. Updated `cmd/genrify/main.go`
Added thread locking in the init function:

```go
func init() {
    // Lock the main goroutine to the OS thread.
    // This is required for GTK on macOS (Quartz backend) which requires
    // all GUI operations to happen on the main thread.
    runtime.LockOSThread()
}
```

This ensures that from the very start of the program, the main goroutine is locked to the OS thread, which is required for macOS GUI applications.

### 2. Created `run-gui.sh`
Launch script that sets required environment variables:

```bash
#!/bin/bash
export GDK_PIXBUF_MODULEDIR="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders"
export GDK_PIXBUF_MODULE_FILE="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache"
export GTK_THEME=Adwaita
exec ./genrify "$@"
```

### 3. Updated Documentation
- Enhanced `INSTALL_MACOS.md` with environment setup steps
- Added troubleshooting sections for common GTK3 issues on macOS

## Why These Issues Occur on macOS

### GTK3 on macOS Uses Quartz Backend
Unlike Linux (which uses X11 or Wayland), macOS uses the Quartz backend for GTK3. This backend bridges GTK calls to native macOS Cocoa/NSWindow APIs.

### macOS GUI Thread Requirements
macOS has strict requirements that all GUI operations must happen on the main thread. This is enforced by the Cocoa framework (NSWindow, NSView, etc.). Violating this causes the app to crash.

### Go's Runtime vs macOS Requirements
Go's runtime scheduler can move goroutines between OS threads. Without `runtime.LockOSThread()`, the main goroutine could move to a different thread, causing GUI operations to fail.

## How to Run the GUI

### Quick Start
```bash
# Update loaders cache (one-time setup)
gdk-pixbuf-query-loaders --update-cache

# Run with the launch script
./run-gui.sh
```

### Alternative: Set Environment Variables
```bash
export GDK_PIXBUF_MODULEDIR="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders"
export GDK_PIXBUF_MODULE_FILE="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache"
export GTK_THEME=Adwaita

./genrify
```

### Permanent Setup
Add to your `~/.zshrc`:
```bash
export GDK_PIXBUF_MODULEDIR="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders"
export GDK_PIXBUF_MODULE_FILE="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache"
export GTK_THEME=Adwaita
```

## Testing

```bash
# Build the GUI version
make build

# Test with launch script
./run-gui.sh version

# Launch GUI
./run-gui.sh
```

## Platform Differences

| Platform | Threading Requirements | Pixbuf Setup |
|----------|----------------------|--------------|
| Linux | No special requirements | Usually auto-configured |
| macOS | Must lock main thread | Requires manual environment vars |
| Windows | COM initialization required | Different setup needed |

## Additional Notes

- The thread locking has no negative effect on CLI-only usage
- These fixes are specific to macOS with the GTK3 Quartz backend
- Linux builds don't require these workarounds
- The `run-gui.sh` script is safe to use even if environment vars are already set

## References

- [GTK+ macOS Installation Guide](https://www.gtk.org/docs/installations/macos)
- [gotk3 GitHub Issues on macOS](https://github.com/gotk3/gotk3/issues)
- [Go runtime.LockOSThread documentation](https://pkg.go.dev/runtime#LockOSThread)
