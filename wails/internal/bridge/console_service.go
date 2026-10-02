package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ConsoleService is the interactive Go console of the bottom panel. Both methods are
// synchronous; a snippet that fails is reported inside the result.
type ConsoleService struct {
	console app.Console
}

// NewConsoleService creates the service.
func NewConsoleService(console app.Console) *ConsoleService {
	return &ConsoleService{console: console}
}

// Eval runs a snippet in the console session and returns its value, output and error.
func (s *ConsoleService) Eval(code string) domain.ConsoleResult {
	return s.console.Eval(code)
}

// Reset forgets every variable and function of the session.
func (s *ConsoleService) Reset() {
	s.console.Reset()
}
