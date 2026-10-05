package cmake

import (
	"regexp"
	"strings"
	"sync"
	"time"
)

// What the build prints that a student does not need: Ninja's progress ("[3/7] Building CXX
// object ...", "[0/2] Re-checking globbed directories..."), CMake's status lines ("-- The CXX
// compiler identification is Clang 23.1.2") and Ninja's own summaries.
var (
	progressLine = regexp.MustCompile(`^\[\d+/\d+\] `)
	ninjaLine    = regexp.MustCompile(`^ninja: (?:build stopped|no work to do|Entering directory)`)
	failedLine   = regexp.MustCompile(`^FAILED: `)
)

const noticeDelay = 2 * time.Second

// Filter hides the noise of a CMake and Ninja run and keeps what the student needs: the
// compiler's diagnostics (so the C++ error parser still turns them into Problems), CMake's
// errors and warnings, and what the libraries' installation prints. Ninja's progress is also
// what would stop the "Compiling..." notice from ever showing, so the filter prints it itself
// once the build has gone on for Delay.
type Filter struct {
	// Notice returns the already translated "Compiling..." text; nil shows none.
	Notice func() string
	// Delay is how long the build runs before the notice is shown. Default: 2 seconds.
	Delay time.Duration

	mu       sync.Mutex
	started  time.Time
	noticed  bool
	skipNext bool
	now      func() time.Time // replaced in tests
}

// Line is process.Job.OutputFilter of the configure stage: what is kept stays on its stream.
func (f *Filter) Line(stream, line string) (shownStream, text string, keep bool) {
	return f.filter(stream, line, false)
}

// BuildLine is process.Job.OutputFilter of the build stage. Ninja prints everything the compiler
// says on its own stdout, so what is kept (the compiler's warnings and errors) goes to stderr, as
// it would from the compiler itself: that is where the Assistant looks for them.
func (f *Filter) BuildLine(stream, line string) (shownStream, text string, keep bool) {
	return f.filter(stream, line, true)
}

func (f *Filter) filter(stream, line string, toStderr bool) (string, string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	hidden := f.hides(strings.TrimRight(line, "\r"))
	if !hidden {
		if toStderr {
			stream = "stderr"
		}
		return stream, line, true
	}
	if notice := f.notice(); notice != "" {
		return "stdout", notice, true // shown in place of a hidden line
	}
	return stream, "", false
}

// Text applies the filter to captured output (a build that ran outside the supervisor).
func (f *Filter) Text(captured string) string {
	var shown strings.Builder
	lines := strings.Split(strings.ReplaceAll(captured, "\r\n", "\n"), "\n")
	for index, line := range lines {
		if f.hidesCaptured(line) {
			continue
		}
		shown.WriteString(line)
		if index < len(lines)-1 {
			shown.WriteString("\n")
		}
	}
	return shown.String()
}

func (f *Filter) hidesCaptured(line string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hides(line)
}

// hides reports whether the line is noise. After Ninja's "FAILED:" line comes the whole compiler
// command, which is noise too.
func (f *Filter) hides(line string) bool {
	if f.skipNext {
		f.skipNext = false
		return true
	}
	switch {
	case failedLine.MatchString(line):
		f.skipNext = true
		return true
	case progressLine.MatchString(line), ninjaLine.MatchString(line), line == "Re-running CMake...":
		return true
	case strings.HasPrefix(line, "-- "):
		return !strings.Contains(strings.ToLower(line), "vcpkg")
	}
	return false
}

// notice is the "Compiling..." text with its line break when it is time to show it, else "". It
// is asked for on every hidden line: the first one starts the clock.
func (f *Filter) notice() string {
	if f.Notice == nil || f.noticed {
		return ""
	}
	now := f.clock()
	if f.started.IsZero() {
		f.started = now
		return ""
	}
	delay := f.Delay
	if delay <= 0 {
		delay = noticeDelay
	}
	if now.Sub(f.started) < delay {
		return ""
	}
	f.noticed = true
	return f.Notice() + "\n"
}

func (f *Filter) clock() time.Time {
	if f.now != nil {
		return f.now()
	}
	return time.Now()
}
