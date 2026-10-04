package lldb

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// launchArguments are the fields lldb-dap reads from the DAP "launch" request.
type launchArguments struct {
	Request       string            `json:"request"`
	Program       string            `json:"program"`
	Cwd           string            `json:"cwd"`
	Args          []string          `json:"args"`
	Env           map[string]string `json:"env,omitempty"`
	StopOnEntry   bool              `json:"stopOnEntry"`
	RunInTerminal bool              `json:"runInTerminal"`
	InitCommands  []string          `json:"initCommands,omitempty"`
}

func launchRaw(options Options, config domain.RunConfiguration, env map[string]string) (json.RawMessage, error) {
	arguments := launchArguments{
		Request: "launch", Program: options.Program, Cwd: options.workingDirectory(),
		Args: append([]string{}, config.ProgramArgs...), Env: env,
		RunInTerminal: options.RunInTerminal, InitCommands: options.InitCommands,
	}
	raw, err := json.Marshal(arguments)
	if err != nil {
		return nil, fmt.Errorf("encoding launch arguments: %w", err)
	}
	return raw, nil
}

func (o Options) workingDirectory() string {
	if o.Dir != "" {
		return o.Dir
	}
	return filepath.Dir(o.Program)
}
