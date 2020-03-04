package networkmanager

import (
	"fmt"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

type AccessPoint struct {
	obj *utils.DBusObject
}

const (
	accessPointIface = nmIface + ".AccessPoint"

	accessPointPropSSID     = accessPointIface + ".Ssid"
	accessPointPropStrength = accessPointIface + ".Strength"
)

func newAccessPoint(path dbus.ObjectPath) (*AccessPoint, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	ap := &AccessPoint{
		obj: utils.NewDBusObject(bus, nmIface, path),
	}

	return ap, nil
}

// SSID returns the Service Set Identifier identifying the access point.
func (a *AccessPoint) SSID() ([]byte, error) {
	ssid, err := a.obj.PropertyByteSlice(accessPointPropSSID)
	if err != nil {
		return nil, fmt.Errorf("failed to read the Ssid property: %w", err)
	}

	return ssid, nil
}

// Strength returns the current signal quality of the access point, in percent.
func (a *AccessPoint) Strength() (uint8, error) {
	strength, err := a.obj.PropertyByte(accessPointPropStrength)
	if err != nil {
		return 0, fmt.Errorf("failed to read the Strength property: %w", err)
	}

	return strength, nil
}
