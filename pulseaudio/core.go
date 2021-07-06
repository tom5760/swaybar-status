package pulseaudio

// https://gavv.github.io/articles/pulseaudio-under-the-hood/#d-bus-api
// https://www.freedesktop.org/wiki/Software/PulseAudio/Documentation/Developer/Clients/DBus/

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/lawl/pulseaudio"
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
	client *pulseaudio.Client
}

func New() (*Core, error) {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create pulseaudio client: %w", err)
	}

	return &Core{
		client: client,
	}, nil
}

func (c *Core) Close() {
	return c.client.Close()
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
			log.Printf("failed to stop listening for signal %s: %v", name, err)
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
	if err := c.obj.Call(coreMethodStopListeningForSignal, 0, name).Store(); err != nil {
		return fmt.Errorf("failed to make dbus call: %w", err)
	}

	return nil
}

func (c *Core) SubscribeFallbackSinkUpdated() (<-chan *Device, func(), error) {
	sigChan, unsub, err := c.signalSubscribe(coreSigFallbackSinkUpdated)
	if err != nil {
		return nil, nil, err
	}

	sinkChan := make(chan *Device, 1)

	go func() {
		defer close(sinkChan)
		for sig := range sigChan {
			var objPath dbus.ObjectPath
			if err := dbus.Store(sig.Body, &objPath); err != nil {
				log.Println("failed to store signal:", err)
				continue
			}

			sinkChan <- newDevice(c, objPath)
		}
	}()

	return sinkChan, unsub, nil
}
