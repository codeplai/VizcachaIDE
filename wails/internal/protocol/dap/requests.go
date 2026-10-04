package dap

import (
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

const clientID = "vizcacha"

// newRequest builds the header every DAP request starts with.
func newRequest(command string) dap.Request {
	return dap.Request{ProtocolMessage: dap.ProtocolMessage{Type: "request"}, Command: command}
}

func initializeRequest(adapterID string) *dap.InitializeRequest {
	return &dap.InitializeRequest{
		Request: newRequest("initialize"),
		Arguments: dap.InitializeRequestArguments{
			ClientID:                     clientID,
			ClientName:                   "VizcachaIDE",
			AdapterID:                    adapterID,
			LinesStartAt1:                true,
			ColumnsStartAt1:              true,
			PathFormat:                   "path",
			SupportsVariableType:         true,
			SupportsRunInTerminalRequest: true,
		},
	}
}

func setBreakpointsRequest(file string, points []domain.Breakpoint) *dap.SetBreakpointsRequest {
	wanted := make([]dap.SourceBreakpoint, 0, len(points))
	for _, point := range points {
		wanted = append(wanted, dap.SourceBreakpoint{Line: point.Location.Line, Condition: point.Condition})
	}
	return &dap.SetBreakpointsRequest{
		Request: newRequest("setBreakpoints"),
		Arguments: dap.SetBreakpointsArguments{
			Source:      dap.Source{Name: filepath.Base(file), Path: file},
			Breakpoints: wanted,
		},
	}
}

func exceptionBreakpointsRequest(filters []string) *dap.SetExceptionBreakpointsRequest {
	return &dap.SetExceptionBreakpointsRequest{
		Request:   newRequest("setExceptionBreakpoints"),
		Arguments: dap.SetExceptionBreakpointsArguments{Filters: filters},
	}
}
