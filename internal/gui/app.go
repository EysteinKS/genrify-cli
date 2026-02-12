//go:build !nogui

package gui

import (
	"context"
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/config"
)

// App represents the GUI application state.
type App struct {
	cfg    config.Config
	ctx    context.Context
	window *Window
	client SpotifyClient

	// Views.
	loginView      *LoginView
	playlistsView  *PlaylistsView
	tracksView     *TracksView
	createView     *CreateView
	addTracksView  *AddTracksView
	mergeView      *MergeView

	// Dependencies injected from CLI layer.
	opts Options
}

// Options contains dependencies injected from the CLI layer.
type Options struct {
	DoLogin          func(context.Context, config.Config) (string, error)
	NewSpotifyClient func(config.Config) (SpotifyClient, error)
	LoadConfig       func() (config.Config, error)
	SaveConfig       func(config.Config) (string, error)
}

// Run initializes GTK and runs the GUI application.
func Run(cfg config.Config, opts Options) error {
	// Note: runtime.LockOSThread() is called in main.go init() for macOS compatibility.
	gtk.Init(nil)

	app := &App{
		cfg:  cfg,
		ctx:  context.Background(),
		opts: opts,
	}

	// Show config dialog if not configured.
	if !config.IsConfigured(cfg) {
		if err := app.showConfigDialog(); err != nil {
			return fmt.Errorf("configuration: %w", err)
		}
	}

	// Build main window.
	window, err := NewWindow(app)
	if err != nil {
		return fmt.Errorf("create window: %w", err)
	}
	app.window = window

	// Create and add views.
	if err := app.setupViews(); err != nil {
		return fmt.Errorf("setup views: %w", err)
	}

	// Show window and run main loop.
	window.Show()
	gtk.Main()

	return nil
}

func (app *App) setupViews() error {
	// Login view.
	loginView, err := NewLoginView(app)
	if err != nil {
		return fmt.Errorf("create login view: %w", err)
	}
	app.loginView = loginView
	app.window.AddView("login", "Login", loginView.Widget())

	// Playlists view.
	playlistsView, err := NewPlaylistsView(app)
	if err != nil {
		return fmt.Errorf("create playlists view: %w", err)
	}
	app.playlistsView = playlistsView
	app.window.AddView("playlists", "Playlists", playlistsView.Widget())

	// Tracks view.
	tracksView, err := NewTracksView(app)
	if err != nil {
		return fmt.Errorf("create tracks view: %w", err)
	}
	app.tracksView = tracksView
	app.window.AddView("tracks", "Tracks", tracksView.Widget())

	// Create view.
	createView, err := NewCreateView(app)
	if err != nil {
		return fmt.Errorf("create playlist view: %w", err)
	}
	app.createView = createView
	app.window.AddView("create", "Create Playlist", createView.Widget())

	// Add tracks view.
	addTracksView, err := NewAddTracksView(app)
	if err != nil {
		return fmt.Errorf("create add tracks view: %w", err)
	}
	app.addTracksView = addTracksView
	app.window.AddView("add-tracks", "Add Tracks", addTracksView.Widget())

	// Merge view.
	mergeView, err := NewMergeView(app)
	if err != nil {
		return fmt.Errorf("create merge view: %w", err)
	}
	app.mergeView = mergeView
	app.window.AddView("merge", "Merge Playlists", mergeView.Widget())

	return nil
}

func (app *App) showConfigDialog() error {
	dialog, err := NewConfigDialog(app.cfg, func(newCfg config.Config) error {
		app.cfg = newCfg
		path, err := app.opts.SaveConfig(newCfg)
		if err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		fmt.Printf("Saved config to: %s\n", path)
		return nil
	})
	if err != nil {
		return err
	}

	response := dialog.Run()
	dialog.Destroy()

	if response != gtk.RESPONSE_OK {
		return fmt.Errorf("configuration cancelled")
	}

	return nil
}

// Client returns the Spotify client, initializing it if needed.
func (app *App) Client() (SpotifyClient, error) {
	if app.client == nil {
		client, err := app.opts.NewSpotifyClient(app.cfg)
		if err != nil {
			return nil, err
		}
		app.client = client
	}
	return app.client, nil
}

// RefreshClient clears the cached client, forcing reinitialization on next access.
func (app *App) RefreshClient() {
	app.client = nil
}
