package lsp

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

// Flavor is everything that differs between language servers: how to find and start the
// executable, which folder is a project, and what to tell the server at the handshake.
type Flavor interface {
	// Command locates the executable and returns it with its arguments. A missing tool is
	// reported as app.MissingTool(id).
	Command(env map[string]string) (executable string, args []string, err error)
	// RootOf is the project folder of a file (the nearest go.mod, pyproject.toml...).
	RootOf(path string) string
	// InitializationOptions go in the initialize request; nil when there are none.
	InitializationOptions() any
	// Configuration is sent as workspace/didChangeConfiguration after initialized; nil when
	// there is none.
	Configuration() any
	// Environment returns the variables the server runs with (may be nil).
	Environment() map[string]string
}

// Options tunes a Server. The zero value is valid.
type Options struct {
	// Name is the Source of the diagnostics ("gopls", "pylsp"...).
	Name string
	// LanguageID is the languageId sent with didOpen ("go", "python"...).
	LanguageID string
	// IdleTimeout shuts the server down (shutdown + exit) when no document stays open for
	// this long. The next OpenDocument starts it again and the status is not "unavailable":
	// that one means the tool is missing. Zero disables it.
	IdleTimeout time.Duration
	// AfterFunc schedules f after d; nil uses time.AfterFunc. Tests inject a fake clock.
	AfterFunc func(d time.Duration, f func()) Timer
}

// Server implements app.LanguageServer with a lazily started language server process.
type Server struct {
	sink   app.EventSink
	flavor Flavor
	opts   Options
	docs   *openDocuments
	busy   atomic.Bool // the last query timed out: the next ones wait less

	mu      sync.Mutex
	state   serverState
	conn    *connection
	folders map[string]bool
	ready   chan struct{} // closed when the server is ready or unavailable
	idle    idleShutdown
}

var _ app.LanguageServer = (*Server)(nil)

// New creates the server. The process is not started until the first document is opened.
func New(sink app.EventSink, flavor Flavor, options Options) *Server {
	if options.AfterFunc == nil {
		options.AfterFunc = func(d time.Duration, f func()) Timer { return time.AfterFunc(d, f) }
	}
	return &Server{sink: sink, flavor: flavor, opts: options, docs: newOpenDocuments(), folders: map[string]bool{}}
}

// OpenDocument registers the file and starts the server on the first call.
func (s *Server) OpenDocument(_ context.Context, path, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelIdleLocked()
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

// CloseDocument tells the server the file is no longer open.
func (s *Server) CloseDocument(_ context.Context, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.docs.close(path) && s.state == stateReady {
		s.notifyLocked("textDocument/didClose", didCloseParams(path))
	}
	s.armIdleLocked()
	return nil
}

// Shutdown stops the server (shutdown + exit) and forgets the documents.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	s.cancelIdleLocked()
	conn := s.stopLocked()
	s.mu.Unlock()
	if conn != nil {
		conn.close(ctx)
	}
	return nil
}

// stopLocked returns the server to the idle state and hands back the connection to close
// (outside the lock).
func (s *Server) stopLocked() *connection {
	conn := s.conn
	if s.state == stateStarting {
		close(s.ready)
	}
	s.conn, s.state = nil, stateIdle
	s.docs.reset()
	s.folders = map[string]bool{}
	return conn
}
