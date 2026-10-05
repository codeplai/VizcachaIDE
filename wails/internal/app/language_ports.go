package app

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// PORTS OF A LANGUAGE: CONTRACT v3 (PLAN_NUCLEO_MULTILENGUAJE.md section 3.5).
//
// Every language provides a ProgramRunner, a Debugger, a LanguageServer and an ErrorExplainer.
// Each optional capability is a port of its own that is nil when the language lacks it
// (Console, CodeFormatter, CodeChecker, PackageManager). The only exception is Build, a
// ProgramRunner method that returns ErrUnsupported and is offered only if Capabilities.Build.

// ProgramRunner runs programs of one language. Events: run:started, run:output, run:finished.
//
// Every runner receives the same process supervisor (protocol/process), so only one program
// runs at a time in the whole IDE and Stop, IsRunning and WriteInput act on that single slot.
type ProgramRunner interface {
	// Configure decides what Run would run for a file: the file alone or its project.
	Configure(path string, programArgs []string) domain.RunConfiguration
	// Run starts the configuration. Returns ErrBusy while another program is running.
	Run(ctx context.Context, config domain.RunConfiguration) error
	// Build compiles without running. Languages without a build step return ErrUnsupported.
	Build(ctx context.Context, config domain.RunConfiguration) error
	// RunUntitled runs unsaved source; the extension of path (an untitled name such as
	// "untitled-1.py") is kept for the temporary file.
	RunUntitled(ctx context.Context, path, source string, programArgs []string) (domain.RunConfiguration, error)
	// Stop interrupts the running program (Ctrl+C semantics) and kills its process tree if
	// it is still alive after about 2 s.
	Stop() error
	IsRunning() bool
	// WriteInput sends text to the program's stdin (or its pseudoterminal).
	WriteInput(text string) error
	// Tools reports every tool of the language: where it was found and its version.
	Tools(ctx context.Context) []domain.ToolStatus
	// Environment returns the environment variables the language's tools run with.
	Environment() map[string]string
}

// CodeFormatter formats source text (go/format, ruff format, clang-format).
// Errors wrap ErrFormat or ErrToolNotFound.
type CodeFormatter interface {
	Format(path, text string) (string, error)
}

// CodeChecker runs the language's checker (go vet, ruff check) without events and outside
// the single run slot. It returns the checker's output, "" when it found nothing.
type CodeChecker interface {
	Check(ctx context.Context, config domain.RunConfiguration) (string, error)
}

// PackageManager runs package commands through the shared process slot, so they emit run
// events. Verbs a language lacks return ErrUnsupported; the profile lists the supported ones.
type PackageManager interface {
	Init(ctx context.Context, dir, name string) error
	Add(ctx context.Context, dir, pkg string) error
	Remove(ctx context.Context, dir, pkg string) error
	Tidy(ctx context.Context, dir string) error
	List(ctx context.Context, dir string) error
}

// PackageSearch looks a package up by name in the package index of its language, so the student
// can choose from a list instead of typing an exact name. It never touches the project and emits
// no events. When the index cannot be searched (no network, an unexpected answer) the error wraps
// ErrPackageIndexUnavailable. An empty query gives an empty list without any request.
type PackageSearch interface {
	Search(ctx context.Context, query string) ([]domain.PackageInfo, error)
}

// MemberRunner is a ProgramRunner whose projects can hold several programs: in the root of a
// Cargo workspace, Run answers that the student must choose a member, and ConfigureMember gives
// the configuration of the chosen one (docs/PLAN_RUST.md section 3.2 point 8).
type MemberRunner interface {
	ConfigureMember(path, member string, programArgs []string) (domain.RunConfiguration, error)
}
