package networkmanager

import (
	"fmt"

	"github.com/tom5760/swaybar-status/utils"
)

type WirelessDevice struct {
	obj *utils.DBusObject
}

const (
	wirelessDeviceIface = deviceIface + ".Wireless"

	wirelessDevicePropActiveAccessPoint = wirelessDeviceIface + ".ActiveAccessPoint"
)

// ActiveAccessPoint returns the access point currently used by the wireless
// device.
func (d *WirelessDevice) ActiveAccessPoint() (*AccessPoint, error) {
	path, err := d.obj.PropertyObjectPath(wirelessDevicePropActiveAccessPoint)
	if err != nil {
		return nil, fmt.Errorf("failed to read the ActiveAccessPoint property: %w", err)
	}

	ap, err := newAccessPoint(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create new access point: %w", err)
	}

	return ap, nil
}
