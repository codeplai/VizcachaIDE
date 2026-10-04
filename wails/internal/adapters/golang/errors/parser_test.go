package errors

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const workDir = "/home/ana/hello"

func TestRelativePathsResolveAgainstTheWorkingDirectory(t *testing.T) {
	raw := "# command-line-arguments\n./main.go:5:2: declared and not used: x\n.\\util\\calc.go:7:14: undefined: total\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 2 {
		t.Fatalf("diagnostics = %+v", got)
	}
	if want := filepath.Join(workDir, "main.go"); got[0].Location.File != want || got[0].Location.Line != 5 || got[0].Location.Column != 2 {
		t.Errorf("first location = %+v, want %s", got[0].Location, want)
	}
	if got[0].Message != "declared and not used: x" || got[0].RawText != "./main.go:5:2: declared and not used: x" {
		t.Errorf("first = %+v", got[0])
	}
	if want := filepath.Join(workDir, "util", "calc.go"); got[1].Location.File != want || got[1].Code != "E-UNDEFINED" {
		t.Errorf("second = %+v (location %+v)", got[1], got[1].Location)
	}
}

func TestAbsoluteWindowsPathWithDriveAndSpaces(t *testing.T) {
	raw := "C:\\Users\\Ana Perez\\go\\main.go:3:8: \"os\" imported and not used\r\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 1 || got[0].Location.File != `C:\Users\Ana Perez\go\main.go` {
		t.Fatalf("diagnostics = %+v", got)
	}
	if got[0].Location.Line != 3 || got[0].Location.Column != 8 || got[0].Code != "E-UNUSED-IMPORT" {
		t.Errorf("diagnostic = %+v", got[0])
	}
}

func TestAbsolutePosixPathIsKept(t *testing.T) {
	got := newExplainer(t).Parse("/tmp/x/main.go:10:1: missing return\n", workDir)
	if len(got) != 1 || got[0].Location.File != "/tmp/x/main.go" || got[0].Code != "E-MISSING-RETURN" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestTabLinesContinueThePreviousMessage(t *testing.T) {
	raw := "./main.go:11:18: not enough arguments in call to add\n\thave (number)\n\twant (int)\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 1 || got[0].Message != "not enough arguments in call to add" {
		t.Fatalf("diagnostics = %+v", got)
	}
	if !strings.HasSuffix(got[0].RawText, "\thave (number)\n\twant (int)") {
		t.Errorf("rawText = %q", got[0].RawText)
	}
}

func TestVetHeaderMarksWarnings(t *testing.T) {
	raw := "# command-line-arguments\n# [command-line-arguments]\n./main.go:9:2: unreachable code\n./main.go:7:2: something vet found\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 2 {
		t.Fatalf("diagnostics = %+v", got)
	}
	if got[0].Source != "vet" || got[0].Severity != domain.SeverityWarning || got[0].Code != "V-UNREACHABLE" {
		t.Errorf("first = %+v", got[0])
	}
	if got[1].Source != "vet" || got[1].Code != "" {
		t.Errorf("second = %+v", got[1])
	}
}

func TestRecognisedLinesWithoutLocationAreKept(t *testing.T) {
	raw := "# command-line-arguments\nruntime.main_main·f: function main is undeclared in the main package\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 1 || got[0].Location != nil || got[0].Code != "E-NO-MAIN" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestProgramOutputIsIgnoredButGoCommandErrorsAreKept(t *testing.T) {
	got := newExplainer(t).Parse("hello from the program\ngo: cannot find main module\n", workDir)
	if len(got) != 1 || got[0].Message != "go: cannot find main module" || got[0].Code != "" {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestPanicLocationIsTheFirstUserFrame(t *testing.T) {
	raw := "panic: runtime error: index out of range [5] with length 3\n\ngoroutine 1 [running]:\n" +
		"panic({0x4a1f20?, 0xc000012345?})\n\t/usr/local/go/src/runtime/panic.go:787 +0x132\n" +
		"main.pick(...)\n\t/home/ana/hello/main.go:12 +0x1d\nmain.main()\n\t/home/ana/hello/main.go:20 +0x2a\nexit status 2\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 1 {
		t.Fatalf("diagnostics = %+v", got)
	}
	if got[0].Location == nil || got[0].Location.File != "/home/ana/hello/main.go" || got[0].Location.Line != 12 || got[0].Location.Column != 1 {
		t.Errorf("location = %+v", got[0].Location)
	}
	if got[0].Source != "panic" || got[0].Code != "P-INDEX-RANGE" || got[0].Severity != domain.SeverityError {
		t.Errorf("diagnostic = %+v", got[0])
	}
	if !strings.Contains(got[0].RawText, "goroutine 1 [running]:") || strings.Contains(got[0].RawText, "exit status") {
		t.Errorf("rawText = %q", got[0].RawText)
	}
}

func TestWindowsPanicFrameWithForwardSlashes(t *testing.T) {
	raw := "fatal error: all goroutines are asleep - deadlock!\n\ngoroutine 1 [chan send]:\nmain.main()\n\tC:/Users/ana/hello/main.go:6 +0x28\nexit status 2\n"
	got := newExplainer(t).Parse(raw, workDir)
	if len(got) != 1 || got[0].Code != "P-DEADLOCK" || got[0].Location == nil {
		t.Fatalf("diagnostics = %+v", got)
	}
	if got[0].Location.File != "C:/Users/ana/hello/main.go" || got[0].Location.Line != 6 {
		t.Errorf("location = %+v", got[0].Location)
	}
}
