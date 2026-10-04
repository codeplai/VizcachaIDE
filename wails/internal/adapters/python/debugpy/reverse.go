package debugpy

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

// jobStarter is the part of process.Supervisor the handler needs.
type jobStarter interface {
	Start(ctx context.Context, job process.Job) error
}

// debugEvents sends the output of the debugged program as debug:output. The start and end of
// the run are not announced: debug:terminated says when it is over.
type debugEvents struct {
	sink     app.EventSink
	started  atomic.Bool
	finished atomic.Bool
}

var _ process.JobEvents = (*debugEvents)(nil)

func (e *debugEvents) RunStarted(domain.RunConfiguration) { e.started.Store(true) }
func (e *debugEvents) RunOutput(stream, text string)      { e.sink.DebugOutput(text, stream) }
func (e *debugEvents) RunFinished(int, int64)             { e.finished.Store(true) }

// alive reports whether the debugged program was started and has not ended.
func (e *debugEvents) alive() bool { return e.started.Load() && !e.finished.Load() }

// reverseHandler answers runInTerminal by running debugpy's launcher on the shared supervisor,
// in a pseudoterminal, so the program can read what the user types in Output.
type reverseHandler struct {
	supervisor jobStarter
	events     *debugEvents
	config     domain.RunConfiguration
	base       map[string]string
}

var _ protodap.ReverseHandler = (*reverseHandler)(nil)

// RunInTerminal starts the command. It returns process id 0: the supervisor does not expose the
// pid, and debugpy does not need it because its launcher connects back to the adapter itself.
func (h *reverseHandler) RunInTerminal(args dap.RunInTerminalRequestArguments) (int, error) {
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
func (h *reverseHandler) mergedEnvironment(requested map[string]any) map[string]string {
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
