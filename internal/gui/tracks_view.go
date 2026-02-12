//go:build !nogui

package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/helpers"
	"genrify/internal/spotify"
)

// TracksView displays tracks from a playlist.
type TracksView struct {
	app        *App
	box        *gtk.Box
	idEntry    *gtk.Entry
	limitSpin  *gtk.SpinButton
	loadBtn    *gtk.Button
	treeView   *gtk.TreeView
	listStore  *gtk.ListStore
}

// NewTracksView creates a new tracks view.
func NewTracksView(app *App) (*TracksView, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(DefaultMargin)
	box.SetMarginBottom(DefaultMargin)
	box.SetMarginStart(DefaultMargin)
	box.SetMarginEnd(DefaultMargin)

	// Controls box.
	controlsBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}

	// ID entry.
	idEntry, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	idEntry.SetPlaceholderText("Playlist ID, URI, or URL...")
	controlsBox.PackStart(idEntry, true, true, 0)

	// Limit spin button.
	limitAdj, err := gtk.AdjustmentNew(50, 1, 500, 10, 50, 0)
	if err != nil {
		return nil, err
	}
	limitSpin, err := gtk.SpinButtonNew(limitAdj, 1, 0)
	if err != nil {
		return nil, err
	}
	limitSpin.SetValue(50)
	controlsBox.PackStart(limitSpin, false, false, 0)

	// Load button.
	loadBtn, err := gtk.ButtonNewWithLabel("Load")
	if err != nil {
		return nil, err
	}
	controlsBox.PackStart(loadBtn, false, false, 0)

	box.PackStart(controlsBox, false, false, 0)

	// Create tree view with columns: URI, Name, Artists.
	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		return nil, err
	}

	treeView, err := gtk.TreeViewNewWithModel(listStore)
	if err != nil {
		return nil, err
	}

	// Add columns.
	if err := addColumn(treeView, "URI", 0); err != nil {
		return nil, err
	}
	if err := addColumn(treeView, "Name", 1); err != nil {
		return nil, err
	}
	if err := addColumn(treeView, "Artists", 2); err != nil {
		return nil, err
	}

	// Add tree view to scrolled window.
	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.Add(treeView)
	box.PackStart(scrolled, true, true, 0)

	v := &TracksView{
		app:       app,
		box:       box,
		idEntry:   idEntry,
		limitSpin: limitSpin,
		loadBtn:   loadBtn,
		treeView:  treeView,
		listStore: listStore,
	}

	// Connect signals.
	loadBtn.Connect("clicked", v.handleLoad)
	idEntry.Connect("activate", v.handleLoad)

	return v, nil
}

// Widget returns the root widget.
func (v *TracksView) Widget() *gtk.Box {
	return v.box
}

// SetPlaylistID sets the playlist ID in the entry field and loads the tracks.
func (v *TracksView) SetPlaylistID(id string) {
	v.idEntry.SetText(id)
	v.handleLoad()
}

func (v *TracksView) handleLoad() {
	idText, err := v.idEntry.GetText()
	if err != nil {
		return
	}

	playlistID, err := helpers.NormalizePlaylistID(idText)
	if err != nil {
		v.app.window.StatusBar().SetError(fmt.Sprintf("Invalid playlist ID: %v", err))
		return
	}

	limit := int(v.limitSpin.GetValue())

	RunAsync(
		func() {
			v.loadBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Loading tracks...")
		},
		func() ([]spotify.FullTrack, error) {
			client, err := v.app.Client()
			if err != nil {
				return nil, err
			}
			return client.ListPlaylistTracks(v.app.ctx, playlistID, limit)
		},
		func(tracks []spotify.FullTrack, err error) {
			v.loadBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				return
			}

			v.updateTreeView(tracks)
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Loaded %d tracks", len(tracks)))
		},
	)
}

func (v *TracksView) updateTreeView(tracks []spotify.FullTrack) {
	// Clear existing data.
	v.listStore.Clear()

	// Add rows.
	for _, t := range tracks {
		iter := v.listStore.Append()
		v.listStore.SetValue(iter, 0, t.URI)
		v.listStore.SetValue(iter, 1, t.Name)
		v.listStore.SetValue(iter, 2, helpers.JoinArtistNames(t.Artists))
	}
}
