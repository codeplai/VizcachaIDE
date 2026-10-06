package app

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ErrIntoItself means a folder cannot be copied or moved into itself or one of its subfolders.
var ErrIntoItself = errors.New("a folder cannot go inside itself")

// sameOrInside reports whether path is folder or lives under it (case-insensitive off Linux).
func sameOrInside(path, folder string) bool {
	path, folder = filepath.Clean(path), filepath.Clean(folder)
	if runtime.GOOS != "linux" {
		path, folder = strings.ToLower(path), strings.ToLower(folder)
	}
	return path == folder || strings.HasPrefix(path, folder+string(filepath.Separator))
}

// Copy duplicates a file or a folder (with everything inside it) at a new path. It never
// overwrites: it fails if the target exists, and it refuses a target inside the source.
// A copy that fails halfway is removed, so no half-copied folder is left behind.
func (o *FileOperations) Copy(from, to string) error {
	if err := ValidateEntryName(filepath.Base(to)); err != nil {
		return err
	}
	if sameOrInside(to, from) {
		return fmt.Errorf("%w: %s", ErrIntoItself, from)
	}
	if exists(to) {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, to)
	}
	if err := copyEntry(from, to); err != nil {
		_ = os.RemoveAll(to)
		return fmt.Errorf("copy %s: %w", from, err)
	}
	return nil
}

func copyEntry(from, to string) error {
	info, err := os.Lstat(from)
	if err != nil {
		return err
	}
	switch {
	case info.IsDir():
		return copyFolder(from, to, info.Mode().Perm())
	case info.Mode().IsRegular():
		return copyFile(from, to, info.Mode().Perm())
	default:
		return fmt.Errorf("%s: %w", from, fs.ErrInvalid) // links and devices are not copied
	}
}

func copyFolder(from, to string, mode fs.FileMode) error {
	if err := os.Mkdir(to, mode|0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(from)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := copyEntry(filepath.Join(from, entry.Name()), filepath.Join(to, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(from, to string, mode fs.FileMode) error {
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer func() { _ = source.Close() }()
	target, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode|0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(target, source)
	if closeErr := target.Close(); copyErr == nil {
		copyErr = closeErr
	}
	return copyErr
}
