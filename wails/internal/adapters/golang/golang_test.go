package golang_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestProfileDeclaresGo(t *testing.T) {
	profile := golang.Profile
	if profile.ID != domain.CodeLanguageGo || profile.NameKey != "codeLanguage.go" || profile.Extensions[0] != ".go" {
		t.Errorf("profile = %+v", profile)
	}
	caps := profile.Capabilities
	if !caps.Build || !caps.Console || !caps.Format || !caps.Check || len(caps.PackageActions) != 3 || caps.ThreadsLabel != "debug.goroutines" {
		t.Errorf("capabilities = %+v", caps)
	}
	ids := []string{}
	for _, tool := range profile.Tools {
		ids = append(ids, tool.ID)
	}
	if len(ids) != 3 || ids[0] != "go" || ids[1] != "dlv" || ids[2] != "gopls" {
		t.Errorf("tools = %v", ids)
	}
}

func TestConfigurationRunsTheModuleWhenThereIsOne(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/demo\n\ngo 1.21\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	inModule := golang.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)
	alone := golang.ConfigurationForFile(filepath.Join(t.TempDir(), "main.go"), []string{"a"})

	if inModule.Mode != domain.RunProject || inModule.Project == nil || inModule.Project.Name != "example.com/demo" {
		t.Errorf("module configuration = %+v", inModule)
	}
	if golang.GoTargetArgument(inModule) != "." || golang.GoTargetArgument(alone) != "main.go" {
		t.Error("a module runs \".\" and a file runs its own name")
	}
	if alone.Mode != domain.RunFile || alone.Project != nil || len(alone.ProgramArgs) != 1 {
		t.Errorf("file configuration = %+v", alone)
	}
	if got := golang.GoExecutableName(inModule, true); got != filepath.Base(dir)+".exe" {
		t.Errorf("module executable = %q", got)
	}
	if got := golang.GoExecutableName(alone, false); got != "main" {
		t.Errorf("file executable = %q", got)
	}
}
