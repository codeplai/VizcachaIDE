package cmake

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// ErrConfigureFailed is wrapped when CMake (or its libraries' installation) exits with an error.
// The output is returned beside it: it is what the Assistant explains.
var ErrConfigureFailed = errors.New("CMake could not configure the project")

// ErrBuildFailed is wrapped when the build exits with an error (the compiler rejected the code).
var ErrBuildFailed = errors.New("the build failed")

// inputs are the files whose change makes the build folder stale.
var inputs = []string{ListsFile, ManifestFile, PresetsFile}

// Configure configures the Debug build folder (root/build) the way the IDE runs, debugs and
// checks: Ninja, Debug, the compiler and Ninja of the Builder and, when there are Dependencies,
// their arguments and environment. It does nothing when the folder is already configured and
// newer than CMakeLists.txt, vcpkg.json and CMakePresets.json. The output is what CMake and the
// libraries' installation printed; a failure wraps ErrConfigureFailed.
func (b *Builder) Configure(ctx context.Context, root string) (output string, err error) {
	return b.configure(ctx, root, false)
}

// ConfigureRelease is Configure for root/build/release, the folder "Build executable" uses.
func (b *Builder) ConfigureRelease(ctx context.Context, root string) (string, error) {
	return b.configure(ctx, root, true)
}

func (b *Builder) configure(ctx context.Context, root string, release bool) (string, error) {
	plan, err := b.Plan(ctx, root, release)
	if err != nil || plan.Configure == nil {
		return "", err
	}
	return capture(ctx, *plan.Configure, ErrConfigureFailed)
}

// Build configures if needed and builds the project, waiting for it. It returns everything the
// tools printed; a failure wraps ErrConfigureFailed or ErrBuildFailed.
func (b *Builder) Build(ctx context.Context, root string, release bool) (string, error) {
	plan, err := b.Plan(ctx, root, release)
	if err != nil {
		return "", err
	}
	var output string
	if plan.Configure != nil {
		if output, err = capture(ctx, *plan.Configure, ErrConfigureFailed); err != nil {
			return output, err
		}
	}
	built, err := capture(ctx, plan.Build, ErrBuildFailed)
	return output + built, err
}

// capture runs a command outside the supervisor and returns its stdout and stderr together. A
// non-zero exit wraps failure.
func capture(ctx context.Context, command Command, failure error) (string, error) {
	cmd := exec.CommandContext(ctx, command.Path, command.Args...)
	cmd.Dir, cmd.Env = command.Dir, command.Env
	process.HideConsole(cmd)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return output.String(), fmt.Errorf("%w (exit code %d)", failure, exit.ExitCode())
	}
	if err != nil {
		return output.String(), fmt.Errorf("start %s: %w", filepath.Base(command.Path), err)
	}
	return output.String(), nil
}

// needsConfigure reports whether the build folder has to be configured: it never was, CMake has
// not written the File API reply, the project moved, or one of its input files is newer.
func needsConfigure(root, buildDir string) bool {
	cache := filepath.Join(buildDir, "CMakeCache.txt")
	info, err := os.Stat(cache)
	if err != nil || latestReply(buildDir) == "" {
		return true
	}
	if movedProject(cache, root) {
		_ = os.RemoveAll(buildDir) // a disposable folder: CMake refuses to reuse it in another place
		return true
	}
	for _, name := range inputs {
		if input, err := os.Stat(filepath.Join(root, name)); err == nil && !info.ModTime().After(input.ModTime()) {
			return true
		}
	}
	return false
}

// movedProject reports whether the cache was made for another source folder (the student renamed
// or copied the project).
func movedProject(cache, root string) bool {
	data, err := os.ReadFile(cache)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if home, ok := strings.CutPrefix(strings.TrimSpace(line), "CMAKE_HOME_DIRECTORY:INTERNAL="); ok {
			return !sameFolder(home, root)
		}
	}
	return false
}
