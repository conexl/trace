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

type BufferedSink interface {
	Sink
	ReadBatch(limit int) ([]collectors.Snapshot, error)
	Ack(count int) error
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

func (b *JSONLBuffer) ReadBatch(limit int) ([]collectors.Snapshot, error) {
	if limit <= 0 {
		return nil, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.file.Sync(); err != nil {
		return nil, err
	}
	lines, err := readJSONLLines(b.path)
	if err != nil {
		return nil, err
	}
	if len(lines) > limit {
		lines = lines[:limit]
	}
	batch := make([]collectors.Snapshot, 0, len(lines))
	for _, line := range lines {
		var snapshot collectors.Snapshot
		if err := json.Unmarshal(line, &snapshot); err != nil {
			return nil, fmt.Errorf("decode buffered snapshot: %w", err)
		}
		batch = append(batch, snapshot)
	}
	return batch, nil
}

func (b *JSONLBuffer) Ack(count int) error {
	if count <= 0 {
		return nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := b.file.Sync(); err != nil {
		return err
	}
	lines, err := readJSONLLines(b.path)
	if err != nil {
		return err
	}
	if count >= len(lines) {
		lines = nil
	} else {
		lines = lines[count:]
	}
	return b.rewriteLocked(lines)
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
	lines, err := readJSONLLines(b.path)
	if err != nil {
		return err
	}
	if len(lines) > b.maxEvents {
		lines = lines[len(lines)-b.maxEvents:]
	}
	return b.rewriteLocked(lines)
}

func (b *JSONLBuffer) rewriteLocked(lines [][]byte) error {
	if err := b.file.Truncate(0); err != nil {
		return err
	}
	if _, err := b.file.Seek(0, 0); err != nil {
		return err
	}
	if len(lines) > 0 {
		payload := append(bytes.Join(lines, []byte("\n")), '\n')
		if _, err := b.file.Write(payload); err != nil {
			return err
		}
	}
	b.events = len(lines)
	return nil
}

func countJSONLLines(path string) (int, error) {
	lines, err := readJSONLLines(path)
	if err != nil {
		return 0, err
	}
	return len(lines), nil
}

func readJSONLLines(path string) ([][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	data = bytes.TrimRight(data, "\n")
	if len(data) == 0 {
		return nil, nil
	}
	raw := bytes.Split(data, []byte("\n"))
	lines := make([][]byte, 0, len(raw))
	for _, line := range raw {
		lines = append(lines, bytes.Clone(line))
	}
	return lines, nil
}
