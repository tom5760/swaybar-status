package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/lawl/pulseaudio"
)

const (
	speakerMuted = "🔇"
	speakerLow   = "🔈"
	speakerMed   = "🔉"
	speakerHigh  = "🔊"

	volumeScrollDelta = .02
)

func statusVolume(ctx context.Context, sb *StatusBar) error {
	client, err := pulseaudio.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create pulseaudio client: %w", err)
	}

	defer client.Close()

	updates, err := client.Updates()
	if err != nil {
		return fmt.Errorf("failed to subscribe to pulseaudio updates: %w", err)
	}

	block := Block{
		Name: "10-volume",
	}

	sb.OnClick(block.Key(), func(evt ClickEvent) {
		switch evt.Button {
		case 1:
			if _, err := client.ToggleMute(); err != nil {
				log.Println("failed to toggle mute:", err)
			}

		case 3:
			if err := exec.Command("swaymsg", "exec", "pavucontrol").Start(); err != nil {
				log.Println("failed to start pavucontrol:", err)
			}

		case 4:
			setVolume(client, volumeScrollDelta)

		case 5:
			setVolume(client, -volumeScrollDelta)
		}
	})

	updateVolumeBlock(sb, client)

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-updates:
			updateVolumeBlock(sb, client)
		}
	}
}

func updateVolumeBlock(sb *StatusBar, client *pulseaudio.Client) {
	block := Block{
		Name:     "10-volume",
		FullText: "Error",
	}

	volume, err := client.Volume()
	if err != nil {
		log.Printf("failed to get volume: %v", err)
		sb.Update(block)
		return
	}

	muted, err := client.Mute()
	if err != nil {
		log.Printf("failed to get mute state: %v", err)
		sb.Update(block)
		return
	}

	icon := speakerMuted

	if !muted {
		switch {
		case volume == 0:
			icon = speakerLow
		case volume <= .5:
			icon = speakerMed
		default:
			icon = speakerHigh
		}
	}

	block.FullText = fmt.Sprintf("%s%.0f%%", icon, volume*100)

	sb.Update(block)
}

func setVolume(client *pulseaudio.Client, diff float32) {
	volume, err := client.Volume()
	if err != nil {
		log.Printf("failed to get volume: %v", err)
		return
	}

	if err := client.SetVolume(volume + diff); err != nil {
		log.Printf("failed to set volume: %v", err)
		return
	}
}
