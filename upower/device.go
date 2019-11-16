package upower

import (
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/utils"
)

const (
	deviceIface = upowerIface + ".Device"

	devicePropNativePath       = deviceIface + ".NativePath"
	devicePropVendor           = deviceIface + ".Vendor"
	devicePropModel            = deviceIface + ".Model"
	devicePropSerial           = deviceIface + ".Serial"
	devicePropUpdateTime       = deviceIface + ".UpdateTime"
	devicePropType             = deviceIface + ".Type"
	devicePropPowerSupply      = deviceIface + ".PowerSupply"
	devicePropHasHistory       = deviceIface + ".HasHistory"
	devicePropHasStatistics    = deviceIface + ".HasStatistics"
	devicePropOnline           = deviceIface + ".Online"
	devicePropEnergy           = deviceIface + ".Energy"
	devicePropEnergyEmpty      = deviceIface + ".EnergyEmpty"
	devicePropEnergyFull       = deviceIface + ".EnergyFull"
	devicePropEnergyFullDesign = deviceIface + ".EnergyFullDesign"
	devicePropEnergyRate       = deviceIface + ".EnergyRate"
	devicePropVoltage          = deviceIface + ".Voltage"
	devicePropLuminosity       = deviceIface + ".Luminosity"
	devicePropTimeToEmpty      = deviceIface + ".TimeToEmpty"
	devicePropTimeToFull       = deviceIface + ".TimeToFull"
	devicePropPercentage       = deviceIface + ".Percentage"
	devicePropTemperature      = deviceIface + ".Temperature"
	devicePropIsPresent        = deviceIface + ".IsPresent"
	devicePropState            = deviceIface + ".State"
	devicePropIsRechargeable   = deviceIface + ".IsRechargeable"
	devicePropCapacity         = deviceIface + ".Capacity"
	devicePropTechnology       = deviceIface + ".Technology"
	devicePropWarningLevel     = deviceIface + ".WarningLevel"
	devicePropBatteryLevel     = deviceIface + ".BatteryLevel"
	devicePropIconName         = deviceIface + ".IconName"

	deviceMethodRefresh       = deviceIface + ".Refresh"
	deviceMethodGetHistory    = deviceIface + ".GetHistory"
	deviceMethodGetStatistics = deviceIface + ".GetStatistics"
)

// DeviceType is the type of power source.
type DeviceType uint32

// Valid values for DeviceType.
const (
	DeviceTypeUnknown DeviceType = iota
	DeviceTypeLinePower
	DeviceTypeBattery
	DeviceTypeUPS
	DeviceTypeMonitor
	DeviceTypeMouse
	DeviceTypeKeyboard
	DeviceTypePDA
	DeviceTypePhone
)

// DeviceState is the battery power state.
type DeviceState uint32

// Valid values for DeviceState.
const (
	DeviceStateUnknown DeviceState = iota
	DeviceStateCharging
	DeviceStateDischarging
	DeviceStateEmpty
	DeviceStateFullyCharged
	DeviceStatePendingCharge
	DeviceStatePendingDischarge
)

// DeviceTechnology is the technology used in the battery.
type DeviceTechnology uint32

// Valid values for DeviceTechnology.
const (
	DeviceTechnologyUnknown DeviceTechnology = iota
	DeviceTechnologyLithiumIon
	DeviceTechnologyLithiumPolymer
	DeviceTechnologyLithiumIronPhosphate
	DeviceTechnologyLeadAcid
	DeviceTechnologyNickelCadmium
	DeviceTechnologyNickelMetalHydride
)

// DeviceWarningLevel is the warning level of the battery.
type DeviceWarningLevel uint32

// Valid values for DeviceWarningLevel.
const (
	DeviceWarningLevelUnknown DeviceWarningLevel = iota
	DeviceWarningLevelNone
	DeviceWarningLevelDischarging
	DeviceWarningLevelLow
	DeviceWarningLevelCritical
	DeviceWarningLevelAction
)

// DeviceBatteryLevel is the level of the battery.
type DeviceBatteryLevel uint32

// Valid values for DeviceBatteryLevel.
const (
	DeviceBatteryLevelUnknown DeviceBatteryLevel = iota
	DeviceBatteryLevelNone
	DeviceBatteryLevelLow
	DeviceBatteryLevelCritical
	DeviceBatteryLevelNormal
	DeviceBatteryLevelHigh
	DeviceBatteryLevelFull
)

// DeviceHistoryType is the type of history to request from GetHistory.
type DeviceHistoryType string

// Valid values for DeviceHistoryType.
const (
	DeviceHistoryRate   DeviceHistoryType = "rate"
	DeviceHistoryCharge DeviceHistoryType = "charge"
)

// DeviceStatisticsType is the type of statistics to request from
// GetStatistics.
type DeviceStatisticsType string

// Valid values for DeviceHistoryType.
const (
	DeviceStatisticsCharging    DeviceStatisticsType = "charging"
	DeviceStatisticsDischarging DeviceStatisticsType = "discharging"
)

// DeviceHistoryRecord is a history record from GetHistory.
type DeviceHistoryRecord struct {
	// The time value in seconds from the gettimeofday() method.
	Time time.Time

	// The data value, for instance the rate in W or the charge in %.
	Value float64

	// The state of the device, for instance charging or discharging.
	State DeviceState
}

// DeviceStatisticsRecord is a statistics record from GetStatistics.
type DeviceStatisticsRecord struct {
	// The value of the percentage point, usually in seconds
	Value float64

	// The accuracy of the prediction in percent.
	Accuracy float64
}

// Device represents a power device.
type Device struct {
	obj *utils.DBusObject
}

func newDevice(path dbus.ObjectPath) (*Device, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create system bus: %w", err)
	}

	return &Device{
		obj: utils.NewDBusObject(conn, upowerIface, path),
	}, nil
}

// NativePath is the OS specific native path of the power source. On Linux this
// is the sysfs path, for example
// /sys/devices/LNXSYSTM:00/device:00/PNP0C0A:00/power_supply/BAT0. Is blank if
// the device is being driven by a user space driver.
func (d *Device) NativePath() (string, error) {
	return d.obj.PropertyString(devicePropNativePath)
}

// Vendor is the name of the vendor of the battery.
func (d *Device) Vendor() (string, error) {
	return d.obj.PropertyString(devicePropVendor)
}

// Model is the name of the model of this battery.
func (d *Device) Model() (string, error) {
	return d.obj.PropertyString(devicePropModel)
}

// Serial is the unique serial number of the battery.
func (d *Device) Serial() (string, error) {
	return d.obj.PropertyString(devicePropSerial)
}

// UpdateTime is the point in time that data was read from the power source.
func (d *Device) UpdateTime() (time.Time, error) {
	t, err := d.obj.PropertyUint64(devicePropUpdateTime)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(t), 0), nil
}

// Type is the type of power source.
func (d *Device) Type() (DeviceType, error) {
	v, err := d.obj.PropertyUint32(devicePropType)
	if err != nil {
		return DeviceTypeUnknown, err
	}

	return DeviceType(v), nil
}

// PowerSupply reports if the power device is used to supply the system. This
// would be set TRUE for laptop batteries and UPS devices, but set FALSE for
// wireless mice or PDAs.
func (d *Device) PowerSupply() (bool, error) {
	return d.obj.PropertyBool(devicePropPowerSupply)
}

// HasHistory reports if the power device has history.
func (d *Device) HasHistory() (bool, error) {
	return d.obj.PropertyBool(devicePropHasHistory)
}

// HasStatistics reports if the power device has statistics.
func (d *Device) HasStatistics() (bool, error) {
	return d.obj.PropertyBool(devicePropHasStatistics)
}

// Online reports whether power is currently being provided through line power.
// This property is only valid if the property type has the value "line-power".
func (d *Device) Online() (bool, error) {
	return d.obj.PropertyBool(devicePropOnline)
}

// Energy returns the amount of energy (measured in Wh) currently available in
// the power source.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) Energy() (float64, error) {
	return d.obj.PropertyFloat64(devicePropEnergy)
}

// EnergyEmpty returns the amount of energy (measured in Wh) in the power
// source when it's considered to be empty.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) EnergyEmpty() (float64, error) {
	return d.obj.PropertyFloat64(devicePropEnergyEmpty)
}

// EnergyFull returns the amount of energy (measured in Wh) in the power source
// when it's considered full.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) EnergyFull() (float64, error) {
	return d.obj.PropertyFloat64(devicePropEnergyFull)
}

// EnergyFullDesign returns the amount of energy (measured in Wh) the power
// source is designed to hold when it's considered full.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) EnergyFullDesign() (float64, error) {
	return d.obj.PropertyFloat64(devicePropEnergyFullDesign)
}

// EnergyRate returns the amount of energy being drained from the source,
// measured in W. If positive, the source is being discharged, if negative it's
// being charged.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) EnergyRate() (float64, error) {
	return d.obj.PropertyFloat64(devicePropEnergyRate)
}

// Voltage returns the voltage in the Cell or being recorded by the meter.
func (d *Device) Voltage() (float64, error) {
	return d.obj.PropertyFloat64(devicePropVoltage)
}

// Luminosity returns the luminosity being recorded by the meter.
func (d *Device) Luminosity() (float64, error) {
	return d.obj.PropertyFloat64(devicePropLuminosity)
}

// TimeToEmpty returns the number of seconds until the power source is
// considered empty. Is set to 0 if unknown.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) TimeToEmpty() (int64, error) {
	return d.obj.PropertyInt64(devicePropTimeToEmpty)
}

// TimeToFull returns the number of seconds until the power source is
// considered full. Is set to 0 if unknown.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) TimeToFull() (int64, error) {
	return d.obj.PropertyInt64(devicePropTimeToFull)
}

// Percentage returns the amount of energy left in the power source expressed
// as a percentage between 0 and 100. Typically this is the same as (energy -
// energy-empty) / (energy-full - energy-empty). However, some primitive power
// sources are capable of only reporting percentages and in this case the
// energy-* properties will be unset while this property is set.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) Percentage() (float64, error) {
	return d.obj.PropertyFloat64(devicePropPercentage)
}

// Temperature returns the temperature of the device in degrees Celsius. This
// property is only valid if the property type has the value "battery".
func (d *Device) Temperature() (float64, error) {
	return d.obj.PropertyFloat64(devicePropTemperature)
}

// IsPresent reports if the power source is present in the bay. This field is
// required as some batteries are hot-removable, for example expensive UPS and
// most laptop batteries.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) IsPresent() (bool, error) {
	return d.obj.PropertyBool(devicePropIsPresent)
}

// State
func (d *Device) State() (DeviceState, error) {
	v, err := d.obj.PropertyUint32(devicePropState)
	if err != nil {
		return DeviceStateUnknown, err
	}

	return DeviceState(v), nil
}

// IsRechargeable reports if the power source is rechargeable.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) IsRechargeable() (bool, error) {
	return d.obj.PropertyBool(devicePropIsRechargeable)
}

// Capacity returns the capacity of the power source expressed as a percentage
// between 0 and 100. The capacity of the battery will reduce with age. A
// capacity value less than 75% is usually a sign that you should renew your
// battery. Typically this value is the same as (full-design / full) * 100.
// However, some primitive power sources are not capable reporting capacity and
// in this case the capacity property will be unset.
//
// This property is only valid if the property type has the value "battery".
func (d *Device) Capacity() (float64, error) {
	return d.obj.PropertyFloat64(devicePropCapacity)
}

// Technology returns the technology used in the battery.
func (d *Device) Technology() (DeviceTechnology, error) {
	v, err := d.obj.PropertyUint32(devicePropTechnology)
	if err != nil {
		return DeviceTechnologyUnknown, err
	}

	return DeviceTechnology(v), nil
}

// WarningLevel returns the warning level of the battery.
func (d *Device) WarningLevel() (DeviceWarningLevel, error) {
	v, err := d.obj.PropertyUint32(devicePropWarningLevel)
	if err != nil {
		return DeviceWarningLevelUnknown, err
	}

	return DeviceWarningLevel(v), nil
}

// BatteryLevel returns the level of the battery.
func (d *Device) BatteryLevel() (DeviceBatteryLevel, error) {
	v, err := d.obj.PropertyUint32(devicePropBatteryLevel)
	if err != nil {
		return DeviceBatteryLevelUnknown, err
	}

	return DeviceBatteryLevel(v), nil
}

// IconName returns an icon name, following the Icon Naming Specification.
func (d *Device) IconName() (string, error) {
	return d.obj.PropertyString(devicePropIconName)
}

// Refresh refreshes the data collected from the power source.
func (d *Device) Refresh() error {
	err := d.obj.Call(deviceMethodRefresh, 0).Store()
	if err != nil {
		return fmt.Errorf("failed to make dbus call: %w", err)
	}

	return nil
}

// GetHistory gets history for the power device that is persistent across
// reboots.
//
// type: The type of history. Valid types are rate or charge.
// timespan: The amount of data to return in seconds, or 0 for all.
// resolution: The approximate number of points to return. A higher resolution
//             is more accurate, at the expense of plotting speed.
func (d *Device) GetHistory(
	typ DeviceHistoryType, timespan time.Duration, resolution uint32,
) ([]DeviceHistoryRecord, error) {
	var values [][]interface{}

	ts := uint32(timespan.Seconds())

	err := d.obj.
		Call(deviceMethodGetHistory, 0, typ, ts, resolution).
		Store(&values)
	if err != nil {
		return nil, fmt.Errorf("failed to make dbus call: %w", err)
	}

	records := make([]DeviceHistoryRecord, len(values))

	for i, value := range values {
		if len(value) != 3 {
			return nil, fmt.Errorf("unexpected result length")
		}

		record := &records[i]

		t, ok := value[0].(uint32)
		if !ok {
			return nil, fmt.Errorf("unexpected time format")
		}

		record.Time = time.Unix(int64(t), 0)

		record.Value, ok = value[1].(float64)
		if !ok {
			return nil, fmt.Errorf("unexpected value format")
		}

		state, ok := value[2].(uint32)
		if !ok {
			return nil, fmt.Errorf("unexpected state format")
		}

		record.State = DeviceState(state)
	}

	return records, nil
}

// GetStatistics gets statistics for the power device that may be interesting
// to show on a graph in the session.
func (d *Device) GetStatistics(
	typ DeviceStatisticsType,
) ([]DeviceStatisticsRecord, error) {
	var values [][]interface{}

	err := d.obj.
		Call(deviceMethodGetStatistics, 0, typ).
		Store(&values)
	if err != nil {
		return nil, fmt.Errorf("failed to make dbus call: %w", err)
	}

	records := make([]DeviceStatisticsRecord, len(values))

	for i, value := range values {
		if err := dbus.Store([]interface{}{value}, &records[i]); err != nil {
			return nil, fmt.Errorf("failed to store result: %w", err)
		}
	}

	return records, nil
}
