package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/sync/errgroup"
)

var (
	inputReader io.Reader = os.Stdin

	statusFuncs = []func(context.Context, *StatusBar) error{
		statusBattery,
		statusNetwork,
		//    statusPlayer,
		statusTime,
		statusVolume,
	}
)

func main() {
	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run() error {
	sb := NewStatusBar(os.Stdout)

	if err := sb.Open(); err != nil {
		return fmt.Errorf("failed to open status bar: %w", err)
	}

	defer func() {
		if err := sb.Close(); err != nil {
			log.Println("failed to close status bar:", err)
		}
	}()

	group, ctx := errgroup.WithContext(context.Background())

	group.Go(func() error {
		return recv(ctx, sb)
	})

	for _, statusFunc := range statusFuncs {
		fn := statusFunc
		group.Go(func() error { return fn(ctx, sb) })
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("status failed: %w", err)
	}

	return nil
}

func recv(ctx context.Context, sb *StatusBar) error {
	decoder := json.NewDecoder(inputReader)

	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read initial input token: %w", err)
	}

	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("unexpected initial input token: %v", tok)
	}

	for ctx.Err() == nil {
		var evt ClickEvent
		if err := decoder.Decode(&evt); err != nil {
			return fmt.Errorf("failed to decode click event: %w", err)
		}

		sb.Click(evt)
	}

	return nil
}
