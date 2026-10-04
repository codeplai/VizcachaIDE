package python

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// moduleDistributions maps each module tool to the pip distribution that provides it.
var moduleDistributions = map[string]string{
	ToolDebugpy: "debugpy",
	ToolPylsp:   "python-lsp-server",
	ToolRuff:    "ruff",
}

const modulesScript = `import json, sys
from importlib import metadata
found = {}
for tool, dist in json.loads(sys.argv[1]).items():
    try:
        found[tool] = metadata.version(dist)
    except metadata.PackageNotFoundError:
        found[tool] = ""
print(json.dumps(found))`

// ModuleVersions asks the interpreter, in one call, which versions of debugpy, pylsp and ruff it
// has: tool id -> version, "" for a module that is not installed. It returns nil when the
// interpreter does not answer.
func ModuleVersions(ctx context.Context, interpreter Interpreter, env map[string]string) map[string]string {
	wanted, err := json.Marshal(moduleDistributions)
	if err != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, interpreter.Path, "-c", modulesScript, string(wanted))
	cmd.Env = EnvironmentList(env)
	process.HideConsole(cmd)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	versions := map[string]string{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &versions); err != nil {
		return nil
	}
	return versions
}
