//go:build !nogui

package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

// Window represents the main application window.
type Window struct {
	window    *gtk.Window
	headerBar *gtk.HeaderBar
	paned     *gtk.Paned
	stack     *gtk.Stack
	sidebar   *Sidebar
	statusBar *StatusBar
}

// NewWindow creates the main application window with header, sidebar, and content area.
func NewWindow(app *App) (*Window, error) {
	// Create main window.
	window, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return nil, err
	}
	window.SetTitle("Genrify")
	window.SetDefaultSize(WindowDefaultWidth, WindowDefaultHeight)
	window.SetSizeRequest(WindowMinWidth, WindowMinHeight)

	// Create header bar.
	headerBar, err := gtk.HeaderBarNew()
	if err != nil {
		return nil, err
	}
	headerBar.SetTitle("Genrify")
	headerBar.SetShowCloseButton(true)

	// Settings button.
	settingsBtn, err := gtk.ButtonNewFromIconName("preferences-system", gtk.ICON_SIZE_BUTTON)
	if err != nil {
		return nil, err
	}
	settingsBtn.SetTooltipText("Settings")
	headerBar.PackEnd(settingsBtn)

	window.SetTitlebar(headerBar)

	// Create main layout.
	mainBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
	}

	// Create paned layout for sidebar and content.
	paned, err := gtk.PanedNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	paned.SetPosition(SidebarWidth)

	// Create stack for content views.
	stack, err := gtk.StackNew()
	if err != nil {
		return nil, err
	}

	// Create sidebar items.
	items := []SidebarItem{
		{Name: "login", Label: "Login"},
		{Name: "playlists", Label: "Playlists"},
		{Name: "tracks", Label: "Tracks"},
		{Name: "create", Label: "Create Playlist"},
		{Name: "add-tracks", Label: "Add Tracks"},
		{Name: "merge", Label: "Merge Playlists"},
	}

	// Create sidebar with selection handler.
	sidebar, err := NewSidebar(items, func(index int, item SidebarItem) {
		stack.SetVisibleChildName(item.Name)
	})
	if err != nil {
		return nil, err
	}

	// Create status bar.
	statusBar, err := NewStatusBar()
	if err != nil {
		return nil, err
	}

	// Add sidebar and stack to paned layout.
	paned.Add1(sidebar.Widget())
	paned.Add2(stack)

	// Assemble main layout.
	mainBox.PackStart(paned, true, true, 0)
	mainBox.PackEnd(statusBar.Widget(), false, false, 0)

	window.Add(mainBox)

	// Connect settings button.
	settingsBtn.Connect("clicked", func() {
		app.showConfigDialog()
	})

	// Connect window close signal.
	window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	w := &Window{
		window:    window,
		headerBar: headerBar,
		paned:     paned,
		stack:     stack,
		sidebar:   sidebar,
		statusBar: statusBar,
	}

	return w, nil
}

// Show displays the window.
func (w *Window) Show() {
	w.window.ShowAll()
}

// AddView adds a view to the content stack.
func (w *Window) AddView(name, title string, widget gtk.IWidget) {
	w.stack.AddTitled(widget, name, title)
}

// StatusBar returns the status bar.
func (w *Window) StatusBar() *StatusBar {
	return w.statusBar
}

// Stack returns the content stack.
func (w *Window) Stack() *gtk.Stack {
	return w.stack
}
