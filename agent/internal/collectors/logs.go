package collectors

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"agent/internal/config"
)

type LogCollector struct{}

func NewLogCollector() *LogCollector {
	return &LogCollector{}
}

func (c *LogCollector) Collect(ctx context.Context, streams []config.LogStream) []LogChunk {
	chunks := make([]LogChunk, 0, len(streams))
	for _, stream := range streams {
		select {
		case <-ctx.Done():
			return chunks
		default:
		}
		data, truncated, err := tailFile(stream.Path, stream.MaxBytes)
		chunk := LogChunk{Name: stream.Name, Path: stream.Path, Data: data, Truncated: truncated, Collected: time.Now()}
		if err != nil {
			chunk.Error = err.Error()
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func tailFile(path string, maxBytes int64) (string, bool, error) {
	if maxBytes <= 0 {
		maxBytes = 16 * 1024
	}
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", false, err
	}
	if info.IsDir() {
		return "", false, errors.New("log path is a directory")
	}

	start := int64(0)
	truncated := false
	if info.Size() > maxBytes {
		start = info.Size() - maxBytes
		truncated = true
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return "", false, err
	}
	buf, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		return "", false, err
	}
	return string(buf), truncated, nil
}
