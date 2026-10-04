package process

import (
	"strings"
	"sync"
)

// OutputFilter transforms or drops one line a stage prints (without its line break). It returns
// the text to show and whether to show it. Rust compiles with --error-format=json and shows each
// diagnostic's "rendered" text instead of the JSON (docs/PLAN_RUST.md sections 3.1 and 3.2).
type OutputFilter func(stream, line string) (string, bool)

// filteredEvents passes the output of a Pipes stage through an OutputFilter one complete line at
// a time; start and end go through unchanged.
type filteredEvents struct {
	JobEvents
	filter  OutputFilter
	mu      sync.Mutex
	pending map[string]string
}

func newFilteredEvents(inner JobEvents, filter OutputFilter) *filteredEvents {
	return &filteredEvents{JobEvents: inner, filter: filter, pending: map[string]string{}}
}

// RunOutput keeps an unfinished last line until its break arrives.
func (f *filteredEvents) RunOutput(stream, text string) {
	f.mu.Lock()
	data := f.pending[stream] + text
	cut := strings.LastIndexByte(data, '\n')
	f.pending[stream] = data[cut+1:]
	f.mu.Unlock()
	if cut < 0 {
		return
	}
	for _, line := range strings.SplitAfter(data[:cut+1], "\n") {
		f.emit(stream, line)
	}
}

// flush sends what is left without a line break when the stage ends.
func (f *filteredEvents) flush() {
	f.mu.Lock()
	rest := f.pending
	f.pending = map[string]string{}
	f.mu.Unlock()
	for _, stream := range []string{"stdout", "stderr"} {
		f.emit(stream, rest[stream])
	}
}

func (f *filteredEvents) emit(stream, line string) {
	if line == "" {
		return
	}
	body := strings.TrimRight(line, "\r\n")
	shown, keep := f.filter(stream, body)
	if !keep {
		return
	}
	if strings.HasSuffix(line, "\n") && !strings.HasSuffix(shown, "\n") {
		shown += "\n"
	}
	f.JobEvents.RunOutput(stream, shown)
}
