package pulseaudio

// https://gavv.github.io/articles/pulseaudio-under-the-hood/#d-bus-api
// https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/Developer/Clients/DBus/

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

const (
	sessionIface = "org.PulseAudio1"
	serverIface  = "org.PulseAudio"

	lookupPath        = "/org/pulseaudio/server_lookup1"
	lookupIface       = serverIface + ".ServerLookup1"
	lookupPropAddress = lookupIface + ".Address"

	corePath  = "/org/pulseaudio/core1"
	coreIface = serverIface + ".Core1"

	corePropName         = coreIface + ".Name"
	corePropFallbackSink = coreIface + ".FallbackSink"
	corePropSinks        = coreIface + ".Sinks"

	coreMethodListenForSignal        = coreIface + ".ListenForSignal"
	coreMethodStopListeningForSignal = coreIface + ".StopListeningForSignal"

	coreSigFallbackSinkUpdated = coreIface + ".FallbackSinkUpdated"
)

type Core struct {
	conn *dbus.Conn
	obj  *utils.DBusObject
}

func New() (*Core, error) {
	sessionbus, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to session bus: %w", err)
	}

	lookup := utils.NewDBusObject(sessionbus, sessionIface, lookupPath)

	addr, err := lookup.PropertyString(lookupPropAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup server address: %w", err)
	}

	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to pulseaudio bus: %w", err)
	}

	if err := conn.Auth(nil); err != nil {
		return nil, fmt.Errorf("failed to authenticate to pulseaudio bus: %w", err)
	}

	return &Core{
		conn: conn,
		obj:  utils.NewDBusObject(conn, coreIface, corePath),
	}, nil
}

func (c *Core) Name() (string, error) {
	return c.obj.PropertyString(corePropName)
}

func (c *Core) FallbackSink() (*Device, error) {
	path, err := c.obj.PropertyObjectPath(corePropFallbackSink)
	if err != nil {
		return nil, fmt.Errorf("failed to get object path: %w", err)
	}

	return newDevice(c, path), nil
}

func (c *Core) Sinks() ([]*Device, error) {
	paths, err := c.obj.PropertySliceObjectPath(corePropSinks)
	if err != nil {
		return nil, fmt.Errorf("failed to read Sinks property: %w", err)
	}

	sinks := make([]*Device, len(paths))

	for i, path := range paths {
		sinks[i] = newDevice(c, path)
	}

	return sinks, nil
}

func (c *Core) signalSubscribe(name string) (<-chan *dbus.Signal, func(), error) {
	if err := c.listenForSignal(name); err != nil {
		return nil, nil, err
	}

	sigChan := make(chan *dbus.Signal, 1)
	c.conn.Signal(sigChan)

	filterSigChan := make(chan *dbus.Signal, 1)

	go func() {
		for sig := range sigChan {
			if sig.Name != name {
				continue
			}
			filterSigChan <- sig
		}
	}()

	unsub := func() {
		if err := c.stopListeningForSignal(name); err != nil {
			log.Println("failed to stop listening for signal:", err)
		}

		c.conn.RemoveSignal(sigChan)
		close(sigChan)
		close(filterSigChan)
	}

	return filterSigChan, unsub, nil
}

func (c *Core) listenForSignal(name string) error {
	if err := c.obj.Call(coreMethodListenForSignal, 0, name, []dbus.ObjectPath{}).Store(); err != nil {
		return fmt.Errorf("failed to start listening for signal: %w", err)
	}

	return nil
}

func (c *Core) stopListeningForSignal(name string) error {
	if err := c.obj.Call(coreMethodStopListeningForSignal, 0, name, []dbus.ObjectPath{}).Store(); err != nil {
		return fmt.Errorf("failed to make dbus call: %w", err)
	}

	return nil
}

func (c *Core) SubscribeFallbackSinkUpdated() (<-chan *Device, func(), error) {
	sigChan, unsubSig, err := c.signalSubscribe(coreSigFallbackSinkUpdated)
	if err != nil {
		return nil, nil, err
	}

	sinkChan := make(chan *Device, 1)

	go func() {
		for sig := range sigChan {
			var objPath dbus.ObjectPath
			if err := dbus.Store(sig.Body, &objPath); err != nil {
				log.Println("failed to store signal:", err)
				continue
			}

			sinkChan <- newDevice(c, objPath)
		}
	}()

	unsub := func() {
		unsubSig()
		close(sinkChan)
	}

	return sinkChan, unsub, nil
}
