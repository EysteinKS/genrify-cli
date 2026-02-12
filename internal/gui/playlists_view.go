//go:build !nogui

package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/helpers"
	"genrify/internal/spotify"
)

// PlaylistsView displays user's playlists.
type PlaylistsView struct {
	app       *App
	box       *gtk.Box
	searchEntry *gtk.SearchEntry
	limitSpin *gtk.SpinButton
	refreshBtn *gtk.Button
	treeView  *gtk.TreeView
	listStore *gtk.ListStore
}

// NewPlaylistsView creates a new playlists view.
func NewPlaylistsView(app *App) (*PlaylistsView, error) {
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

	// Search entry.
	searchEntry, err := gtk.SearchEntryNew()
	if err != nil {
		return nil, err
	}
	searchEntry.SetPlaceholderText("Filter by name...")
	controlsBox.PackStart(searchEntry, true, true, 0)

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

	// Refresh button.
	refreshBtn, err := gtk.ButtonNewWithLabel("Refresh")
	if err != nil {
		return nil, err
	}
	controlsBox.PackStart(refreshBtn, false, false, 0)

	box.PackStart(controlsBox, false, false, 0)

	// Create tree view with columns: ID, Name, Tracks, Owner.
	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT, glib.TYPE_STRING)
	if err != nil {
		return nil, err
	}

	treeView, err := gtk.TreeViewNewWithModel(listStore)
	if err != nil {
		return nil, err
	}

	// Add columns.
	if err := addColumn(treeView, "ID", 0); err != nil {
		return nil, err
	}
	if err := addColumn(treeView, "Name", 1); err != nil {
		return nil, err
	}
	if err := addColumnInt(treeView, "Tracks", 2); err != nil {
		return nil, err
	}
	if err := addColumn(treeView, "Owner", 3); err != nil {
		return nil, err
	}

	// Add tree view to scrolled window.
	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.Add(treeView)
	box.PackStart(scrolled, true, true, 0)

	v := &PlaylistsView{
		app:       app,
		box:       box,
		searchEntry: searchEntry,
		limitSpin: limitSpin,
		refreshBtn: refreshBtn,
		treeView:  treeView,
		listStore: listStore,
	}

	// Connect signals.
	refreshBtn.Connect("clicked", v.handleRefresh)
	searchEntry.Connect("search-changed", v.handleSearchChanged)
	treeView.Connect("row-activated", v.handleRowActivated)

	return v, nil
}

// Widget returns the root widget.
func (v *PlaylistsView) Widget() *gtk.Box {
	return v.box
}

func (v *PlaylistsView) handleRefresh() {
	limit := int(v.limitSpin.GetValue())

	RunAsync(
		func() {
			v.refreshBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Loading playlists...")
		},
		func() ([]spotify.SimplifiedPlaylist, error) {
			client, err := v.app.Client()
			if err != nil {
				return nil, err
			}
			return client.ListCurrentUserPlaylists(v.app.ctx, limit)
		},
		func(playlists []spotify.SimplifiedPlaylist, err error) {
			v.refreshBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				return
			}

			v.updateTreeView(playlists)
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Loaded %d playlists", len(playlists)))
		},
	)
}

func (v *PlaylistsView) handleSearchChanged() {
	// Filter is applied locally, so just trigger a re-filter.
	// For simplicity, we'll reload the data. In a real app, we'd cache and filter.
	v.handleRefresh()
}

func (v *PlaylistsView) handleRowActivated(tv *gtk.TreeView, path *gtk.TreePath) {
	// Get the selected playlist ID.
	model, err := tv.GetModel()
	if err != nil {
		return
	}
	iter, err := model.ToTreeModel().GetIter(path)
	if err != nil {
		return
	}

	val, err := model.ToTreeModel().GetValue(iter, 0)
	if err != nil {
		return
	}
	playlistID, err := val.GetString()
	if err != nil {
		return
	}

	// Navigate to tracks view with this playlist ID.
	v.app.window.Stack().SetVisibleChildName("tracks")
	if v.app.tracksView != nil {
		v.app.tracksView.SetPlaylistID(playlistID)
	}
}

func (v *PlaylistsView) updateTreeView(playlists []spotify.SimplifiedPlaylist) {
	// Apply filter.
	filterText, _ := v.searchEntry.GetText()
	filtered := helpers.FilterPlaylistsByName(playlists, filterText)

	// Clear existing data.
	v.listStore.Clear()

	// Add rows.
	for _, p := range filtered {
		iter := v.listStore.Append()
		v.listStore.SetValue(iter, 0, p.ID)
		v.listStore.SetValue(iter, 1, p.Name)
		v.listStore.SetValue(iter, 2, p.Tracks.Total)
		v.listStore.SetValue(iter, 3, p.Owner.ID)
	}
}

func addColumn(tv *gtk.TreeView, title string, id int) error {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return err
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		return err
	}
	column.SetResizable(true)
	column.SetSortColumnID(id)

	tv.AppendColumn(column)
	return nil
}

func addColumnInt(tv *gtk.TreeView, title string, id int) error {
	cellRenderer, err := gtk.CellRendererTextNew()
	if err != nil {
		return err
	}

	column, err := gtk.TreeViewColumnNewWithAttribute(title, cellRenderer, "text", id)
	if err != nil {
		return err
	}
	column.SetResizable(true)
	column.SetSortColumnID(id)

	tv.AppendColumn(column)
	return nil
}
