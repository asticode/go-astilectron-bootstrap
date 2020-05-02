package bootstrap

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astilectron"
	astibundler "github.com/asticode/go-astilectron-bundler"
)

// Run runs the bootstrap
func Run(o Options) (err error) {
	// Create logger
	l := astikit.AdaptStdLogger(o.Logger)

	// Create astilectron
	var a *astilectron.Astilectron
	if a, err = astilectron.New(o.Logger, o.AstilectronOptions); err != nil {
		return fmt.Errorf("creating new astilectron failed: %w", err)
	}
	defer a.Close()

	// Handle signals
	a.HandleSignals(astikit.LoggerSignalHandler(l, o.IgnoredSignals...))

	// Adapt astilectron
	if o.Adapter != nil {
		o.Adapter(a)
	}

	// Set provisioner
	if o.Asset != nil {
		a.SetProvisioner(astibundler.NewProvisioner(o.Asset, o.Logger))
	}

	// Get relative and absolute resources path
	var relativeResourcesPath = o.ResourcesPath
	if len(relativeResourcesPath) == 0 {
		relativeResourcesPath = "resources"
	}

	// Restore resources
	if o.RestoreAssets != nil {
		if err = restoreResources(l, a, o.Asset, o.AssetDir, o.RestoreAssets, relativeResourcesPath); err != nil {
			err = fmt.Errorf("restoring resources failed: %w", err)
			return
		}
	}

	// Start
	if err = a.Start(); err != nil {
		return fmt.Errorf("starting astilectron failed: %w", err)
	}

	// Init windows
	var w = make([]*astilectron.Window, len(o.Windows))
	for i, wo := range o.Windows {
		var url = wo.Homepage
		if !strings.Contains(url, "://") && !strings.HasPrefix(url, string(filepath.Separator)) {
			url = filepath.Join(absoluteResourcesPath(a, relativeResourcesPath), "app", url)
		}
		if w[i], err = a.NewWindow(url, wo.Options); err != nil {
			return fmt.Errorf("new window failed: %w", err)
		}

		// Handle messages
		if wo.MessageHandler != nil {
			w[i].OnMessage(handleMessages(w[i], wo.MessageHandler, l))
		}

		// Adapt window
		if wo.Adapter != nil {
			wo.Adapter(w[i])
		}

		// Create window
		if err = w[i].Create(); err != nil {
			return fmt.Errorf("creating window failed: %w", err)
		}
	}

	// Create menu options
	mo := o.MenuOptions
	if o.MenuOptionsFunc != nil {
		mo = o.MenuOptionsFunc(a)
	}

	// Debug
	if o.Debug {
		// Create menu item
		var debug bool
		mi := &astilectron.MenuItemOptions{
			Accelerator: astilectron.NewAccelerator("Control", "d"),
			Label:       astikit.StrPtr("Debug"),
			OnClick: func(e astilectron.Event) (deleteListener bool) {
				for i, window := range w {
					width := *o.Windows[i].Options.Width
					if debug {
						if err := window.CloseDevTools(); err != nil {
							l.Error(fmt.Errorf("closing dev tools failed: %w", err))
						}
						if err := window.Resize(width, *o.Windows[i].Options.Height); err != nil {
							l.Error(fmt.Errorf("resizing window failed: %w", err))
						}
					} else {
						if err := window.OpenDevTools(); err != nil {
							l.Error(fmt.Errorf("opening dev tools failed: %w", err))
						}
						if err := window.Resize(width+700, *o.Windows[i].Options.Height); err != nil {
							l.Error(fmt.Errorf("resizing window failed: %w", err))
						}
					}
				}
				debug = !debug
				return
			},
			Type: astilectron.MenuItemTypeCheckbox,
		}

		// Add menu item
		if len(mo) == 0 {
			mo = []*astilectron.MenuItemOptions{{SubMenu: []*astilectron.MenuItemOptions{mi}}}
		} else {
			if len(mo[0].SubMenu) > 0 {
				mo[0].SubMenu = append(mo[0].SubMenu, &astilectron.MenuItemOptions{Type: astilectron.MenuItemTypeSeparator})
			}
			mo[0].SubMenu = append(mo[0].SubMenu, mi)
		}
	}

	// Menu
	var m *astilectron.Menu
	if len(mo) > 0 {
		// Init menu
		m = a.NewMenu(mo)

		// Create menu
		if err = m.Create(); err != nil {
			return fmt.Errorf("creating menu failed: %w", err)
		}
	}

	// Tray
	var t *astilectron.Tray
	var tm *astilectron.Menu
	if o.TrayOptions != nil {
		// Make sure path to image is absolute
		if o.TrayOptions.Image != nil && !filepath.IsAbs(*o.TrayOptions.Image) {
			*o.TrayOptions.Image = filepath.Join(a.Paths().DataDirectory(), *o.TrayOptions.Image)
		}

		// Init tray
		t = a.NewTray(o.TrayOptions)

		// Create tray
		if err = t.Create(); err != nil {
			return fmt.Errorf("creating tray failed: %w", err)
		}

		// Tray menu
		if len(o.TrayMenuOptions) > 0 {
			// Init tray menu
			tm = t.NewMenu(o.TrayMenuOptions)

			// Create tray menu
			if err = tm.Create(); err != nil {
				return fmt.Errorf("creating tray menu failed: %w", err)
			}
		}
	}

	// On wait
	if o.OnWait != nil {
		if err = o.OnWait(a, w, m, t, tm); err != nil {
			return fmt.Errorf("onwait failed: %w", err)
		}
	}

	// Blocking pattern
	a.Wait()
	return
}

func absoluteResourcesPath(a *astilectron.Astilectron, relativeResourcesPath string) string {
	return filepath.Join(a.Paths().DataDirectory(), relativeResourcesPath)
}

func restoreResources(l astikit.SeverityLogger, a *astilectron.Astilectron, asset Asset, assetDir AssetDir, assetRestorer RestoreAssets, relativeResourcesPath string) (err error) {
	// Check resources
	var restore bool
	var computedChecksums map[string]string
	var checksumsPath string
	if restore, computedChecksums, checksumsPath, err = checkResources(l, a, asset, assetDir, relativeResourcesPath); err != nil {
		err = fmt.Errorf("checking resources failed: %w", err)
		return
	}

	// Restore resources
	if restore {
		if err = restoreResourcesFunc(l, a, relativeResourcesPath, assetRestorer, computedChecksums, checksumsPath); err != nil {
			err = fmt.Errorf("restoring resources failed: %w", err)
			return
		}
	} else {
		l.Debug("Skipping restoring resources...")
	}
	return
}

func checkResources(l astikit.SeverityLogger, a *astilectron.Astilectron, asset Asset, assetDir AssetDir, relativeResourcesPath string) (restore bool, computedChecksums map[string]string, checksumsPath string, err error) {
	// Compute checksums
	arp := absoluteResourcesPath(a, relativeResourcesPath)
	checksumsPath = filepath.Join(arp, "checksums.json")
	if asset != nil && assetDir != nil {
		computedChecksums = make(map[string]string)
		if err = checksumAssets(asset, assetDir, relativeResourcesPath, computedChecksums); err != nil {
			err = fmt.Errorf("getting checksum of assets failed: %w", err)
			return
		}
	}

	// Stat resources
	if _, err = os.Stat(arp); err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("stating %s failed: %w", arp, err)
		return
	} else if os.IsNotExist(err) {
		l.Debug("Resources folder doesn't exist, restoring resources...")
		err = nil
		restore = true
		return
	}

	// No computed checksums
	if computedChecksums == nil {
		l.Debug("No computed checksums, restoring resources...")
		restore = true
		return
	}

	// Stat checksums file
	if _, err = os.Stat(checksumsPath); err != nil && !os.IsNotExist(err) {
		err = fmt.Errorf("stating %s failed: %w", checksumsPath, err)
		return
	} else if os.IsNotExist(err) {
		l.Debug("Checksums file doesn't exist, restoring resources...")
		err = nil
		restore = true
		return
	}

	// Open resources checksums
	var f *os.File
	if f, err = os.Open(checksumsPath); err != nil {
		err = fmt.Errorf("opening %s failed: %w", checksumsPath, err)
		return
	}
	defer f.Close()

	// Unmarshal checksums
	var unmarshaledChecksums map[string]string
	if err = json.NewDecoder(f).Decode(&unmarshaledChecksums); err != nil {
		err = fmt.Errorf("unmarshaling checksums failed: %w", err)
		return
	}

	// Check number of paths
	if len(unmarshaledChecksums) != len(computedChecksums) {
		l.Debugf("%d paths in unmarshaled checksums != %d paths in computed checksums, restoring resources...", len(unmarshaledChecksums), len(computedChecksums))
		restore = true
		return
	}

	// Loop through computed checksums
	for p, c := range computedChecksums {
		// Path doesn't exist in unmarshaled checksums
		v, ok := unmarshaledChecksums[p]
		if !ok {
			l.Debugf("Path %s doesn't exist in unmarshaled checksums, restoring resources...", p)
			restore = true
			return
		}

		// Checksums are different
		if c != v {
			l.Debugf("Unmarshaled checksum (%s) != computed checksum (%s) for path %s, restoring resources...", v, c, p)
			restore = true
			return
		}
	}
	return
}

func checksumAssets(asset Asset, assetDir AssetDir, name string, m map[string]string) (err error) {
	// Get children
	children, errDir := assetDir(name)

	// File
	if errDir != nil {
		// Get checksum
		var h string
		if h, err = checksumAsset(asset, name); err != nil {
			err = fmt.Errorf("getting checksum of %s failed: %w", name, err)
			return
		}
		m[name] = h
		return
	}

	// Dir
	for _, child := range children {
		if err = checksumAssets(asset, assetDir, filepath.Join(name, child), m); err != nil {
			err = fmt.Errorf("getting checksum of assets in %s failed: %w", name, err)
			return
		}
	}
	return
}

func checksumAsset(asset Asset, name string) (o string, err error) {
	// Get data
	var b []byte
	if b, err = asset(name); err != nil {
		err = fmt.Errorf("getting data from asset %s failed: %w", name, err)
		return
	}

	// Hash
	h := md5.New()
	if _, err = h.Write(b); err != nil {
		err = fmt.Errorf("writing data of asset %s to hash failed: %w", name, err)
		return
	}
	o = base64.StdEncoding.EncodeToString(h.Sum(nil))
	return
}

func restoreResourcesFunc(l astikit.SeverityLogger, a *astilectron.Astilectron, relativeResourcesPath string, assetRestorer RestoreAssets, computedChecksums map[string]string, checksumsPath string) (err error) {
	// Remove resources
	arp := absoluteResourcesPath(a, relativeResourcesPath)
	l.Debugf("Removing %s", arp)
	if err = os.RemoveAll(arp); err != nil {
		err = fmt.Errorf("removing %s failed: %w", arp, err)
		return
	}

	// Restore resources
	l.Debugf("Restoring resources in %s", arp)
	if err = assetRestorer(a.Paths().DataDirectory(), relativeResourcesPath); err != nil {
		err = fmt.Errorf("restoring resources in %s failed: %w", arp, err)
		return
	}

	// Write checksums
	if computedChecksums != nil {
		// Create checksums file
		var f *os.File
		if f, err = os.Create(checksumsPath); err != nil {
			err = fmt.Errorf("creating %s failed: %w", checksumsPath, err)
			return
		}
		defer f.Close()

		// Marshal
		if err = json.NewEncoder(f).Encode(computedChecksums); err != nil {
			err = fmt.Errorf("marshaling checksums failed: %w", err)
			return
		}
	}
	return
}
