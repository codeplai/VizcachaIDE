// Package clippy is the Rust checker: the lints of clippy (and the warnings of rustc) as the JSON
// lines the compiler prints, for the Assistant to read after a run (docs/PLAN_RUST.md sections 3
// and 4.5). A crate is checked with cargo, a loose .rs file with clippy-driver, which runs rustc
// with the clippy lints and needs no manifest.
package clippy

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const checkTimeout = 60 * time.Second

// Tool implements app.CodeChecker with clippy.
type Tool struct{ locator *rust.Locator }

var _ app.CodeChecker = (*Tool)(nil)

// New creates the checker.
func New(locator *rust.Locator) *Tool { return &Tool{locator: locator} }

// Check runs the checker on the crate around config.Target, or on the loose file, without
// events and outside the run slot. It returns the JSON lines of the diagnostics ("" when the
// code is clean, or when the 60 seconds run out). Findings are output, not an error. Without
// clippy the same is done with plain cargo check or rustc, which only warn what the compiler
// warns. A missing rustc or cargo is app.MissingTool.
func (t *Tool) Check(ctx context.Context, config domain.RunConfiguration) (string, error) {
	if info, err := os.Stat(config.WorkingDir); err != nil || !info.IsDir() {
		return "", nil // for example an untitled run, whose folder is already deleted
	}
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	start := config.Target
	if config.Project != nil && config.Project.Kind == domain.ProjectCargo {
		start = filepath.Join(config.Project.Root, "Cargo.toml") // FindProject starts at the folder of this
	}
	project, found, err := rust.FindProject(start)
	if err != nil {
		return "", err
	}
	var output string
	if found {
		output, err = t.checkCrate(ctx, project)
	} else {
		output, err = t.checkFile(ctx, config)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "", nil
	}
	return output, err
}

// hasClippy says whether the clippy component is installed.
func (t *Tool) hasClippy() bool {
	return t.locator.Tool(rust.ToolClippy).Source != domain.ToolMissing
}
