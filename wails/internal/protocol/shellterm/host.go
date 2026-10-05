package shellterm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// ErrUnknownSession means no terminal has that id (it ended or was closed).
var ErrUnknownSession = errors.New("unknown terminal")

// Options configures a Host. Only Sink is required.
type Options struct {
	Sink app.TerminalSink
	// Folders returns the folders to put first in PATH (default: none).
	Folders func(ctx context.Context) []string
	// Base is the environment the shells start from (default: os.Environ()).
	Base []string
	// Shell is the shell to start (default: ChooseShell for this machine).
	Shell func() Command
}

// Host implements app.TerminalHost.
type Host struct {
	options Options
	mu      sync.Mutex
	next    int
	live    map[string]*session
}

var _ app.TerminalHost = (*Host)(nil)

// NewHost creates a host with no sessions.
func NewHost(options Options) *Host {
	if options.Base == nil {
		options.Base = os.Environ()
	}
	if options.Shell == nil {
		options.Shell = func() Command { return ChooseShell(runtime.GOOS, exec.LookPath, os.Getenv) }
	}
	return &Host{options: options, live: map[string]*session{}}
}

// Start implements app.TerminalHost.
func (h *Host) Start(dir string, cols, rows int) (string, error) {
	var folders []string
	if h.options.Folders != nil {
		folders = h.options.Folders(context.Background())
	}
	shell := h.options.Shell()
	h.mu.Lock()
	h.next++
	id := "t" + strconv.Itoa(h.next)
	h.mu.Unlock()
	s, err := startSession(id, pty.Program{
		Command: shell.Path, Args: shell.Args, Dir: startDirectory(dir),
		Env:         Environment(h.options.Base, folders),
		Interactive: true, Columns: cols, Rows: rows,
	}, h.options.Sink)
	if err != nil {
		return "", err
	}
	h.mu.Lock()
	h.live[id] = s
	h.mu.Unlock()
	go s.watch(func() { h.forget(id) })
	return id, nil
}

// startDirectory is dir when it is a folder, else the user's home.
func startDirectory(dir string) string {
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

func (h *Host) forget(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.live, id)
}

func (h *Host) find(id string) (*session, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.live[id]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownSession, id)
	}
	return s, nil
}

// Write implements app.TerminalHost: keyboard input, exactly as typed.
func (h *Host) Write(id, data string) error {
	s, err := h.find(id)
	if err != nil {
		return err
	}
	if _, err := s.terminal.Write([]byte(data)); err != nil {
		return fmt.Errorf("write to the terminal: %w", err)
	}
	return nil
}

// Resize implements app.TerminalHost.
func (h *Host) Resize(id string, cols, rows int) error {
	s, err := h.find(id)
	if err != nil {
		return err
	}
	return s.terminal.Resize(max(cols, 1), max(rows, 1))
}

// Close implements app.TerminalHost.
func (h *Host) Close(id string) error {
	s, err := h.find(id)
	if err != nil {
		return nil // already ended
	}
	s.close()
	return nil
}

// CloseAll implements app.TerminalHost.
func (h *Host) CloseAll() {
	h.mu.Lock()
	sessions := make([]*session, 0, len(h.live))
	for _, s := range h.live {
		sessions = append(sessions, s)
	}
	h.mu.Unlock()
	for _, s := range sessions {
		s.close()
	}
}
