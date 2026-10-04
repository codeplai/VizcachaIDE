package process

import (
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const (
	waitDelay = 3 * time.Second
	killGrace = 2 * time.Second
)

// streamWriter sends what a process prints to the EventSink as it arrives.
// It never splits a UTF-8 character between two events.
type streamWriter struct {
	sink    app.EventSink
	stream  string
	seen    *atomic.Bool
	pending []byte
}

func newStreamWriter(sink app.EventSink, stream string, seen *atomic.Bool) *streamWriter {
	return &streamWriter{sink: sink, stream: stream, seen: seen}
}

// Write implements io.Writer. It always accepts everything.
func (w *streamWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.seen.Store(true)
	data := append(w.pending, p...)
	complete, rest := splitIncompleteRune(data)
	w.pending = append([]byte(nil), rest...)
	if len(complete) > 0 {
		w.sink.RunOutput(w.stream, string(complete))
	}
	return len(p), nil
}

// Flush sends the bytes held back at the end of the process.
func (w *streamWriter) Flush() {
	if len(w.pending) == 0 {
		return
	}
	w.sink.RunOutput(w.stream, string(w.pending))
	w.pending = nil
}

// splitIncompleteRune separates a trailing, still unfinished UTF-8 sequence.
func splitIncompleteRune(data []byte) (complete, rest []byte) {
	for back := 1; back < utf8.UTFMax && back <= len(data); back++ {
		tail := data[len(data)-back:]
		if !utf8.RuneStart(tail[0]) {
			continue
		}
		if !utf8.FullRune(tail) {
			return data[:len(data)-back], tail
		}
		break
	}
	return data, nil
}
