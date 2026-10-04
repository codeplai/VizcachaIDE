package python

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// Probe asks an interpreter for its version ("3.12.14"); it fails when the program does not
// answer like a Python (the Microsoft Store alias, a broken venv, a file that is not Python).
type Probe func(ctx context.Context, path string) (string, error)

const probeTimeout = 5 * time.Second

// Minimum supported version: 3.10.
const minimumMajor, minimumMinor = 3, 10

const versionScript = "import sys; print('.'.join(map(str, sys.version_info[:3])))"

// runProbe runs `<python> -c <versionScript>` with a time limit.
func runProbe(ctx context.Context, path string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "-c", versionScript)
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("probing %s: %w", path, err)
	}
	version := strings.TrimSpace(string(output))
	if _, _, ok := majorMinor(version); !ok {
		return "", fmt.Errorf("probing %s: unexpected answer %q", path, version)
	}
	return version, nil
}

// launcherInterpreter asks the Windows "py" launcher which Python 3 it would run, or "".
func launcherInterpreter(ctx context.Context, probe Probe) string {
	launcher, err := exec.LookPath("py")
	if err != nil {
		return ""
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, launcher, "-3", "-c", "import sys; print(sys.executable)")
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	path := strings.TrimSpace(string(output))
	if _, err := probe(ctx, path); err != nil {
		return ""
	}
	return path
}

// supportedVersion reports whether "3.12.14" is at least the minimum version.
func supportedVersion(version string) bool {
	major, minor, ok := majorMinor(version)
	if !ok {
		return false
	}
	return major > minimumMajor || (major == minimumMajor && minor >= minimumMinor)
}

func majorMinor(version string) (int, int, bool) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}
	major, errMajor := strconv.Atoi(parts[0])
	minor, errMinor := strconv.Atoi(parts[1])
	return major, minor, errMajor == nil && errMinor == nil
}
