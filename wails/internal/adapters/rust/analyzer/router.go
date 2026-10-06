package analyzer

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

// Router is the Rust language server of the IDE: one rust-analyzer per Cargo workspace and one
// for the loose .rs files. rust-analyzer started on a loose file does not load a crate opened
// later, nor the other way round (found in the M3 QA), so each kind of root gets its own
// process. Each one stops after the idle timeout when its files are closed.
type Router struct {
	sink    app.EventSink
	cfg     Config
	options lsp.Options
	mu      sync.Mutex
	servers map[string]*lsp.Server // by workspace folder; "" holds the loose files
}

var _ app.LanguageServer = (*Router)(nil)

// NewRouter creates the router; no rust-analyzer starts until a file is opened.
func NewRouter(sink app.EventSink, cfg Config, options lsp.Options) *Router {
	return &Router{sink: sink, cfg: cfg, options: options, servers: map[string]*lsp.Server{}}
}

// serverFor is the rust-analyzer of the file's workspace, created on first use.
func (r *Router) serverFor(path string) *lsp.Server {
	key := ""
	if project, found, err := rust.FindProject(path); err == nil && found {
		key = strings.ToLower(filepath.Clean(project.Workspace))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	server, ok := r.servers[key]
	if !ok {
		server = New(r.sink, r.cfg, r.options)
		r.servers[key] = server
	}
	return server
}

// OpenDocument implements app.LanguageServer.
func (r *Router) OpenDocument(ctx context.Context, path, text string) error {
	return r.serverFor(path).OpenDocument(ctx, path, text)
}

// ChangeDocument implements app.LanguageServer.
func (r *Router) ChangeDocument(ctx context.Context, path, text string, version int) error {
	return r.serverFor(path).ChangeDocument(ctx, path, text, version)
}

// CloseDocument implements app.LanguageServer.
func (r *Router) CloseDocument(ctx context.Context, path string) error {
	return r.serverFor(path).CloseDocument(ctx, path)
}

// Completion implements app.LanguageServer.
func (r *Router) Completion(ctx context.Context, at domain.SourceLocation) ([]domain.CompletionItem, error) {
	return r.serverFor(at.File).Completion(ctx, at)
}

// Hover implements app.LanguageServer.
func (r *Router) Hover(ctx context.Context, at domain.SourceLocation) (string, error) {
	return r.serverFor(at.File).Hover(ctx, at)
}

// Definition implements app.LanguageServer.
func (r *Router) Definition(ctx context.Context, at domain.SourceLocation) (*domain.SourceLocation, error) {
	return r.serverFor(at.File).Definition(ctx, at)
}

// SignatureHelp implements app.LanguageServer.
func (r *Router) SignatureHelp(ctx context.Context, at domain.SourceLocation) (*domain.SignatureHelp, error) {
	return r.serverFor(at.File).SignatureHelp(ctx, at)
}

// DocumentHighlights implements app.LanguageServer.
func (r *Router) DocumentHighlights(ctx context.Context, at domain.SourceLocation) ([]domain.SourceRange, error) {
	return r.serverFor(at.File).DocumentHighlights(ctx, at)
}

// DocumentSymbols implements app.LanguageServer.
func (r *Router) DocumentSymbols(ctx context.Context, path string) ([]domain.DocumentSymbol, error) {
	return r.serverFor(path).DocumentSymbols(ctx, path)
}

// InlayHints implements app.LanguageServer.
func (r *Router) InlayHints(ctx context.Context, visible domain.SourceRange) ([]domain.InlayHint, error) {
	return r.serverFor(visible.Start.File).InlayHints(ctx, visible)
}

// PrepareRename implements app.LanguageServer.
func (r *Router) PrepareRename(ctx context.Context, at domain.SourceLocation) (domain.RenameTarget, error) {
	return r.serverFor(at.File).PrepareRename(ctx, at)
}

// Rename implements app.LanguageServer.
func (r *Router) Rename(ctx context.Context, at domain.SourceLocation, newName string) (domain.RenameResult, error) {
	return r.serverFor(at.File).Rename(ctx, at, newName)
}

// References implements app.LanguageServer.
func (r *Router) References(ctx context.Context, at domain.SourceLocation) ([]domain.Reference, error) {
	return r.serverFor(at.File).References(ctx, at)
}

// Shutdown stops every rust-analyzer.
func (r *Router) Shutdown(ctx context.Context) error {
	r.mu.Lock()
	servers := make([]*lsp.Server, 0, len(r.servers))
	for _, server := range r.servers {
		servers = append(servers, server)
	}
	r.mu.Unlock()
	var all error
	for _, server := range servers {
		all = errors.Join(all, server.Shutdown(ctx))
	}
	return all
}
