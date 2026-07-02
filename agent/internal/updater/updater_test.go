package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyDownloadsVerifiesAndReplacesTarget(t *testing.T) {
	payload := []byte("new binary")
	sum := sha256.Sum256(payload)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()
	target := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := New().Apply(context.Background(), server.URL, hex.EncodeToString(sum[:]), target)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !result.Updated || result.Bytes != int64(len(payload)) {
		t.Fatalf("result = %#v", result)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("target = %q", got)
	}
}

func TestApplyRejectsSHA256Mismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("new binary"))
	}))
	defer server.Close()
	target := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := New().Apply(context.Background(), server.URL, "bad", target); err == nil {
		t.Fatal("Apply() expected checksum error")
	}
}
