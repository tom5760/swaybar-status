package main

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

const (
	writeDebounceTime = 250 * time.Millisecond
)

var header = Header{
	Version:     1,
	ClickEvents: true,
}

type StatusBar struct {
	sf   singleflight.Group
	lock sync.Mutex

	w       io.Writer
	encoder *json.Encoder

	blockMap  map[BlockKey]Block
	blockList []Block

	clickMap map[BlockKey]func(ClickEvent)
}

func NewStatusBar(w io.Writer) *StatusBar {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	return &StatusBar{
		w:        w,
		encoder:  encoder,
		blockMap: make(map[BlockKey]Block),
		clickMap: make(map[BlockKey]func(ClickEvent)),
	}
}

func (s *StatusBar) Open() error {
	if err := s.encoder.Encode(header); err != nil {
		return fmt.Errorf("failed to encode header: %w", err)
	}

	if _, err := s.w.Write([]byte("[{}\n")); err != nil {
		return fmt.Errorf("failed to write body start: %w", err)
	}

	return nil
}

func (s *StatusBar) Close() error {
	if _, err := s.w.Write([]byte{']', '\n'}); err != nil {
		return fmt.Errorf("failed to write body array end: %w", err)
	}

	if cl, ok := s.w.(io.Closer); ok {
		return cl.Close()
	}

	return nil
}

// Update inserts or updates a block to the status bar.  Blocks are sorted
// lexographically by Name-Instance.
func (s *StatusBar) Update(block Block) {
	if block.Name == "" {
		panic("block has no name")
	}

	key := block.Key()

	s.lock.Lock()
	defer s.lock.Unlock()

	prevBlock := s.blockMap[key]

	// Ignore blocks if they haven't changed.
	if block == prevBlock {
		return
	}

	s.blockMap[key] = block
	s.sort()
	s.write()
}

// Removes a block based on its Name-Instance key.
func (s *StatusBar) Remove(key BlockKey) {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.blockMap, key)
	s.sort()
	s.write()
}

func (s *StatusBar) OnClick(key BlockKey, fn func(ClickEvent)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.clickMap[key] = fn
}

func (s *StatusBar) Click(evt ClickEvent) {
	key := BlockKey{
		Name:     evt.Name,
		Instance: evt.Instance,
	}

	if fn, ok := s.clickMap[key]; ok {
		fn(evt)
	}
}

func (s *StatusBar) write() error {
	if _, err := s.w.Write([]byte{','}); err != nil {
		return fmt.Errorf("failed to write body array separator: %w", err)
	}

	if err := s.encoder.Encode(s.blockList); err != nil {
		return fmt.Errorf("failed to encode blocks: %w", err)
	}

	return nil
}

func (s *StatusBar) sort() {
	s.blockList = s.blockList[:0]

	for _, block := range s.blockMap {
		s.blockList = append(s.blockList, block)
	}

	sort.Slice(s.blockList, func(i, j int) bool {
		a := s.blockList[i]
		b := s.blockList[j]

		if a.Name > b.Name {
			return true
		}
		if a.Name < b.Name {
			return false
		}
		return a.Instance > b.Instance
	})
}
