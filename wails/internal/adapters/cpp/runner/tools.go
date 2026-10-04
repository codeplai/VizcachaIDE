package runner

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const versionTimeout = 5 * time.Second

var versionNumber = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// Tools implements app.ProgramRunner: the compiler, lldb-dap, clangd and clang-format, with where
// they were found and the version number each one reports.
func (r *Runner) Tools(ctx context.Context) []domain.ToolStatus {
	statuses := make([]domain.ToolStatus, 0, len(cpp.Profile.Tools))
	for _, spec := range cpp.Profile.Tools {
		statuses = append(statuses, r.status(ctx, spec))
	}
	return statuses
}

func (r *Runner) status(ctx context.Context, spec domain.ToolSpec) domain.ToolStatus {
	if spec.ID == cpp.ToolCompiler {
		return r.compilerStatus(ctx, spec)
	}
	status := r.locator.Tool(ctx, spec.ID)
	if status.Source != domain.ToolMissing {
		status.Version = versionOf(ctx, status.Path, r.base)
	}
	return status
}

func (r *Runner) compilerStatus(ctx context.Context, spec domain.ToolSpec) domain.ToolStatus {
	status := domain.ToolStatus{ID: spec.ID, CodeLanguage: domain.CodeLanguageCpp, Role: spec.Role, Source: domain.ToolMissing}
	compiler, err := r.locator.Compiler(ctx)
	if err != nil {
		return status
	}
	status.Path, status.Source, status.Version = compiler.Path, compiler.Source, versionNumberOf(compiler.Version)
	return status
}

// versionOf runs "<tool> --version" and returns its version number ("" when it fails).
func versionOf(ctx context.Context, path string, env []string) string {
	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "--version")
	cmd.Env = env
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return versionNumberOf(string(output))
}

// versionNumberOf finds the number after the first "version " of a --version text ("clang
// version 23.1.2", "  LLVM version 23.1.2" in lldb-dap), or the first number of the first line
// when the word never appears (g++ prints "g++ (Ubuntu 13.2.0) 13.2.0").
func versionNumberOf(text string) string {
	text = strings.TrimSpace(text)
	if _, after, found := strings.Cut(text, "version "); found {
		line, _, _ := strings.Cut(after, "\n")
		return versionNumber.FindString(line)
	}
	line, _, _ := strings.Cut(text, "\n")
	return versionNumber.FindString(line)
}
