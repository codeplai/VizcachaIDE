package updates

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// errChecksum means the downloaded file is not the one the release published.
var errChecksum = errors.New("the downloaded file does not match its published checksum")

// progressFunc is told how many bytes arrived so far and the expected total (0 when unknown).
type progressFunc func(done, total int64)

// download saves url to path (through a .part file, renamed at the end) and returns its SHA-256.
func download(ctx context.Context, client *http.Client, url, path string, progress progressFunc) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("downloading the update: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading the update: %s", response.Status)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	partial := path + ".part"
	file, err := os.Create(partial)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	counter := &countingWriter{total: response.ContentLength, progress: progress}
	_, copyErr := io.Copy(io.MultiWriter(file, hash, counter), response.Body)
	closeErr := file.Close()
	if err := errors.Join(copyErr, closeErr); err != nil {
		_ = os.Remove(partial)
		return "", fmt.Errorf("downloading the update: %w", err)
	}
	if err := os.Rename(partial, path); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// fetchText downloads a small text file (the checksum list).
func fetchText(ctx context.Context, client *http.Client, url string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading %s: %s", url, response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	return string(body), err
}

// expectedSum finds name in a "sha256  name" list (the format of sha256sum and build_release.py).
func expectedSum(list, name string) (string, bool) {
	scanner := bufio.NewScanner(strings.NewReader(list))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			return strings.ToLower(fields[0]), true
		}
	}
	return "", false
}

// countingWriter reports the progress of a download.
type countingWriter struct {
	done, total int64
	progress    progressFunc
}

func (c *countingWriter) Write(p []byte) (int, error) {
	c.done += int64(len(p))
	if c.progress != nil {
		c.progress(c.done, c.total)
	}
	return len(p), nil
}
