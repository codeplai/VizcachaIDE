package bridge

import (
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// ConsoleService is the interactive console of the bottom panel, one session per language.
// Both methods are synchronous; a snippet that fails is reported inside the result. A language
// without a console answers app.ErrUnsupported.
type ConsoleService struct {
	supportRouter
}

// NewConsoleService creates the service.
func NewConsoleService(registry *app.LanguageRegistry) *ConsoleService {
	return &ConsoleService{supportRouter{registry: registry}}
}

// Eval runs a snippet in the console session of a language and returns its value, output and error.
func (s *ConsoleService) Eval(codeLanguage domain.CodeLanguage, code string) (domain.ConsoleResult, error) {
	console, err := s.consoleOf(codeLanguage)
	if err != nil {
		return domain.ConsoleResult{}, err
	}
	return console.Eval(code), nil
}

// Reset forgets every variable and function of the session of a language.
func (s *ConsoleService) Reset(codeLanguage domain.CodeLanguage) error {
	console, err := s.consoleOf(codeLanguage)
	if err != nil {
		return err
	}
	console.Reset()
	return nil
}

func (s *ConsoleService) consoleOf(codeLanguage domain.CodeLanguage) (app.Console, error) {
	support, err := s.supportOf(codeLanguage)
	if err != nil {
		return nil, fmt.Errorf("console: %w", err)
	}
	if support.Console == nil {
		return nil, fmt.Errorf("console %s: %w", codeLanguage, app.ErrUnsupported)
	}
	return support.Console, nil
}
