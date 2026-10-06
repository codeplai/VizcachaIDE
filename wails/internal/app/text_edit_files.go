package app

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ApplyEditsToFiles edits files on disk (the ones the editor does not have open) and returns
// the new text of each by path. It is all or nothing for the reading and checking part: every
// file is read and edited in memory first, and only then written, so a bad edit changes nothing.
func ApplyEditsToFiles(files []domain.FileEdit) (map[string]string, error) {
	texts := make(map[string]string, len(files))
	for _, file := range files {
		current, err := ReadSourceFile(file.File)
		if err != nil {
			return nil, err
		}
		edited, err := domain.ApplyTextEdits(current, file.Edits)
		if err != nil {
			return nil, fmt.Errorf("edit %s: %w", file.File, err)
		}
		texts[file.File] = edited
	}
	for path, text := range texts {
		if err := WriteSourceFile(path, text); err != nil {
			return nil, err
		}
	}
	return texts, nil
}
