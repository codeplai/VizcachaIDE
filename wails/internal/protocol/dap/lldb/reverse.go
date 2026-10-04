package lldb

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
	"github.com/google/go-dap"
)

// JobStarter is the part of process.Supervisor the reverse handler needs.
type JobStarter interface {
	Start(ctx context.Context, job process.Job) error
}

// Events sends the output of the debugged program as debug:output. The start and end of the
// run are not announced: debug:terminated says when it is over.
type Events struct {
	sink     app.EventSink
	started  atomic.Bool
	finished atomic.Bool
}

var _ process.JobEvents = (*Events)(nil)

// NewEvents creates the receiver of the program's output.
func NewEvents(sink app.EventSink) *Events { return &Events{sink: sink} }

func (e *Events) RunStarted(domain.RunConfiguration) { e.started.Store(true) }
func (e *Events) RunOutput(stream, text string)      { e.sink.DebugOutput(text, stream) }
func (e *Events) RunFinished(int, int64)             { e.finished.Store(true) }

// Alive reports whether the debugged program was started and has not ended.
func (e *Events) Alive() bool { return e.started.Load() && !e.finished.Load() }

// Reverse answers runInTerminal by running lldb-dap's launcher on the shared supervisor, in a
// pseudoterminal, so the program can read what the user types in Output.
type Reverse struct {
	supervisor JobStarter
	events     *Events
	config     domain.RunConfiguration
	base       map[string]string
}

var _ protodap.ReverseHandler = (*Reverse)(nil)

// NewReverse creates the handler. base is the environment of the program; config is only what
// the supervisor labels the job with.
func NewReverse(supervisor JobStarter, events *Events, config domain.RunConfiguration, base map[string]string) *Reverse {
	return &Reverse{supervisor: supervisor, events: events, config: config, base: base}
}

// RunInTerminal starts the launcher. It returns process id 0: the supervisor does not expose the
// pid and the launcher tells lldb-dap its own through the comm file.
func (h *Reverse) RunInTerminal(args dap.RunInTerminalRequestArguments) (int, error) {
	if len(args.Args) == 0 {
		return 0, errors.New("runInTerminal without a command")
	}
	job := process.Job{
		Config: h.config, Command: args.Args[0], Args: args.Args[1:], Dir: args.Cwd,
		Env: h.mergedEnvironment(args.Env), Mode: process.Terminal, Events: h.events,
	}
	if err := h.supervisor.Start(context.Background(), job); err != nil {
		return 0, err
	}
	return 0, nil
}

// mergedEnvironment is the base environment plus what the adapter asks for; a null value
// removes the variable.
func (h *Reverse) mergedEnvironment(requested map[string]any) map[string]string {
	env := make(map[string]string, len(h.base)+len(requested))
	for name, value := range h.base {
		env[name] = value
	}
	for name, value := range requested {
		text, isText := value.(string)
		if !isText {
			delete(env, name)
			continue
		}
		env[name] = text
	}
	return env
}
