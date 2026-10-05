package runner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// ErrCompileFailed is wrapped by CompileForDebug when the compiler rejected the code; the
// compiler's output is returned beside it.
var ErrCompileFailed = errors.New("the compiler rejected the code")

// compilation is stage 1 of a run: one compiler call.
type compilation struct {
	compiler cpp.Compiler
	sources  []string
	folder   string
	output   string
	console  string // the UTF-8 console helper compiled with the sources on Windows (console.go)
}

// compile prepares the compiler call for a configuration. output decides where the executable
// goes; its folder is created.
func (r *Runner) compile(ctx context.Context, config domain.RunConfiguration, output func(domain.RunConfiguration) string) (compilation, error) {
	sources, err := sourcesOf(config)
	if err != nil {
		return compilation{}, err
	}
	compiler, err := r.locator.Compiler(ctx)
	if err != nil {
		return compilation{}, err
	}
	c := compilation{
		compiler: compiler, sources: sources, folder: folderOf(config), output: output(config),
		console: r.consoleSource(),
	}
	if err := os.MkdirAll(filepath.Dir(c.output), 0o755); err != nil {
		return compilation{}, fmt.Errorf("create the build folder: %w", err)
	}
	return c, nil
}

// arguments are "<flags> -o <output> <sources>".
func (c compilation) arguments() []string {
	args := cpp.CompileFlags(c.compiler.Family, runtime.GOOS)
	args = append(args, "-o", c.output)
	args = append(args, c.sources...)
	if c.console != "" {
		args = append(args, c.console)
	}
	return args
}

// job describes the compiler call for the supervisor: pipes, with the "Compiling..." notice.
func (r *Runner) compileJob(c compilation, config domain.RunConfiguration) process.Job {
	job := process.Job{
		Config: config, Command: c.compiler.Path, Args: c.arguments(), Dir: c.folder,
		Env: r.Environment(), Mode: process.Pipes,
	}
	if r.options.CompilingNotice != nil {
		job.Notice = &process.SilentNotice{Text: r.options.CompilingNotice, Delay: r.options.NoticeDelay}
	}
	return job
}

// cacheOutput is <UserCacheDir>/VizcachaIDE/build/<hash of the folder>/<name>(.exe): the
// student's folder stays clean.
func (r *Runner) cacheOutput(config domain.RunConfiguration) string {
	return filepath.Join(r.buildFolder(folderOf(config)), executableName(programName(config)))
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

// besideSource puts the executable next to the code, like "go build".
func besideSource(config domain.RunConfiguration) string {
	return filepath.Join(folderOf(config), executableName(programName(config)))
}

func executableName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

// CompileForDebug is stage 1 for the debugger (track C2): it compiles the configuration
// synchronously, without run events, into the same cache folder Run uses, and returns the
// executable. output is the compiler's text (warnings too). When the compiler rejects the code,
// exePath is "" and err wraps ErrCompileFailed; the output is what the Assistant explains. Other
// errors are app.MissingTool("cxx"), ErrHeaderOnly or a failure to start the compiler.
func (r *Runner) CompileForDebug(ctx context.Context, config domain.RunConfiguration) (exePath, output string, err error) {
	c, err := r.compile(ctx, config, r.cacheOutput)
	if err != nil {
		return "", "", err
	}
	text, err := capture(ctx, c.compiler.Path, c.folder, r.base, c.arguments())
	if err != nil {
		return "", text, err
	}
	return c.output, text, nil
}

// capture runs a compiler call outside the supervisor and returns its stdout and stderr
// together. A non-zero exit wraps ErrCompileFailed.
func capture(ctx context.Context, path, dir string, env, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir, cmd.Env = dir, env
	process.HideConsole(cmd)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return output.String(), fmt.Errorf("%w (exit code %d)", ErrCompileFailed, exit.ExitCode())
	}
	if err != nil {
		return output.String(), fmt.Errorf("start the compiler: %w", err)
	}
	return output.String(), nil
}
