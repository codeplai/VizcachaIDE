package clippy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// compilerMessage marks the lines of cargo's JSON stream that carry a diagnostic; the others
// announce artifacts and the end of the build.
const compilerMessage = `"reason":"compiler-message"`

// checkCrate runs "cargo clippy" (or "cargo check" without clippy) on the crate. In a workspace
// it runs from the workspace folder, whose target/ is shared, with "-p <crate>".
func (t *Tool) checkCrate(ctx context.Context, project rust.CargoProject) (string, error) {
	cargo := t.locator.Tool(rust.ToolCargo)
	if cargo.Source == domain.ToolMissing {
		return "", app.MissingTool(rust.ToolCargo)
	}
	verb := "check"
	if t.hasClippy() {
		verb = "clippy"
	}
	args := []string{verb, "--message-format=json", "--quiet", "--manifest-path", project.Manifest}
	if project.Name != "" && project.Workspace != project.Root {
		args = append(args, "-p", project.Name)
	}
	stdout, stderr, err := t.run(ctx, project.Workspace, cargo.Path, args)
	if lines := keepLines(stdout, compilerMessage); lines != "" {
		return lines, nil
	}
	if err != nil && !isExit(err) {
		return "", fmt.Errorf("cargo %s: %w", verb, err)
	}
	// No diagnostics but a failure: cargo could not even start (a bad Cargo.toml, a dependency
	// that is not there) and says why in plain text.
	if err != nil {
		return strings.TrimSpace(stderr), nil
	}
	return "", nil
}

// checkFile runs clippy-driver (rustc without clippy) on a loose file, with --emit=metadata so
// nothing is linked, into a folder that is deleted afterwards. rustc prints its JSON on stderr.
func (t *Tool) checkFile(ctx context.Context, config domain.RunConfiguration) (string, error) {
	toolchain, err := t.locator.Toolchain(ctx)
	if err != nil {
		return "", err
	}
	executable := t.driver(toolchain)
	out, err := os.MkdirTemp("", "vizcacha-rust-check-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(out) }()
	args := []string{
		"--edition", toolchain.Edition(), "--crate-type", "bin", "--emit=metadata",
		"--out-dir", out, "--error-format=json", config.Target,
	}
	_, stderr, err := t.run(ctx, config.WorkingDir, executable, args)
	if err != nil && !isExit(err) {
		return "", fmt.Errorf("%s: %w", filepath.Base(executable), err)
	}
	return keepLines(stderr, "{"), nil
}

// driver is clippy-driver, which sits next to cargo-clippy, or rustc when clippy is missing.
func (t *Tool) driver(toolchain rust.Toolchain) string {
	if clippy := t.locator.Tool(rust.ToolClippy); clippy.Source != domain.ToolMissing {
		name := "clippy-driver"
		if strings.HasSuffix(strings.ToLower(clippy.Path), ".exe") {
			name += ".exe"
		}
		if driver := filepath.Join(filepath.Dir(clippy.Path), name); isFile(driver) {
			return driver
		}
	}
	return toolchain.Rustc
}

// run starts a tool in dir with the Rust environment and returns what it printed.
func (t *Tool) run(ctx context.Context, dir, executable string, args []string) (stdout, stderr string, err error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = dir
	cmd.Env = t.locator.Environment()
	process.HideConsole(cmd)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err = cmd.Run()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return out.String(), errOut.String(), err
}

// keepLines keeps the lines that contain marker, as lines of their own.
func keepLines(text, marker string) string {
	kept := []string{}
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if strings.Contains(line, marker) && strings.HasPrefix(strings.TrimSpace(line), "{") {
			kept = append(kept, strings.TrimSpace(line))
		}
	}
	if len(kept) == 0 {
		return ""
	}
	return strings.Join(kept, "\n") + "\n"
}

func isExit(err error) bool {
	var exit *exec.ExitError
	return errors.As(err, &exit)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
