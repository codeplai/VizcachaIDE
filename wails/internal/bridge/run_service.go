package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// RunService runs the user's program. The language comes from the extension of the file: its
// runner decides what to run (Configure) and emits run:started, run:output and run:finished.
// Every runner shares one process slot, so there is only one program running in the whole IDE.
type RunService struct {
	supportRouter

	mu     sync.Mutex
	active app.ProgramRunner // runner of the last program started
}

// NewRunService creates the service.
func NewRunService(registry *app.LanguageRegistry) *RunService {
	return &RunService{supportRouter: supportRouter{registry: registry}}
}

// Run compiles and runs one file and returns the configuration it used.
// The program outlives this call, so it runs on its own context.
func (s *RunService) Run(path string, programArgs []string) (domain.RunConfiguration, error) {
	support, err := s.supportFor(path)
	if err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("run %s: %w", path, err)
	}
	config := support.Runner.Configure(path, programArgs)
	if err := support.Runner.Run(context.Background(), config); err != nil {
		return config, fmt.Errorf("run %s: %w", path, err)
	}
	s.remember(support.Runner)
	return config, nil
}

// RunMember runs the member of a Cargo workspace the student chose after Run answered
// run.chooseMember. Languages without members answer app.ErrUnsupported.
func (s *RunService) RunMember(path, member string, programArgs []string) (domain.RunConfiguration, error) {
	support, err := s.supportFor(path)
	if err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("run %s: %w", member, err)
	}
	members, ok := support.Runner.(app.MemberRunner)
	if !ok {
		return domain.RunConfiguration{}, fmt.Errorf("run %s: %w", member, app.ErrUnsupported)
	}
	config, err := members.ConfigureMember(path, member, programArgs)
	if err != nil {
		return config, fmt.Errorf("run %s: %w", member, err)
	}
	if err := support.Runner.Run(context.Background(), config); err != nil {
		return config, fmt.Errorf("run %s: %w", member, err)
	}
	s.remember(support.Runner)
	return config, nil
}

// Build compiles one file without running it. Languages without a build step answer
// app.ErrUnsupported.
func (s *RunService) Build(path string, programArgs []string) (domain.RunConfiguration, error) {
	support, err := s.supportFor(path)
	if err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("build %s: %w", path, err)
	}
	config := support.Runner.Configure(path, programArgs)
	if err := support.Runner.Build(context.Background(), config); err != nil {
		return config, fmt.Errorf("build %s: %w", path, err)
	}
	s.remember(support.Runner)
	return config, nil
}

// RunUntitled runs unsaved source from a temporary folder. The extension of the untitled name
// ("untitled-1.py") decides the language.
func (s *RunService) RunUntitled(path, source string, programArgs []string) (domain.RunConfiguration, error) {
	support, err := s.supportFor(path)
	if err != nil {
		return domain.RunConfiguration{}, fmt.Errorf("run untitled %s: %w", path, err)
	}
	config, err := support.Runner.RunUntitled(context.Background(), path, source, programArgs)
	if err != nil {
		return config, fmt.Errorf("run untitled %s: %w", path, err)
	}
	s.remember(support.Runner)
	return config, nil
}

// Check runs the checker of config.CodeLanguage (go vet) on the target of a finished run and
// returns its output. It works in the background: it neither blocks Run nor emits run events.
func (s *RunService) Check(config domain.RunConfiguration) (string, error) {
	support, err := s.supportOf(config.CodeLanguage)
	if err != nil {
		return "", fmt.Errorf("check %s: %w", config.Target, err)
	}
	if support.Checker == nil {
		return "", fmt.Errorf("check %s: %w", config.Target, app.ErrUnsupported)
	}
	output, err := support.Checker.Check(context.Background(), config)
	if err != nil {
		return "", fmt.Errorf("check %s: %w", config.Target, err)
	}
	return output, nil
}

// SplitArguments splits the "program arguments" text like a shell (quotes group).
func (s *RunService) SplitArguments(text string) ([]string, error) {
	return app.SplitProgramArguments(text)
}

// Stop asks the running program to finish (so its defers and signal handlers run) and
// kills it and everything it started if it does not within about two seconds.
func (s *RunService) Stop() error { return s.activeRunner().Stop() }

// WriteInput sends text, followed by Enter, to the stdin of the running program.
func (s *RunService) WriteInput(text string) error { return s.activeRunner().WriteInput(text) }

// Format returns the formatted text of a file, in the language of its extension.
func (s *RunService) Format(path, text string) (string, error) {
	support, err := s.supportFor(path)
	if err != nil {
		return "", fmt.Errorf("format %s: %w", path, err)
	}
	if support.Formatter == nil {
		return "", fmt.Errorf("format %s: %w", path, app.ErrUnsupported)
	}
	return support.Formatter.Format(path, text)
}

// remember keeps the runner of a program that did start; a refused run (busy, unsupported)
// must not hide the runner that is really running.
func (s *RunService) remember(runner app.ProgramRunner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = runner
}

// activeRunner is the runner of the last program started. All runners share the process slot,
// so before any run the one that reports a running program (or the default) serves as well.
func (s *RunService) activeRunner() app.ProgramRunner {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active != nil {
		return s.active
	}
	for _, support := range s.registry.All() {
		if support.Runner.IsRunning() {
			return support.Runner
		}
	}
	return s.registry.Default().Runner
}
