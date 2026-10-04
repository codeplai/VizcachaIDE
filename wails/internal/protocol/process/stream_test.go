package process

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

type collectingSink struct {
	app.EventSink
	text strings.Builder
}

func (s *collectingSink) RunOutput(_, text string) { s.text.WriteString(text) }

func TestSplitIncompleteRune(t *testing.T) {
	text := []byte("añ") // ñ is two bytes
	complete, rest := splitIncompleteRune(text[:2])
	if string(complete) != "a" || len(rest) != 1 {
		t.Errorf("complete=%q rest=%v", complete, rest)
	}
	complete, rest = splitIncompleteRune(text)
	if string(complete) != "añ" || rest != nil {
		t.Errorf("complete=%q rest=%v", complete, rest)
	}
}

func TestStreamWriterNeverSplitsACharacter(t *testing.T) {
	sink := &collectingSink{}
	writer := newStreamWriter(sink, "stdout", &atomic.Bool{})
	text := []byte("canción")
	_, _ = writer.Write(text[:5]) // cuts ó in half
	_, _ = writer.Write(text[5:])
	writer.Flush()
	if sink.text.String() != "canción" {
		t.Errorf("stdout = %q", sink.text.String())
	}
}
