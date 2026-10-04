package delve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const debugBinaryName = "vizcacha_debug_bin"

// launchArguments are the fields Delve reads from the DAP "launch" request.
type launchArguments struct {
	Request     string            `json:"request"`
	Mode        string            `json:"mode"`
	Program     string            `json:"program"`
	Cwd         string            `json:"cwd"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env,omitempty"`
	Output      string            `json:"output"`
	OutputMode  string            `json:"outputMode"`
	StopOnEntry bool              `json:"stopOnEntry"`
}

// launchRaw encodes the launch arguments Delve reads.
func launchRaw(config domain.RunConfiguration, environment map[string]string, output string) (json.RawMessage, error) {
	arguments := launchArguments{
		Request:    "launch",
		Mode:       "debug",
		Program:    programPath(config),
		Cwd:        workingDirectory(config),
		Args:       append([]string{}, config.ProgramArgs...),
		Env:        environment,
		Output:     output,
		OutputMode: "remote", // the program's stdout and stderr arrive as DAP output events
	}
	raw, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("encoding launch arguments: %w", err)
	}
	return raw, nil
}

func workingDirectory(config domain.RunConfiguration) string {
	if config.WorkingDir != "" {
		return config.WorkingDir
	}
	return filepath.Dir(config.Target)
}

// programPath is the file or package folder Delve builds.
func programPath(config domain.RunConfiguration) string {
	path := filepath.Join(workingDirectory(config), app.GoTargetArgument(config))
	if absolute, err := filepath.Abs(path); err == nil {
		return absolute
	}
	return path
}

// debugBinaryPath is where Delve writes the debug build, so it never litters
// the user's folder.
func debugBinaryPath(processID int) string {
	name := debugBinaryName + "_" + strconv.Itoa(processID)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(os.TempDir(), name)
}

// removeStaleDebugBinaries deletes the debug builds that earlier sessions left
// behind. Files still locked by another running IDE are skipped.
func removeStaleDebugBinaries(currentProcessID int, directory string) []string {
	matches, err := filepath.Glob(filepath.Join(directory, debugBinaryName+"_*"))
	if err != nil {
		return nil
	}
	own := debugBinaryName + "_" + strconv.Itoa(currentProcessID)
	var removed []string
	for _, path := range matches {
		stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if stem == own {
			continue
		}
		if os.Remove(path) == nil {
			removed = append(removed, path)
		}
	}
	return removed
}
