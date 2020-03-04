package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tom5760/swaybar-status/networkmanager"
)

const (
	networkIconEthernet = "🖧"
	networkIconWireless = "📶"
)

func statusNetwork(ctx context.Context, blockChan chan<- Block) func() error {
	return func() error {
		nm, err := networkmanager.New()
		if err != nil {
			return fmt.Errorf("failed to create networkmanager: %w", err)
		}

		conns, err := nm.ActiveConnections()
		if err != nil {
			return fmt.Errorf("failed to get active connections: %w", err)
		}

		for i, conn := range conns {
			block := Block{
				Name:     "30-networking",
				Instance: fmt.Sprintf("%v", i),
			}

			typ, err := conn.Type()
			if err != nil {
				return fmt.Errorf("failed to get connection type: %w", err)
			}

			state, err := conn.State()
			if err != nil {
				return fmt.Errorf("failed to get connection state: %w", err)
			}

			devices, err := conn.Devices()
			if err != nil {
				return fmt.Errorf("failed to get connection devices: %w", err)
			}

			for _, device := range devices {
				typ, err := device.Type()
				if err != nil {
					return fmt.Errorf("failed to get device type: %w", err)
				}

				log.Printf("device %v: %v", device, typ)
			}

			switch typ {
			case networkmanager.ActiveConnectionEthernet:
				var label string
				switch state {
				case networkmanager.ActiveConnectionStateActivating:
					label = "activating..."
				case networkmanager.ActiveConnectionStateActivated:
					label = "up"
				case networkmanager.ActiveConnectionStateDeactivating:
					label = "deactivating..."
				case networkmanager.ActiveConnectionStateDeactivated:
					label = "down"
				}
				block.FullText = fmt.Sprintf("%s %s", networkIconEthernet, label)

			case networkmanager.ActiveConnectionWireless:
				status, err := getWifiStatus(conn)
				if err != nil {
					log.Println("failed to get wifi status:", err)
				}

				var label string
				switch state {
				case networkmanager.ActiveConnectionStateActivating:
					label = "activating..."
				case networkmanager.ActiveConnectionStateActivated:
					label = ""
				case networkmanager.ActiveConnectionStateDeactivating:
					label = "deactivating..."
				case networkmanager.ActiveConnectionStateDeactivated:
					label = "down"
				}

				block.FullText = fmt.Sprintf("%s %s%s", networkIconWireless, status, label)

			default:
				log.Println("unexpected connection type:", typ)
				continue
			}

			blockChan <- block
		}

		return nil
	}
}

func getWifiStatus(conn *networkmanager.ActiveConnection) (string, error) {
	dev, err := findWifiDev(conn)
	if err != nil {
		return "", fmt.Errorf("failed to find wifi device: %w", err)
	}

	wifi := dev.WirelessDevice()

	ap, err := wifi.ActiveAccessPoint()
	if err != nil {
		return "", fmt.Errorf("failed to get active access point: %w", err)
	}

	ssid, err := ap.SSID()
	if err != nil {
		return "", fmt.Errorf("failed to get SSID: %w", err)
	}

	strength, err := ap.Strength()
	if err != nil {
		return "", fmt.Errorf("failed to get Strength: %w", err)
	}

	return fmt.Sprintf("%s (%v%%)", string(ssid), strength), nil
}

func findWifiDev(conn *networkmanager.ActiveConnection) (*networkmanager.Device, error) {
	devices, err := conn.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection devices: %w", err)
	}

	for _, device := range devices {
		typ, err := device.Type()
		if err != nil {
			return nil, fmt.Errorf("failed to get device type: %w", err)
		}

		if typ == networkmanager.DeviceTypeWifi {
			return device, nil
		}
	}

	return nil, fmt.Errorf("wifi device not found")
}
