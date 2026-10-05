package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cmake"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// projectCheckTimeout is long: the first build of a project may install its libraries.
const projectCheckTimeout = 15 * time.Minute

// debugBuild is CompileForDebug for a CMake project: the Debug build run uses, whose program has
// debug information (CMAKE_BUILD_TYPE=Debug). output is what the tools printed without the
// build's noise; a rejected build wraps ErrCompileFailed.
func (r *Runner) debugBuild(ctx context.Context, config domain.RunConfiguration, project cmake.Project) (exePath, output string, err error) {
	tools, err := r.cmakeTools(ctx)
	if err != nil {
		return "", "", err
	}
	raw, err := tools.builder.Build(ctx, project.Root, false)
	output = r.filter().Text(raw)
	if errors.Is(err, cmake.ErrConfigureFailed) || errors.Is(err, cmake.ErrBuildFailed) {
		return "", output, fmt.Errorf("%w: %w", ErrCompileFailed, err)
	}
	if err != nil {
		return "", output, err
	}
	build := cmakeBuild{runner: r, project: project, config: config}
	build.plan.BuildDir = cmake.BuildDir(project.Root, false)
	exePath, err = build.executable()
	if err != nil {
		return "", output, err
	}
	return exePath, output, nil
}

// checkProject is Check for a CMake project: a build, whose warnings and errors are the findings.
func (r *Runner) checkProject(ctx context.Context, project cmake.Project) (string, error) {
	tools, err := r.cmakeTools(ctx)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, projectCheckTimeout)
	defer cancel()
	raw, err := tools.builder.Build(ctx, project.Root, false)
	if err != nil && !errors.Is(err, cmake.ErrConfigureFailed) && !errors.Is(err, cmake.ErrBuildFailed) {
		return "", err
	}
	return r.filter().Text(raw), nil
}
