package logger

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agent/internal/collectors"
	"agent/internal/config"
)

func TestJSONLBufferKeepsLastEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buffer.jsonl")
	buffer, err := NewJSONLBuffer(config.BufferConfig{Path: path, MaxEvents: 2})
	if err != nil {
		t.Fatal(err)
	}
	defer buffer.Close()

	for _, name := range []string{"one", "two", "three"} {
		if err := buffer.PublishSnapshot(context.Background(), collectors.Snapshot{AgentName: name, Collected: time.Now()}); err != nil {
			t.Fatalf("PublishSnapshot() error = %v", err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "one") {
		t.Fatalf("buffer retained old event: %s", content)
	}
	if !strings.Contains(content, "two") || !strings.Contains(content, "three") {
		t.Fatalf("buffer did not retain latest events: %s", content)
	}
}

func TestJSONLBufferReadBatchAndAck(t *testing.T) {
	path := filepath.Join(t.TempDir(), "buffer.jsonl")
	buffer, err := NewJSONLBuffer(config.BufferConfig{Path: path, MaxEvents: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer buffer.Close()

	for _, name := range []string{"one", "two", "three"} {
		if err := buffer.PublishSnapshot(context.Background(), collectors.Snapshot{AgentName: name, Collected: time.Now()}); err != nil {
			t.Fatalf("PublishSnapshot() error = %v", err)
		}
	}

	batch, err := buffer.ReadBatch(2)
	if err != nil {
		t.Fatalf("ReadBatch() error = %v", err)
	}
	if len(batch) != 2 || batch[0].AgentName != "one" || batch[1].AgentName != "two" {
		t.Fatalf("ReadBatch() = %#v", batch)
	}
	if err := buffer.Ack(2); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}
	batch, err = buffer.ReadBatch(10)
	if err != nil {
		t.Fatalf("ReadBatch() after ack error = %v", err)
	}
	if len(batch) != 1 || batch[0].AgentName != "three" {
		t.Fatalf("ReadBatch() after ack = %#v", batch)
	}
}
