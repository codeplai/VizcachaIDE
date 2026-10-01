package delve

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

const (
	connectTimeout    = 15 * time.Second
	requestTimeout    = 30 * time.Second
	disconnectTimeout = time.Second
	failedExitCode    = 1
)

// session is one debugging run: one Delve process, one connection, one program.
// A new session is created for every Start, so late signals of an old one are harmless.
type session struct {
	sink     app.EventSink
	book     *app.BreakpointBook
	tracker  *app.ChangeTracker
	process  *adapterProcess
	text     func(key string) string
	onFinish func()

	mu         sync.Mutex
	client     *Client
	threadID   int
	exitCode   int
	configured bool
	finished   bool
	stopCtx    context.Context // alive while the program is paused; nil while it runs
	stopCancel context.CancelFunc

	ready     chan struct{} // closed by the DAP "initialized" event
	readyOnce sync.Once
	done      chan struct{} // closed when the session finishes
}

func newSession(sink app.EventSink, book *app.BreakpointBook, tracker *app.ChangeTracker, process *adapterProcess) *session {
	return &session{
		sink:    sink,
		book:    book,
		tracker: tracker,
		process: process,
		ready:   make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// begin connects to Delve and walks initialize, launch, breakpoints, configurationDone.
func (s *session) begin(config domain.RunConfiguration, environment map[string]string) {
	if err := s.connect(); err != nil {
		s.fail(err)
		return
	}
	if err := s.initialize(); err != nil {
		s.fail(err)
		return
	}
	launch, err := launchRequest(config, environment, debugBinaryPath(osProcessID()))
	if err != nil {
		s.fail(err)
		return
	}
	// Delve answers launch after "initialized"; a build error shows up here.
	go func() {
		if _, err := s.call(context.Background(), launch); err != nil {
			s.fail(err)
		}
	}()
	select {
	case <-s.ready:
	case <-s.done:
		return
	}
	if err := s.configure(); err != nil {
		s.fail(err)
	}
}

func (s *session) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	address, err := s.process.waitAddress(ctx)
	if err != nil {
		return err
	}
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("connecting to delve: %w", err)
	}
	client := NewClient(conn, s.handleEvent, s.connectionClosed)
	s.mu.Lock()
	s.client = client
	s.mu.Unlock()
	return nil
}

func (s *session) initialize() error {
	_, err := s.call(context.Background(), initializeRequest())
	return err
}

func (s *session) configure() error {
	s.mu.Lock()
	s.configured = true
	s.mu.Unlock()
	for _, file := range s.book.Files() {
		if err := s.sendBreakpoints(file); err != nil {
			return err
		}
	}
	configurationDone := &dap.ConfigurationDoneRequest{Request: newRequest("configurationDone")}
	_, err := s.call(context.Background(), configurationDone)
	return err
}

func (s *session) sendBreakpoints(file string) error {
	_, err := s.call(context.Background(), setBreakpointsRequest(file, s.book.For(file)))
	return err
}

// call sends one request with a timeout.
func (s *session) call(ctx context.Context, request dap.RequestMessage) (dap.ResponseMessage, error) {
	s.mu.Lock()
	client := s.client
	s.mu.Unlock()
	if client == nil {
		return nil, errConnectionClosed
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	return client.Call(ctx, request)
}

func (s *session) isConfigured() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.configured && !s.finished
}
