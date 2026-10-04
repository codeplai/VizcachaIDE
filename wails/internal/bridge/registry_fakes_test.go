package bridge

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakeRunner records what the services ask of a ProgramRunner.
type fakeRunner struct {
	language domain.CodeLanguage
	tools    []domain.ToolStatus

	ran, built, stopped int
	runConfig           domain.RunConfiguration
	untitledPath        string
	inputs              []string
	running             bool
	runErr              error
}

func (f *fakeRunner) Configure(path string, args []string) domain.RunConfiguration {
	return domain.RunConfiguration{CodeLanguage: f.language, Target: path, WorkingDir: "wd-" + string(f.language), ProgramArgs: args}
}

func (f *fakeRunner) Run(_ context.Context, c domain.RunConfiguration) error {
	f.ran++
	f.runConfig = c
	return f.runErr
}

func (f *fakeRunner) Build(context.Context, domain.RunConfiguration) error {
	f.built++
	return nil
}

func (f *fakeRunner) RunUntitled(_ context.Context, path, _ string, args []string) (domain.RunConfiguration, error) {
	f.untitledPath = path
	return f.Configure(path, args), nil
}

func (f *fakeRunner) Stop() error                               { f.stopped++; return nil }
func (f *fakeRunner) IsRunning() bool                           { return f.running }
func (f *fakeRunner) WriteInput(text string) error              { f.inputs = append(f.inputs, text); return nil }
func (f *fakeRunner) Tools(context.Context) []domain.ToolStatus { return f.tools }
func (f *fakeRunner) Environment() map[string]string            { return nil }

// fakeDebugger is a Debugger whose session can be switched on and off.
type fakeDebugger struct {
	app.Debugger
	active    bool
	started   []domain.RunConfiguration
	breakons  []domain.Breakpoint
	steps     int
	stopped   int
	startErr  error
	blockOnce chan struct{} // when set, Start waits for it to be closed
	entered   chan struct{}
}

func (f *fakeDebugger) Start(_ context.Context, c domain.RunConfiguration, b []domain.Breakpoint) error {
	if f.entered != nil {
		close(f.entered)
	}
	if f.blockOnce != nil {
		<-f.blockOnce
	}
	if f.startErr != nil {
		return f.startErr
	}
	f.active = true
	f.started = append(f.started, c)
	f.breakons = b
	return nil
}

func (f *fakeDebugger) StepOver() error { f.steps++; return nil }
func (f *fakeDebugger) Stop() error     { f.stopped++; f.active = false; return nil }
func (f *fakeDebugger) IsActive() bool  { return f.active }

// fakeServer records the paths a LanguageServer received.
type fakeServer struct {
	app.LanguageServer
	opened []string
	hovers []string
}

func (f *fakeServer) OpenDocument(_ context.Context, path, _ string) error {
	f.opened = append(f.opened, path)
	return nil
}

func (f *fakeServer) Hover(_ context.Context, at domain.SourceLocation) (string, error) {
	f.hovers = append(f.hovers, at.File)
	return "hover", nil
}

// fakeExplainer parses every line as a diagnostic and explains it as "<language>:<message>".
type fakeExplainer struct{ tag string }

func (f fakeExplainer) Parse(rawOutput, _ string) []domain.Diagnostic {
	return []domain.Diagnostic{
		{Severity: domain.SeverityError, Message: rawOutput},
		{Severity: domain.SeverityError, Message: rawOutput},
		{Severity: domain.SeverityHint, Message: "hint"},
	}
}

func (f fakeExplainer) Explain(d domain.Diagnostic, language string) *domain.ErrorExplanation {
	return &domain.ErrorExplanation{ExplanationID: "X", Title: f.tag + ":" + language + ":" + d.Message}
}

type fakeConsole struct{ resets int }

func (f *fakeConsole) Eval(code string) domain.ConsoleResult {
	return domain.ConsoleResult{Result: code}
}
func (f *fakeConsole) Reset() { f.resets++ }

type fakeFormatter struct{}

func (fakeFormatter) Format(path, text string) (string, error) { return path + "|" + text, nil }

type fakeChecker struct{ configs []domain.RunConfiguration }

func (f *fakeChecker) Check(_ context.Context, c domain.RunConfiguration) (string, error) {
	f.configs = append(f.configs, c)
	return "vet output", nil
}

type fakePackages struct{ calls []string }

func (f *fakePackages) Init(_ context.Context, dir, name string) error {
	f.calls = append(f.calls, "init "+dir+" "+name)
	return nil
}
func (f *fakePackages) Add(_ context.Context, dir, pkg string) error {
	f.calls = append(f.calls, "add "+dir+" "+pkg)
	return nil
}
func (f *fakePackages) Remove(context.Context, string, string) error { return app.ErrUnsupported }
func (f *fakePackages) Tidy(_ context.Context, dir string) error {
	f.calls = append(f.calls, "tidy "+dir)
	return nil
}
func (f *fakePackages) List(context.Context, string) error { return app.ErrUnsupported }
