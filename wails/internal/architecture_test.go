package internal_test

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const maxFileLines = 200

var bannedNames = []string{"utils", "helpers", "common", "misc"}

var ignoredDirs = map[string]bool{"node_modules": true, "wailsjs": true, "dist": true, "build": true, "locales": true}

// walkProject visits every file of the wails/ folder except generated and vendored ones.
func walkProject(t *testing.T, visit func(path string, entry fs.DirEntry)) {
	t.Helper()
	root := ".."
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (ignoredDirs[entry.Name()] || strings.HasPrefix(entry.Name(), ".") && path != root) {
			return filepath.SkipDir
		}
		visit(path, entry)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Domain names only: PLAN_WAILS.md section 2.4 bans utils, helpers, common and misc.
func TestNoGenericNames(t *testing.T) {
	walkProject(t, func(path string, entry fs.DirEntry) {
		stem := strings.ToLower(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		for _, banned := range bannedNames {
			if stem == banned {
				t.Errorf("%s: %q is not a domain name", path, banned)
			}
		}
	})
}

// Files stay under 200 lines (generated and data files excluded).
func TestSourceFilesAreShort(t *testing.T) {
	walkProject(t, func(path string, entry fs.DirEntry) {
		ext := filepath.Ext(entry.Name())
		if entry.IsDir() || (ext != ".go" && ext != ".ts" && ext != ".svelte") {
			return
		}
		if lines := countLines(t, path); lines >= maxFileLines {
			t.Errorf("%s has %d lines, the limit is %d", path, lines, maxFileLines)
		}
	})
}

func countLines(t *testing.T, path string) int {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	count := 0
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		count++
	}
	return count
}
