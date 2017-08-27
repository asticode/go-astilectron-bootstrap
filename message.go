package bootstrap

import (
	"encoding/json"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

// MessageOut represents a message going out
type MessageOut struct {
	CallbackID *int        `json:"callbackId,omitempty"`
	Name       string      `json:"name"`
	Payload    interface{} `json:"payload"`
}

// MessageIn represents a message going in
type MessageIn struct {
	CallbackID *int            `json:"callbackId,omitempty"`
	Name       string          `json:"name"`
	Payload    json.RawMessage `json:"payload"`
}

// handleMessages handles messages
func handleMessages(w *astilectron.Window, messageHandler MessageHandler) astilectron.Listener {
	return func(e astilectron.Event) (deleteListener bool) {
		// Unmarshal message
		var m MessageIn
		var err error
		if err = e.Message.Unmarshal(&m); err != nil {
			astilog.Error(errors.Wrapf(err, "unmarshaling message %+v failed", *e.Message))
			return
		}

		// Handle message
		var p interface{}
		if p, err = messageHandler(w, m); err != nil {
			astilog.Error(errors.Wrapf(err, "handling message %+v failed", m))
			return
		}

		// Send message
		if p != nil && m.CallbackID != nil {
			var m = MessageOut{CallbackID: m.CallbackID, Name: m.Name, Payload: p}
			if err = w.Send(m); err != nil {
				astilog.Error(errors.Wrapf(err, "sending message %+v failed", m))
				return
			}
		}
		return
	}
}
