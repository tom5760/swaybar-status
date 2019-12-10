package networkmanager

import (
	"fmt"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

// https://developer.gnome.org/NetworkManager/stable/spec.html

const (
	nmIface = "org.freedesktop.NetworkManager"

	nmPath = "/org/freedesktop/NetworkManager"

	nmPropActiveConnections = nmIface + ".ActiveConnections"
)

// NetworkManager provides access to the connection manager.
type NetworkManager struct {
	obj *utils.DBusObject
}

// New creates a new instance of the NetworkManager interface.
func New() (*NetworkManager, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	nm := &NetworkManager{
		obj: utils.NewDBusObject(conn, nmIface, nmPath),
	}

	return nm, nil
}

// ActiveConnections returns the list of active connection object paths.
func (n *NetworkManager) ActiveConnections() ([]*ActiveConnection, error) {
	paths, err := n.obj.PropertySliceObjectPath(nmPropActiveConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to read the ActiveConnections property: %w", err)
	}

	conns := make([]*ActiveConnection, len(paths))

	for i, path := range paths {
		conn, err := newActiveConnection(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create new connection: %w", err)
		}
		conns[i] = conn
	}

	return conns, nil
}
