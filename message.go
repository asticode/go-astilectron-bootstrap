package bootstrap

import (
	"encoding/json"
	"fmt"

	"github.com/asticode/go-astikit"
	"github.com/asticode/go-astilectron"
)

// MessageOut represents a message going out
type MessageOut struct {
	Name    string      `json:"name"`
	Payload interface{} `json:"payload,omitempty"`
}

// MessageIn represents a message going in
type MessageIn struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func handleMessages(w *astilectron.Window, messageHandler MessageHandler, l astikit.SeverityLogger) astilectron.ListenerMessage {
	return func(m *astilectron.EventMessage) (v interface{}) {
		// Unmarshal message
		var i MessageIn
		var err error
		if err = m.Unmarshal(&i); err != nil {
			l.Error(fmt.Errorf("unmarshaling message %+v failed: %w", *m, err))
			return
		}

		// Handle message
		var p interface{}
		if p, err = messageHandler(w, i); err != nil {
			l.Error(fmt.Errorf("handling message %+v failed: %w", i, err))
		}

		// Return message
		if p != nil {
			o := &MessageOut{Name: i.Name + ".callback", Payload: p}
			if err != nil {
				o.Name = "error"
			}
			v = o
		}
		return
	}
}

// CallbackMessage represents a bootstrap message callback
type CallbackMessage func(m *MessageIn)

// SendMessage sends a message
func SendMessage(w *astilectron.Window, name string, payload interface{}, cs ...CallbackMessage) error {
	var callbacks []astilectron.CallbackMessage
	for _, c := range cs {
		callbacks = append(callbacks, func(e *astilectron.EventMessage) {
			var m *MessageIn
			if e != nil {
				m = &MessageIn{}
				if err := e.Unmarshal(m); err != nil {
					return
				}
			}
			c(m)
		})
	}
	return w.SendMessage(MessageOut{Name: name, Payload: payload}, callbacks...)
}
