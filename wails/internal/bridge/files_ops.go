package bridge

import "github.com/codeplai/VizcachaIDE/wails/internal/app"

// UseShell gives the service the system's Recycle Bin and file manager for the Files panel
// (create and rename need nothing from the system). It returns the service for chaining.
func (s *FilesService) UseShell(shell app.SystemShell) *FilesService {
	s.operations = app.NewFileOperations(shell)
	return s
}

// CreateFile creates a new file with the given text; it fails if the path already exists.
func (s *FilesService) CreateFile(path, text string) error {
	return s.fileOperations().CreateFile(path, text)
}

// CreateFolder creates a new folder; it fails if the path already exists.
func (s *FilesService) CreateFolder(path string) error {
	return s.fileOperations().CreateFolder(path)
}

// Rename renames or moves a file or folder; it fails if the target already exists.
func (s *FilesService) Rename(from, to string) error {
	return s.fileOperations().Rename(from, to)
}

// Copy duplicates a file or folder (recursively); it never overwrites and refuses a target inside the source.
func (s *FilesService) Copy(from, to string) error {
	return s.fileOperations().Copy(from, to)
}

// MoveToTrash sends a file or folder to the Recycle Bin. It never deletes permanently.
func (s *FilesService) MoveToTrash(path string) error {
	return s.fileOperations().MoveToTrash(path)
}

// RevealInExplorer shows the file selected in the system's file manager.
func (s *FilesService) RevealInExplorer(path string) error {
	return s.fileOperations().RevealInExplorer(path)
}

func (s *FilesService) fileOperations() *app.FileOperations {
	if s.operations == nil {
		return app.NewFileOperations(nil) // file operations work without a shell; trash/reveal fail clearly
	}
	return s.operations
}
