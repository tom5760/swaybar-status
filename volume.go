package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os/exec"

	"github.com/tom5760/swaybar-status/pulseaudio"
)

const (
	speakerMuted = "🔇"
	speakerLow   = "🔈"
	speakerMed   = "🔉"
	speakerHigh  = "🔊"

	volumeScrollDelta = 2
)

func volumeToPercent(v []uint32) int {
	if len(v) == 0 {
		return 0
	}

	return int(math.Round(float64(v[0]) / pulseaudio.VolumeNorm * 100))
}

func percentToVolume(p int) []uint32 {
	if p > 100 {
		p = 100
	}

	if p < 0 {
		p = 0
	}

	return []uint32{uint32(math.Round((float64(p) / 100) * pulseaudio.VolumeNorm))}
}

func statusVolume(ctx context.Context, blockChan chan<- Block) func() error {
	return func() error {
		core, err := pulseaudio.New()
		if err != nil {
			return fmt.Errorf("failed to create pulseaudio: %w", err)
		}

		var (
			volChan  <-chan []uint32
			muteChan <-chan bool

			unsubVol, unsubMute func()

			curSink *pulseaudio.Device
			percent int
			muted   bool
		)

		block := Block{
			Name: "10-volume",
			ClickHandler: func(evt ClickEvent) {
				switch evt.Button {
				case 1:
					if err := curSink.SetMute(!muted); err != nil {
						log.Println("failed to toggle mute:", err)
					}

				case 3:
					if err := exec.Command("swaymsg", "exec", "pavucontrol").Start(); err != nil {
						log.Println("failed to start pavucontrol:", err)
					}

				case 4:
					v := percentToVolume(percent + volumeScrollDelta)
					if err := curSink.SetVolume(v); err != nil {
						log.Println("failed to raise volume:", err)
					}

				case 5:
					v := percentToVolume(percent - volumeScrollDelta)
					if err := curSink.SetVolume(v); err != nil {
						log.Println("failed to lower volume:", err)
					}
				}
			},
		}

		defer func() {
			if unsubVol != nil {
				unsubVol()
			}
			if unsubMute != nil {
				unsubMute()
			}
		}()

		updateStatus := func() {
			var icon string
			switch {
			case percent == 0:
				icon = speakerLow
			case percent <= 50:
				icon = speakerMed
			default:
				icon = speakerHigh
			}

			if muted {
				icon = speakerMuted
			}

			block.FullText = fmt.Sprintf("%s%v%%", icon, percent)
			blockChan <- block
		}

		reloadSink := func(sink *pulseaudio.Device) error {
			if unsubVol != nil {
				unsubVol()
				unsubVol = nil
			}
			if unsubMute != nil {
				unsubMute()
				unsubMute = nil
			}

			var err error

			if sink == nil {
				if sink, err = core.FallbackSink(); err != nil {
					return fmt.Errorf("failed to get fallback sink: %w", err)
				}
			}

			volume, err := sink.Volume()
			if err != nil {
				return fmt.Errorf("failed to get sink volume: %w", err)
			}

			percent = volumeToPercent(volume)

			if muted, err = sink.Mute(); err != nil {
				return fmt.Errorf("failed to get sink mute: %w", err)
			}

			volChan, unsubVol, err = sink.SubscribeVolumeUpdated()
			if err != nil {
				return fmt.Errorf("failed to subscribe to volume events: %w", err)
			}

			muteChan, unsubMute, err = sink.SubscribeMuteUpdated()
			if err != nil {
				return fmt.Errorf("failed to subscribe to mute events: %w", err)
			}

			curSink = sink

			updateStatus()
			return nil
		}

		if err := reloadSink(nil); err != nil {
			return err
		}

		sinkChan, sinkUnsub, err := core.SubscribeFallbackSinkUpdated()
		if err != nil {
			return fmt.Errorf("failed to subscribe to fallback sink changes: %w", err)
		}
		defer sinkUnsub()

	loop:
		for ctx.Err() == nil {
			select {
			case sink := <-sinkChan:
				if err := reloadSink(sink); err != nil {
					return err
				}

			case volume := <-volChan:
				percent = volumeToPercent(volume)
				updateStatus()

			case muted = <-muteChan:
				updateStatus()

			case <-ctx.Done():
				break loop
			}
		}

		return nil
	}
}
