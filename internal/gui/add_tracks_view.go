//go:build !nogui

package gui

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/helpers"
)

// AddTracksView handles adding tracks to a playlist.
type AddTracksView struct {
	app         *App
	box         *gtk.Box
	idEntry     *gtk.Entry
	tracksTextView *gtk.TextView
	addBtn      *gtk.Button
	resultLabel *gtk.Label
}

// NewAddTracksView creates a new add tracks view.
func NewAddTracksView(app *App) (*AddTracksView, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(DefaultMargin)
	box.SetMarginBottom(DefaultMargin)
	box.SetMarginStart(DefaultMargin)
	box.SetMarginEnd(DefaultMargin)

	// Playlist ID entry.
	idLabel, err := gtk.LabelNew("Playlist ID, URI, or URL:")
	if err != nil {
		return nil, err
	}
	idLabel.SetHAlign(gtk.ALIGN_START)
	box.PackStart(idLabel, false, false, 0)

	idEntry, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	idEntry.SetPlaceholderText("Playlist ID, URI, or URL...")
	box.PackStart(idEntry, false, false, 0)

	// Track URIs text view.
	tracksLabel, err := gtk.LabelNew("Track URIs/URLs (one per line or comma-separated):")
	if err != nil {
		return nil, err
	}
	tracksLabel.SetHAlign(gtk.ALIGN_START)
	box.PackStart(tracksLabel, false, false, DefaultPadding)

	tracksTextView, err := gtk.TextViewNew()
	if err != nil {
		return nil, err
	}
	tracksTextView.SetWrapMode(gtk.WRAP_WORD)

	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.SetSizeRequest(-1, 200)
	scrolled.Add(tracksTextView)
	box.PackStart(scrolled, true, true, 0)

	// Add button.
	addBtn, err := gtk.ButtonNewWithLabel("Add Tracks")
	if err != nil {
		return nil, err
	}
	box.PackStart(addBtn, false, false, 0)

	// Result label.
	resultLabel, err := gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	resultLabel.SetHAlign(gtk.ALIGN_START)
	resultLabel.SetLineWrap(true)
	box.PackStart(resultLabel, false, false, DefaultPadding)

	v := &AddTracksView{
		app:         app,
		box:         box,
		idEntry:     idEntry,
		tracksTextView: tracksTextView,
		addBtn:      addBtn,
		resultLabel: resultLabel,
	}

	// Connect signals.
	addBtn.Connect("clicked", v.handleAdd)

	return v, nil
}

// Widget returns the root widget.
func (v *AddTracksView) Widget() *gtk.Box {
	return v.box
}

func (v *AddTracksView) handleAdd() {
	idText, err := v.idEntry.GetText()
	if err != nil {
		return
	}

	playlistID, err := helpers.NormalizePlaylistID(idText)
	if err != nil {
		v.resultLabel.SetMarkup(fmt.Sprintf(`<span foreground="red">Invalid playlist ID: %v</span>`, err))
		return
	}

	// Get tracks text.
	buffer, err := v.tracksTextView.GetBuffer()
	if err != nil {
		return
	}
	start, end := buffer.GetBounds()
	tracksText, err := buffer.GetText(start, end, false)
	if err != nil {
		return
	}

	// Parse track URIs.
	uris, warnings := parseTrackURIs(tracksText)
	if len(uris) == 0 {
		v.resultLabel.SetText("Please enter at least one track URI")
		return
	}

	RunAsync(
		func() {
			v.addBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Adding tracks...")
			v.resultLabel.SetText("")
		},
		func() (string, error) {
			client, err := v.app.Client()
			if err != nil {
				return "", err
			}
			return client.AddTracksToPlaylist(v.app.ctx, playlistID, uris)
		},
		func(snapshotID string, err error) {
			v.addBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				v.resultLabel.SetMarkup(fmt.Sprintf(`<span foreground="red">Error: %v</span>`, err))
				return
			}

			result := fmt.Sprintf(`<span foreground="green">Added %d tracks to playlist</span>`, len(uris))
			if len(warnings) > 0 {
				result += "\n\nWarnings:\n" + strings.Join(warnings, "\n")
			}
			v.resultLabel.SetMarkup(result)
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Added %d tracks", len(uris)))

			// Clear form.
			buffer, _ := v.tracksTextView.GetBuffer()
			buffer.SetText("")
		},
	)
}

func parseTrackURIs(text string) ([]string, []string) {
	var uris []string
	var warnings []string

	// Split by newlines and commas.
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		parts := strings.Split(line, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			uri, err := helpers.NormalizeTrackURI(part)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("Invalid track URI: %s (%v)", part, err))
				continue
			}
			uris = append(uris, uri)
		}
	}

	return uris, warnings
}
