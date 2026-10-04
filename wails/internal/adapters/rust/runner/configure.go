package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var (
	// ErrNoBinary means the crate (or the workspace) has nothing to run: a library. It wraps
	// app.ErrUnsupported and carries the key errors.rustNoBinary, which the frontend translates.
	ErrNoBinary = fmt.Errorf("%w: errors.rustNoBinary", app.ErrUnsupported)
	// ErrChooseMember means the file is in the root of a virtual workspace that has several
	// members with programs; the error is a *ChooseMemberError that lists them. It wraps
	// app.ErrUnsupported and carries the key run.chooseMember. The frontend asks which member to
	// run and calls Configure again with a file inside that member.
	ErrChooseMember = fmt.Errorf("%w: run.chooseMember", app.ErrUnsupported)
)

// ChooseMemberError is the ErrChooseMember of a workspace: Members are the package names (what
// "cargo run -p" takes) and Roots their folders, in the same order.
type ChooseMemberError struct {
	Members []string
	Roots   []string
}

func (e *ChooseMemberError) Error() string {
	return ErrChooseMember.Error() + ": " + strings.Join(e.Members, ", ")
}

// Unwrap makes errors.Is(err, ErrChooseMember) true.
func (e *ChooseMemberError) Unwrap() error { return ErrChooseMember }

// Configure implements app.ProgramRunner. A file inside a Cargo crate runs that crate (Mode
// project, Project from the crate); in project mode Target stays the open file, because it picks
// the binary (src/bin/<name>.rs). A loose file runs alone with rustc. The root of a virtual
// workspace is also a project: Run picks the member or answers ErrChooseMember.
func (r *Runner) Configure(path string, programArgs []string) domain.RunConfiguration {
	config := domain.NewFileRunConfiguration(domain.CodeLanguageRust, path, programArgs)
	project, ok, err := rust.FindProject(path)
	if err != nil || !ok {
		return config
	}
	config.Mode = domain.RunProject
	config.WorkingDir = project.Root
	config.Project = project.Context()
	return config
}

// cargoProject is the crate a project configuration runs, and the program in it.
type cargoProject struct {
	rust.CargoProject
	binary rust.Binary
}

// resolveCrate finds the crate and program of a project configuration. The root of a virtual
// workspace gives its only member with programs, or ErrChooseMember.
func resolveCrate(config domain.RunConfiguration) (cargoProject, error) {
	project, ok, err := rust.FindProject(anchor(config.Target))
	if err != nil {
		return cargoProject{}, fmt.Errorf("read Cargo.toml: %w", err)
	}
	if !ok {
		return cargoProject{}, ErrNoBinary
	}
	if project.Virtual {
		return memberToRun(project)
	}
	binary, ok := project.BinaryFor(config.Target)
	if !ok {
		return cargoProject{}, ErrNoBinary
	}
	return cargoProject{CargoProject: project, binary: binary}, nil
}

// memberToRun chooses among the members of a virtual workspace.
func memberToRun(workspace rust.CargoProject) (cargoProject, error) {
	var runnable []cargoProject
	for _, folder := range workspace.Members {
		member, ok, err := rust.FindProject(filepath.Join(folder, "Cargo.toml"))
		if err != nil || !ok || member.Virtual || len(member.Binaries) == 0 {
			continue
		}
		binary, _ := member.BinaryFor("")
		runnable = append(runnable, cargoProject{CargoProject: member, binary: binary})
	}
	switch len(runnable) {
	case 0:
		return cargoProject{}, ErrNoBinary
	case 1:
		return runnable[0], nil
	}
	choice := &ChooseMemberError{}
	for _, member := range runnable {
		choice.Members, choice.Roots = append(choice.Members, member.Name), append(choice.Roots, member.Root)
	}
	return cargoProject{}, choice
}

// anchor makes FindProject start in a folder when the target is the folder itself.
func anchor(target string) string {
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		return filepath.Join(target, "Cargo.toml")
	}
	return target
}

// isMember reports whether cargo needs "-p": the crate lives in a workspace with another root.
func (p cargoProject) isMember() bool {
	return !strings.EqualFold(filepath.Clean(p.Workspace), filepath.Clean(p.Root))
}
