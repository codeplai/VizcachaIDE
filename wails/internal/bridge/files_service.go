package bridge

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ContextSource gives the Wails application context, available after startup.
// WailsEventSink implements it.
type ContextSource interface {
	Context() context.Context
}

// FilesService reads and writes the user's files and remembers the last folder.
type FilesService struct {
	context  ContextSource
	settings app.SettingsStore
}

// NewFilesService creates the service. settings may be nil (the last folder is then not remembered).
func NewFilesService(source ContextSource, settings app.SettingsStore) *FilesService {
	return &FilesService{context: source, settings: settings}
}

// OpenFolder lets the user pick a folder and returns its tree. When the user
// cancels the dialog it returns the tree of the last folder (or an empty node).
func (s *FilesService) OpenFolder() (domain.FileNode, error) {
	ctx := s.context.Context()
	if ctx == nil {
		return s.ListTree("")
	}
	chosen, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		DefaultDirectory: s.lastFolder(),
	})
	if err != nil {
		return domain.FileNode{}, fmt.Errorf("choose folder: %w", err)
	}
	return s.ListTree(chosen)
}

// ListTree returns the tree of a folder and remembers it as the last one.
// An empty root means "the last folder"; with none, the result is an empty node.
func (s *FilesService) ListTree(root string) (domain.FileNode, error) {
	if root == "" {
		root = s.lastFolder()
	}
	if root == "" {
		return domain.FileNode{Children: []domain.FileNode{}}, nil
	}
	tree, err := app.BuildFileTree(root)
	if err != nil {
		return domain.FileNode{}, err
	}
	s.rememberFolder(root)
	return tree, nil
}

// ReadFile returns the text of a file.
func (s *FilesService) ReadFile(path string) (string, error) { return app.ReadSourceFile(path) }

// SaveFile writes the text of a file as UTF-8.
func (s *FilesService) SaveFile(path, text string) error { return app.WriteSourceFile(path, text) }

func (s *FilesService) lastFolder() string {
	if s.settings == nil {
		return ""
	}
	current, err := s.settings.Load()
	if err != nil {
		return ""
	}
	return current.LastFolder
}

func (s *FilesService) rememberFolder(root string) {
	if s.settings == nil {
		return
	}
	current, err := s.settings.Load()
	if err != nil || current.LastFolder == root {
		return
	}
	current.LastFolder = root
	_ = s.settings.Save(current) // remembering is a convenience: a failure must not block opening
}
