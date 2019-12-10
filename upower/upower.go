package upower

// https://upower.freedesktop.org/docs/ref-dbus.html

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

const (
	upowerIface = "org.freedesktop.UPower"

	upowerPath = "/org/freedesktop/UPower"

	upowerPropDaemonVersion = upowerIface + ".DaemonVersion"
	upowerPropOnBattery     = upowerIface + ".OnBattery"
	upowerPropLidIsClosed   = upowerIface + ".LidIsClosed"
	upowerPropLidIsPresent  = upowerIface + ".LidIsPresent"

	upowerMethodEnumerateDevices  = upowerIface + ".EnumerateDevices"
	upowerMethodGetDisplayDevice  = upowerIface + ".GetDisplayDevice"
	upowerMethodGetCriticalAction = upowerIface + ".GetCriticalAction"

	upowerSigDeviceAdded   = upowerIface + ".DeviceAdded"
	upowerSigDeviceRemoved = upowerIface + ".DeviceRemoved"
)

// UPower provides a wrapper around the UPower dbus service.
// See https://upower.freedesktop.org/docs/UPower.html for more info.
type UPower struct {
	obj *utils.DBusObject
}

// New creates a new instance of the UPower interface.
func New() (*UPower, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	upower := &UPower{
		obj: utils.NewDBusObject(conn, upowerIface, upowerPath),
	}

	return upower, nil
}

// DaemonVersion returns the version of the running daemon, e.g. 002.
func (u *UPower) DaemonVersion() (string, error) {
	return u.obj.PropertyString(upowerPropDaemonVersion)
}

// OnBattery indicates whether the system is running on battery power. This
// property is provided for convenience.
func (u *UPower) OnBattery() (bool, error) {
	return u.obj.PropertyBool(upowerPropOnBattery)
}

// LidIsClosed indicates if the laptop lid is closed where the display cannot
// be seen.
func (u *UPower) LidIsClosed() (bool, error) {
	return u.obj.PropertyBool(upowerPropLidIsClosed)
}

// LidIsPresent indicates if the system has a lid device.
func (u *UPower) LidIsPresent() (bool, error) {
	return u.obj.PropertyBool(upowerPropLidIsPresent)
}

// EnumerateDevices enumerates all power objects on the system.
func (u *UPower) EnumerateDevices() ([]*Device, error) {
	var paths []dbus.ObjectPath

	err := u.obj.Call(upowerMethodEnumerateDevices, 0).Store(&paths)
	if err != nil {
		return nil, fmt.Errorf("failed to make dbus call: %w", err)
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

// GetDisplayDevice returns the object to the "display device", a composite
// device that represents the status icon to show in desktop environments.
//
// The following standard org.freedesktop.UPower.Device properties will be
// defined (only IsPresent takes a special meaning):
//
// * Type: the type of the display device, UPS or Battery. Note that this value
//         can change, as opposed to real devices.
// * State: the power state of the display device, such as Charging or
//          Discharging.
// * Percentage: the amount of energy left on the device.
// * Energy: Amount of energy (measured in Wh) currently available in the power
//           source.
// * EnergyFull: Amount of energy (measured in Wh) in the power source when
//               it's considered full.
// * EnergyRate: Amount of energy being drained from the source, measured in W.
//               If positive, the source is being discharged, if negative it's
//               being charged.
// * TimeToEmpty: Number of seconds until the power source is considered empty.
// * TimeToFull: Number of seconds until the power source is considered full.
// * IsPresent: Whether a status icon using this information should be
//              presented.
// * IconName: An icon name representing the device state.
// * WarningLevel: The same as the overall WarningLevel
func (u *UPower) GetDisplayDevice() (*Device, error) {
	var path dbus.ObjectPath

	err := u.obj.Call(upowerMethodGetDisplayDevice, 0).Store(&path)
	if err != nil {
		return nil, fmt.Errorf("failed to make dbus call: %w", err)
	}

	return newDevice(path)
}

// GetCriticalAction returns the action the system will take when the system's
// power supply is critical (critically low batteries or UPS).
//
// Possible values are:
//
// * HybridSleep
// * Hibernate
// * PowerOff
func (u *UPower) GetCriticalAction() (string, error) {
	var action string

	err := u.obj.Call(upowerMethodGetCriticalAction, 0).Store(&action)
	if err != nil {
		return "", fmt.Errorf("failed to make dbus call: %w", err)
	}

	return action, err
}

// Subscribes to a signal emitted when a device is added.  Returns a channel to
// receive added devices, and a unsubscription function.
func (u *UPower) SubscribeDeviceAdded() (<-chan *Device, func(), error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	sigChan, unsubSig, err := utils.
		DBusSignalSubscribe(conn, upowerSigDeviceAdded)

	if err != nil {
		return nil, nil, err
	}

	devChan := make(chan *Device, 1)

	go deviceSigChanLoop(sigChan, devChan)

	unsub := func() {
		close(devChan)
		unsubSig()
	}

	return devChan, unsub, nil
}

// Subscribes to a signal emitted when a device is removed.  Returns a channel to
// receive removed devices, and a unsubscription function.
func (u *UPower) SubscribeDeviceRemoved() (<-chan *Device, func(), error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	sigChan, unsubSig, err := utils.
		DBusSignalSubscribe(conn, upowerSigDeviceRemoved)

	if err != nil {
		return nil, nil, err
	}

	devChan := make(chan *Device, 1)

	go deviceSigChanLoop(sigChan, devChan)

	unsub := func() {
		close(devChan)
		unsubSig()
	}

	return devChan, unsub, nil
}

func deviceSigChanLoop(sigChan <-chan *dbus.Signal, devChan chan<- *Device) {
	for sig := range sigChan {
		var objPath dbus.ObjectPath
		if err := dbus.Store(sig.Body, &objPath); err != nil {
			log.Println("failed to store signal:", err)
			continue
		}

		dev, err := newDevice(objPath)
		if err != nil {
			log.Println("failed to create new device:", err)
			continue
		}

		devChan <- dev
	}
}
