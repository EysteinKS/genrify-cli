//go:build !nogui

package gui

import (
	"github.com/gotk3/gotk3/gtk"
)

// SidebarItem represents a navigation item in the sidebar.
type SidebarItem struct {
	Name  string
	Label string
}

// Sidebar provides navigation between different views.
type Sidebar struct {
	listBox *gtk.ListBox
	items   []SidebarItem
}

// NewSidebar creates a new sidebar with the given items.
func NewSidebar(items []SidebarItem, onSelect func(index int, item SidebarItem)) (*Sidebar, error) {
	listBox, err := gtk.ListBoxNew()
	if err != nil {
		return nil, err
	}
	listBox.SetSelectionMode(gtk.SELECTION_SINGLE)

	// Add items to the list box.
	for _, item := range items {
		label, err := gtk.LabelNew(item.Label)
		if err != nil {
			return nil, err
		}
		label.SetHAlign(gtk.ALIGN_START)
		label.SetMarginTop(DefaultPadding)
		label.SetMarginBottom(DefaultPadding)
		label.SetMarginStart(DefaultMargin)
		label.SetMarginEnd(DefaultMargin)

		listBox.Add(label)
	}

	// Connect row-selected signal.
	if onSelect != nil {
		listBox.Connect("row-selected", func(lb *gtk.ListBox, row *gtk.ListBoxRow) {
			if row == nil {
				return
			}
			index := row.GetIndex()
			if index >= 0 && index < len(items) {
				onSelect(index, items[index])
			}
		})
	}

	sidebar := &Sidebar{
		listBox: listBox,
		items:   items,
	}

	return sidebar, nil
}

// Widget returns the underlying GTK widget.
func (s *Sidebar) Widget() *gtk.ListBox {
	return s.listBox
}

// SelectItem selects the item at the given index.
func (s *Sidebar) SelectItem(index int) {
	if index >= 0 && index < len(s.items) {
		row := s.listBox.GetRowAtIndex(index)
		if row != nil {
			s.listBox.SelectRow(row)
		}
	}
}
