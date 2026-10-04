package clangd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// realCompilers runs a check with the llvm-mingw clang++ and with the GCC g++ (query-driver).
func realCompilers(t *testing.T, check func(t *testing.T, compilerBin string)) {
	t.Helper()
	t.Run("clang++", func(t *testing.T) { check(t, cpptest.LLVMBin(t)) })
	t.Run("g++", func(t *testing.T) { check(t, cpptest.GCCBin(t)) })
}

func TestRealClangdSeesOnlyTheUndeclaredName(t *testing.T) {
	realCompilers(t, func(t *testing.T, bin string) {
		s := openSample(t, NewFlavor(Config{Locator: locatorFor(bin, cpptest.LLVMBin(t))}), true)
		// The sample has one real error (missing) and the incomplete "std::vec" line: nothing
		// may complain about the headers.
		reported := false
		for _, d := range s.sink.diagnosticsOf(s.file) {
			lower := strings.ToLower(d.Message)
			if strings.Contains(lower, "file not found") || strings.Contains(lower, "undeclared identifier 'cout'") ||
				strings.Contains(lower, "undeclared identifier 'vector'") {
				t.Errorf("false diagnostic: %+v", d)
			}
			reported = reported || (strings.Contains(d.Message, "undeclared identifier 'missing'") &&
				d.Severity == domain.SeverityError && d.Location.Line == 17)
		}
		if !reported {
			t.Errorf("missing was not reported: %+v", s.sink.diagnosticsOf(s.file))
		}
	})
}

func TestRealClangdCompletesAndExplainsPushBack(t *testing.T) {
	realCompilers(t, func(t *testing.T, bin string) {
		s := openSample(t, NewFlavor(Config{Locator: locatorFor(bin, cpptest.LLVMBin(t))}), true)
		members := eventually(t, func() ([]domain.CompletionItem, bool) {
			items, _ := s.server.Completion(context.Background(), atMemberPush(s.file))
			return items, len(items) > 0
		})
		if !strings.Contains(members[0].Label, "push_back") {
			t.Errorf("members = %+v", members)
		}
		hover := eventually(t, func() (string, bool) {
			text, _ := s.server.Hover(context.Background(), atPushBack(s.file))
			return text, text != ""
		})
		if !strings.Contains(hover, "push_back") {
			t.Errorf("hover = %q", hover)
		}
	})
}

// A reloaded window opens the same file again: clangd must publish its diagnostics again.
func TestRealClangdRepublishesWhenTheFileIsOpenedAgain(t *testing.T) {
	llvm := cpptest.LLVMBin(t)
	s := openSample(t, NewFlavor(Config{Locator: locatorFor(llvm, llvm)}), true)
	s.sink.mu.Lock()
	delete(s.sink.diagnostics, s.file)
	s.sink.mu.Unlock()
	text, err := os.ReadFile(s.file)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.server.OpenDocument(context.Background(), s.file, string(text)); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, func() bool { return len(s.sink.diagnosticsOf(s.file)) > 0 })
}
