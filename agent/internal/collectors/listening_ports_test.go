package collectors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseProcNetListeningTCP(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tcp")
	content := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode\n" +
		"   0: 0100007F:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000 1000 0 4242 1 0000000000000000 100 0 0 10 0\n" +
		"   1: 0100007F:1F91 00000000:0000 01 00000000:00000000 00:00000000 00000000 1000 0 4243 1 0000000000000000 100 0 0 10 0\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	entries := parseProcNet(path, "tcp4")
	if len(entries) != 1 {
		t.Fatalf("entries = %#v", entries)
	}
	if entries[0].address != "127.0.0.1" || entries[0].port != 8080 || entries[0].inode != "4242" {
		t.Fatalf("entry = %#v", entries[0])
	}
}
