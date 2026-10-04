package lldb

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/google/go-dap"
)

// Flavor holds what is particular to lldb-dap.
type Flavor struct {
	options Options
	roots   []string
}

var _ protodap.Flavor = Flavor{}

// New creates the flavor of one debugging session.
func New(options Options) Flavor {
	roots := make([]string, 0, len(options.Roots))
	for _, root := range options.Roots {
		roots = append(roots, normalize(root))
	}
	return Flavor{options: options, roots: roots}
}

func (Flavor) AdapterID() string { return "lldb" }

func (f Flavor) Launch(config domain.RunConfiguration, env map[string]string) (json.RawMessage, error) {
	return launchRaw(f.options, config, env)
}

// ExceptionFilters: crashes and signals stop the program without any filter. Only the optional
// C++ throw filter needs one.
func (f Flavor) ExceptionFilters() []string {
	if f.options.CatchThrow {
		return []string{"cpp_throw"}
	}
	return nil
}

// StopReason maps lldb-dap's reasons ("breakpoint", "step", "exception", "signal"...). A crash
// shows the same phrase as the runner followed by LLDB's own text.
func (Flavor) StopReason(event *dap.StoppedEvent) (domain.StopReason, string) {
	description := protodap.StopDescription(event)
	if event.Body.Reason == "signal" || event.Body.Reason == "exception" {
		return domain.StopException, crashDescription(description)
	}
	return protodap.StandardStopReason(event.Body.Reason), description
}

// KeepFrame hides frames without a source file and, when there are roots, frames outside them.
func (f Flavor) KeepFrame(frame dap.StackFrame) bool {
	if frame.Source == nil || frame.Source.Path == "" {
		return false
	}
	if len(f.roots) == 0 {
		return true
	}
	path := normalize(frame.Source.Path)
	for _, root := range f.roots {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

func (Flavor) KeepVariable(dap.Variable) bool { return true }

func (Flavor) IsLocalsScope(name string) bool { return name == "Locals" }

func (Flavor) Output(event *dap.OutputEvent) (string, string, bool) {
	return protodap.StandardOutput(event)
}

// normalize makes paths comparable: clean, forward slashes, lower case on Windows.
func normalize(path string) string {
	path = strings.TrimRight(filepath.ToSlash(filepath.Clean(path)), "/")
	if runtime.GOOS == "windows" {
		return strings.ToLower(path)
	}
	return path
}
