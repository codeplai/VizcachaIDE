package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Errors of the file operations. The frontend validates names first and shows its own
// translated text; these are the backend's second line of defense.
var (
	// ErrInvalidName means the name is empty or has characters Windows does not allow.
	ErrInvalidName = errors.New("invalid file name")
	// ErrAlreadyExists means a file or folder with that name already exists.
	ErrAlreadyExists = errors.New("a file with that name already exists")
	// ErrTrashUnavailable means the Recycle Bin could not be used. Nothing was deleted.
	ErrTrashUnavailable = errors.New("the Recycle Bin is not available; nothing was deleted")
)

// SystemShell is what the operating system does for the Files panel: the Recycle Bin
// and "Show in Explorer". Implemented by adapters/filesystem.
type SystemShell interface {
	// MoveToTrash sends a file or folder to the Recycle Bin. It never deletes permanently:
	// when the bin is unavailable it returns an error wrapping ErrTrashUnavailable.
	MoveToTrash(path string) error
	// Reveal shows the file selected in the system's file manager.
	Reveal(path string) error
}

// FileOperations creates, renames and deletes the files of the Files panel.
type FileOperations struct{ shell SystemShell }

// NewFileOperations creates the use cases. shell may be nil (trash and reveal then fail).
func NewFileOperations(shell SystemShell) *FileOperations { return &FileOperations{shell: shell} }

// ValidateEntryName checks one file or folder name (not a path).
func ValidateEntryName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return fmt.Errorf("%w: empty", ErrInvalidName)
	}
	if strings.ContainsAny(name, "/\\:*?\"<>|") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("%w: %q", ErrInvalidName, name)
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// CreateFile writes a new file; it fails if the path already exists.
func (o *FileOperations) CreateFile(path, text string) error {
	if err := ValidateEntryName(filepath.Base(path)); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, path)
	}
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	_, writeErr := file.WriteString(text)
	if closeErr := file.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		return fmt.Errorf("write %s: %w", path, writeErr)
	}
	return nil
}

// CreateFolder makes a new folder; it fails if the path already exists.
func (o *FileOperations) CreateFolder(path string) error {
	if err := ValidateEntryName(filepath.Base(path)); err != nil {
		return err
	}
	if exists(path) {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, path)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		return fmt.Errorf("create folder %s: %w", path, err)
	}
	return nil
}

// Rename moves a file or folder; it fails if the target exists (a change of letter case is allowed).
func (o *FileOperations) Rename(from, to string) error {
	if err := ValidateEntryName(filepath.Base(to)); err != nil {
		return err
	}
	if !strings.EqualFold(from, to) && exists(to) {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, to)
	}
	if err := os.Rename(from, to); err != nil {
		return fmt.Errorf("rename %s: %w", from, err)
	}
	return nil
}

// MoveToTrash sends the path to the Recycle Bin.
func (o *FileOperations) MoveToTrash(path string) error {
	if o.shell == nil {
		return ErrTrashUnavailable
	}
	return o.shell.MoveToTrash(path)
}

// RevealInExplorer shows the path in the system's file manager.
func (o *FileOperations) RevealInExplorer(path string) error {
	if o.shell == nil {
		return fmt.Errorf("reveal %s: no file manager", path)
	}
	return o.shell.Reveal(path)
}
