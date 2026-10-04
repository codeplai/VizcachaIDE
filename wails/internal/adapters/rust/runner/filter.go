package runner

import (
	"encoding/json"
	"strings"
	"sync"
)

// diagnostics turns the JSON of the compile stage into the text rustc and cargo would print, and
// remembers the executable cargo built. rustc (--error-format=json) prints one object per
// diagnostic with a "rendered" field; cargo (--message-format=json) wraps rustc's object in a
// "compiler-message" and adds "compiler-artifact", "build-finished" and so on.
type diagnostics struct {
	binary     string // the bin target whose executable is wanted ("" for rustc)
	mu         sync.Mutex
	executable string
}

type jsonLine struct {
	Reason     string          `json:"reason"`
	Rendered   *string         `json:"rendered"`
	Message    json.RawMessage `json:"message"`
	Executable *string         `json:"executable"`
	Target     struct {
		Name string `json:"name"`
	} `json:"target"`
}

// Filter is process.Job.OutputFilter: the rendered text of a diagnostic on stderr (cargo prints
// its JSON on stdout, but the Assistant reads compiler errors from stderr), nothing for the other
// JSON lines and the line itself when it is not JSON.
func (d *diagnostics) Filter(stream, line string) (string, string, bool) {
	if !strings.HasPrefix(strings.TrimSpace(line), "{") {
		return stream, line, true
	}
	var parsed jsonLine
	if err := json.Unmarshal([]byte(line), &parsed); err != nil {
		return stream, line, true
	}
	switch parsed.Reason {
	case "":
		return rendered(parsed.Rendered)
	case "compiler-message":
		var inner struct {
			Rendered *string `json:"rendered"`
		}
		if json.Unmarshal(parsed.Message, &inner) != nil {
			return stream, "", false
		}
		return rendered(inner.Rendered)
	case "compiler-artifact":
		d.remember(parsed)
	}
	return stream, "", false
}

func rendered(text *string) (string, string, bool) {
	if text == nil || *text == "" {
		return "stderr", "", false
	}
	return "stderr", *text, true // it ends with its own line break and a blank line
}

func (d *diagnostics) remember(parsed jsonLine) {
	if parsed.Executable == nil || *parsed.Executable == "" || parsed.Target.Name != d.binary {
		return
	}
	d.mu.Lock()
	d.executable = *parsed.Executable
	d.mu.Unlock()
}

// built is the executable of the bin target ("" when cargo did not report one).
func (d *diagnostics) built() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.executable
}

// text applies the filter to captured output (CompileForDebug).
func (d *diagnostics) text(captured string) string {
	var shown strings.Builder
	for _, line := range strings.Split(strings.ReplaceAll(captured, "\r\n", "\n"), "\n") {
		if line == "" {
			continue
		}
		if _, text, keep := d.Filter("", line); keep {
			shown.WriteString(text)
			if !strings.HasSuffix(text, "\n") {
				shown.WriteString("\n")
			}
		}
	}
	return shown.String()
}
