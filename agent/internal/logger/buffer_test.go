package logger

import (
	"context"
	"encoding/json"
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
	if err := buffer.PublishSnapshot(context.Background(), collectors.Snapshot{AgentName: "four", Collected: time.Now()}); err != nil {
		t.Fatalf("PublishSnapshot() after ack error = %v", err)
	}
	batch, err = buffer.ReadBatch(10)
	if err != nil {
		t.Fatalf("ReadBatch() after publish error = %v", err)
	}
	if len(batch) != 2 || batch[0].AgentName != "three" || batch[1].AgentName != "four" {
		t.Fatalf("ReadBatch() after publish = %#v", batch)
	}
}

func TestJSONLBufferQuarantinesCorruptLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "buffer.jsonl")
	first, _ := json.Marshal(collectors.Snapshot{AgentName: "one", Collected: time.Now()})
	second, _ := json.Marshal(collectors.Snapshot{AgentName: "two", Collected: time.Now()})
	if err := os.WriteFile(path, []byte(string(first)+"\nnot-json\n"+string(second)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	buffer, err := NewJSONLBuffer(config.BufferConfig{Path: path, MaxEvents: 10})
	if err != nil {
		t.Fatal(err)
	}
	defer buffer.Close()

	batch, err := buffer.ReadBatch(10)
	if err != nil {
		t.Fatalf("ReadBatch() error = %v", err)
	}
	if len(batch) != 2 || batch[0].AgentName != "one" || batch[1].AgentName != "two" {
		t.Fatalf("batch = %#v", batch)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "not-json") {
		t.Fatalf("main buffer still contains corrupt line: %s", string(data))
	}
	matches, err := filepath.Glob(path + ".corrupt.*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("corrupt quarantine files = %#v", matches)
	}
	corrupt, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(corrupt), "not-json") {
		t.Fatalf("quarantine missing corrupt line: %s", string(corrupt))
	}
}
