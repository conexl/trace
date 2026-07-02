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
