package bootstrap

import (
	"os"
	"path/filepath"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// Run runs the bootstrap
func Run(o Options) (err error) {
	// Create astilectron
	var a *astilectron.Astilectron
	if a, err = astilectron.New(o.AstilectronOptions); err != nil {
		return errors.Wrap(err, "creating new astilectron failed")
	}
	defer a.Close()
	a.HandleSignals()

	// TODO Set bindata provisioner

	// Restore resources
	if o.RestoreAssets != nil {
		var rp = filepath.Join(a.Paths().BaseDirectory(), "resources")
		if _, err = os.Stat(rp); os.IsNotExist(err) {
			astilog.Debugf("Restoring resources in %s", rp)
			if err = o.RestoreAssets(a.Paths().BaseDirectory(), "resources"); err != nil {
				err = errors.Wrapf(err, "restoring resources in %s failed", rp)
				return
			}
		} else if err != nil {
			err = errors.Wrapf(err, "stating %s failed", rp)
			return
		} else {
			astilog.Debugf("%s already exists, skipping restoring resources...", rp)
		}
	}

	// Start
	if err = a.Start(); err != nil {
		return errors.Wrap(err, "starting astilectron failed")
	}

	// Debug
	if o.Debug {
		o.WindowOptions.Width = astilectron.PtrInt(*o.WindowOptions.Width + 700)
	}

	// Init window
	var w *astilectron.Window
	if w, err = a.NewWindow(filepath.Join(a.Paths().BaseDirectory(), "resources", "app", o.Homepage), o.WindowOptions); err != nil {
		return errors.Wrap(err, "new window failed")
	}

	// Handle messages
	if o.MessageHandler != nil {
		w.On(astilectron.EventNameWindowEventMessage, handleMessages(w, o.MessageHandler))
	}

	// Create window
	if err = w.Create(); err != nil {
		return errors.Wrap(err, "creating window failed")
	}

	// Debug
	if o.Debug {
		if err = w.OpenDevTools(); err != nil {
			return errors.Wrap(err, "opening dev tools failed")
		}
	}

	// Menu
	var m *astilectron.Menu
	if len(o.MenuOptions) > 0 {
		// Init menu
		m = a.NewMenu(o.MenuOptions)

		// Create menu
		if err = m.Create(); err != nil {
			return errors.Wrap(err, "creating menu failed")
		}
	}

	// Tray
	var t *astilectron.Tray
	var tm *astilectron.Menu
	if o.TrayOptions != nil {
		// Init tray
		t = a.NewTray(o.TrayOptions)

		// Tray menu
		if len(o.TrayMenuOptions) > 0 {
			// Init tray menu
			tm = t.NewMenu(o.TrayMenuOptions)

			// Create tray menu
			if err = tm.Create(); err != nil {
				return errors.Wrap(err, "creating tray menu failed")
			}
		}

		// Create tray
		if err = t.Create(); err != nil {
			return errors.Wrap(err, "creating tray failed")
		}
	}

	// On wait
	if o.OnWait != nil {
		if err = o.OnWait(a, w, m, t, tm); err != nil {
			return errors.Wrap(err, "onwait failed")
		}
	}

	// Blocking pattern
	a.Wait()
	return
}
