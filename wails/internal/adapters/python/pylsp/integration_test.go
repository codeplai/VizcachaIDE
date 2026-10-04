package pylsp

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python/pythontest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// realSample opens main.py in the real python-lsp-server. Diagnostics need pyflakes, which
// python-lsp-server does not install by itself ("python-lsp-server[pyflakes]").
func realSample(t *testing.T, withDiagnostics bool) sample {
	t.Helper()
	path := pythontest.Interpreter(t)
	if withDiagnostics && exec.Command(path, "-c", "import pyflakes").Run() != nil {
		t.Skip("the test Python has no pyflakes: pip install \"python-lsp-server[pyflakes]\"")
	}
	return openSample(t, NewFlavor(configAt(path)), withDiagnostics)
}

func TestRealPylspReportsTheUndefinedName(t *testing.T) {
	s := realSample(t, true)
	var undefined *domain.Diagnostic
	for _, d := range s.sink.diagnosticsOf(s.file) {
		if d.Message == "undefined name 'x'" {
			undefined = &d
		}
	}
	if undefined == nil {
		t.Fatalf("no diagnostic for x: %+v", s.sink.diagnosticsOf(s.file))
	}
	if undefined.Location.Line != 6 || undefined.Location.Column != 7 || undefined.Source != serverName {
		t.Errorf("diagnostic = %+v at %+v", *undefined, *undefined.Location)
	}
	if s.sink.lastStatus() != domain.ServerReady {
		t.Errorf("status = %q, want ready", s.sink.lastStatus())
	}
}

func TestRealPylspCompletesPriWithPrintAndExplainsLen(t *testing.T) {
	s := realSample(t, false)
	items := eventually(t, func() ([]domain.CompletionItem, bool) {
		items, _ := s.server.Completion(context.Background(), atPri(s.file))
		return items, len(items) > 0
	})
	found := false
	for _, item := range items {
		found = found || strings.HasPrefix(item.Label, "print")
	}
	if !found {
		t.Errorf("print not suggested: %+v", items)
	}
	hover := eventually(t, func() (string, bool) {
		text, _ := s.server.Hover(context.Background(), atLen(s.file))
		return text, text != ""
	})
	if !strings.Contains(hover, "len(") {
		t.Errorf("hover = %q", hover)
	}
	help := eventually(t, func() (*domain.SignatureHelp, bool) {
		help, _ := s.server.SignatureHelp(context.Background(), insideGreet(s.file))
		return help, help != nil
	})
	if help.Label != "greet(name)" {
		t.Errorf("signature help = %+v", help)
	}
}
