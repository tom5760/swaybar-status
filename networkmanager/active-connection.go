package networkmanager

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

type (
	ActiveConnectionType        string
	ActiveConnectionState       uint32
	ActiveConnectionStateReason uint32

	ActiveConnectionStateChange struct {
		State  ActiveConnectionState
		Reason ActiveConnectionStateReason
	}

	ActiveConnection struct {
		obj *utils.DBusObject
	}
)

const (
	activeConnectionIface = nmIface + ".Connection.Active"

	activeConnectionPropDevices = activeConnectionIface + ".Devices"
	activeConnectionPropType    = activeConnectionIface + ".Type"
	activeConnectionPropState   = activeConnectionIface + ".State"

	activeConnectionSigStateChanged = activeConnectionIface + ".StateChanged"
)

const (
	ActiveConnectionEthernet ActiveConnectionType = "802-3-ethernet"
	ActiveConnectionWireless ActiveConnectionType = "802-11-wireless"
)

const (
	// The state of the connection is unknown
	ActiveConnectionStateUnknown ActiveConnectionState = iota
	// A network connection is being prepared
	ActiveConnectionStateActivating
	// There is a connection to the network
	ActiveConnectionStateActivated
	// The network connection is being torn down and cleaned up
	ActiveConnectionStateDeactivating
	// The network connection is disconnected and will be removed
	ActiveConnectionStateDeactivated
)

const (
	// The reason for the active connection state change is unknown.
	ActiveConnectionStateReasonUnknown ActiveConnectionStateReason = iota
	// No reason was given for the active connection state change.
	ActiveConnectionStateReasonNone
	// The active connection changed state because the user disconnected it.
	ActiveConnectionStateReasonUserDisconnected
	// The active connection changed state because the device it was using was disconnected.
	ActiveConnectionStateReasonDeviceDisconnected
	// The service providing the VPN connection was stopped.
	ActiveConnectionStateReasonServiceStopped
	// The IP config of the active connection was invalid.
	ActiveConnectionStateReasonIpConfigInvalid
	// The connection attempt to the VPN service timed out.
	ActiveConnectionStateReasonConnectTimeout
	// A timeout occurred while starting the service providing the VPN connection.
	ActiveConnectionStateReasonServiceStartTimeout
	// Starting the service providing the VPN connection failed.
	ActiveConnectionStateReasonServiceStartFailed
	// Necessary secrets for the connection were not provided.
	ActiveConnectionStateReasonNoSecrets
	// Authentication to the server failed.
	ActiveConnectionStateReasonLoginFailed
	// The connection was deleted from settings.
	ActiveConnectionStateReasonConnectionRemoved
	// Master connection of this connection failed to activate.
	ActiveConnectionStateReasonDependencyFailed
	// Could not create the software device link.
	ActiveConnectionStateReasonDeviceRealizeFailed
	// The device this connection depended on disappeared.
	ActiveConnectionStateReasonDeviceRemoved
)

func newActiveConnection(path dbus.ObjectPath) (*ActiveConnection, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	activeConn := &ActiveConnection{
		obj: utils.NewDBusObject(bus, nmIface, path),
	}

	return activeConn, nil
}

// Devices returns an array of devices which are part of this active
// connection.
func (c *ActiveConnection) Devices() ([]*Device, error) {
	paths, err := c.obj.PropertySliceObjectPath(activeConnectionPropDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to read the Devices property: %w", err)
	}

	devices := make([]*Device, len(paths))

	for i, path := range paths {
		device, err := newDevice(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create new device: %w", err)
		}
		devices[i] = device
	}

	return devices, nil
}

// Type returns the type of the connection, provided as a convenience so that
// clients do not have to retrieve all connection details.
func (c *ActiveConnection) Type() (ActiveConnectionType, error) {
	typ, err := c.obj.PropertyString(activeConnectionPropType)
	if err != nil {
		return "", fmt.Errorf("failed to read the Type property: %w", err)
	}

	return ActiveConnectionType(typ), nil
}

// State returns the state of this active connection.
func (c *ActiveConnection) State() (ActiveConnectionState, error) {
	state, err := c.obj.PropertyUint32(activeConnectionPropState)
	if err != nil {
		return 0, fmt.Errorf("failed to read the State property: %w", err)
	}

	return ActiveConnectionState(state), nil
}

// Subscribes to a signal emitted when the state of the active connection has
// changed.
func (c *ActiveConnection) SubscribeStateChanged() (<-chan ActiveConnectionStateChange, func(), error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	sigChan, unsubSig, err := utils.
		DBusSignalSubscribe(conn, activeConnectionSigStateChanged)

	if err != nil {
		return nil, nil, err
	}

	stateChan := make(chan ActiveConnectionStateChange, 1)

	go func() {
		for sig := range sigChan {
			var change ActiveConnectionStateChange
			if err := dbus.Store(sig.Body, &change.State, &change.Reason); err != nil {
				log.Println("failed to store signal:", err)
				continue
			}
			stateChan <- change
		}
	}()

	unsub := func() {
		close(stateChan)
		unsubSig()
	}

	return stateChan, unsub, nil
}
