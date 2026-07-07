package updater

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckOptionsDownloadsAndVerifies(t *testing.T) {
	payload := []byte("agent-binary-v2")
	expectedSHA := sha256.Sum256(payload)
	expectedSHAHex := hex.EncodeToString(expectedSHA[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	u := New()
	result, err := u.CheckOptions(context.Background(), Options{URL: server.URL, ExpectedSHA256: expectedSHAHex})
	if err != nil {
		t.Fatalf("CheckOptions() error = %v", err)
	}
	if result.Updated {
		t.Fatalf("CheckOptions should never set Updated=true")
	}
	if result.SHA256 != expectedSHAHex {
		t.Fatalf("sha256 mismatch: got %s want %s", result.SHA256, expectedSHAHex)
	}
	if result.Bytes != int64(len(payload)) {
		t.Fatalf("bytes mismatch: got %d want %d", result.Bytes, len(payload))
	}
}

func TestCheckOptionsRejectsSHA256Mismatch(t *testing.T) {
	payload := []byte("agent-binary-v2")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	u := New()
	_, err := u.CheckOptions(context.Background(), Options{URL: server.URL, ExpectedSHA256: "0000000000000000000000000000000000000000000000000000000000000000"})
	if err == nil {
		t.Fatalf("expected sha256 mismatch error")
	}
}

func TestCheckOptionsRejectsEmptyURL(t *testing.T) {
	u := New()
	_, err := u.CheckOptions(context.Background(), Options{})
	if err == nil {
		t.Fatalf("expected empty url error")
	}
}

func TestCheckOptionsVerifiesSignature(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	payload := []byte("signed-agent-binary")
	signature := ed25519.Sign(privateKey, payload)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bin":
			if _, err := w.Write(payload); err != nil {
				t.Fatalf("write binary: %v", err)
			}
		case "/sig":
			if _, err := io.WriteString(w, base64.StdEncoding.EncodeToString(signature)); err != nil {
				t.Fatalf("write signature: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	u := New()
	result, err := u.CheckOptions(context.Background(), Options{
		URL:              server.URL + "/bin",
		SignatureURL:     server.URL + "/sig",
		Ed25519PublicKey: base64.StdEncoding.EncodeToString(publicKey),
	})
	if err != nil {
		t.Fatalf("CheckOptions() error = %v", err)
	}
	if !result.SignatureVerified {
		t.Fatalf("expected signature to be verified")
	}
}

func TestCheckOptionsRejectsBadSignature(t *testing.T) {
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	payload := []byte("signed-agent-binary")
	publicKey := privateKey.Public().(ed25519.PublicKey)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bin":
			if _, err := w.Write(payload); err != nil {
				t.Fatalf("write binary: %v", err)
			}
		case "/sig":
			if _, err := io.WriteString(w, base64.StdEncoding.EncodeToString([]byte("invalid-signature"))); err != nil {
				t.Fatalf("write signature: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	u := New()
	_, err = u.CheckOptions(context.Background(), Options{
		URL:              server.URL + "/bin",
		SignatureURL:     server.URL + "/sig",
		Ed25519PublicKey: base64.StdEncoding.EncodeToString(publicKey),
	})
	if err == nil {
		t.Fatalf("expected signature verification error")
	}
}

func TestCheckOptionsCleansTempFile(t *testing.T) {
	payload := []byte("temp-binary")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	before, err := countUpdateCheckTempFiles()
	if err != nil {
		t.Fatalf("count temp files before: %v", err)
	}

	u := New()
	if _, err := u.CheckOptions(context.Background(), Options{URL: server.URL}); err != nil {
		t.Fatalf("CheckOptions() error = %v", err)
	}

	after, err := countUpdateCheckTempFiles()
	if err != nil {
		t.Fatalf("count temp files after: %v", err)
	}
	if after != before {
		t.Fatalf("temp file leak: before=%d after=%d", before, after)
	}
}

func countUpdateCheckTempFiles() (int, error) {
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), ".homelytics-update-check-*"))
	if err != nil {
		return 0, err
	}
	return len(matches), nil
}

func TestApplyOptionsUpdatesTarget(t *testing.T) {
	payload := []byte("agent-binary-v2")
	expectedSHA := sha256.Sum256(payload)
	expectedSHAHex := hex.EncodeToString(expectedSHA[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	target := t.TempDir() + "/agent"
	u := New()
	result, err := u.ApplyOptions(context.Background(), Options{URL: server.URL, ExpectedSHA256: expectedSHAHex}, target)
	if err != nil {
		t.Fatalf("ApplyOptions() error = %v", err)
	}
	if !result.Updated {
		t.Fatalf("expected Updated=true")
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(data) != string(payload) {
		t.Fatalf("target content mismatch")
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat target: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("target is not executable")
	}
}

func TestApplyOptionsRejectsSHA256Mismatch(t *testing.T) {
	payload := []byte("agent-binary-v2")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(payload); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	target := t.TempDir() + "/agent"
	u := New()
	_, err := u.ApplyOptions(context.Background(), Options{URL: server.URL, ExpectedSHA256: "0000000000000000000000000000000000000000000000000000000000000000"}, target)
	if err == nil {
		t.Fatalf("expected sha256 mismatch error")
	}
}

func TestCurrentExecutableSHA256(t *testing.T) {
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("executable: %v", err)
	}
	expected, err := hashFile(path)
	if err != nil {
		t.Fatalf("hash file: %v", err)
	}
	got, err := CurrentExecutableSHA256()
	if err != nil {
		t.Fatalf("CurrentExecutableSHA256() error = %v", err)
	}
	if got != expected {
		t.Fatalf("hash mismatch: got %s want %s", got, expected)
	}
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
