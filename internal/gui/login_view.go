//go:build !nogui

package gui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
)

// LoginView handles Spotify authentication.
type LoginView struct {
	app       *App
	box       *gtk.Box
	statusLbl *gtk.Label
	loginBtn  *gtk.Button
	logoutBtn *gtk.Button
}

// NewLoginView creates a new login view.
func NewLoginView(app *App) (*LoginView, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, DefaultSpacing)
	if err != nil {
		return nil, err
	}
	box.SetMarginTop(DefaultMargin)
	box.SetMarginBottom(DefaultMargin)
	box.SetMarginStart(DefaultMargin)
	box.SetMarginEnd(DefaultMargin)

	// Status label.
	statusLbl, err := gtk.LabelNew("Not logged in")
	if err != nil {
		return nil, err
	}
	statusLbl.SetHAlign(gtk.ALIGN_START)
	box.PackStart(statusLbl, false, false, DefaultPadding)

	// Login button.
	loginBtn, err := gtk.ButtonNewWithLabel("Login with Spotify")
	if err != nil {
		return nil, err
	}
	box.PackStart(loginBtn, false, false, 0)

	// Logout button.
	logoutBtn, err := gtk.ButtonNewWithLabel("Logout")
	if err != nil {
		return nil, err
	}
	box.PackStart(logoutBtn, false, false, 0)

	v := &LoginView{
		app:       app,
		box:       box,
		statusLbl: statusLbl,
		loginBtn:  loginBtn,
		logoutBtn: logoutBtn,
	}

	// Connect signals.
	loginBtn.Connect("clicked", v.handleLogin)
	logoutBtn.Connect("clicked", v.handleLogout)

	// Update initial state.
	v.updateState()

	return v, nil
}

// Widget returns the root widget.
func (v *LoginView) Widget() *gtk.Box {
	return v.box
}

func (v *LoginView) handleLogin() {
	RunAsync(
		func() {
			v.loginBtn.SetSensitive(false)
			v.app.window.StatusBar().SetLoading(true)
			v.app.window.StatusBar().SetStatus("Logging in...")
		},
		func() (string, error) {
			return v.app.opts.DoLogin(v.app.ctx, v.app.cfg)
		},
		func(username string, err error) {
			v.loginBtn.SetSensitive(true)
			v.app.window.StatusBar().SetLoading(false)

			if err != nil {
				v.app.window.StatusBar().SetError(fmt.Sprintf("Login failed: %v", err))
				return
			}

			v.app.RefreshClient()
			v.statusLbl.SetText(fmt.Sprintf("Logged in as: %s", username))
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Logged in as %s", username))
			v.updateState()
		},
	)
}

func (v *LoginView) handleLogout() {
	// Clear the cached client.
	v.app.RefreshClient()
	v.statusLbl.SetText("Not logged in")
	v.app.window.StatusBar().SetStatus("Logged out")
	v.updateState()
}

func (v *LoginView) updateState() {
	// Check if we have a client (logged in).
	client, err := v.app.Client()
	loggedIn := err == nil && client != nil

	if loggedIn {
		// Try to get user info to confirm login.
		user, err := client.GetMe(v.app.ctx)
		if err == nil && user.ID != "" {
			v.statusLbl.SetText(fmt.Sprintf("Logged in as: %s", user.DisplayName))
			v.app.window.StatusBar().SetStatus(fmt.Sprintf("Logged in as %s", user.DisplayName))
		}
	}

	v.loginBtn.SetVisible(!loggedIn)
	v.logoutBtn.SetVisible(loggedIn)
}
