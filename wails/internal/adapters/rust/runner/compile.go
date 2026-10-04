package runner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// compilation is stage 1 of a run: one rustc or cargo call.
type compilation struct {
	command string
	args    []string
	dir     string            // where the tool runs
	env     map[string]string // the environment of the tool
	// executable is where the program will be: the file rustc writes, or where cargo puts the
	// crate's binary unless its compiler-artifact message says otherwise.
	executable string
	filter     *diagnostics
	crate      bool
}

// program is the executable stage 2 runs: cargo reported its path, or the expected one.
func (c compilation) program() string {
	if built := c.filter.built(); built != "" {
		return built
	}
	return c.executable
}

// outputFor says where a loose file's executable goes.
type outputFor func(domain.RunConfiguration) string

// compile prepares the compiler call of a configuration. output decides where a loose file's
// executable goes (a crate always builds into its own target/).
func (r *Runner) compile(ctx context.Context, config domain.RunConfiguration, output outputFor) (compilation, error) {
	if config.Mode == domain.RunProject {
		return r.cargoBuild(ctx, config)
	}
	return r.rustcBuild(ctx, config, output)
}

func (r *Runner) rustcBuild(ctx context.Context, config domain.RunConfiguration, output outputFor) (compilation, error) {
	toolchain, err := r.locator.Toolchain(ctx)
	if err != nil {
		return compilation{}, err
	}
	out := output(config)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return compilation{}, fmt.Errorf("create the build folder: %w", err)
	}
	args := []string{"--edition", toolchain.Edition(), "-g", "--error-format=json"}
	env := r.Environment()
	if linker := r.linkerFor(toolchain); linker != "" {
		args = append(args, "-C", "linker="+linker)
	}
	if dlltool := r.locator.Dlltool(toolchain); dlltool != "" {
		args = append(args, "-C", "dlltool="+dlltool)
	}
	args = append(args, "-o", out, config.Target)
	return compilation{
		command: toolchain.Rustc, args: args, dir: config.WorkingDir, env: env, executable: out,
		filter: &diagnostics{},
	}, nil
}

func (r *Runner) cargoBuild(ctx context.Context, config domain.RunConfiguration) (compilation, error) {
	crate, err := resolveCrate(config)
	if err != nil {
		return compilation{}, err
	}
	cargo := r.locator.Tool(rust.ToolCargo)
	if cargo.Source == domain.ToolMissing {
		return compilation{}, app.MissingTool(rust.ToolCargo)
	}
	toolchain, err := r.locator.Toolchain(ctx)
	if err != nil {
		return compilation{}, err
	}
	args := []string{"build", "--message-format=json", "--manifest-path", crate.Manifest}
	if crate.isMember() {
		args = append(args, "-p", crate.Name)
	}
	args = append(args, "--bin", crate.binary.Name)
	env := r.Environment()
	if linker := r.linkerFor(toolchain); linker != "" {
		env[cargoLinkerVariable(toolchain.Host)] = linker
	}
	expected := filepath.Join(crate.Workspace, "target", "debug", executableName(crate.binary.Name))
	return compilation{
		command: cargo.Path, args: args, dir: crate.Root, env: env, executable: expected,
		filter: &diagnostics{binary: crate.binary.Name}, crate: true,
	}, nil
}

// compileJob describes the compiler call for the supervisor: pipes, the diagnostics as their
// rendered text and the "Compiling..." notice.
func (r *Runner) compileJob(c compilation, config domain.RunConfiguration) process.Job {
	job := process.Job{
		Config: config, Command: c.command, Args: c.args, Dir: c.dir, Env: c.env,
		Mode: process.Pipes, OutputFilter: c.filter.Filter,
	}
	if r.options.CompilingNotice != nil {
		job.Notice = &process.SilentNotice{Text: r.options.CompilingNotice, Delay: r.options.NoticeDelay}
	}
	return job
}

// cacheOutput is <UserCacheDir>/VizcachaIDE/build/<hash of the folder>/<name>(.exe): the
// student's folder stays clean.
func (r *Runner) cacheOutput(config domain.RunConfiguration) string {
	return filepath.Join(r.buildFolder(filepath.Dir(config.Target)), executableName(fileName(config.Target)))
}

// besideSource puts the executable next to the code (main.exe, main), like "go build".
func besideSource(config domain.RunConfiguration) string {
	return filepath.Join(filepath.Dir(config.Target), executableName(fileName(config.Target)))
}

// buildFolder is the cache folder of a source folder.
func (r *Runner) buildFolder(sourceFolder string) string {
	root := r.options.CacheDir
	if root == "" {
		var err error
		if root, err = os.UserCacheDir(); err != nil {
			root = os.TempDir()
		}
	}
	key := filepath.Clean(sourceFolder)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(root, "VizcachaIDE", "build", hex.EncodeToString(sum[:8]))
}

func fileName(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
