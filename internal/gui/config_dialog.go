//go:build !nogui

package gui

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"genrify/internal/certs"
	"genrify/internal/config"
)

// ConfigDialog presents a modal dialog for configuring Spotify settings.
type ConfigDialog struct {
	dialog *gtk.Dialog

	clientIDEntry *gtk.Entry
	redirectEntry *gtk.Entry
	scopesEntry   *gtk.Entry
	certEntry     *gtk.Entry
	keyEntry      *gtk.Entry
	autoGenCheck  *gtk.CheckButton
	certKeyBox    *gtk.Box

	onSave func(config.Config) error
}

// NewConfigDialog creates a configuration dialog.
func NewConfigDialog(cfg config.Config, onSave func(config.Config) error) (*ConfigDialog, error) {
	dialog, err := gtk.DialogNew()
	if err != nil {
		return nil, err
	}
	dialog.SetTitle("Genrify Configuration")
	dialog.SetDefaultSize(500, 400)
	dialog.SetModal(true)

	// Add buttons.
	if _, err := dialog.AddButton("Cancel", gtk.RESPONSE_CANCEL); err != nil {
		return nil, err
	}
	if _, err := dialog.AddButton("Save", gtk.RESPONSE_OK); err != nil {
		return nil, err
	}

	// Create content area.
	contentArea, err := dialog.GetContentArea()
	if err != nil {
		return nil, err
	}
	contentArea.SetSpacing(DefaultSpacing)
	contentArea.SetMarginTop(DefaultMargin)
	contentArea.SetMarginBottom(DefaultMargin)
	contentArea.SetMarginStart(DefaultMargin)
	contentArea.SetMarginEnd(DefaultMargin)

	// Client ID field.
	clientIDEntry, err := createLabeledEntry(contentArea, "Spotify Client ID:", cfg.SpotifyClientID)
	if err != nil {
		return nil, err
	}

	// Redirect URI field.
	redirectEntry, err := createLabeledEntry(contentArea, "Redirect URI:", cfg.SpotifyRedirect)
	if err != nil {
		return nil, err
	}

	// Scopes field.
	scopesEntry, err := createLabeledEntry(contentArea, "Scopes (space/comma separated):", strings.Join(cfg.SpotifyScopes, " "))
	if err != nil {
		return nil, err
	}

	// TLS section (visible only for HTTPS redirects).
	certKeyBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}

	// Auto-generate checkbox.
	autoGenCheck, err := gtk.CheckButtonNewWithLabel("Auto-generate self-signed certificates")
	if err != nil {
		return nil, err
	}
	certKeyBox.PackStart(autoGenCheck, false, false, 0)

	// Cert file field.
	certEntry, err := createLabeledEntry(certKeyBox, "TLS Certificate File:", cfg.SpotifyTLSCert)
	if err != nil {
		return nil, err
	}

	// Key file field.
	keyEntry, err := createLabeledEntry(certKeyBox, "TLS Key File:", cfg.SpotifyTLSKey)
	if err != nil {
		return nil, err
	}

	contentArea.PackStart(certKeyBox, false, false, 0)

	d := &ConfigDialog{
		dialog:        dialog,
		clientIDEntry: clientIDEntry,
		redirectEntry: redirectEntry,
		scopesEntry:   scopesEntry,
		certEntry:     certEntry,
		keyEntry:      keyEntry,
		autoGenCheck:  autoGenCheck,
		certKeyBox:    certKeyBox,
		onSave:        onSave,
	}

	// Update TLS field visibility based on redirect URI.
	d.updateTLSVisibility()
	redirectEntry.Connect("changed", func() {
		d.updateTLSVisibility()
	})

	// Toggle cert/key fields when auto-generate is checked.
	autoGenCheck.Connect("toggled", func() {
		active := autoGenCheck.GetActive()
		certEntry.SetSensitive(!active)
		keyEntry.SetSensitive(!active)
	})

	dialog.ShowAll()

	return d, nil
}

// Run runs the dialog and returns the response.
func (d *ConfigDialog) Run() gtk.ResponseType {
	for {
		response := d.dialog.Run()
		if response != gtk.RESPONSE_OK {
			return response
		}
		if err := d.save(); err != nil {
			showErrorDialog(d.dialog, "Error", err.Error())
			d.dialog.ShowAll()
			continue
		}
		return response
	}
}

// Destroy destroys the dialog.
func (d *ConfigDialog) Destroy() {
	d.dialog.Destroy()
}

func (d *ConfigDialog) updateTLSVisibility() {
	redirectText, err := d.redirectEntry.GetText()
	if err != nil {
		return
	}
	isHTTPS := strings.HasPrefix(strings.ToLower(strings.TrimSpace(redirectText)), "https://")
	if isHTTPS {
		d.certKeyBox.Show()
	} else {
		d.certKeyBox.Hide()
	}
}

func (d *ConfigDialog) save() error {
	clientID, err := d.clientIDEntry.GetText()
	if err != nil {
		return err
	}
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return fmt.Errorf("client ID is required")
	}

	redirect, err := d.redirectEntry.GetText()
	if err != nil {
		return err
	}
	redirect = strings.TrimSpace(redirect)
	if redirect == "" {
		redirect = config.Default().SpotifyRedirect
	}

	scopesText, err := d.scopesEntry.GetText()
	if err != nil {
		return err
	}
	scopes := parseScopes(scopesText)
	if len(scopes) == 0 {
		scopes = config.Default().SpotifyScopes
	}

	cfg := config.Config{
		SpotifyClientID: clientID,
		SpotifyRedirect: redirect,
		SpotifyScopes:   scopes,
	}

	// Handle TLS certs if HTTPS.
	if strings.HasPrefix(strings.ToLower(redirect), "https://") {
		if d.autoGenCheck.GetActive() {
			// Auto-generate certificates.
			certPath, keyPath, err := certs.EnsureCerts("", "")
			if err != nil {
				return err
			}
			cfg.SpotifyTLSCert = certPath
			cfg.SpotifyTLSKey = keyPath
		} else {
			certPath, err := d.certEntry.GetText()
			if err != nil {
				return err
			}
			keyPath, err := d.keyEntry.GetText()
			if err != nil {
				return err
			}
			cfg.SpotifyTLSCert = strings.TrimSpace(certPath)
			cfg.SpotifyTLSKey = strings.TrimSpace(keyPath)

			if cfg.SpotifyTLSCert == "" || cfg.SpotifyTLSKey == "" {
				return fmt.Errorf("https redirect requires TLS certificate and key files")
			}
		}
	}

	return d.onSave(cfg)
}

func createLabeledEntry(container *gtk.Box, labelText, defaultValue string) (*gtk.Entry, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultPadding)
	if err != nil {
		return nil, err
	}

	label, err := gtk.LabelNew(labelText)
	if err != nil {
		return nil, err
	}
	label.SetHAlign(gtk.ALIGN_START)

	entry, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	if defaultValue != "" {
		entry.SetText(defaultValue)
	}

	box.PackStart(label, false, false, 0)
	box.PackStart(entry, false, false, 0)
	container.PackStart(box, false, false, 0)

	return entry, nil
}

func parseScopes(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ','
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func showErrorDialog(parent *gtk.Dialog, title, message string) {
	dialog := gtk.MessageDialogNew(parent, gtk.DIALOG_MODAL, gtk.MESSAGE_ERROR, gtk.BUTTONS_OK, "%s", message)
	dialog.SetTitle(title)
	dialog.Run()
	dialog.Destroy()
}
