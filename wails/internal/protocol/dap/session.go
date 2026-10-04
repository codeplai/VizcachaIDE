package dap

import (
	"context"
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

// Session is one debugging run: one adapter process, one connection, one program.
// A new session is created for every Start, so late signals of an old one are harmless.
type Session struct {
	sink      app.EventSink
	book      *app.BreakpointBook
	tracker   *app.ChangeTracker
	transport Transport
	flavor    Flavor
	reverse   ReverseHandler
	text      func(key string) string
	onFinish  func()

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

// NewSession creates a session; call Begin to start it.
func NewSession(deps SessionDeps) *Session {
	return &Session{
		sink:      deps.Sink,
		book:      deps.Book,
		tracker:   deps.Tracker,
		transport: deps.Transport,
		flavor:    deps.Flavor,
		reverse:   deps.Reverse,
		text:      deps.Texts,
		onFinish:  func() {},
		ready:     make(chan struct{}),
		done:      make(chan struct{}),
	}
}

// OnFinish registers what runs once when the session ends.
func (s *Session) OnFinish(finished func()) { s.onFinish = finished }

// Begin connects to the adapter and walks initialize, launch, breakpoints, configurationDone.
func (s *Session) Begin(config domain.RunConfiguration, environment map[string]string) {
	if err := s.connect(); err != nil {
		s.fail(err)
		return
	}
	if err := s.initialize(); err != nil {
		s.fail(err)
		return
	}
	launch, err := s.launchRequest(config, environment)
	if err != nil {
		s.fail(err)
		return
	}
	// The adapter answers launch after "initialized"; a build error shows up here.
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

func (s *Session) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()
	conn, err := s.transport.Open(ctx)
	if err != nil {
		return err
	}
	client := NewClient(conn, s.handleEvent, s.connectionClosed, s.reverse)
	s.mu.Lock()
	s.client = client
	s.mu.Unlock()
	return nil
}

func (s *Session) initialize() error {
	_, err := s.call(context.Background(), initializeRequest(s.flavor.AdapterID()))
	return err
}

func (s *Session) configure() error {
	s.mu.Lock()
	s.configured = true
	s.mu.Unlock()
	for _, file := range s.book.Files() {
		if err := s.sendBreakpoints(file); err != nil {
			return err
		}
	}
	if filters := s.flavor.ExceptionFilters(); len(filters) > 0 {
		if _, err := s.call(context.Background(), exceptionBreakpointsRequest(filters)); err != nil {
			return err
		}
	}
	configurationDone := &dap.ConfigurationDoneRequest{Request: newRequest("configurationDone")}
	_, err := s.call(context.Background(), configurationDone)
	return err
}

func (s *Session) sendBreakpoints(file string) error {
	_, err := s.call(context.Background(), setBreakpointsRequest(file, s.book.For(file)))
	return err
}

func (s *Session) launchRequest(config domain.RunConfiguration, environment map[string]string) (*dap.LaunchRequest, error) {
	arguments, err := s.flavor.Launch(config, environment)
	if err != nil {
		return nil, err
	}
	return &dap.LaunchRequest{Request: newRequest("launch"), Arguments: arguments}, nil
}

// call sends one request with a timeout.
func (s *Session) call(ctx context.Context, request dap.RequestMessage) (dap.ResponseMessage, error) {
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

func (s *Session) isConfigured() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.configured && !s.finished
}
