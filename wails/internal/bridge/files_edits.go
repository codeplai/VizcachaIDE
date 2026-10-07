package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ApplyTextEdits edits files on disk: the files a rename touches that the editor has not open.
// The watcher is told about each new text, so none of them looks changed from outside.
func (s *FilesService) ApplyTextEdits(files []domain.FileEdit) (domain.EditSummary, error) {
	texts, err := app.ApplyEditsToFiles(files)
	if err != nil {
		return domain.EditSummary{}, err
	}
	summary := domain.EditSummary{}
	for _, file := range files {
		summary.Files++
		summary.Edits += len(file.Edits)
		if s.watcher != nil {
			s.watcher.Remember(file.File, texts[file.File])
		}
	}
	return summary, nil
}
