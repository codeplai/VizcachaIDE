package dap

import (
	"context"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/google/go-dap"
)

const (
	stackLevels          = 50
	threadLocationLimit  = 50
	threadLocationLevels = 1
)

// inspect builds the DebugState after a stop: threads, then the stack, scopes and
// Locals of the top user frame, plus the top frame of up to 50 threads.
// If the program resumes first (ctx cancelled) nothing is emitted.
func (s *Session) inspect(ctx context.Context, event *dap.StoppedEvent) {
	reason, description := s.flavor.StopReason(event)
	threads := s.threads(ctx)
	threadID := event.Body.ThreadId
	if threadID == 0 && len(threads) > 0 {
		threadID = threads[0].ThreadID
	}
	locations := s.threadLocations(ctx, threads, threadID)
	frames := s.frames(ctx, threadID, reason == domain.StopException)
	variables := s.locals(ctx, frames)
	if ctx.Err() != nil {
		return
	}
	variables = s.tracker.Mark(frames, variables)
	for index := range threads {
		threads[index].Location = locations[threads[index].ThreadID]
		if threads[index].ThreadID == threadID && len(frames) > 0 {
			threads[index].Location = frames[0].Location // the user frame, not the panic machinery
		}
	}
	s.sink.DebugStopped(domain.DebugState{
		Reason:        reason,
		Frames:        frames,
		Variables:     variables,
		Threads:       threads,
		CurrentThread: &threadID,
		Description:   description,
	})
}

func (s *Session) threads(ctx context.Context) []domain.Thread {
	response, err := s.call(ctx, &dap.ThreadsRequest{Request: newRequest("threads")})
	if err != nil {
		return []domain.Thread{}
	}
	return mapThreads(response.(*dap.ThreadsResponse).Body.Threads)
}

// threadLocations asks for the top frame of each thread, in parallel.
func (s *Session) threadLocations(ctx context.Context, threads []domain.Thread, current int) map[int]*domain.SourceLocation {
	locations := map[int]*domain.SourceLocation{}
	var mu sync.Mutex
	var group sync.WaitGroup
	for _, thread := range threads[:min(len(threads), threadLocationLimit)] {
		if thread.ThreadID == current {
			continue
		}
		group.Add(1)
		go func() {
			defer group.Done()
			frames := mapFrames(s.stack(ctx, thread.ThreadID, threadLocationLevels), s.flavor, false)
			if len(frames) == 0 {
				return
			}
			mu.Lock()
			locations[thread.ThreadID] = frames[0].Location
			mu.Unlock()
		}()
	}
	group.Wait()
	return locations
}

func (s *Session) stack(ctx context.Context, threadID, levels int) []dap.StackFrame {
	request := &dap.StackTraceRequest{
		Request:   newRequest("stackTrace"),
		Arguments: dap.StackTraceArguments{ThreadId: threadID, Levels: levels},
	}
	response, err := s.call(ctx, request)
	if err != nil {
		return nil
	}
	return response.(*dap.StackTraceResponse).Body.StackFrames
}

func (s *Session) frames(ctx context.Context, threadID int, skipHidden bool) []domain.StackFrame {
	return mapFrames(s.stack(ctx, threadID, stackLevels), s.flavor, skipHidden)
}

// locals returns the Locals of the top frame, or an empty list.
func (s *Session) locals(ctx context.Context, frames []domain.StackFrame) []domain.Variable {
	if len(frames) == 0 {
		return []domain.Variable{}
	}
	scopes := &dap.ScopesRequest{Request: newRequest("scopes"), Arguments: dap.ScopesArguments{FrameId: frames[0].FrameID}}
	response, err := s.call(ctx, scopes)
	if err != nil {
		return []domain.Variable{}
	}
	reference := localsReference(response.(*dap.ScopesResponse).Body.Scopes, s.flavor)
	if reference == 0 {
		return []domain.Variable{}
	}
	return s.children(ctx, reference)
}

// children asks for the variables behind a reference.
func (s *Session) children(ctx context.Context, reference int) []domain.Variable {
	request := &dap.VariablesRequest{
		Request:   newRequest("variables"),
		Arguments: dap.VariablesArguments{VariablesReference: reference},
	}
	response, err := s.call(ctx, request)
	if err != nil {
		return []domain.Variable{}
	}
	return mapVariables(response.(*dap.VariablesResponse).Body.Variables, s.flavor)
}
