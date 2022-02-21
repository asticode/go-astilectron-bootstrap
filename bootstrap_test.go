package bootstrap

import (
	"github.com/asticode/go-astilectron"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	err := Run(Options{
		AstilectronOptions: astilectron.Options{},
		Debug:              false,
		WithOns: map[string]WithOn{astilectron.EventNameAppCmdQuit: func(e astilectron.Event) (deleteListener bool) {
			t.Log("-----", time.Now(), "-----")
			return
		}},
		Windows: []*Window{{
			Homepage: "app/index.html",
			Options:  &astilectron.WindowOptions{},
		}},
	})

	if err != nil {
		panic(err)
	}
}
