package lldbdap

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	protodap "github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/dap/lldb"
	"github.com/google/go-dap"
)

// panicBreakpointID is the id LLDB gives the breakpoint of InitCommands: it is the first one.
const panicBreakpointID = 1

// flavor is the LLDB flavor plus the stop at panic!: the hit on rust_panic becomes an exception
// whose description is the panic message. Frames outside the user's folder (/rustc/<hash>/
// library, the sysroot's rust-src, the cargo registry) are hidden by the Roots of the base.
type flavor struct {
	lldb.Flavor
	panics *panicWatch
}

var _ protodap.Flavor = flavor{}

func newFlavor(options lldb.Options, panics *panicWatch) flavor {
	return flavor{Flavor: lldb.New(options), panics: panics}
}

func (f flavor) StopReason(event *dap.StoppedEvent) (domain.StopReason, string) {
	if !hitPanic(event) {
		return f.Flavor.StopReason(event)
	}
	return domain.StopException, f.panics.description()
}

func hitPanic(event *dap.StoppedEvent) bool {
	for _, id := range event.Body.HitBreakpointIds {
		if id == panicBreakpointID {
			return true
		}
	}
	return false
}

// panicWatch remembers the last panic message the program wrote to stderr. The runtime prints
// it before it calls rust_panic, but the pseudoterminal delivers it asynchronously, so
// description waits a moment for it.
type panicWatch struct {
	app.EventSink // the real sink: only DebugOutput is observed

	mu      sync.Mutex
	message string
	tail    string // incomplete last line
	// expectMessage is true right after the "panicked at" line: the next line is the message.
	expectMessage bool
}

// DebugOutput passes the text on and looks for "thread 'main' panicked at file:L:C:" followed by
// the message line.
func (p *panicWatch) DebugOutput(text, category string) {
	p.observe(text)
	p.EventSink.DebugOutput(text, category)
}

var panicHeader = regexp.MustCompile(`thread '.*' (?:\(\d+\) )?panicked at `)

func (p *panicWatch) observe(text string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	lines := strings.Split(strings.ReplaceAll(p.tail+text, "\r\n", "\n"), "\n")
	p.tail = lines[len(lines)-1]
	complete := lines[:len(lines)-1]
	for _, line := range complete {
		switch {
		case p.expectMessage:
			p.message, p.expectMessage = strings.TrimSpace(line), false
		case panicHeader.MatchString(line):
			p.expectMessage = true
		}
	}
}

func (p *panicWatch) description() string {
	deadline := time.Now().Add(500 * time.Millisecond)
	for {
		p.mu.Lock()
		message := p.message
		p.mu.Unlock()
		if message != "" || time.Now().After(deadline) {
			if message == "" {
				return "panicked"
			}
			return "panicked: " + message
		}
		time.Sleep(20 * time.Millisecond)
	}
}
