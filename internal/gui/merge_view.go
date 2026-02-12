//go:build !nogui

package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/playlist"
	"genrify/internal/spotify"
)

// MergeView handles merging multiple playlists into one.
type MergeView struct {
	app *App
	box *gtk.Box

	// Step 1: Find playlists.
	patternEntry *gtk.Entry
	findBtn      *gtk.Button
	treeView     *gtk.TreeView
	listStore    *gtk.ListStore

	// Step 2: Merge configuration.
	nameEntry   *gtk.Entry
	descTextView *gtk.TextView
	publicCheck *gtk.CheckButton
	dedupeCheck *gtk.CheckButton
	mergeBtn    *gtk.Button
	progress    *gtk.ProgressBar

	// Step 3: Results.
	resultLabel *gtk.Label
	deleteBtn   *gtk.Button

	// State.
	matchedPlaylists []spotify.SimplifiedPlaylist
	mergeResult      *playlist.MergeResult
}

// NewMergeView creates a new merge playlists view.
func NewMergeView(app *App) (*MergeView, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(DefaultMargin)
	box.SetMarginBottom(DefaultMargin)
	box.SetMarginStart(DefaultMargin)
	box.SetMarginEnd(DefaultMargin)

	// Step 1: Find matches.
	step1Label, err := gtk.LabelNew("Step 1: Find Playlists")
	if err != nil {
		return nil, err
	}
	step1Label.SetHAlign(gtk.ALIGN_START)
	step1Label.SetMarkup("<b>Step 1: Find Playlists</b>")
	box.PackStart(step1Label, false, false, DefaultPadding)

	patternBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}

	patternEntry, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	patternEntry.SetPlaceholderText("Regex pattern (e.g., ^My Playlist)")
	patternBox.PackStart(patternEntry, true, true, 0)

	findBtn, err := gtk.ButtonNewWithLabel("Find Matches")
	if err != nil {
		return nil, err
	}
	patternBox.PackStart(findBtn, false, false, 0)

	box.PackStart(patternBox, false, false, 0)

	// Tree view for matched playlists.
	listStore, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING, glib.TYPE_INT)
	if err != nil {
		return nil, err
	}

	treeView, err := gtk.TreeViewNewWithModel(listStore)
	if err != nil {
		return nil, err
	}

	// Add columns: ID, Name, Tracks.
	if err := addColumn(treeView, "ID", 0); err != nil {
		return nil, err
	}
	if err := addColumn(treeView, "Name", 1); err != nil {
		return nil, err
	}
	if err := addColumnInt(treeView, "Tracks", 2); err != nil {
		return nil, err
	}

	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.SetSizeRequest(-1, 150)
	scrolled.Add(treeView)
	box.PackStart(scrolled, false, false, 0)

	// Step 2: Merge configuration.
	step2Label, err := gtk.LabelNew("Step 2: Merge Configuration")
	if err != nil {
		return nil, err
	}
	step2Label.SetHAlign(gtk.ALIGN_START)
	step2Label.SetMarkup("<b>Step 2: Merge Configuration</b>")
	box.PackStart(step2Label, false, false, DefaultPadding)

	// Target name.
	nameLabel, err := gtk.LabelNew("Target Playlist Name:")
	if err != nil {
		return nil, err
	}
	nameLabel.SetHAlign(gtk.ALIGN_START)
	box.PackStart(nameLabel, false, false, 0)

	nameEntry, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	nameEntry.SetPlaceholderText("Merged Playlist")
	box.PackStart(nameEntry, false, false, 0)

	// Description.
	descLabel, err := gtk.LabelNew("Description (optional):")
	if err != nil {
		return nil, err
	}
	descLabel.SetHAlign(gtk.ALIGN_START)
	box.PackStart(descLabel, false, false, DefaultPadding)

	descTextView, err := gtk.TextViewNew()
	if err != nil {
		return nil, err
	}
	descTextView.SetWrapMode(gtk.WRAP_WORD)

	descScrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	descScrolled.SetSizeRequest(-1, 60)
	descScrolled.Add(descTextView)
	box.PackStart(descScrolled, false, false, 0)

	// Options.
	publicCheck, err := gtk.CheckButtonNewWithLabel("Public playlist")
	if err != nil {
		return nil, err
	}
	box.PackStart(publicCheck, false, false, 0)

	dedupeCheck, err := gtk.CheckButtonNewWithLabel("Remove duplicate tracks")
	if err != nil {
		return nil, err
	}
	dedupeCheck.SetActive(true)
	box.PackStart(dedupeCheck, false, false, 0)

	// Merge button and progress.
	mergeBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}

	mergeBtn, err := gtk.ButtonNewWithLabel("Merge Playlists")
	if err != nil {
		return nil, err
	}
	mergeBtn.SetSensitive(false) // Disabled until playlists are found.
	mergeBox.PackStart(mergeBtn, false, false, 0)

	progress, err := gtk.ProgressBarNew()
	if err != nil {
		return nil, err
	}
	progress.SetShowText(true)
	progress.SetNoShowAll(true)
	mergeBox.PackStart(progress, true, true, 0)

	box.PackStart(mergeBox, false, false, DefaultPadding)

	// Step 3: Results.
	step3Label, err := gtk.LabelNew("Step 3: Results")
	if err != nil {
		return nil, err
	}
	step3Label.SetHAlign(gtk.ALIGN_START)
	step3Label.SetMarkup("<b>Step 3: Results</b>")
	box.PackStart(step3Label, false, false, DefaultPadding)

	resultLabel, err := gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	resultLabel.SetHAlign(gtk.ALIGN_START)
	resultLabel.SetLineWrap(true)
	resultLabel.SetSelectable(true)
	box.PackStart(resultLabel, false, false, 0)

	// Delete source playlists button.
	deleteBtn, err := gtk.ButtonNewWithLabel("Delete Source Playlists")
	if err != nil {
		return nil, err
	}
	deleteBtn.SetSensitive(false)
	box.PackStart(deleteBtn, false, false, DefaultPadding)

	v := &MergeView{
		app:          app,
		box:          box,
		patternEntry: patternEntry,
		findBtn:      findBtn,
		treeView:     treeView,
		listStore:    listStore,
		nameEntry:    nameEntry,
		descTextView: descTextView,
		publicCheck:  publicCheck,
		dedupeCheck:  dedupeCheck,
		mergeBtn:     mergeBtn,
		progress:     progress,
		resultLabel:  resultLabel,
		deleteBtn:    deleteBtn,
	}

	// Connect signals.
	findBtn.Connect("clicked", v.handleFindMatches)
	mergeBtn.Connect("clicked", v.handleMerge)
	deleteBtn.Connect("clicked", v.handleDelete)

	return v, nil
}

// Widget returns the root widget.
func (v *MergeView) Widget() *gtk.Box {
	return v.box
}

func (v *MergeView) handleFindMatches() {
	pattern, err := v.patternEntry.GetText()
	if err != nil {
		return
	}

	RunAsync(
		func() {
			v.findBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Finding playlists...")
		},
		func() ([]spotify.SimplifiedPlaylist, error) {
			client, err := v.app.Client()
			if err != nil {
				return nil, err
			}
			svc := playlist.NewService(client)
			return svc.FindPlaylistsByPattern(v.app.ctx, pattern)
		},
		func(playlists []spotify.SimplifiedPlaylist, err error) {
			v.findBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				return
			}

			v.matchedPlaylists = playlists
			v.updateTreeView(playlists)
			v.mergeBtn.SetSensitive(len(playlists) > 0)
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Found %d playlists", len(playlists)))
		},
	)
}

func (v *MergeView) handleMerge() {
	if len(v.matchedPlaylists) == 0 {
		v.resultLabel.SetText("No playlists to merge")
		return
	}

	name, err := v.nameEntry.GetText()
	if err != nil {
		return
	}
	if name == "" {
		v.resultLabel.SetText("Please enter a target playlist name")
		return
	}

	// Get description.
	buffer, err := v.descTextView.GetBuffer()
	if err != nil {
		return
	}
	start, end := buffer.GetBounds()
	description, err := buffer.GetText(start, end, false)
	if err != nil {
		return
	}

	public := v.publicCheck.GetActive()
	dedupe := v.dedupeCheck.GetActive()

	sourceIDs := make([]string, len(v.matchedPlaylists))
	for i, p := range v.matchedPlaylists {
		sourceIDs[i] = p.ID
	}

	RunAsync(
		func() {
			v.mergeBtn.SetSensitive(false)
			v.progress.Show()
			v.progress.SetFraction(0.0)
			v.progress.SetText("Merging...")
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Merging playlists...")
			v.resultLabel.SetText("")
		},
		func() (*playlist.MergeResult, error) {
			client, err := v.app.Client()
			if err != nil {
				return nil, err
			}
			svc := playlist.NewService(client)
			return svc.MergePlaylists(v.app.ctx, sourceIDs, name, playlist.MergeOptions{
				Deduplicate: dedupe,
				Public:      public,
				Description: description,
			})
		},
		func(result *playlist.MergeResult, err error) {
			v.mergeBtn.SetSensitive(true)
			v.progress.Hide()
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				v.resultLabel.SetMarkup(fmt.Sprintf(`<span foreground="red">Error: %v</span>`, err))
				return
			}

			v.mergeResult = result
			resultText := fmt.Sprintf(`<span foreground="green">✓ Merge complete!</span>

New playlist ID: %s
Total tracks: %d
Duplicates removed: %d
Verified: %v`, result.NewPlaylistID, result.TrackCount, result.DuplicatesRemoved, result.Verified)

			if len(result.MissingURIs) > 0 {
				resultText += fmt.Sprintf("\n\nWarning: %d tracks missing in verification", len(result.MissingURIs))
			}

			v.resultLabel.SetMarkup(resultText)
			v.app.window.StatusBar().SetStatus("Merge complete")

			// Enable delete button only if verified.
			v.deleteBtn.SetSensitive(result.Verified)
		},
	)
}

func (v *MergeView) handleDelete() {
	if len(v.matchedPlaylists) == 0 || v.mergeResult == nil {
		return
	}

	// Confirm deletion.
	dialog := gtk.MessageDialogNew(
		nil,
		gtk.DIALOG_MODAL,
		gtk.MESSAGE_QUESTION,
		gtk.BUTTONS_YES_NO,
		fmt.Sprintf("Delete %d source playlists? This cannot be undone.", len(v.matchedPlaylists)),
	)
	dialog.SetTitle("Confirm Deletion")
	response := dialog.Run()
	dialog.Destroy()

	if response != gtk.RESPONSE_YES {
		return
	}

	sourceIDs := make([]string, len(v.matchedPlaylists))
	for i, p := range v.matchedPlaylists {
		sourceIDs[i] = p.ID
	}

	RunAsync(
		func() {
			v.deleteBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Deleting playlists...")
		},
		func() (struct{}, error) {
			client, err := v.app.Client()
			if err != nil {
				return struct{}{}, err
			}
			svc := playlist.NewService(client)
			return struct{}{}, svc.DeletePlaylists(v.app.ctx, sourceIDs)
		},
		func(_ struct{}, err error) {
			v.deleteBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				return
			}

			v.app.window.StatusBar().SetStatus("Source playlists deleted")
			v.resultLabel.SetMarkup(v.resultLabel.GetLabel() + "\n\n<span foreground=\"green\">✓ Source playlists deleted</span>")

			// Clear state.
			v.matchedPlaylists = nil
			v.listStore.Clear()
			v.mergeBtn.SetSensitive(false)
			v.deleteBtn.SetSensitive(false)
		},
	)
}

func (v *MergeView) updateTreeView(playlists []spotify.SimplifiedPlaylist) {
	// Clear existing data.
	v.listStore.Clear()

	// Add rows.
	for _, p := range playlists {
		iter := v.listStore.Append()
		v.listStore.SetValue(iter, 0, p.ID)
		v.listStore.SetValue(iter, 1, p.Name)
		v.listStore.SetValue(iter, 2, p.Tracks.Total)
	}
}
