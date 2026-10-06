package bridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// UntitledFile returns where a file not saved yet lives while it is open: an absolute path in the
// system's temporary folder, one folder per IDE process and a subfolder per file
// (<temp>/VizcachaIDE/untitled/<pid>/main/main.go). The subfolder and an empty file are created
// (gopls reports "No packages found" for an open file that is not on disk); the text reaches the
// language server with the document.
//
// A relative "untitled/main.go" left gopls without a workspace (no completion, no problems on an
// unsaved file), and one folder for every unsaved file would make two Go files one package with
// two main functions.
func (s *LanguageService) UntitledFile(name string) (string, error) {
	if name == "" || filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("untitled file %q: not a plain file name", name)
	}
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	folder := filepath.Join(os.TempDir(), "VizcachaIDE", "untitled", strconv.Itoa(os.Getpid()), stem)
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", fmt.Errorf("create the folder of the unsaved file: %w", err)
	}
	path := filepath.Join(folder, name)
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		return "", fmt.Errorf("create the unsaved file: %w", err)
	}
	return path, nil
}
