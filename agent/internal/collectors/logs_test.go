package collectors

import (
	"os"
	"path/filepath"
	"testing"
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
