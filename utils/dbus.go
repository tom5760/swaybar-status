package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/godbus/dbus/v5"
)

// DBusObject wraps dbus.BusObject and adds a few more convenience methods.
type DBusObject struct {
	obj dbus.BusObject
}

// NewDBusObject creates and wraps the given dbus object.
func NewDBusObject(conn *dbus.Conn, dest string, path dbus.ObjectPath) *DBusObject {
	return &DBusObject{conn.Object(dest, path)}
}

func (o *DBusObject) Call(method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return o.obj.Call(method, flags, args...)
}

func (o *DBusObject) Property(name string) (interface{}, error) {
	variant, err := o.obj.GetProperty(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	return variant.Value(), nil
}

func (o *DBusObject) SetProperty(name string, value interface{}) error {
	return o.obj.SetProperty(name, dbus.MakeVariant(value))
}

func (o *DBusObject) PropertyInt64(name string) (int64, error) {
	v, err := o.Property(name)
	if err != nil {
		return 0, err
	}

	x, ok := v.(int64)
	if !ok {
		return 0, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyUint32(name string) (uint32, error) {
	v, err := o.Property(name)
	if err != nil {
		return 0, err
	}

	x, ok := v.(uint32)
	if !ok {
		return 0, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyUint64(name string) (uint64, error) {
	v, err := o.Property(name)
	if err != nil {
		return 0, err
	}

	x, ok := v.(uint64)
	if !ok {
		return 0, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyFloat64(name string) (float64, error) {
	v, err := o.Property(name)
	if err != nil {
		return 0, err
	}

	x, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyBool(name string) (bool, error) {
	v, err := o.Property(name)
	if err != nil {
		return false, err
	}

	x, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyString(name string) (string, error) {
	v, err := o.Property(name)
	if err != nil {
		return "", err
	}

	x, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyByte(name string) (byte, error) {
	v, err := o.Property(name)
	if err != nil {
		return 0, err
	}

	x, ok := v.(byte)
	if !ok {
		return 0, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyByteSlice(name string) ([]byte, error) {
	v, err := o.Property(name)
	if err != nil {
		return nil, err
	}

	x, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertyObjectPath(name string) (dbus.ObjectPath, error) {
	v, err := o.Property(name)
	if err != nil {
		return "", err
	}

	x, ok := v.(dbus.ObjectPath)
	if !ok {
		return "", fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertySliceUint32(name string) ([]uint32, error) {
	v, err := o.Property(name)
	if err != nil {
		return nil, err
	}

	x, ok := v.([]uint32)
	if !ok {
		return nil, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertySliceString(name string) ([]string, error) {
	v, err := o.Property(name)
	if err != nil {
		return nil, err
	}

	x, ok := v.([]string)
	if !ok {
		return nil, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func (o *DBusObject) PropertySliceObjectPath(name string) ([]dbus.ObjectPath, error) {
	v, err := o.Property(name)
	if err != nil {
		return nil, err
	}

	x, ok := v.([]dbus.ObjectPath)
	if !ok {
		return nil, fmt.Errorf("unexpected variant type; got %T; expected %T", v, x)
	}

	return x, nil
}

func DBusSignalSubscribe(conn *dbus.Conn, name string) (<-chan *dbus.Signal, func(), error) {
	iface := ""

	i := strings.LastIndex(name, ".")
	if i != -1 {
		iface = name[:i]
	}

	member := name[i+1:]

	matchOptions := []dbus.MatchOption{
		dbus.WithMatchInterface(iface),
		dbus.WithMatchMember(member),
	}

	if err := conn.AddMatchSignal(matchOptions...); err != nil {
		return nil, nil, fmt.Errorf("failed to add match signal: %w", err)
	}

	sigChan := make(chan *dbus.Signal, 1)
	conn.Signal(sigChan)

	cancelChan := make(chan struct{})
	filterSigChan := make(chan *dbus.Signal, 1)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				if sig == nil {
					return
				}

				if sig.Name != name {
					continue
				}

				filterSigChan <- sig

			case <-cancelChan:
				return
			}
		}
	}()

	cancel := func() {
		conn.RemoveSignal(sigChan)

		if err := conn.RemoveMatchSignal(matchOptions...); err != nil {
			log.Println("failed to remove match signal: %w", err)
		}

		close(sigChan)
		close(cancelChan)
		close(filterSigChan)
	}

	return filterSigChan, cancel, nil
}
