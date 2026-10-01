package gopls

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const (
	initializeTimeout = 60 * time.Second
	notifyTimeout     = 2 * time.Second
)

type serverState int

const (
	stateIdle serverState = iota
	stateStarting
	stateReady
	stateUnavailable
)

// Config says where gopls is and how to run it. The zero value searches for gopls.
type Config struct {
	// Executable returns the path of gopls each time the server starts. Nil or an empty
	// result means "search for it".
	Executable func() string
	// Environment returns the variables of the Go toolchain (may be nil).
	Environment func() map[string]string
}

// Server implements app.LanguageServer with a lazily started "gopls serve".
type Server struct {
	sink app.EventSink
	cfg  Config
	docs *openDocuments
	busy atomic.Bool // the last query timed out: the next ones wait less

	mu      sync.Mutex
	state   serverState
	conn    *connection
	folders map[string]bool
	ready   chan struct{} // closed when the server is ready or unavailable
}

var _ app.LanguageServer = (*Server)(nil)

// New creates the server. gopls is not started until the first document is opened.
func New(sink app.EventSink, cfg Config) *Server {
	return &Server{sink: sink, cfg: cfg, docs: newOpenDocuments(), folders: map[string]bool{}}
}

func (s *Server) configuredExecutable() string {
	if s.cfg.Executable == nil {
		return ""
	}
	return s.cfg.Executable()
}

func (s *Server) environment()map[string]string {
	if s.cfg.Environment == nil {
		return nil
	}
	return s.cfg.Environment()
}

// OpenDocument registers the file and starts gopls on the first call.
func (s *Server) OpenDocument(_ context.Context, path, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc := s.docs.open(path, text)
	s.ensureStartedLocked(path)
	if s.state == stateReady {
		s.openLocked(doc)
	}
	return nil
}

// ChangeDocument sends the complete new text with the next version.
func (s *Server) ChangeDocument(_ context.Context, path, text string, _ int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, open := s.docs.change(path, text)
	if open && s.state == stateReady {
		s.notifyLocked("textDocument/didChange", didChangeParams(doc))
	}
	return nil
}

// CloseDocument tells gopls the file is no longer open.
func (s *Server) CloseDocument(_ context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.docs.close(path) && s.state == stateReady {
		s.notifyLocked("textDocument/didClose", didCloseParams(path))
	}
	return nil
}

// Shutdown stops gopls (shutdown + exit) and forgets the documents.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	conn := s.conn
	if s.state == stateStarting {
		close(s.ready)
	}
	s.conn, s.state = nil, stateIdle
	s.docs.reset()
	s.folders = map[string]bool{}
	s.mu.Unlock()
	if conn != nil {
		conn.close(ctx)
	}
	return nil
}
