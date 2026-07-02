package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"agent/internal/collectors"
	"agent/internal/config"
)

type Sink interface {
	PublishSnapshot(ctx context.Context, snapshot collectors.Snapshot) error
	Close() error
}

type JSONLBuffer struct {
	mu        sync.Mutex
	file      *os.File
	path      string
	stdout    bool
	maxEvents int
	events    int
}

func NewJSONLBuffer(cfg config.BufferConfig) (*JSONLBuffer, error) {
	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, err
	}
	events, err := countJSONLLines(cfg.Path)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &JSONLBuffer{file: file, path: cfg.Path, stdout: cfg.MirrorToStdout, maxEvents: cfg.MaxEvents, events: events}, nil
}

func (b *JSONLBuffer) PublishSnapshot(ctx context.Context, snapshot collectors.Snapshot) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, err := b.file.Write(append(payload, '\n')); err != nil {
		return err
	}
	b.events++
	if b.maxEvents > 0 && b.events > b.maxEvents {
		if err := b.compactLocked(); err != nil {
			return err
		}
	}
	if b.stdout {
		fmt.Println(string(payload))
	}
	return nil
}

func (b *JSONLBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.file.Close()
}

func (b *JSONLBuffer) compactLocked() error {
	if err := b.file.Sync(); err != nil {
		return err
	}
	data, err := os.ReadFile(b.path)
	if err != nil {
		return err
	}
	data = bytes.TrimRight(data, "\n")
	if len(data) == 0 {
		b.events = 0
		return nil
	}
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) > b.maxEvents {
		lines = lines[len(lines)-b.maxEvents:]
	}
	rotated := append(bytes.Join(lines, []byte("\n")), '\n')
	if err := b.file.Truncate(0); err != nil {
		return err
	}
	if _, err := b.file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := b.file.Write(rotated); err != nil {
		return err
	}
	b.events = len(lines)
	return nil
}

func countJSONLLines(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	data = bytes.TrimRight(data, "\n")
	if len(data) == 0 {
		return 0, nil
	}
	return len(bytes.Split(data, []byte("\n"))), nil
}
