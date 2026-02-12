//go:build !nogui

package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

// StatusBar displays application status at the bottom of the window.
type StatusBar struct {
	box     *gtk.Box
	label   *gtk.Label
	spinner *gtk.Spinner
}

// NewStatusBar creates a new status bar widget.
func NewStatusBar() (*StatusBar, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(DefaultPadding)
	box.SetMarginBottom(DefaultPadding)
	box.SetMarginStart(DefaultMargin)
	box.SetMarginEnd(DefaultMargin)

	label, err := gtk.LabelNew("Ready")
	if err != nil {
		return nil, err
	}
	label.SetHAlign(gtk.ALIGN_START)
	label.SetEllipsize(3) // PANGO_ELLIPSIZE_END

	spinner, err := gtk.SpinnerNew()
	if err != nil {
		return nil, err
	}

	box.PackStart(label, true, true, 0)
	box.PackEnd(spinner, false, false, 0)

	return &StatusBar{
		box:     box,
		label:   label,
		spinner: spinner,
	}, nil
}

// Widget returns the underlying GTK widget.
func (s *StatusBar) Widget() *gtk.Box {
	return s.box
}

// SetStatus sets the status message.
func (s *StatusBar) SetStatus(text string) {
	s.label.SetText(text)
	s.label.SetMarkup(text) // Allow basic markup
}

// SetError sets an error message (displayed in red).
func (s *StatusBar) SetError(text string) {
	s.label.SetMarkup(`<span foreground="red">` + text + `</span>`)
}

// SetLoading shows or hides the loading spinner.
func (s *StatusBar) SetLoading(loading bool) {
	if loading {
		s.spinner.Start()
		s.spinner.Show()
	} else {
		s.spinner.Stop()
		s.spinner.Hide()
	}
}

// Clear resets the status bar to default state.
func (s *StatusBar) Clear() {
	s.label.SetText("Ready")
	s.SetLoading(false)
}
