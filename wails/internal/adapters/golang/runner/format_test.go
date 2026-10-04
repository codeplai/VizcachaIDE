package runner

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

func TestFormatSourceFormats(t *testing.T) {
	got, err := New(process.New(newTestSink()), Options{}).Format("main.go", "package main\nfunc main(){x:=1;_=x}\n")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "func main() {") {
		t.Errorf("formatted = %q", got)
	}
}

func TestFormatSourceSyntaxErrorNamesTheLine(t *testing.T) {
	_, err := New(process.New(newTestSink()), Options{}).Format("main.go", "package main\n\nfunc main() {\n\tx := \n}\n")
	if !errors.Is(err, app.ErrFormat) {
		t.Fatalf("error = %v, want ErrFormat", err)
	}
	// "main.go:5:1: ..." is what lineFromFormatError (frontend saving.ts) reads.
	if !regexp.MustCompile(`main\.go:5:\d+: `).MatchString(err.Error()) {
		t.Errorf("error %q must name line 5 as main.go:5:<column>", err)
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
