package bridge

import (
	"errors"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ErrFileNotFound is returned when a stub is asked for a file it does not know.
var ErrFileNotFound = errors.New("file not found")

// FilesService reads and writes the user's files. W0 STUB: owned by track G1.
type FilesService struct{}

// NewFilesService creates the service.
func NewFilesService() *FilesService { return &FilesService{} }

// OpenFolder lets the user pick a folder and returns its tree.
func (s *FilesService) OpenFolder() (domain.FileNode, error) { return sampleTree(), nil }

// ListTree returns the tree of a folder.
func (s *FilesService) ListTree(root string) (domain.FileNode, error) { return sampleTree(), nil }

// ReadFile returns the text of a file.
func (s *FilesService) ReadFile(path string) (string, error) {
	switch path {
	case sampleMain:
		return sampleMainSource, nil
	case sampleCalc:
		return sampleCalcSource, nil
	}
	return "", ErrFileNotFound
}

// SaveFile writes the text of a file.
func (s *FilesService) SaveFile(path, text string) error { return nil }
