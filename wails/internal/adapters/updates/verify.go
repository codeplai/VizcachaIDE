package updates

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// fetchVerified returns the verified file of the offer, downloading it unless a good copy is
// already in the cache (from an earlier session).
func (u *Updater) fetchVerified(ctx context.Context, found *offer) (string, error) {
	sums, err := fetchText(ctx, u.options.Client, found.sumsURL)
	if err != nil {
		return "", err
	}
	want, ok := expectedSum(sums, found.release.Asset)
	if !ok {
		return "", errChecksum
	}
	path := filepath.Join(u.options.CacheDir, found.release.Asset)
	if got, err := sumOfFile(path); err == nil && got == want {
		return path, nil
	}
	got, err := download(ctx, u.options.Client, found.fileURL, path, u.progress)
	if err != nil {
		return "", err
	}
	if got != want {
		_ = os.Remove(path)
		return "", errChecksum
	}
	return path, nil
}

func sumOfFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// removeOthers deletes older downloads, keeping only the current one.
func removeOthers(dir, keep string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if path := filepath.Join(dir, entry.Name()); path != keep {
			_ = os.Remove(path)
		}
	}
}
