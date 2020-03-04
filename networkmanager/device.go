package networkmanager

import (
	"fmt"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

type (
	DeviceType uint32

	Device struct {
		obj *utils.DBusObject
	}
)

const (
	deviceIface = nmIface + ".Device"

	devicePropDeviceType = deviceIface + ".DeviceType"
)

const (
	// unknown device
	DeviceTypeUnknown DeviceType = iota
	// a wired ethernet device
	DeviceTypeEthernet
	// an 802.11 Wi-Fi device
	DeviceTypeWifi
	// not used
	DeviceTypeUnused1
	// not used
	DeviceTypeUnused2
	// a Bluetooth device supporting PAN or DUN access protocols
	DeviceTypeBT
	// an OLPC XO mesh networking device
	DeviceTypeOLPCMesh
	// an 802.16e Mobile WiMAX broadband device
	DeviceTypeWiMAX
	// a modem supporting analog telephone, CDMA/EVDO, GSM/UMTS, or LTE network access protocols
	DeviceTypeModem
	// an IP-over-InfiniBand device
	DeviceTypeInfiniband
	// a bond master interface
	DeviceTypeBond
	// an 802.1Q VLAN interface
	DeviceTypeVLAN
	// ADSL modem
	DeviceTypeADSL
	// a bridge master interface
	DeviceTypeBridge
	// generic support for unrecognized device types
	DeviceTypeGeneric
	// a team master interface
	DeviceTypeTeam
	// a TUN or TAP interface
	DeviceTypeTUN
	// a IP tunnel interface
	DeviceTypeIPTunnel
	// a MACVLAN interface
	DeviceTypeMACVLAN
	// a VXLAN interface
	DeviceTypeVXLAN
	// a VETH interface
	DeviceTypeVETH
	// a MACsec interface
	DeviceTypeMACSec
	// a dummy interface
	DeviceTypeDummy
	// a PPP interface
	DeviceTypePPP
	// a Open vSwitch interface
	DeviceTypeOvSInterface
	// a Open vSwitch port
	DeviceTypeOvSPort
	// a Open vSwitch bridge
	DeviceTypeOvSBridge
	// a IEEE 802.15.4 (WPAN) MAC Layer Device
	DeviceTypeWPAN
	// 6LoWPAN interface
	DeviceType6LoWPAN
	// a WireGuard interface
	DeviceTypeWireguard
	// an 802.11 Wi-Fi P2P device (Since: 1.16)
	DeviceTypeWifiP2P
)

func newDevice(path dbus.ObjectPath) (*Device, error) {
	bus, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	device := &Device{
		obj: utils.NewDBusObject(bus, nmIface, path),
	}

	return device, nil
}

// Type returns the general type of the network device; ie Ethenet, Wi-Fi, etc.
func (d *Device) Type() (DeviceType, error) {
	typ, err := d.obj.PropertyUint32(devicePropDeviceType)
	if err != nil {
		return DeviceTypeUnknown, fmt.Errorf("failed to read the DeviceType property: %w", err)
	}

	return DeviceType(typ), nil
}

// WirelessDevice casts this device to a WirelessDevice.
func (d *Device) WirelessDevice() *WirelessDevice {
	return &WirelessDevice{obj: d.obj}
}
