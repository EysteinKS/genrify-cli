//go:build !nogui

package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/spotify"
)

// CreateView handles playlist creation.
type CreateView struct {
	app         *App
	box         *gtk.Box
	nameEntry   *gtk.Entry
	descTextView *gtk.TextView
	publicCheck *gtk.CheckButton
	createBtn   *gtk.Button
	resultLabel *gtk.Label
}

// NewCreateView creates a new create playlist view.
func NewCreateView(app *App) (*CreateView, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(DefaultMargin)
	box.SetMarginBottom(DefaultMargin)
	box.SetMarginStart(DefaultMargin)
	box.SetMarginEnd(DefaultMargin)

	// Name entry.
	nameLabel, err := gtk.LabelNew("Playlist Name:")
	if err != nil {
		return nil, err
	}
	nameLabel.SetHAlign(gtk.ALIGN_START)
	box.PackStart(nameLabel, false, false, 0)

	nameEntry, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	nameEntry.SetPlaceholderText("My Awesome Playlist")
	box.PackStart(nameEntry, false, false, 0)

	// Description text view.
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

	scrolled, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	scrolled.SetSizeRequest(-1, 100)
	scrolled.Add(descTextView)
	box.PackStart(scrolled, false, false, 0)

	// Public checkbox.
	publicCheck, err := gtk.CheckButtonNewWithLabel("Public playlist")
	if err != nil {
		return nil, err
	}
	box.PackStart(publicCheck, false, false, DefaultPadding)

	// Create button.
	createBtn, err := gtk.ButtonNewWithLabel("Create Playlist")
	if err != nil {
		return nil, err
	}
	box.PackStart(createBtn, false, false, 0)

	// Result label.
	resultLabel, err := gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	resultLabel.SetHAlign(gtk.ALIGN_START)
	resultLabel.SetLineWrap(true)
	box.PackStart(resultLabel, false, false, DefaultPadding)

	v := &CreateView{
		app:         app,
		box:         box,
		nameEntry:   nameEntry,
		descTextView: descTextView,
		publicCheck: publicCheck,
		createBtn:   createBtn,
		resultLabel: resultLabel,
	}

	// Connect signals.
	createBtn.Connect("clicked", v.handleCreate)

	return v, nil
}

// Widget returns the root widget.
func (v *CreateView) Widget() *gtk.Box {
	return v.box
}

func (v *CreateView) handleCreate() {
	name, err := v.nameEntry.GetText()
	if err != nil {
		return
	}
	if name == "" {
		v.resultLabel.SetText("Please enter a playlist name")
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

	RunAsync(
		func() {
			v.createBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Creating playlist...")
			v.resultLabel.SetText("")
		},
		func() (spotify.SimplifiedPlaylist, error) {
			client, err := v.app.Client()
			if err != nil {
				return spotify.SimplifiedPlaylist{}, err
			}

			// Get user ID.
			user, err := client.GetMe(v.app.ctx)
			if err != nil {
				return spotify.SimplifiedPlaylist{}, fmt.Errorf("get user: %w", err)
			}

			return client.CreatePlaylist(v.app.ctx, user.ID, name, description, public)
		},
		func(playlist spotify.SimplifiedPlaylist, err error) {
			v.createBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Error: %v", err))
				v.resultLabel.SetMarkup(fmt.Sprintf(`<span foreground="red">Error: %v</span>`, err))
				return
			}

			v.resultLabel.SetMarkup(fmt.Sprintf(`<span foreground="green">Created playlist: %s (ID: %s)</span>`, playlist.Name, playlist.ID))
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Created playlist: %s", playlist.Name))

			// Clear form.
			v.nameEntry.SetText("")
			buffer, _ := v.descTextView.GetBuffer()
			buffer.SetText("")
			v.publicCheck.SetActive(false)
		},
	)
}
