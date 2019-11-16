package pulseaudio

import (
	"fmt"
	"log"
	"math"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

const (
	VolumeMuted = 0
	VolumeNorm  = 0x10000
	VolumeMax   = math.MaxUint32 - 1
)

const (
	deviceIface = coreIface + ".Device"

	devicePropName   = deviceIface + ".Name"
	devicePropVolume = deviceIface + ".Volume"
	devicePropMute   = deviceIface + ".Mute"

	deviceSigVolumeUpdated = deviceIface + ".VolumeUpdated"
	deviceSigMuteUpdated   = deviceIface + ".MuteUpdated"
)

type Device struct {
	core *Core
	obj  *utils.DBusObject
}

func newDevice(core *Core, path dbus.ObjectPath) *Device {
	return &Device{
		core: core,
		obj:  utils.NewDBusObject(core.conn, deviceIface, path),
	}
}

func (d *Device) Name() (string, error) {
	return d.obj.PropertyString(devicePropName)
}

func (d *Device) Volume() ([]uint32, error) {
	return d.obj.PropertySliceUint32(devicePropVolume)
}

func (d *Device) SetVolume(volume []uint32) error {
	return d.obj.SetProperty(devicePropVolume, volume)
}

func (d *Device) Mute() (bool, error) {
	return d.obj.PropertyBool(devicePropMute)
}

func (d *Device) SetMute(mute bool) error {
	return d.obj.SetProperty(devicePropMute, mute)
}

func (d *Device) SubscribeVolumeUpdated() (<-chan []uint32, func(), error) {
	sigChan, unsubSig, err := d.core.signalSubscribe(deviceSigVolumeUpdated)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start listenting for signal: %w", err)
	}

	volChan := make(chan []uint32, 1)

	go func() {
		for sig := range sigChan {
			var volume []uint32
			if err := dbus.Store(sig.Body, &volume); err != nil {
				log.Println("failed to store signal:", err)
				continue
			}

			volChan <- volume
		}
	}()

	unsub := func() {
		unsubSig()
		close(volChan)
	}

	return volChan, unsub, nil
}

func (d *Device) SubscribeMuteUpdated() (<-chan bool, func(), error) {
	sigChan, unsubSig, err := d.core.signalSubscribe(deviceSigMuteUpdated)
	if err != nil {
		return nil, nil, err
	}

	muteChan := make(chan bool, 1)

	go func() {
		for sig := range sigChan {
			var mute bool
			if err := dbus.Store(sig.Body, &mute); err != nil {
				log.Println("failed to store signal:", err)
				continue
			}

			muteChan <- mute
		}
	}()

	unsub := func() {
		unsubSig()
		close(muteChan)
	}

	return muteChan, unsub, nil
}
