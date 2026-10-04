package debugpy

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Console values of the debugpy launch request.
const (
	consoleTerminal = "integratedTerminal" // debugpy asks the IDE to run the program: runInTerminal
	consoleInternal = "internalConsole"    // debugpy runs it itself, without a keyboard
)

// launchArguments are the fields debugpy reads from the DAP "launch" request.
type launchArguments struct {
	Request        string            `json:"request"`
	Type           string            `json:"type"`
	Name           string            `json:"name"`
	Program        string            `json:"program"`
	Cwd            string            `json:"cwd"`
	Args           []string          `json:"args"`
	Python         []string          `json:"python"`
	Console        string            `json:"console"`
	JustMyCode     bool              `json:"justMyCode"`
	StopOnEntry    bool              `json:"stopOnEntry"`
	RedirectOutput bool              `json:"redirectOutput"`
	Env            map[string]string `json:"env,omitempty"`
}

func launchRaw(config domain.RunConfiguration, env map[string]string, interpreter, console string) (json.RawMessage, error) {
	arguments := launchArguments{
		Request: "launch", Type: "python", Name: "VizcachaIDE",
		Program: programPath(config), Cwd: workingDirectory(config),
		Args:   append([]string{}, config.ProgramArgs...),
		Python: []string{interpreter}, Console: console,
		JustMyCode: true, Env: env,
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

func programPath(config domain.RunConfiguration) string {
	if absolute, err := filepath.Abs(config.Target); err == nil {
		return absolute
	}
	return config.Target
}
