// Package process runs and supervises the one program the IDE runs at a time, whatever its
// language: a compiler stage followed by the program, or a package command.
//
// A Supervisor has a single slot, so "one program at a time" holds for the whole IDE by
// construction: a second Start gets app.ErrBusy no matter which language asks. The supervisor
// emits the run events (run:started, run:output, run:finished) through the EventSink.
package process

import (
	"context"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Mode says how the program is connected to the IDE.
type Mode int

// Mode values.
const (
	// Pipes connects stdin, stdout and stderr through pipes (the Go runner).
	Pipes Mode = iota
	// Terminal runs the program in a pseudoterminal (Python and C++): its output is line
	// buffered, it survives a crash and the terminal echoes what the user types.
	Terminal
)

const defaultNoticeDelay = 5 * time.Second

// SilentNotice is printed as stdout when nothing arrived after Delay (Go's "first build",
// C++'s "Compiling...").
type SilentNotice struct {
	// Text returns the already translated text.
	Text func() string
	// Delay defaults to 5 seconds.
	Delay time.Duration
}

// Job is one stage of a run: a command with its arguments.
type Job struct {
	// Config is what run:started announces.
	Config  domain.RunConfiguration
	Command string
	Args    []string
	Dir     string
	// Env is the full environment of the command; nil means the IDE's own environment.
	Env  map[string]string
	Mode Mode
	// Cleanup runs once when this stage ends or fails to start (it may be nil).
	Cleanup func()
	// Notice is shown when the stage stays silent for a while; nil when there is none.
	Notice *SilentNotice
	// Then is called with the exit code of this stage. It returns the next stage (a compile
	// stage is followed by the program), or false to finish. run:finished carries the exit code
	// of the last stage that ran.
	Then func(exitCode int) (*Job, bool)
	// Events receives the start, output and end of the run instead of the supervisor's sink
	// (run:started, run:output, run:finished). A debug adapter's runInTerminal passes one that
	// turns the output into debug:output. Only the first stage's Events counts. Nil = the sink.
	Events JobEvents
}

// JobEvents is what a run reports. app.EventSink satisfies it, so the default is the sink.
type JobEvents interface {
	RunStarted(config domain.RunConfiguration)
	RunOutput(stream, text string)
	RunFinished(exitCode int, durationMs int64)
}

// session is one running stage.
type session interface {
	// write sends a line typed by the user.
	write(text string) error
	// stop asks the program to end (Ctrl+C) and kills its process tree after a grace period.
	stop()
	// seen reports whether the program has printed anything.
	seen() bool
	// wait blocks until the program ended and its output was delivered; it returns the exit code.
	wait() int
}

// run is the occupant of the slot.
type run struct {
	current session
	stopped bool
	events  JobEvents
}

// Supervisor owns the single slot.
type Supervisor struct {
	sink app.EventSink
	mu   sync.Mutex
	run  *run
}

// New creates the supervisor every runner and package manager of the IDE shares.
func New(sink app.EventSink) *Supervisor { return &Supervisor{sink: sink} }

// Start launches the first stage and supervises the run in the background. It returns
// app.ErrBusy while another run is alive.
func (s *Supervisor) Start(ctx context.Context, job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.run != nil {
		cleanup(job)
		return app.ErrBusy
	}
	events := s.eventsOf(job)
	first, err := startSession(ctx, job, events)
	if err != nil {
		cleanup(job)
		return err
	}
	occupant := &run{current: first, events: events}
	s.run = occupant
	events.RunStarted(job.Config)
	go s.supervise(ctx, occupant, job)
	return nil
}

// eventsOf returns where the job reports: its own Events, or the supervisor's sink.
func (s *Supervisor) eventsOf(job Job) JobEvents {
	if job.Events != nil {
		return job.Events
	}
	return s.sink
}

// IsRunning reports whether a run is alive.
func (s *Supervisor) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.run != nil
}

// Stop interrupts the program (Ctrl+C semantics: defers and signal handlers run) and kills its
// whole process tree if it is still alive after about 2 seconds. It does nothing when idle.
func (s *Supervisor) Stop() error {
	s.mu.Lock()
	occupant := s.run
	if occupant == nil {
		s.mu.Unlock()
		return nil
	}
	occupant.stopped = true
	current := occupant.current
	s.mu.Unlock()
	current.stop()
	return nil
}

// WriteInput types text and Enter into the running program.
func (s *Supervisor) WriteInput(text string) error {
	s.mu.Lock()
	var current session
	if s.run != nil {
		current = s.run.current
	}
	s.mu.Unlock()
	if current == nil {
		return nil
	}
	return current.write(text)
}

func cleanup(job Job) {
	if job.Cleanup != nil {
		job.Cleanup()
	}
}
