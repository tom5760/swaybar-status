package main

import (
	"context"
	"time"
)

const timeFormat = "Mon Jan 2, 2006 3:04PM"

func statusTime(ctx context.Context, sb *StatusBar) error {
	block := Block{
		Name: "00-time",
	}

	timer := time.NewTimer(0)

	for ctx.Err() == nil {
		select {
		case <-timer.C:
			block.FullText = time.Now().Format(timeFormat)
			sb.Update(block)
			timer.Reset(1 * time.Second)

		case <-ctx.Done():
			return nil
		}
	}

	return nil
}
