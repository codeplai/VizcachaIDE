package bridge

import "github.com/codeplai/VizcachaIDE/wails/internal/app"

// TerminalService is the integrated terminal of the bottom panel: real shells in a
// pseudoterminal, several at once, independent of the program the IDE runs with F5.
// Events: terminal:output {id, data} and terminal:exit {id, exitCode}.
type TerminalService struct {
	host app.TerminalHost
}

// NewTerminalService creates the service.
func NewTerminalService(host app.TerminalHost) *TerminalService {
	return &TerminalService{host: host}
}

// Start opens a shell in dir (the open folder; the home folder when dir is empty or missing)
// and returns the id of the session.
func (s *TerminalService) Start(dir string, cols, rows int) (string, error) {
	return s.host.Start(dir, cols, rows)
}

// Write sends what the user typed (or pasted) to the shell, unchanged.
func (s *TerminalService) Write(id, data string) error { return s.host.Write(id, data) }

// Resize tells the shell the size of the terminal in characters.
func (s *TerminalService) Resize(id string, cols, rows int) error {
	return s.host.Resize(id, cols, rows)
}

// Close ends the shell and everything it started.
func (s *TerminalService) Close(id string) error { return s.host.Close(id) }
