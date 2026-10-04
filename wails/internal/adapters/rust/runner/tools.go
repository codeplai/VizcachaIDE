package runner

import (
	"context"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const versionTimeout = 10 * time.Second

var versionNumber = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// Tools implements app.ProgramRunner: rustc, cargo, rust-analyzer, lldb-dap, clippy and rustfmt,
// with where they were found and the version number each one reports.
func (r *Runner) Tools(ctx context.Context) []domain.ToolStatus {
	statuses := make([]domain.ToolStatus, len(rust.Profile.Tools))
	var wait sync.WaitGroup
	for i, spec := range rust.Profile.Tools {
		wait.Add(1)
		go func() {
			defer wait.Done()
			statuses[i] = r.status(ctx, spec)
		}()
	}
	wait.Wait()
	return statuses
}

func (r *Runner) status(ctx context.Context, spec domain.ToolSpec) domain.ToolStatus {
	status := r.locator.Tool(spec.ID)
	if status.Source == domain.ToolMissing && spec.ProvidedBy != "" {
		status = r.besideCargo(spec, status)
	}
	if status.Source != domain.ToolMissing {
		status.Version = r.versionOf(ctx, spec.ID, status.Path)
	}
	return status
}

// besideCargo looks for a component of the toolchain (clippy, rustfmt) in the folder of the
// cargo that is in use, which is where rustup installs it even when cargo was chosen by hand.
func (r *Runner) besideCargo(spec domain.ToolSpec, missing domain.ToolStatus) domain.ToolStatus {
	cargo := r.locator.Tool(spec.ProvidedBy)
	if cargo.Source == domain.ToolMissing {
		return missing
	}
	name := spec.ID
	if spec.ID == rust.ToolClippy {
		name = "cargo-clippy"
	}
	path := filepath.Join(filepath.Dir(cargo.Path), executableName(name))
	if !isFile(path) {
		return missing
	}
	missing.Path, missing.Source = path, cargo.Source
	return missing
}

// versionOf asks the tool for its version: "<tool> --version", except clippy ("cargo clippy
// --version") whose own executable is only the cargo subcommand.
func (r *Runner) versionOf(ctx context.Context, id, path string) string {
	command, args := path, []string{"--version"}
	if id == rust.ToolClippy {
		cargo := r.locator.Tool(rust.ToolCargo)
		if cargo.Source == domain.ToolMissing {
			return ""
		}
		command, args = cargo.Path, []string{"clippy", "--version"}
	}
	ctx, cancel := context.WithTimeout(ctx, versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Env = environmentList(r.Environment())
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return versionNumberOf(string(output))
}

// versionNumberOf finds the number after the first "version " of a --version text ("LLVM version
// 23.1.2" in lldb-dap), or the first number of the first line when the word never appears
// ("rustc 1.99.0 (b940084d7 2026-09-28)", "clippy 0.1.99 (...)").
func versionNumberOf(text string) string {
	text = strings.TrimSpace(text)
	if _, after, found := strings.Cut(text, "version "); found {
		line, _, _ := strings.Cut(after, "\n")
		return versionNumber.FindString(line)
	}
	line, _, _ := strings.Cut(text, "\n")
	return versionNumber.FindString(line)
}

// CommandJob describes "cargo <args>" run in dir as a project command ("cargo add"). The package
// manager starts it on the shared supervisor. A missing cargo is app.MissingTool("cargo").
func (r *Runner) CommandJob(dir string, args []string) (process.Job, error) {
	cargo := r.locator.Tool(rust.ToolCargo)
	if cargo.Source == domain.ToolMissing {
		return process.Job{}, app.MissingTool(rust.ToolCargo)
	}
	config := domain.RunConfiguration{
		CodeLanguage: domain.CodeLanguageRust,
		Target:       strings.Join(append([]string{"cargo"}, args...), " "),
		WorkingDir:   dir,
		Mode:         domain.RunProject,
		ProgramArgs:  []string{},
	}
	job := process.Job{
		Config: config, Command: cargo.Path, Args: args, Dir: dir, Env: r.Environment(), Mode: process.Pipes,
	}
	if r.options.CompilingNotice != nil {
		job.Notice = &process.SilentNotice{Text: r.options.CompilingNotice, Delay: r.options.NoticeDelay}
	}
	return job, nil
}
