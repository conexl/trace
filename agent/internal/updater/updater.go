package updater

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Updater struct {
	client *http.Client
}

type Options struct {
	URL              string
	ExpectedSHA256   string
	SignatureURL     string
	Ed25519PublicKey string
}

type Result struct {
	Target            string `json:"target"`
	Bytes             int64  `json:"bytes"`
	SHA256            string `json:"sha256"`
	SignatureVerified bool   `json:"signature_verified"`
	Updated           bool   `json:"updated"`
}

func New() *Updater {
	return &Updater{client: &http.Client{Timeout: 30 * time.Second}}
}

func NewWithTLS(tlsConfig *tls.Config) *Updater {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}
	return &Updater{client: &http.Client{Timeout: 30 * time.Second, Transport: transport}}
}

func CurrentExecutableSHA256() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open executable: %w", err)
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", fmt.Errorf("hash executable: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (u *Updater) Apply(ctx context.Context, url string, expectedSHA256 string, target string) (Result, error) {
	return u.ApplyOptions(ctx, Options{URL: url, ExpectedSHA256: expectedSHA256}, target)
}

// CheckOptions downloads the update to a temporary file, verifies SHA-256 and
// optional Ed25519 signature, then discards the file. It never replaces the
// running binary. Use it for policy="check".
func (u *Updater) CheckOptions(ctx context.Context, opts Options) (Result, error) {
	if strings.TrimSpace(opts.URL) == "" {
		return Result{}, fmt.Errorf("update url is empty")
	}
	tmp, err := os.CreateTemp("", ".homelytics-update-check-*")
	if err != nil {
		return Result{}, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	bytesWritten, digest, err := u.download(ctx, opts.URL, tmp)
	closeErr := tmp.Close()
	if err != nil {
		return Result{}, err
	}
	if closeErr != nil {
		return Result{}, closeErr
	}
	actual := hex.EncodeToString(digest)
	if opts.ExpectedSHA256 != "" && !strings.EqualFold(opts.ExpectedSHA256, actual) {
		return Result{}, fmt.Errorf("sha256 mismatch: expected %s, got %s", opts.ExpectedSHA256, actual)
	}

	signatureVerified, err := u.verifySignature(ctx, opts, tmpPath)
	if err != nil {
		return Result{}, err
	}
	return Result{Bytes: bytesWritten, SHA256: actual, SignatureVerified: signatureVerified, Updated: false}, nil
}

func (u *Updater) ApplyOptions(ctx context.Context, opts Options, target string) (Result, error) {
	if strings.TrimSpace(opts.URL) == "" {
		return Result{}, fmt.Errorf("update url is empty")
	}
	if target == "" {
		current, err := os.Executable()
		if err != nil {
			return Result{}, fmt.Errorf("resolve current executable: %w", err)
		}
		target = current
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".homelytics-update-*")
	if err != nil {
		return Result{}, err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	bytesWritten, digest, err := u.download(ctx, opts.URL, tmp)
	closeErr := tmp.Close()
	if err != nil {
		return Result{}, err
	}
	if closeErr != nil {
		return Result{}, closeErr
	}
	actual := hex.EncodeToString(digest)
	if opts.ExpectedSHA256 != "" && !strings.EqualFold(opts.ExpectedSHA256, actual) {
		return Result{}, fmt.Errorf("sha256 mismatch: expected %s, got %s", opts.ExpectedSHA256, actual)
	}

	signatureVerified, err := u.verifySignature(ctx, opts, tmpPath)
	if err != nil {
		return Result{}, err
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return Result{}, err
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return Result{}, err
	}
	return Result{Target: target, Bytes: bytesWritten, SHA256: actual, SignatureVerified: signatureVerified, Updated: true}, nil
}

func (u *Updater) verifySignature(ctx context.Context, opts Options, artifactPath string) (bool, error) {
	if strings.TrimSpace(opts.Ed25519PublicKey) == "" && strings.TrimSpace(opts.SignatureURL) == "" {
		return false, nil
	}
	if strings.TrimSpace(opts.Ed25519PublicKey) == "" || strings.TrimSpace(opts.SignatureURL) == "" {
		return false, fmt.Errorf("ed25519_public_key and signature_url must be set together")
	}
	publicKey, err := decodeBase64("ed25519 public key", opts.Ed25519PublicKey, ed25519.PublicKeySize)
	if err != nil {
		return false, err
	}
	signatureText, err := u.fetchString(ctx, opts.SignatureURL, 4<<10)
	if err != nil {
		return false, fmt.Errorf("download signature: %w", err)
	}
	signature, err := decodeBase64("ed25519 signature", signatureText, ed25519.SignatureSize)
	if err != nil {
		return false, err
	}
	artifact, err := os.ReadFile(artifactPath)
	if err != nil {
		return false, err
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), artifact, signature) {
		return false, fmt.Errorf("ed25519 signature verification failed")
	}
	return true, nil
}

func (u *Updater) download(ctx context.Context, url string, dst io.Writer) (int64, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, nil, fmt.Errorf("download update failed: %s", resp.Status)
	}
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(dst, hash), resp.Body)
	if err != nil {
		return written, nil, err
	}
	return written, hash.Sum(nil), nil
}

func (u *Updater) fetchString(ctx context.Context, url string, maxBytes int64) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return strings.TrimSpace(string(data)), nil
}

func decodeBase64(name string, value string, expectedLen int) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(strings.TrimSpace(value))
	}
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	if len(decoded) != expectedLen {
		return nil, fmt.Errorf("%s length = %d, want %d", name, len(decoded), expectedLen)
	}
	return decoded, nil
}
