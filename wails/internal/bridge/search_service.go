package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// SearchService searches and replaces text in the files of the open folder.
type SearchService struct {
	watcher app.FileWatcher
}

// NewSearchService creates the service. watcher may be nil; when given, it is told what
// replace wrote so those writes are not reported as changes made outside the IDE.
func NewSearchService(watcher app.FileWatcher) *SearchService {
	return &SearchService{watcher: watcher}
}

// Search finds the occurrences of query in the text files under root.
func (s *SearchService) Search(root, query string, options domain.SearchOptions) (domain.SearchResult, error) {
	return app.SearchFolder(root, query, options)
}

// Replace rewrites the listed files on disk. line and column (both 1-based) limit it to one
// occurrence; line 0 replaces every occurrence in each file.
func (s *SearchService) Replace(root, query string, options domain.SearchOptions, replacement string, paths []string, line, column int) (domain.ReplaceResult, error) {
	var remember func(path, text string)
	if s.watcher != nil {
		remember = s.watcher.Remember
	}
	return app.ReplaceInFolder(app.ReplaceRequest{
		Root: root, Query: query, Options: options, Replacement: replacement,
		Paths: paths, Line: line, Column: column,
	}, remember)
}
