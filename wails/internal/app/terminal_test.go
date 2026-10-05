package app

import (
	"context"
	"slices"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type fixedShellPaths []string

func (f fixedShellPaths) ShellPaths(context.Context) []string { return f }

func TestRegistryGathersShellPathsInLanguageOrderWithoutRepeats(t *testing.T) {
	first := supportOf(domain.CodeLanguageGo, ".go")
	first.Shell = fixedShellPaths{"go-bin", "shared"}
	second := supportOf(domain.CodeLanguagePython, ".py")
	second.Shell = fixedShellPaths{"shared", "py", ""}
	none := supportOf(domain.CodeLanguageCpp, ".cpp") // a language without the optional port
	registry, err := NewLanguageRegistry(domain.CodeLanguageGo, first, second, none)
	if err != nil {
		t.Fatal(err)
	}
	got := registry.ShellPaths(context.Background())
	if want := []string{"go-bin", "shared", "py"}; !slices.Equal(got, want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
}
