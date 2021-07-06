package main

import (
	"context"
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"

	"github.com/tom5760/swaybar-status/mpris"
	"github.com/tom5760/swaybar-status/utils"
)

const (
	playerIcon = "♪"

	playerStatusPlaying = "▶️"
	playerStatusPaused  = "⏸️"
	playerStatusStopped = "⏹️"
)

func playerBlock(player *mpris.Player) (Block, error) {
	status, err := player.PlaybackStatus()
	if err != nil {
		return Block{}, fmt.Errorf("failed to get player '%s' playback status: %w", player.Name, err)
	}

	metadata, err := player.Metadata()
	if err != nil {
		return Block{}, fmt.Errorf("failed to get player '%s' metadata: %w", player.Name, err)
	}

	title, err := metadata.Title()
	if err != nil {
		return Block{}, fmt.Errorf("failed to get player '%s' title: %w", player.Name, err)
	}

	artists, err := metadata.Artist()
	if err != nil {
		return Block{}, fmt.Errorf("failed to get player '%s' artist: %w", player.Name, err)
	}

	var artist string
	if len(artists) > 0 {
		artist = " - " + artists[0]
	}

	if title == "" || artist == "" {
		return Block{}, nil
	}

	var icon string
	switch status {
	case mpris.PlaybackStatusPlaying:
		icon = playerStatusPlaying
	case mpris.PlaybackStatusPaused:
		icon = playerStatusPaused
	case mpris.PlaybackStatusStopped:
		icon = playerStatusStopped
	default:
		icon = playerIcon
	}

	return Block{
		Name:     "40-player",
		Instance: player.Name,
		FullText: fmt.Sprintf("%v %v%v", icon, title, artist),
	}, nil
}

func statusPlayer(ctx context.Context, sb *StatusBar) error {
	sessionBus, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	var (
		players   []*mpris.Player
		instances = make(map[string]bool)
	)

	playersChangeChan, playersChangeUnsub, err := mpris.SubscribePlayers(sessionBus)
	if err != nil {
		return fmt.Errorf("failed to subscribe to player changes: %w", err)
	}
	defer playersChangeUnsub()

	propertyChangeChan, propertyChangeUnsub, err := utils.DBusSubscribePropertyChanges(sessionBus)
	if err != nil {
		return fmt.Errorf("failed to subscribe to player property changes: %w", err)
	}
	defer propertyChangeUnsub()

	for ctx.Err() == nil {
		players, err = mpris.Players(sessionBus)
		if err != nil {
			return fmt.Errorf("failed to list players: %w", err)
		}

		for i := range instances {
			instances[i] = false
		}

		for _, player := range players {
			block, err := playerBlock(player)
			if err != nil {
				log.Println("failed to make player block:", err)
				continue
			}

			if block.FullText == "" {
				continue
			}

			instances[player.Name] = true

			sb.Update(block)

			p := player
			sb.OnClick(block.Key(), func(evt ClickEvent) {
				switch evt.Button {
				case 1:
					if err := p.PlayPause(); err != nil {
						log.Printf("failed to play/pause player '%v': %v", p.Name, err)
					}

				case 3:
					if err := p.Next(); err != nil {
						log.Printf("failed to next player '%v': %v", p.Name, err)
					}
				}
			})
		}

		for name, exists := range instances {
			if !exists {
				sb.Remove(BlockKey{
					Name:     "40-player",
					Instance: name,
				})
				delete(instances, name)
			}
		}

		select {
		case change := <-playersChangeChan:
			log.Printf("PLAYER CHANGED: %#v", change)

		case change := <-propertyChangeChan:
			log.Printf("PROPERTY CHANGED: %#v", change)

		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
