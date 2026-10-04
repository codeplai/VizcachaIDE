package gopls

import (
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// recordingSink keeps the lsp events. The other EventSink methods come from the nil embed
// and must never be called by this adapter.
type recordingSink struct {
	app.EventSink
	mu          sync.Mutex
	statuses    []domain.ServerStatus
	diagnostics map[string][]domain.Diagnostic
}

func newRecordingSink() *recordingSink {
	return &recordingSink{diagnostics: map[string][]domain.Diagnostic{}}
}

func (r *recordingSink) LanguageServerStatus(status domain.ServerStatus) {
	r.mu.Lock()
	r.statuses = append(r.statuses, status)
	r.mu.Unlock()
}

func (r *recordingSink) Diagnostics(path string, diagnostics []domain.Diagnostic) {
	r.mu.Lock()
	r.diagnostics[path] = diagnostics
	r.mu.Unlock()
}

func (r *recordingSink) lastStatus() domain.ServerStatus {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.statuses) == 0 {
		return ""
	}
	return r.statuses[len(r.statuses)-1]
}

func (r *recordingSink) diagnosticsOf(path string) []domain.Diagnostic {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.diagnostics[path]
}
