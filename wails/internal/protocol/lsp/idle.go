package lsp

import (
	"context"
)

// Timer is the part of *time.Timer that the idle shutdown uses.
type Timer interface{ Stop() bool }

// idleShutdown tracks the pending idle timer. generation tells a stale expiry (the timer
// fired just as a document was opened) from the current one.
type idleShutdown struct {
	timer      Timer
	generation int
}

// armIdleLocked starts the idle countdown when the last document has just been closed.
func (s *Server) armIdleLocked() {
	if s.opts.IdleTimeout <= 0 || !s.docs.empty() {
		return
	}
	s.cancelIdleLocked()
	generation := s.idle.generation
	s.idle.timer = s.opts.AfterFunc(s.opts.IdleTimeout, func() { s.idleExpired(generation) })
}

// cancelIdleLocked stops the countdown (a document was opened or the server shuts down).
func (s *Server) cancelIdleLocked() {
	s.idle.generation++
	if s.idle.timer != nil {
		s.idle.timer.Stop()
		s.idle.timer = nil
	}
}

// idleExpired shuts the running server down. A server that is missing or has died stays
// "unavailable": idleness never changes the status.
func (s *Server) idleExpired(generation int) {
	s.mu.Lock()
	running := s.state == stateStarting || s.state == stateReady
	if generation != s.idle.generation || !s.docs.empty() || !running {
		s.mu.Unlock()
		return
	}
	s.idle.timer = nil
	conn := s.stopLocked()
	s.mu.Unlock()
	if conn != nil {
		conn.close(context.Background())
	}
}
