package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestConfigurationForFileWithoutModuleRunsTheFile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "main.go"), "package main")
	config := ConfigurationForFile(filepath.Join(dir, "main.go"), nil)
	if config.Mode != domain.RunFile || config.Project != nil || config.WorkingDir != dir {
		t.Errorf("config = %+v", config)
	}
}

func TestConfigurationForFileInsideModuleRunsThePackage(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "go.mod"), "module example.com/app\n\ngo 1.21\n")
	file := filepath.Join(root, "cmd", "main.go")
	write(t, file, "package main")

	config := ConfigurationForFile(file, []string{"-v"})

	if config.Mode != domain.RunProject || config.WorkingDir != filepath.Dir(file) {
		t.Errorf("config = %+v", config)
	}
	if config.Project == nil || config.Project.Root != root || config.Project.Name != "example.com/app" || config.Project.Kind != domain.ProjectGoModule {
		t.Errorf("project = %+v", config.Project)
	}
	if GoTargetArgument(config) != "." {
		t.Errorf("target = %q", GoTargetArgument(config))
	}
}

func TestParseModulePath(t *testing.T) {
	cases := map[string]string{
		"module example.com/x\n\ngo 1.21": "example.com/x",
		"// c\nmodule \"quoted/mod\"\n":   "quoted/mod",
		"go 1.21":                         "",
	}
	for text, want := range cases {
		if got := ParseModulePath(text); got != want {
			t.Errorf("ParseModulePath(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestGoTargetArgumentAndExecutableName(t *testing.T) {
	file := domain.NewFileRunConfiguration(domain.CodeLanguageGo, "proj/main.go", nil)
	if got := GoTargetArgument(file); got != "main.go" {
		t.Errorf("GoTargetArgument(file) = %q, want main.go", got)
	}
	if got := GoExecutableName(file, true); got != "main.exe" {
		t.Errorf("GoExecutableName(file, true) = %q, want main.exe", got)
	}
	if got := GoExecutableName(file, false); got != "main" {
		t.Errorf("GoExecutableName(file, false) = %q, want main", got)
	}
	project := domain.RunConfiguration{Target: "proj/main.go", WorkingDir: "proj", Mode: domain.RunProject}
	if got := GoTargetArgument(project); got != "." {
		t.Errorf("GoTargetArgument(project) = %q, want .", got)
	}
	if got := GoExecutableName(project, false); got != "proj" {
		t.Errorf("GoExecutableName(project, false) = %q, want proj", got)
	}
}
