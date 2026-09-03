package usb

import (
	"github.com/alexballas/portal/internal/apis"
	"github.com/alexballas/portal/internal/session"
	"github.com/godbus/dbus/v5"
)

const createSessionCallName = interfaceName + ".CreateSession"

// CreateSession creates a USB monitoring session.
// This is only necessary to receive device events, like device being plugged or unplugged.
//
// A successfully created session can at any time be closed,
// or may at any time be closed by the portal implementation.
func CreateSession() (*Session, error) {
	data := map[string]dbus.Variant{"session_handle_token": session.GenerateToken()}

	result, err := apis.Call(createSessionCallName, data)
	if err != nil {
		return nil, err
	}

	return &Session{path: result.(dbus.ObjectPath)}, nil
}
