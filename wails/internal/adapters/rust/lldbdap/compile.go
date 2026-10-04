package lldbdap

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec // a short folder name, not security
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// ErrBuildFailed is returned with the compiler's output when the program does not compile.
var ErrBuildFailed = errors.New("the program does not compile")

// Compiler builds the program with debug information before it is debugged. The runner's own
// build stage can replace the default one.
type Compiler interface {
	// CompileForDebug returns the executable. When the compiler rejects the code it returns its
	// output with an error that wraps ErrBuildFailed; a missing rustc is app.MissingTool.
	CompileForDebug(ctx context.Context, config domain.RunConfiguration) (exe, output string, err error)
}

// buildCompiler is the default Compiler: rustc for a loose file, cargo for a crate.
type buildCompiler struct{ locator *rust.Locator }

var _ Compiler = buildCompiler{}

func (b buildCompiler) CompileForDebug(ctx context.Context, config domain.RunConfiguration) (string, string, error) {
	toolchain, err := b.locator.Toolchain(ctx)
	if err != nil {
		return "", "", err
	}
	project, isCrate, err := rust.FindProject(projectProbe(config))
	if err != nil {
		return "", "", err
	}
	env := b.locator.Environment()
	if isCrate && !project.Virtual {
		return buildCrate(ctx, b.locator, project, config, env)
	}
	return buildFile(ctx, toolchain, config, env)
}

// projectProbe is a path inside the folder whose Cargo.toml matters: FindProject starts at the
// folder of the path it receives.
func projectProbe(config domain.RunConfiguration) string {
	if config.Mode == domain.RunProject {
		return filepath.Join(config.Target, "main.rs")
	}
	return config.Target
}

// buildFile is "rustc --edition N -g -o <cache>/<name>.exe file.rs".
func buildFile(ctx context.Context, toolchain rust.Toolchain, config domain.RunConfiguration, env []string) (string, string, error) {
	target, err := filepath.Abs(config.Target)
	if err != nil {
		return "", "", err
	}
	exe, err := debugExecutable(filepath.Dir(target), strings.TrimSuffix(filepath.Base(target), filepath.Ext(target)))
	if err != nil {
		return "", "", err
	}
	command := exec.CommandContext(ctx, toolchain.Rustc, "--edition", toolchain.Edition(), "-g", "-o", exe, target)
	command.Dir, command.Env = filepath.Dir(target), env
	output, err := run(ctx, command)
	if err != nil {
		return "", output, err
	}
	return exe, output, nil
}

// buildCrate is "cargo build" of the binary the file belongs to; the executable is read from
// cargo's JSON messages, which keeps the rendered diagnostics on stderr.
func buildCrate(ctx context.Context, locator *rust.Locator, project rust.CargoProject, config domain.RunConfiguration, env []string) (string, string, error) {
	cargo := locator.Tool(rust.ToolCargo)
	if cargo.Path == "" {
		return "", "", fmt.Errorf("cargo: %w", errors.New("not found"))
	}
	args := []string{"build", "--manifest-path", project.Manifest, "--message-format=json-render-diagnostics"}
	binary, hasBinary := project.BinaryFor(config.Target)
	if hasBinary {
		args = append(args, "--bin", binary.Name)
	}
	if project.Workspace != project.Root {
		args = append(args, "-p", project.Name)
	}
	command := exec.CommandContext(ctx, cargo.Path, args...)
	command.Dir, command.Env = project.Root, env
	var messages bytes.Buffer
	command.Stdout = &messages
	output, err := run(ctx, command)
	if err != nil {
		return "", output, err
	}
	exe := executableFrom(messages.String(), binary.Name)
	if exe == "" {
		return "", output, fmt.Errorf("%w: cargo did not build a binary", ErrBuildFailed)
	}
	return exe, output, nil
}

// run executes the command and returns what it wrote to stderr (and stdout, unless the caller
// took it).
func run(ctx context.Context, command *exec.Cmd) (string, error) {
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if command.Stdout == nil {
		command.Stdout = &stderr
	}
	process.HideConsole(command)
	err := command.Run()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if err != nil {
		return stderr.String(), fmt.Errorf("%w: %w", ErrBuildFailed, err)
	}
	return stderr.String(), nil
}

// executableFrom finds the executable of the binary in cargo's JSON lines ("compiler-artifact").
func executableFrom(messages, binary string) string {
	found := ""
	for _, line := range strings.Split(messages, "\n") {
		var artifact struct {
			Reason     string `json:"reason"`
			Executable string `json:"executable"`
			Target     struct {
				Name string `json:"name"`
			} `json:"target"`
		}
		if json.Unmarshal([]byte(line), &artifact) != nil || artifact.Reason != "compiler-artifact" || artifact.Executable == "" {
			continue
		}
		if found == "" || artifact.Target.Name == binary {
			found = artifact.Executable
		}
	}
	return found
}

// debugExecutable is <cache>/VizcachaIDE/build/debug/<hash of the folder>/<name>, away from the
// student's folder.
func debugExecutable(folder, name string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	sum := sha1.Sum([]byte(strings.ToLower(folder))) //nolint:gosec // see import
	dir := filepath.Join(cache, "VizcachaIDE", "build", "debug", "rust-"+hex.EncodeToString(sum[:6]))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name), nil
}
