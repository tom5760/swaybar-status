package main

import (
	"context"
	"time"
)

const timeFormat = "Mon Jan _2, 2006 3:04PM"

func statusTime(ctx context.Context, blockChan chan<- Block) func() error {
	return func() error {
		block := Block{
			Name: "00-time",
		}

		timer := time.NewTimer(0)

	loop:
		for ctx.Err() == nil {
			select {
			case <-timer.C:
				block.FullText = time.Now().Format(timeFormat)

				blockChan <- block

				timer.Reset(1 * time.Second)

			case <-ctx.Done():
				break loop
			}
		}

		return nil
	}
}
