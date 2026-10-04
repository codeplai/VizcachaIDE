package dap

import (
	"context"
	"encoding/json"
	"io"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

// Transport connects to a debug adapter: TCP for Delve, stdio for lldb-dap and debugpy.adapter.
type Transport interface {
	// Open starts or reaches the adapter and returns the connection to speak DAP over.
	Open(ctx context.Context) (io.ReadWriteCloser, error)
	// Close stops the adapter process; it is safe to call more than once.
	Close() error
}

// ReverseHandler answers requests the adapter sends to the client.
type ReverseHandler interface {
	RunInTerminal(args dap.RunInTerminalRequestArguments) (processID int, err error)
}

// Flavor is everything that differs between debug adapters.
type Flavor interface {
	AdapterID() string // "go", "python", "lldb"
	Launch(config domain.RunConfiguration, env map[string]string) (json.RawMessage, error)
	ExceptionFilters() []string // debugpy: ["uncaught"]; Delve, lldb: nil
	// StopReason maps a stop to the domain reason and the text shown to the user.
	StopReason(event *dap.StoppedEvent) (domain.StopReason, string)
	KeepFrame(frame dap.StackFrame) bool     // false drops a frame (Delve's "subtle" runtime frames)
	KeepVariable(variable dap.Variable) bool // false hides a variable (debugpy's "special variables")
	IsLocalsScope(name string) bool          // the scope whose variables are the frame's locals
	Output(event *dap.OutputEvent) (text, category string, ok bool)
}

// SessionDeps groups what a session needs.
type SessionDeps struct {
	Sink      app.EventSink
	Book      *app.BreakpointBook
	Tracker   *app.ChangeTracker
	Transport Transport
	Flavor    Flavor
	Reverse   ReverseHandler          // nil: every reverse request is answered "unsupported"
	Texts     func(key string) string // backend i18n, injected: the protocol knows keys, not languages
}
