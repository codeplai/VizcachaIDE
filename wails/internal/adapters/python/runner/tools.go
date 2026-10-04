package runner

import (
	"context"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Interpreter finds the Python for the programs of folder ("" when there is none yet). It fails
// with app.MissingTool("python").
func (r *Runner) Interpreter(ctx context.Context, folder string) (python.Interpreter, error) {
	return r.locator.Find(ctx, folder)
}

// Tools implements app.ProgramRunner: python and the modules that ride on it (debugpy, pylsp,
// ruff), with where they were found and their versions.
func (r *Runner) Tools(ctx context.Context) []domain.ToolStatus {
	interpreter, err := r.locator.Find(ctx, "")
	var versions map[string]string
	if err == nil {
		versions = python.ModuleVersions(ctx, interpreter, python.Environment(r.base, interpreter))
	}
	statuses := make([]domain.ToolStatus, 0, len(python.Profile.Tools))
	for _, spec := range python.Profile.Tools {
		status := domain.ToolStatus{ID: spec.ID, CodeLanguage: domain.CodeLanguagePython, Role: spec.Role, Source: domain.ToolMissing}
		if err == nil {
			status.Path, status.Source, status.Version = interpreter.Path, interpreter.Source, versions[spec.ID]
			if spec.ID == python.ToolPython {
				status.Version = interpreter.Version
			}
		}
		statuses = append(statuses, status)
	}
	return statuses
}

// Environment implements app.ProgramRunner: the variables of the found interpreter, or the
// plain Python environment when there is none.
func (r *Runner) Environment() map[string]string {
	interpreter, _ := r.locator.Find(context.Background(), "")
	return python.Environment(r.base, interpreter)
}

// CommandJob describes "python <args>" run in dir as a project command ("-m pip list"). The
// package manager starts it on the shared supervisor; its output goes through pipes.
func (r *Runner) CommandJob(ctx context.Context, dir string, args []string) (process.Job, error) {
	interpreter, err := r.locator.Find(ctx, dir)
	if err != nil {
		return process.Job{}, err
	}
	config := domain.RunConfiguration{
		CodeLanguage: domain.CodeLanguagePython,
		Target:       strings.Join(append([]string{"python"}, args...), " "),
		WorkingDir:   dir,
		Mode:         domain.RunProject,
		ProgramArgs:  []string{},
	}
	return process.Job{
		Config: config, Command: interpreter.Path, Args: args, Dir: dir,
		Env: python.Environment(r.base, interpreter), Mode: process.Pipes,
	}, nil
}
