package dap

import (
	"bufio"
	"errors"
	"net"
	"testing"

	"github.com/google/go-dap"
)

type terminalHandler struct {
	got       chan dap.RunInTerminalRequestArguments
	processID int
	err       error
}

func (h *terminalHandler) RunInTerminal(args dap.RunInTerminalRequestArguments) (int, error) {
	h.got <- args
	return h.processID, h.err
}

// reverseRequest sends request from a fake adapter and returns the client's answer.
func reverseRequest(t *testing.T, handler ReverseHandler, request dap.RequestMessage) dap.ResponseMessage {
	t.Helper()
	clientSide, adapterSide := net.Pipe()
	t.Cleanup(func() { _ = clientSide.Close(); _ = adapterSide.Close() })
	NewClient(clientSide, func(dap.EventMessage) {}, func() {}, handler)
	request.GetRequest().Seq = 7
	go func() { _ = dap.WriteProtocolMessage(adapterSide, request) }()
	message, err := dap.ReadProtocolMessage(bufio.NewReader(adapterSide))
	if err != nil {
		t.Fatal(err)
	}
	response, ok := message.(dap.ResponseMessage)
	if !ok {
		t.Fatalf("answer = %T, want a response", message)
	}
	return response
}

func runInTerminalRequest() *dap.RunInTerminalRequest {
	return &dap.RunInTerminalRequest{
		Request:   newRequest("runInTerminal"),
		Arguments: dap.RunInTerminalRequestArguments{Cwd: "work", Args: []string{"python", "main.py"}},
	}
}

func TestRunInTerminalCallsTheHandlerAndAnswersWithTheProcessID(t *testing.T) {
	handler := &terminalHandler{got: make(chan dap.RunInTerminalRequestArguments, 1), processID: 4242}

	response := reverseRequest(t, handler, runInTerminalRequest())

	args := <-handler.got
	if args.Cwd != "work" || len(args.Args) != 2 {
		t.Errorf("handler received %+v", args)
	}
	answer, ok := response.(*dap.RunInTerminalResponse)
	if !ok || !answer.Success || answer.Body.ProcessId != 4242 || answer.RequestSeq != 7 {
		t.Errorf("response = %+v", response)
	}
}

func TestRunInTerminalHandlerErrorIsAnErrorResponse(t *testing.T) {
	handler := &terminalHandler{got: make(chan dap.RunInTerminalRequestArguments, 1), err: errors.New("no terminal")}

	response := reverseRequest(t, handler, runInTerminalRequest())

	if _, ok := response.(*dap.ErrorResponse); !ok || response.GetResponse().Success || response.GetResponse().Message != "no terminal" {
		t.Errorf("response = %+v", response)
	}
}

func TestReverseRequestsWithoutHandlerAreUnsupported(t *testing.T) {
	response := reverseRequest(t, nil, runInTerminalRequest())

	base := response.GetResponse()
	if _, ok := response.(*dap.ErrorResponse); !ok || base.Success || base.Message != "unsupported" || base.Command != "runInTerminal" {
		t.Errorf("response = %+v", response)
	}
}

func TestOtherReverseRequestsAreUnsupportedEvenWithAHandler(t *testing.T) {
	handler := &terminalHandler{got: make(chan dap.RunInTerminalRequestArguments, 1)}
	request := &dap.StartDebuggingRequest{Request: newRequest("startDebugging")}

	response := reverseRequest(t, handler, request)

	if response.GetResponse().Success || response.GetResponse().Message != "unsupported" {
		t.Errorf("response = %+v", response)
	}
	if len(handler.got) != 0 {
		t.Error("the handler must not be called for other requests")
	}
}
