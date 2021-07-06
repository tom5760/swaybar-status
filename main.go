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
		//statusPlayer,
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	go recv(ctx, cancel, sb)

	for i, statusFunc := range statusFuncs {
		n := i
		fn := statusFunc
		group.Go(func() error {
			if err := fn(ctx, sb); err != nil {
				return fmt.Errorf("status function %v failed: %w", n, err)
			}

			log.Printf("function %v finished", n)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	return nil
}

func recv(ctx context.Context, cancel context.CancelFunc, sb *StatusBar) error {
	defer cancel()

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
