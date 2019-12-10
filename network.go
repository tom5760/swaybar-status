package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tom5760/swaybar-status/networkmanager"
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
				block.FullText = fmt.Sprintf("🖧 %s", label)

			case networkmanager.ActiveConnectionWireless:
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

				block.FullText = fmt.Sprintf("📶 %s", label)

			default:
				log.Println("unexpected connection type:", typ)
				continue
			}

			blockChan <- block
		}

		return nil
	}
}
