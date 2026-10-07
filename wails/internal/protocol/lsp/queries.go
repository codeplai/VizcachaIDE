package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	requestTimeout = 1500 * time.Millisecond
	busyTimeout    = 150 * time.Millisecond // after a timeout the server is probably still loading
)

// request sends a query and returns its raw result. ok is false when the server is missing,
// not ready in time, failed or timed out: callers then answer with an empty result.
func (s *Server) request(ctx context.Context, method string, params any) (raw json.RawMessage, ok bool) {
	return s.requestIf(ctx, method, params, nil)
}

// requestIf is request that, once the server is ready, does not ask when needs is false: the
// capability the method depends on was not announced in the initialize result.
func (s *Server) requestIf(ctx context.Context, method string, params any, needs *atomic.Bool) (raw json.RawMessage, ok bool) {
	limit := requestTimeout
	if s.busy.Load() {
		limit = busyTimeout
	}
	raw, err := s.exchange(ctx, limit, method, params, needs)
	return raw, err == nil
}

var (
	// errNotAnswered means the server is missing, not ready in time or timed out.
	errNotAnswered = errors.New("the language server did not answer")
	// errNotAnnounced means the server did not announce the capability of the method.
	errNotAnnounced = errors.New("the language server does not support this")
)

// exchange sends a query within limit. Besides the transport errors above, it returns the error
// the server answered with (a *jsonrpc2.Error).
func (s *Server) exchange(ctx context.Context, limit time.Duration, method string, params any, needs *atomic.Bool) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	conn, ready, usable := s.snapshot()
	if !usable {
		return nil, errNotAnswered
	}
	select {
	case <-ready:
	case <-ctx.Done():
		s.busy.Store(true)
		return nil, errNotAnswered
	}
	conn, usable = s.readyConnection(conn)
	if !usable {
		return nil, errNotAnswered
	}
	if needs != nil && !needs.Load() {
		return nil, errNotAnnounced
	}
	s.noticeManifests(ctx, conn)
	s.noticeSources(ctx, conn)
	result, err := conn.call(ctx, method, params)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.busy.Store(true)
			return nil, errNotAnswered
		}
		return nil, err
	}
	s.busy.Store(false)
	return result, nil
}

// awaitReady waits until the server is ready (its capabilities are then known); false when it
// is missing, unavailable or did not get ready within limit.
func (s *Server) awaitReady(ctx context.Context, limit time.Duration) bool {
	ctx, cancel := context.WithTimeout(ctx, limit)
	defer cancel()
	conn, ready, usable := s.snapshot()
	if !usable {
		return false
	}
	select {
	case <-ready:
	case <-ctx.Done():
		return false
	}
	_, usable = s.readyConnection(conn)
	return usable
}

// snapshot returns the connection and the channel closed when the server is ready.
func (s *Server) snapshot() (*connection, <-chan struct{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state != stateStarting && s.state != stateReady {
		return nil, nil, false
	}
	return s.conn, s.ready, true
}

// readyConnection checks, after waiting, that the same connection is still ready.
func (s *Server) readyConnection(previous *connection) (*connection, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn, s.state == stateReady && s.conn == previous
}

// atPosition runs a position query on an open document.
func (s *Server) atPosition(ctx context.Context, method string, at domain.SourceLocation) (raw json.RawMessage, doc document, ok bool) {
	doc, open := s.docs.get(at.File)
	if !open {
		return nil, doc, false
	}
	raw, ok = s.request(ctx, method, positionParams(doc, at.Line, at.Column))
	return raw, doc, ok
}

// Completion returns code suggestions at a position.
func (s *Server) Completion(ctx context.Context, at domain.SourceLocation) ([]domain.CompletionItem, error) {
	raw, _, ok := s.atPosition(ctx, "textDocument/completion", at)
	if !ok {
		return []domain.CompletionItem{}, nil
	}
	return toCompletionItems(raw), nil
}

// Hover returns the documentation at a position, or "".
func (s *Server) Hover(ctx context.Context, at domain.SourceLocation) (string, error) {
	raw, _, ok := s.atPosition(ctx, "textDocument/hover", at)
	if !ok {
		return "", nil
	}
	return toHoverText(raw), nil
}

// Definition returns where the symbol is declared, or nil.
func (s *Server) Definition(ctx context.Context, at domain.SourceLocation) (*domain.SourceLocation, error) {
	raw, _, ok := s.atPosition(ctx, "textDocument/definition", at)
	if !ok {
		return nil, nil
	}
	return toDefinition(raw, s.docs.textOf), nil
}

// SignatureHelp returns the call tip at a position, or nil.
func (s *Server) SignatureHelp(ctx context.Context, at domain.SourceLocation) (*domain.SignatureHelp, error) {
	raw, _, ok := s.atPosition(ctx, "textDocument/signatureHelp", at)
	if !ok {
		return nil, nil
	}
	return toSignatureHelp(raw), nil
}

// DocumentHighlights returns the occurrences of the symbol at a position.
func (s *Server) DocumentHighlights(ctx context.Context, at domain.SourceLocation) ([]domain.SourceRange, error) {
	raw, doc, ok := s.atPosition(ctx, "textDocument/documentHighlight", at)
	if !ok {
		return []domain.SourceRange{}, nil
	}
	return toHighlightRanges(raw, doc.path, doc.text), nil
}

// DocumentSymbols returns the nested declarations of an open document.
func (s *Server) DocumentSymbols(ctx context.Context, path string) ([]domain.DocumentSymbol, error) {
	doc, open := s.docs.get(path)
	if !open {
		return []domain.DocumentSymbol{}, nil
	}
	raw, ok := s.request(ctx, "textDocument/documentSymbol", documentParams(doc.path))
	if !ok {
		return []domain.DocumentSymbol{}, nil
	}
	return toDocumentSymbols(raw, doc.path, doc.text), nil
}
