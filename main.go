package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"golang.org/x/sync/errgroup"
)

var (
	inputReader  = io.Reader(os.Stdin)
	outputWriter = io.Writer(os.Stdout)

	header = Header{
		Version:     1,
		ClickEvents: true,
	}
)

func main() {
	os.Exit(run())
}

func run() int {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			panic(r)
		}
	}()

	group, ctx := errgroup.WithContext(context.Background())

	clickChan := make(chan ClickEvent, 1)
	blocksChan := make(chan []Block, 1)

	group.Go(recv(ctx, clickChan))
	group.Go(send(ctx, blocksChan))
	group.Go(status(ctx, blocksChan, clickChan))

	if err := group.Wait(); err != nil {
		log.Println(err)
		return 1
	}

	return 0
}

func recv(ctx context.Context, clickChan chan<- ClickEvent) func() error {
	return func() error {
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

			clickChan <- evt
		}

		close(clickChan)

		return nil
	}
}

func send(_ context.Context, blocksChan <-chan []Block) func() error {
	return func() error {
		encoder := json.NewEncoder(outputWriter)
		encoder.SetEscapeHTML(false)

		if err := encoder.Encode(header); err != nil {
			return fmt.Errorf("failed to encode header: %w", err)
		}

		firstBlock := true

		if _, err := outputWriter.Write([]byte{'[', '\n'}); err != nil {
			return fmt.Errorf("failed to write body array start: %w", err)
		}

		for blocks := range blocksChan {
			if firstBlock {
				firstBlock = false
			} else if _, err := outputWriter.Write([]byte{',', '\n'}); err != nil {
				return fmt.Errorf("failed to write body array separator: %w", err)
			}

			if err := encoder.Encode(blocks); err != nil {
				return fmt.Errorf("failed to encode blocks: %w", err)
			}
		}

		if _, err := outputWriter.Write([]byte{']', '\n'}); err != nil {
			return fmt.Errorf("failed to write body array end: %w", err)
		}

		return nil
	}
}

type blockKey struct {
	name     string
	instance string
}

func status(ctx context.Context, blocksChan chan<- []Block, clickChan <-chan ClickEvent) func() error {
	return func() error {
		body := make(map[blockKey]Block)
		blockChan := make(chan Block, 1)

		group, statusCtx := errgroup.WithContext(ctx)

		statusFuncs := []func(context.Context, chan<- Block) func() error{
			statusBattery,
			statusTime,
			statusVolume,
		}

		for _, statusFunc := range statusFuncs {
			group.Go(statusFunc(statusCtx, blockChan))
		}

	loop:
		for ctx.Err() == nil {
			select {
			case block := <-blockChan:
				key := blockKey{
					name:     block.Name,
					instance: block.Instance,
				}
				body[key] = block

			case evt := <-clickChan:
				key := blockKey{
					name:     evt.Name,
					instance: evt.Instance,
				}
				block, ok := body[key]
				if !ok {
					log.Println("click event on non-existent key:", key)
					continue loop
				}
				if block.ClickHandler != nil {
					block.ClickHandler(evt)
				}

			case <-statusCtx.Done():
				break loop
			}

			blocks := make([]Block, 0, len(body))

			for _, block := range body {
				blocks = append(blocks, block)
			}

			sort.Slice(blocks, func(i, j int) bool {
				a := blocks[i]
				b := blocks[j]

				if a.Name > b.Name {
					return true
				}
				if a.Name < b.Name {
					return false
				}
				return a.Instance > b.Instance
			})

			blocksChan <- blocks
		}

		close(blocksChan)

		if err := group.Wait(); err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		return nil
	}
}
