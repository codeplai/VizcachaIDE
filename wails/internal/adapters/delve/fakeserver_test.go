package delve

import (
	"bufio"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

// fakeServer plays Delve's side of a connection with the responses of a recorded
// transcript. after lists events to push once a request has been answered.
type fakeServer struct {
	conn      net.Conn
	responses map[string]dap.ResponseMessage
	after     map[string][]dap.Message
	variables [][]dap.Variable // successive answers to "variables"; the last repeats
	mu        sync.Mutex
	asked     []string
}

func newFakeServer(t *testing.T, messages []dap.Message) (*fakeServer, net.Conn) {
	t.Helper()
	clientSide, serverSide := net.Pipe()
	server := &fakeServer{conn: serverSide, responses: map[string]dap.ResponseMessage{}, after: map[string][]dap.Message{}}
	for _, message := range messages {
		if response, ok := message.(dap.ResponseMessage); ok {
			server.responses[response.GetResponse().Command] = response
		}
	}
	go server.serve()
	t.Cleanup(func() { _ = serverSide.Close(); _ = clientSide.Close() })
	return server, clientSide
}

func (f *fakeServer) serve() {
	reader := bufio.NewReader(f.conn)
	for {
		message, err := dap.ReadProtocolMessage(reader)
		if err != nil {
			return
		}
		request, ok := message.(dap.RequestMessage)
		if !ok {
			continue
		}
		f.answer(request)
	}
}

func (f *fakeServer) answer(request dap.RequestMessage) {
	header := request.GetRequest()
	f.mu.Lock()
	f.asked = append(f.asked, header.Command)
	response := f.responses[header.Command]
	f.mu.Unlock()
	if header.Command == "variables" {
		response = f.nextVariables()
	}
	if response == nil {
		response = &dap.ErrorResponse{Response: dap.Response{ProtocolMessage: dap.ProtocolMessage{Type: "response"}, Command: header.Command}}
	}
	response.GetResponse().RequestSeq = header.Seq
	_ = dap.WriteProtocolMessage(f.conn, response)
	for _, event := range f.after[header.Command] {
		_ = dap.WriteProtocolMessage(f.conn, event)
	}
}

func (f *fakeServer) nextVariables() dap.ResponseMessage {
	f.mu.Lock()
	defer f.mu.Unlock()
	body := f.variables[0]
	if len(f.variables) > 1 {
		f.variables = f.variables[1:]
	}
	response := f.responses["variables"].(*dap.VariablesResponse)
	copied := *response
	copied.Body.Variables = body
	return &copied
}

func (f *fakeServer) push(event dap.Message) { _ = dap.WriteProtocolMessage(f.conn, event) }

func (f *fakeServer) requests() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.asked...)
}

// eventSink records what reaches the frontend.
type eventSink struct {
	stops      chan domain.DebugState
	terminated chan int
	outputs    chan string
	variables  chan []domain.Variable
}

var _ app.EventSink = (*eventSink)(nil)

func newEventSink() *eventSink {
	return &eventSink{
		stops: make(chan domain.DebugState, 8), terminated: make(chan int, 4),
		outputs: make(chan string, 16), variables: make(chan []domain.Variable, 4),
	}
}

func (e *eventSink) DebugStopped(state domain.DebugState) { e.stops <- state }
func (e *eventSink) DebugTerminated(code int)             { e.terminated <- code }
func (e *eventSink) DebugOutput(text, _ string)           { e.outputs <- text }
func (e *eventSink) DebugVariables(_ int, variables []domain.Variable) {
	e.variables <- variables
}
func (e *eventSink) RunOutput(string, string)                 {}
func (e *eventSink) RunStarted(domain.RunConfiguration)       {}
func (e *eventSink) RunFinished(int, int64)                   {}
func (e *eventSink) Diagnostics(string, []domain.Diagnostic)  {}
func (e *eventSink) LanguageServerStatus(domain.ServerStatus) {}
func (e *eventSink) Explained([]domain.ExplainedDiagnostic)   {}
func (e *eventSink) SettingsChanged(domain.Settings)          {}

// connectedSession wires a session to the fake server, as if configurationDone was sent.
func connectedSession(t *testing.T, server *fakeServer, conn net.Conn, sink app.EventSink) *session {
	t.Helper()
	s := newSession(sink, app.NewBreakpointBook(), app.NewChangeTracker(), nil)
	s.text = func(key string) string { return englishTexts[key] }
	s.onFinish = func() {}
	s.client = NewClient(conn, s.handleEvent, s.connectionClosed)
	s.configured = true
	return s
}

func receive[T any](t *testing.T, channel <-chan T) T {
	t.Helper()
	select {
	case value := <-channel:
		return value
	// Generous: the real-Delve tests compile a program first, which is slow on a busy machine.
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for an event")
		var zero T
		return zero
	}
}
