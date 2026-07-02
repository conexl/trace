package collectors

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"agent/internal/config"
)

func TestTailFileBoundsRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := os.WriteFile(path, []byte("alpha\nbeta\ngamma\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, truncated, err := tailFile(path, 6)
	if err != nil {
		t.Fatalf("tailFile() error = %v", err)
	}
	if !truncated {
		t.Fatal("tailFile() expected truncated=true")
	}
	if got != "gamma\n" {
		t.Fatalf("tailFile() = %q", got)
	}
}

func TestLogCollectorReadsOnlyNewBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	if err := os.WriteFile(path, []byte("alpha\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	collector := NewLogCollector()
	stream := []config.LogStream{{Name: "app", Path: path, MaxBytes: 1024}}

	first := collector.Collect(context.Background(), stream)
	if len(first) != 1 || first[0].Data != "alpha\n" {
		t.Fatalf("first chunk = %#v", first)
	}
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("beta\n"); err != nil {
		t.Fatal(err)
	}
	_ = file.Close()

	second := collector.Collect(context.Background(), stream)
	if len(second) != 1 || second[0].Data != "beta\n" {
		t.Fatalf("second chunk = %#v", second)
	}
}
