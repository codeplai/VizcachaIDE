package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	maxTreeDepth = 12
	utf8BOM      = "\ufeff"
)

var binaryExtensions = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".a": true, ".o": true,
	".obj": true, ".test": true, ".out": true, ".bin": true, ".pdb": true,
}

// isHiddenFromTree is true for what a beginner should not see: version control
// data, compiled binaries and the debugger's temporary executables.
func isHiddenFromTree(name string) bool {
	if name == ".git" || strings.HasPrefix(name, "__debug_bin") {
		return true
	}
	return binaryExtensions[strings.ToLower(filepath.Ext(name))]
}

// BuildFileTree reads the folder tree under root: folders first, then files, both sorted by name.
func BuildFileTree(root string) (domain.FileNode, error) {
	info, err := os.Stat(root)
	if err != nil {
		return domain.FileNode{}, fmt.Errorf("open folder %s: %w", root, err)
	}
	if !info.IsDir() {
		return domain.FileNode{}, fmt.Errorf("open folder %s: not a folder", root)
	}
	return readNode(root, filepath.Base(root), 0), nil
}

func readNode(path, name string, depth int) domain.FileNode {
	node := domain.FileNode{Name: name, Path: path, IsDir: true, Children: []domain.FileNode{}}
	entries, err := os.ReadDir(path)
	if err != nil || depth >= maxTreeDepth {
		return node
	}
	for _, entry := range entries {
		if isHiddenFromTree(entry.Name()) {
			continue
		}
		child := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			node.Children = append(node.Children, readNode(child, entry.Name(), depth+1))
			continue
		}
		node.Children = append(node.Children,
			domain.FileNode{Name: entry.Name(), Path: child, Children: []domain.FileNode{}})
	}
	sortNodes(node.Children)
	return node
}

func sortNodes(nodes []domain.FileNode) {
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return strings.ToLower(nodes[i].Name) < strings.ToLower(nodes[j].Name)
	})
}

// ReadSourceFile returns the UTF-8 text of a file (a leading BOM is dropped, invalid bytes become U+FFFD).
func ReadSourceFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return strings.ToValidUTF8(strings.TrimPrefix(string(data), "\ufeff"), "�"), nil
}

// WriteSourceFile saves text as UTF-8, creating the folder if needed.
func WriteSourceFile(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create folder for %s: %w", path, err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}
