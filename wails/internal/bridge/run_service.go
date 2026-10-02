package bridge

import (
	"context"
	"fmt"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// RunService runs the user's program. It decides what to run (app.ConfigurationForFile)
// and delegates the process to app.Toolchain, which emits run:started, run:output
// and run:finished.
type RunService struct {
	toolchain app.Toolchain
}

// NewRunService creates the service.
func NewRunService(toolchain app.Toolchain) *RunService {
	return &RunService{toolchain: toolchain}
}

// Run compiles and runs one file and returns the configuration it used.
// The program outlives this call, so it runs on its own context.
func (s *RunService) Run(path string, programArgs []string) (domain.RunConfiguration, error) {
	config := app.ConfigurationForFile(path, programArgs)
	if err := s.toolchain.Run(context.Background(), config); err != nil {
		return config, fmt.Errorf("run %s: %w", path, err)
	}
	return config, nil
}

// Build compiles one file without running it.
func (s *RunService) Build(path string, programArgs []string) (domain.RunConfiguration, error) {
	config := app.ConfigurationForFile(path, programArgs)
	if err := s.toolchain.Build(context.Background(), config); err != nil {
		return config, fmt.Errorf("build %s: %w", path, err)
	}
	return config, nil
}

// RunUntitled runs unsaved source from a temporary folder.
func (s *RunService) RunUntitled(source string, programArgs []string) (domain.RunConfiguration, error) {
	config, err := s.toolchain.RunUntitled(context.Background(), source, programArgs)
	if err != nil {
		return config, fmt.Errorf("run untitled: %w", err)
	}
	return config, nil
}

// Vet runs "go vet" on the target of a finished run and returns its output. It works
// in the background: it neither blocks Run nor emits run events.
func (s *RunService) Vet(config domain.RunConfiguration) (string, error) {
	output, err := s.toolchain.Vet(context.Background(), config)
	if err != nil {
		return "", fmt.Errorf("vet %s: %w", config.Target, err)
	}
	return output, nil
}

// SplitArguments splits the "program arguments" text like a shell (quotes group).
func (s *RunService) SplitArguments(text string) ([]string, error) {
	return app.SplitProgramArguments(text)
}

// Stop asks the running program to finish (so its defers and signal handlers run) and
// kills it and everything it started if it does not within about two seconds.
func (s *RunService) Stop() error { return s.toolchain.Stop() }

// WriteInput sends text, followed by Enter, to the stdin of the running program.
func (s *RunService) WriteInput(text string) error { return s.toolchain.WriteInput(text) }

// Format returns the gofmt-formatted text.
func (s *RunService) Format(text string) (string, error) { return s.toolchain.FormatSource(text) }

// Toolchain reports which Go tools were found and their versions.
func (s *RunService) Toolchain() domain.ToolchainInfo {
	return s.toolchain.Info(context.Background())
}

// ModInit runs "go mod init <modulePath>" in workingDir.
func (s *RunService) ModInit(workingDir, modulePath string) error {
	args, err := app.ModInitArguments(modulePath)
	if err != nil {
		return err
	}
	return s.goCommand(workingDir, args)
}

// ModGet runs "go get <pkg>" in workingDir.
func (s *RunService) ModGet(workingDir, pkg string) error {
	args, err := app.GetArguments(pkg)
	if err != nil {
		return err
	}
	return s.goCommand(workingDir, args)
}

// ModTidy runs "go mod tidy" in workingDir.
func (s *RunService) ModTidy(workingDir string) error {
	return s.goCommand(workingDir, app.ModTidyArguments())
}

func (s *RunService) goCommand(workingDir string, args []string) error {
	if err := s.toolchain.RunGoCommand(context.Background(), workingDir, args); err != nil {
		return fmt.Errorf("go %v: %w", args, err)
	}
	return nil
}
