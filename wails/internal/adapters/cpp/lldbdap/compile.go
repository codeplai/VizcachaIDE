package lldbdap

import (
	"context"
	"crypto/sha1" //nolint:gosec // a short folder name, not security
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

// ErrBuildFailed is returned with the compiler's output when the program does not compile.
var ErrBuildFailed = errors.New("the program does not compile")

// Compiler builds the program with debug information before it is debugged. The runner's own
// build stage can replace the default one.
type Compiler interface {
	// CompileForDebug returns the executable. When the compiler rejects the code it returns its
	// output with an error that wraps ErrBuildFailed; a missing compiler is app.MissingTool.
	CompileForDebug(ctx context.Context, config domain.RunConfiguration) (exe, output string, err error)
}

// sourceExtensions are the C++ source files of a project folder.
var sourceExtensions = map[string]bool{".cpp": true, ".cc": true, ".cxx": true}

// buildCompiler is the default Compiler: one compiler call with the flags of cpp.CompileFlags.
type buildCompiler struct{ locator *cpp.Locator }

var _ Compiler = buildCompiler{}

func (b buildCompiler) CompileForDebug(ctx context.Context, config domain.RunConfiguration) (string, string, error) {
	compiler, err := b.locator.Compiler(ctx)
	if err != nil {
		return "", "", err
	}
	folder, sources, err := sourcesOf(config)
	if err != nil {
		return "", "", err
	}
	exe, err := debugExecutable(folder, config)
	if err != nil {
		return "", "", err
	}
	args := append(cpp.CompileFlags(compiler.Family, runtime.GOOS), "-o", exe)
	command := exec.CommandContext(ctx, compiler.Path, append(args, sources...)...)
	command.Dir = folder
	process.HideConsole(command)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return "", "", ctx.Err()
	}
	if err != nil {
		return "", string(output), fmt.Errorf("%w: %w", ErrBuildFailed, err)
	}
	return exe, string(output), nil
}

// sourcesOf is the folder of the program and the files to compile: the whole folder for a
// project, the one file otherwise.
func sourcesOf(config domain.RunConfiguration) (folder string, sources []string, err error) {
	target, err := filepath.Abs(config.Target)
	if err != nil {
		return "", nil, err
	}
	if config.Mode != domain.RunProject {
		return filepath.Dir(target), []string{target}, nil
	}
	entries, err := os.ReadDir(target)
	if err != nil {
		return "", nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() && sourceExtensions[strings.ToLower(filepath.Ext(entry.Name()))] {
			sources = append(sources, filepath.Join(target, entry.Name()))
		}
	}
	if len(sources) == 0 {
		return "", nil, fmt.Errorf("no C++ sources in %s", target)
	}
	return target, sources, nil
}

// debugExecutable is <cache>/VizcachaIDE/build/debug/<hash of the folder>/<name>, away from the
// student's folder.
func debugExecutable(folder string, config domain.RunConfiguration) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	sum := sha1.Sum([]byte(strings.ToLower(folder))) //nolint:gosec // see import
	dir := filepath.Join(cache, "VizcachaIDE", "build", "debug", hex.EncodeToString(sum[:6]))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	name := filepath.Base(folder)
	if config.Mode != domain.RunProject {
		name = strings.TrimSuffix(filepath.Base(config.Target), filepath.Ext(config.Target))
	}
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name), nil
}
