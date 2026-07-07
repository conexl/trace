package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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
	Count() int
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
	if err := b.file.Sync(); err != nil {
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
	lines, corrupt, err := readJSONLLines(b.path)
	if err != nil {
		return nil, err
	}
	if len(corrupt) > 0 {
		if err := quarantineCorruptLines(b.path, corrupt); err != nil {
			return nil, err
		}
		if err := b.rewriteLocked(lines); err != nil {
			return nil, err
		}
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
	lines, corrupt, err := readJSONLLines(b.path)
	if err != nil {
		return err
	}
	if len(corrupt) > 0 {
		if err := quarantineCorruptLines(b.path, corrupt); err != nil {
			return err
		}
	}
	if count >= len(lines) {
		lines = nil
	} else {
		lines = lines[count:]
	}
	return b.rewriteLocked(lines)
}

func (b *JSONLBuffer) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.events
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
	lines, corrupt, err := readJSONLLines(b.path)
	if err != nil {
		return err
	}
	if len(corrupt) > 0 {
		if err := quarantineCorruptLines(b.path, corrupt); err != nil {
			return err
		}
	}
	if len(lines) > b.maxEvents {
		lines = lines[len(lines)-b.maxEvents:]
	}
	return b.rewriteLocked(lines)
}

func (b *JSONLBuffer) rewriteLocked(lines [][]byte) error {
	dir := filepath.Dir(b.path)
	tmp, err := os.CreateTemp(dir, ".homelytics-buffer-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()
	if len(lines) > 0 {
		payload := append(bytes.Join(lines, []byte("\n")), '\n')
		if _, err := tmp.Write(payload); err != nil {
			_ = tmp.Close()
			return err
		}
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		return err
	}
	if err := b.file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, b.path); err != nil {
		reopened, reopenErr := openBufferFile(b.path)
		if reopenErr == nil {
			b.file = reopened
		}
		if reopenErr != nil {
			return fmt.Errorf("%w; reopen buffer: %v", err, reopenErr)
		}
		return err
	}
	reopened, err := openBufferFile(b.path)
	if err != nil {
		return err
	}
	b.file = reopened
	cleanup = false
	b.events = len(lines)
	return nil
}

func openBufferFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
}

func quarantineCorruptLines(path string, lines [][]byte) error {
	if len(lines) == 0 {
		return nil
	}
	quarantinePath := fmt.Sprintf("%s.corrupt.%d", path, time.Now().UTC().UnixNano())
	payload := append(bytes.Join(lines, []byte("\n")), '\n')
	return os.WriteFile(quarantinePath, payload, 0o600)
}

func countJSONLLines(path string) (int, error) {
	lines, _, err := readJSONLLines(path)
	if err != nil {
		return 0, err
	}
	return len(lines), nil
}

func readJSONLLines(path string) ([][]byte, [][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	data = bytes.TrimRight(data, "\n")
	if len(data) == 0 {
		return nil, nil, nil
	}
	raw := bytes.Split(data, []byte("\n"))
	lines := make([][]byte, 0, len(raw))
	corrupt := make([][]byte, 0)
	for _, line := range raw {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		if !json.Valid(trimmed) {
			corrupt = append(corrupt, bytes.Clone(line))
			continue
		}
		lines = append(lines, bytes.Clone(line))
	}
	return lines, corrupt, nil
}
