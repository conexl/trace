package updater

import (
	"context"
	"crypto/sha256"
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

type Result struct {
	Target  string `json:"target"`
	Bytes   int64  `json:"bytes"`
	SHA256  string `json:"sha256"`
	Updated bool   `json:"updated"`
}

func New() *Updater {
	return &Updater{client: &http.Client{Timeout: 30 * time.Second}}
}

func (u *Updater) Apply(ctx context.Context, url string, expectedSHA256 string, target string) (Result, error) {
	if strings.TrimSpace(url) == "" {
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

	bytesWritten, digest, err := u.download(ctx, url, tmp)
	closeErr := tmp.Close()
	if err != nil {
		return Result{}, err
	}
	if closeErr != nil {
		return Result{}, closeErr
	}
	actual := hex.EncodeToString(digest)
	if expectedSHA256 != "" && !strings.EqualFold(expectedSHA256, actual) {
		return Result{}, fmt.Errorf("sha256 mismatch: expected %s, got %s", expectedSHA256, actual)
	}
	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return Result{}, err
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return Result{}, err
	}
	return Result{Target: target, Bytes: bytesWritten, SHA256: actual, Updated: true}, nil
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
