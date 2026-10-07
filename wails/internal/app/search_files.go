package app

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	maxSearchFileBytes = 2 << 20
	binarySniffBytes   = 8000
)

// searchSkippedFolders are never searched or replaced into: generated output, dependencies
// and version control data.
var searchSkippedFolders = map[string]bool{
	".git": true, "node_modules": true, "build": true, "target": true, "dist": true,
	"__pycache__": true, ".venv": true, "venv": true, ".idea": true, ".vs": true,
}

var nonTextExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true, ".webp": true,
	".pdf": true, ".zip": true, ".gz": true, ".tar": true, ".7z": true, ".woff": true,
	".woff2": true, ".ttf": true, ".pyc": true, ".class": true, ".jar": true, ".lib": true,
}

// textFile is a file whose text can be searched and rewritten.
type textFile struct {
	path string
	text string
	bom  bool
}

func isNonTextName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return binaryExtensions[ext] || nonTextExtensions[ext] || strings.HasPrefix(name, "__debug_bin")
}

// readTextFile reads a file that is small, valid UTF-8 and not binary; ok is false otherwise.
func readTextFile(path string) (file textFile, ok bool) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxSearchFileBytes || isNonTextName(info.Name()) {
		return textFile{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return textFile{}, false
	}
	if bytes.IndexByte(data[:min(len(data), binarySniffBytes)], 0) >= 0 || !utf8.Valid(data) {
		return textFile{}, false
	}
	text := string(data)
	trimmed := strings.TrimPrefix(text, utf8BOM)
	return textFile{path: path, text: trimmed, bom: trimmed != text}, true
}

// walkTextFiles calls visit for every text file under root, skipping the folders above.
// visit returns false to stop.
func walkTextFiles(root string, visit func(textFile) bool) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil // an unreadable file or folder is skipped, not fatal
		}
		if entry.IsDir() {
			if path != root && searchSkippedFolders[entry.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if file, ok := readTextFile(path); ok && !visit(file) {
			return fs.SkipAll
		}
		return nil
	})
}

// insideFolder reports whether path is a file under root (symbolic links resolved) that the
// search itself would have visited: replace never touches anything else.
func insideFolder(root, path string) bool {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(realRoot, realPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	for _, part := range strings.Split(filepath.Dir(relative), string(filepath.Separator)) {
		if searchSkippedFolders[part] {
			return false
		}
	}
	return true
}
