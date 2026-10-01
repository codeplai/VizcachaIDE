package gopls

import (
	"context"
	"encoding/json"

	"go.lsp.dev/protocol"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ensureStartedLocked launches gopls the first time. Failures leave the state unavailable.
func (s *Server) ensureStartedLocked(path string) {
	if s.state != stateIdle {
		return
	}
	env := s.environment()
	executable, err := locateGopls(s.configuredExecutable(), env)
	if err != nil {
		s.makeUnavailableLocked(nil)
		return
	}
	root := moduleRoot(path)
	conn, err := startConnection(executable, env, root, s.onNotification)
	if err != nil {
		s.makeUnavailableLocked(nil)
		return
	}
	s.conn, s.state, s.ready = conn, stateStarting, make(chan struct{})
	s.folders[pathKey(root)] = true
	s.sink.LanguageServerStatus(domain.ServerStarting)
	go s.initialize(conn, root)
	go s.watchExit(conn)
}

// initialize runs the LSP handshake and then opens the documents that are waiting.
func (s *Server) initialize(conn *connection, root string) {
	ctx, cancel := context.WithTimeout(context.Background(), initializeTimeout)
	defer cancel()
	_, err := conn.call(ctx, "initialize", initializeParams(root))
	if err == nil {
		err = conn.notify(ctx, "initialized", protocol.InitializedParams{})
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != conn || s.state != stateStarting {
		return
	}
	if err != nil {
		s.makeUnavailableLocked(conn)
		return
	}
	s.state = stateReady
	close(s.ready)
	for _, doc := range s.docs.all() {
		s.openLocked(doc)
	}
	s.sink.LanguageServerStatus(domain.ServerReady)
}

// watchExit marks the server unavailable if gopls dies on its own.
func (s *Server) watchExit(conn *connection) {
	<-conn.conn.Done()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == conn {
		s.makeUnavailableLocked(conn)
	}
}

func (s *Server) makeUnavailableLocked(conn *connection) {
	if s.state == stateUnavailable {
		return
	}
	if s.state == stateStarting {
		close(s.ready)
	}
	s.state, s.conn = stateUnavailable, nil
	s.sink.LanguageServerStatus(domain.ServerUnavailable)
	if conn != nil {
		go conn.close(context.Background())
	}
}

// openLocked sends didOpen (adding the module folder to the workspace first).
func (s *Server) openLocked(doc document) {
	root := moduleRoot(doc.path)
	if !s.folders[pathKey(root)] {
		s.folders[pathKey(root)] = true
		s.notifyLocked("workspace/didChangeWorkspaceFolders", addFolderParams(root))
	}
	s.notifyLocked("textDocument/didOpen", didOpenParams(doc))
}

func (s *Server) notifyLocked(method string, params any) {
	if s.conn == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()
	_ = s.conn.notify(ctx, method, params) // a dead gopls is noticed by watchExit
}

// onNotification publishes the diagnostics of open documents.
func (s *Server) onNotification(method string, raw json.RawMessage) {
	if method != "textDocument/publishDiagnostics" {
		return
	}
	var params protocol.PublishDiagnosticsParams
	if json.Unmarshal(raw, &params) != nil {
		return
	}
	path := uriToPath(params.URI)
	doc, open := s.docs.get(path)
	if !open {
		return
	}
	s.sink.Diagnostics(doc.path, toDiagnostics(params, doc.path, doc.text))
}
