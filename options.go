package bootstrap

import (
	"github.com/asticode/go-astilectron"
)

// Options represents options
type Options struct {
	AstilectronOptions astilectron.Options
	Debug              bool
	Homepage           string
	MenuOptions        []*astilectron.MenuItemOptions
	MessageHandler     MessageHandler
	OnWait             OnWait
	RestoreAssets      RestoreAssets
	TrayMenuOptions    []*astilectron.MenuItemOptions
	TrayOptions        *astilectron.TrayOptions
	WindowOptions      *astilectron.WindowOptions
}

// MessageHandler is a functions that handles messages
type MessageHandler func(w *astilectron.Window, m MessageIn)

// OnWait is a function that executes custom actions before waiting
type OnWait func(a *astilectron.Astilectron, w *astilectron.Window, m *astilectron.Menu, t *astilectron.Tray, tm *astilectron.Menu) error

// RestoreAssets is a function that restores assets namely the go-bindata's RestoreAssets method
type RestoreAssets func(dir, name string) error
