package bridge

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// RunService runs the user's program. W0 STUB: owned by track G1, which replaces
// the body of each method with calls to app.Toolchain. Emits run:started,
// run:output and run:finished.
type RunService struct {
	sink app.EventSink
}

// NewRunService creates the service.
func NewRunService(sink app.EventSink) *RunService { return &RunService{sink: sink} }

// Run compiles and runs one file and returns the configuration it used.
func (s *RunService) Run(path string, programArgs []string) (domain.RunConfiguration, error) {
	config := domain.NewFileRunConfiguration(path, programArgs)
	s.sink.RunStarted(config)
	s.sink.RunOutput("stdout", "Hola, Go\n")
	s.sink.RunFinished(0, 400)
	return config, nil
}

// RunUntitled runs unsaved source from a temporary folder.
func (s *RunService) RunUntitled(source string, programArgs []string) (domain.RunConfiguration, error) {
	return s.Run("untitled/main.go", programArgs)
}

// Stop kills the running program.
func (s *RunService) Stop() error {
	return nil
}

// WriteInput sends text to the stdin of the running program.
func (s *RunService) WriteInput(text string) error { return nil }

// Format returns the gofmt-formatted text.
func (s *RunService) Format(text string) (string, error) { return text, nil }

// Toolchain reports which Go tools were found.
func (s *RunService) Toolchain() domain.ToolchainInfo {
	return domain.ToolchainInfo{GoVersion: "1.25.5", DelveVersion: "1.27.2", GoplsVersion: "0.21.1"}
}
