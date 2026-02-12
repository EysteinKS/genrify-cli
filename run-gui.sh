#!/bin/bash

# Launch script for Genrify GUI on macOS
# This sets up required environment variables for GTK3

# Set GTK/GDK environment variables
export GDK_PIXBUF_MODULEDIR="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders"
export GDK_PIXBUF_MODULE_FILE="$(brew --prefix)/lib/gdk-pixbuf-2.0/2.10.0/loaders.cache"

# Set GTK theme and settings
export GTK_THEME=Adwaita
export GTK_DATA_PREFIX="$(brew --prefix gtk+3)"

# Launch the GUI
exec ./genrify "$@"
