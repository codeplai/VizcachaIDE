package lldbdap

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// panicBreakpoint is the function where the Rust runtime starts to unwind after it printed the
// message of a panic. It is created first, so it is breakpoint 1 (see flavor.go).
const panicBreakpoint = "rust_panic"

// InitCommands are the LLDB commands run before the program starts. The formatters are what
// `rust-lldb` loads: rustc 1.99 has no lldb_commands file, lldb_lookup.py registers every
// synthetic provider and summary itself in __lldb_init_module, so one "command script import"
// is all it takes. They need the Python of LLDB: python=false skips them (the values then
// show raw). enumsFix, when not "", is the IDE's script that corrects the enum payloads of LLDB 23
// (enums.go); it patches lldb_lookup's providers, so it is imported after them. The breakpoint on
// rust_panic is independent of Python.
func InitCommands(toolchain rust.Toolchain, python bool, enumsFix string) []string {
	commands := []string{"breakpoint set --name " + panicBreakpoint}
	if !python {
		return commands
	}
	lookup := filepath.ToSlash(filepath.Join(toolchain.FormattersDir(), "lldb_lookup.py"))
	commands = append(commands, `command script import "`+lookup+`"`)
	if enumsFix != "" {
		commands = append(commands, `command script import "`+filepath.ToSlash(enumsFix)+`"`)
	}
	return commands
}

// pythonWorks reports whether the lldb next to lldb-dap has Python: "lldb -P" prints the path of
// its Python modules and fails without it. When there is no lldb to ask, it assumes yes (an
// initCommand that fails is harmless).
func pythonWorks(adapter string) bool {
	lldb := filepath.Join(filepath.Dir(adapter), "lldb"+exeSuffix(adapter))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, lldb, "-P")
	process.HideConsole(command)
	output, err := command.Output()
	if err != nil {
		var notStarted *exec.Error
		return errors.As(err, &notStarted)
	}
	return strings.TrimSpace(string(output)) != ""
}

func exeSuffix(adapter string) string {
	if strings.HasSuffix(strings.ToLower(adapter), ".exe") {
		return ".exe"
	}
	return ""
}

// sourceRoot is the folder with the user's source: the workspace of a crate, or the folder of
// the file. Frames outside it (the standard library, the registry) are hidden.
func sourceRoot(config domain.RunConfiguration) string {
	folder := filepath.Dir(config.Target)
	if config.Mode == domain.RunProject {
		folder = config.Target
	}
	if project, ok, err := rust.FindProject(projectProbe(config)); ok && err == nil {
		folder = project.Workspace
	}
	if absolute, err := filepath.Abs(folder); err == nil {
		return absolute
	}
	return folder
}

func workingDirectory(config domain.RunConfiguration) string {
	if config.WorkingDir != "" {
		return config.WorkingDir
	}
	return sourceRoot(config)
}
