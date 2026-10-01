package toolchain

import (
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func TestFormatSourceFormats(t *testing.T) {
	got, err := New(Options{Sink: newTestSink()}).FormatSource("package main\nfunc main(){x:=1;_=x}\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func main() {") {
		t.Errorf("formatted = %q", got)
	}
}

func TestFormatSourceSyntaxErrorNamesTheLine(t *testing.T) {
	_, err := New(Options{Sink: newTestSink()}).FormatSource("package main\n\nfunc main() {\n\tx := \n}\n")
	if !errors.Is(err, app.ErrFormat) {
		t.Fatalf("error = %v, want ErrFormat", err)
	}
	if !strings.Contains(err.Error(), "line 5") {
		t.Errorf("error %q must name line 5", err)
	}
}

func TestParseVersion(t *testing.T) {
	cases := map[string]string{
		"go version go1.25.5 windows/amd64\n":                               "1.25.5",
		"Delve Debugger\nVersion: 1.27.2\nBuild: $Id: abc":                  "1.27.2",
		"golang.org/x/tools/gopls v0.21.1\n    golang.org/x/tools/gopls@v0": "0.21.1",
		"nothing here": "",
	}
	for output, want := range cases {
		if got := parseVersion(output); got != want {
			t.Errorf("parseVersion(%q) = %q, want %q", output, got, want)
		}
	}
}

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
	sink := newTestSink()
	writer := newStreamWriter(sink, "stdout", &atomic.Bool{})
	text := []byte("canción")
	_, _ = writer.Write(text[:5]) // cuts ó in half
	_, _ = writer.Write(text[5:])
	writer.Flush()
	if sink.Stdout() != "canción" {
		t.Errorf("stdout = %q", sink.Stdout())
	}
}
