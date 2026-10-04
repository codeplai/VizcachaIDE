package lsp

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// InlayHints returns the hints of the visible part of an open document. Contract of M3 R0: track
// R6 (docs/PLAN_RUST.md section 9.1) implements textDocument/inlayHint here; until then there are
// none.
func (s *Server) InlayHints(context.Context, domain.SourceRange) ([]domain.InlayHint, error) {
	return []domain.InlayHint{}, nil
}
