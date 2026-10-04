package process

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// startSession starts one stage in the mode the job asks for.
func startSession(ctx context.Context, job Job, sink JobEvents) (session, error) {
	if job.Mode == Terminal {
		return startTerminal(ctx, job, sink)
	}
	return startPipes(ctx, job, sink)
}

// supervise follows the stages of a run, reports the end and frees the slot.
func (s *Supervisor) supervise(ctx context.Context, occupant *run, job Job) {
	began := time.Now()
	exitCode := s.watch(occupant, occupant.current, job)
	for {
		next := s.nextStage(occupant, job, exitCode)
		if next == nil {
			break
		}
		started, err := startSession(ctx, *next, occupant.events)
		if err != nil {
			cleanup(*next)
			occupant.events.RunOutput("stderr", fmt.Sprintf("%v\n", err))
			exitCode = 1
			break
		}
		s.mu.Lock()
		occupant.current = started
		s.mu.Unlock()
		job = *next
		exitCode = s.watch(occupant, started, job)
	}
	s.mu.Lock()
	s.run = nil
	s.mu.Unlock()
	occupant.events.RunFinished(exitCode, time.Since(began).Milliseconds())
}

// nextStage asks the job for its successor, unless the user stopped the run.
func (s *Supervisor) nextStage(occupant *run, job Job, exitCode int) *Job {
	s.mu.Lock()
	stopped := occupant.stopped
	s.mu.Unlock()
	if stopped || job.Then == nil {
		return nil
	}
	next, ok := job.Then(exitCode)
	if !ok {
		return nil
	}
	return next
}

// watch waits for one stage, shows its silent notice if needed and cleans up.
func (s *Supervisor) watch(occupant *run, current session, job Job) int {
	var over atomic.Bool
	var notice *time.Timer
	if job.Notice != nil && job.Notice.Text != nil {
		delay := job.Notice.Delay
		if delay <= 0 {
			delay = defaultNoticeDelay
		}
		notice = time.AfterFunc(delay, func() { announce(occupant.events, current, job.Notice, &over) })
	}
	exitCode := current.wait()
	over.Store(true)
	if notice != nil {
		notice.Stop()
	}
	cleanup(job)
	return exitCode
}

// announce tells the user why nothing has been printed yet. The text travels as ordinary
// stdout output: the event contract has no separate channel.
func announce(events JobEvents, current session, notice *SilentNotice, over *atomic.Bool) {
	if over.Load() || current.seen() {
		return
	}
	events.RunOutput("stdout", notice.Text()+"\n")
}
