package updater

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
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

func TestApplyVerifiesEd25519Signature(t *testing.T) {
	payload := []byte("signed binary")
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signature := ed25519.Sign(privateKey, payload)
	mux := http.NewServeMux()
	mux.HandleFunc("/agent", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	})
	mux.HandleFunc("/agent.sig", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString(signature)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	target := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, err := New().ApplyOptions(context.Background(), Options{
		URL:              server.URL + "/agent",
		SignatureURL:     server.URL + "/agent.sig",
		Ed25519PublicKey: base64.StdEncoding.EncodeToString(publicKey),
	}, target)
	if err != nil {
		t.Fatalf("ApplyOptions() error = %v", err)
	}
	if !result.SignatureVerified {
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

func TestApplyRejectsBadEd25519Signature(t *testing.T) {
	payload := []byte("signed binary")
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	_, otherPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	badSignature := ed25519.Sign(otherPrivateKey, payload)
	mux := http.NewServeMux()
	mux.HandleFunc("/agent", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	})
	mux.HandleFunc("/agent.sig", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(base64.StdEncoding.EncodeToString(badSignature)))
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	target := filepath.Join(t.TempDir(), "agent")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := New().ApplyOptions(context.Background(), Options{
		URL:              server.URL + "/agent",
		SignatureURL:     server.URL + "/agent.sig",
		Ed25519PublicKey: base64.StdEncoding.EncodeToString(publicKey),
	}, target); err == nil {
		t.Fatal("ApplyOptions() expected signature error")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "old" {
		t.Fatalf("target should remain old, got %q", got)
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
