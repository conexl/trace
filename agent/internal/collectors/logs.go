package collectors

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"agent/internal/config"
)

type LogCollector struct {
	mu      sync.Mutex
	offsets map[string]int64
}

func NewLogCollector() *LogCollector {
	return &LogCollector{offsets: make(map[string]int64)}
}

func (c *LogCollector) Collect(ctx context.Context, streams []config.LogStream) []LogChunk {
	chunks := make([]LogChunk, 0, len(streams))
	for _, stream := range streams {
		select {
		case <-ctx.Done():
			return chunks
		default:
		}
		data, offset, truncated, err := c.readNext(stream.Path, stream.MaxBytes)
		chunk := LogChunk{Name: stream.Name, Path: stream.Path, Data: data, Offset: offset, Truncated: truncated, Collected: time.Now()}
		if err != nil {
			chunk.Error = err.Error()
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (c *LogCollector) readNext(path string, maxBytes int64) (string, int64, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, offset, truncated, err := readFromOffset(path, c.offsets[path], maxBytes)
	if err == nil {
		c.offsets[path] = offset
	}
	return data, offset, truncated, err
}

func readFromOffset(path string, offset int64, maxBytes int64) (string, int64, bool, error) {
	if maxBytes <= 0 {
		maxBytes = 16 * 1024
	}
	file, err := os.Open(path)
	if err != nil {
		return "", offset, false, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", offset, false, err
	}
	if info.IsDir() {
		return "", offset, false, errors.New("log path is a directory")
	}
	if offset > info.Size() {
		offset = 0
	}

	available := info.Size() - offset
	start := offset
	truncated := false
	if available > maxBytes {
		start = info.Size() - maxBytes
		truncated = true
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return "", offset, false, err
	}
	buf, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		return "", offset, false, err
	}
	return string(buf), info.Size(), truncated, nil
}

func tailFile(path string, maxBytes int64) (string, bool, error) {
	data, _, truncated, err := readFromOffset(path, 0, maxBytes)
	return data, truncated, err
}
