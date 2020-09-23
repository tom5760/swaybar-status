package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tom5760/swaybar-status/upower"
)

func statusBattery(ctx context.Context, blockChan chan<- Block) func() error {
	return func() error {
		up, err := upower.New()
		if err != nil {
			return fmt.Errorf("failed to create upower: %w", err)
		}

		block := Block{
			Name: "20-battery",
		}

		reloadDev := func() (*upower.Device, error) {
			dev, err := up.GetDisplayDevice()
			if err != nil {
				return nil, fmt.Errorf("failed to get display device: %w", err)
			}
			return dev, nil
		}

		dev, err := reloadDev()
		if err != nil {
			return err
		}

		timer := time.NewTimer(0)

		devAddedChan, devAddedUnsub, err := up.SubscribeDeviceAdded()
		if err != nil {
			return fmt.Errorf("failed to subscribe to device added signals: %w", err)
		}

		devRemovedChan, devRemovedUnsub, err := up.SubscribeDeviceRemoved()
		if err != nil {
			return fmt.Errorf("failed to subscribe to device removed signals: %w", err)
		}

	loop:
		for ctx.Err() == nil {
			select {
			case <-devAddedChan:
				log.Println("device added")
				if dev, err = reloadDev(); err != nil {
					return err
				}

			case <-devRemovedChan:
				log.Println("device removed")
				if dev, err = reloadDev(); err != nil {
					return err
				}

			case <-timer.C:
				if err := dev.Refresh(); err != nil {
					return fmt.Errorf("failed to refresh device: %w", err)
				}

				percent, err := dev.Percentage()
				if err != nil {
					return fmt.Errorf("failed to get percentage: %w", err)
				}

				state, err := dev.State()
				if err != nil {
					return fmt.Errorf("failed to get state: %w", err)
				}

				var label string
				switch state {
				case upower.DeviceStateUnknown:
					label = "unknown"
				case upower.DeviceStateCharging:
					label = "charging"
				case upower.DeviceStateDischarging:
					label = "discharging"
				case upower.DeviceStateEmpty:
					label = "empty"
				case upower.DeviceStateFullyCharged:
					label = "full"
				case upower.DeviceStatePendingCharge:
					label = "pending charge"
				case upower.DeviceStatePendingDischarge:
					label = "pending discharge"
				}

				block.FullText = fmt.Sprintf("🔋%v%% (%s)", percent, label)
				block.Urgent = percent < 15

				blockChan <- block

				timer.Reset(10 * time.Second)

			case <-ctx.Done():
				break loop
			}
		}

		devAddedUnsub()
		devRemovedUnsub()

		return nil
	}
}
