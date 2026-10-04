package runner

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestCompatInfoMapsTheToolStatuses(t *testing.T) {
	r, _ := newTestRunner(t)

	info := r.Compat().Info(context.Background())

	if info.GoSource == "" || info.DelveSource == "" || info.GoplsSource == "" {
		t.Errorf("info = %+v, every tool needs a source", info)
	}
	if info.GoSource != domain.ToolMissing && info.GoVersion == "" {
		t.Errorf("info = %+v, a found go needs a version", info)
	}
}

func TestCompatRunsUntitledSourceAndGoCommands(t *testing.T) {
	r, sink := newTestRunner(t)
	compat := r.Compat()

	if _, err := compat.RunUntitled(context.Background(), helloProgram, nil); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
	dir := writeFiles(t, map[string]string{"main.go": helloProgram})
	if err := compat.RunGoCommand(context.Background(), dir, []string{"mod", "init", "example.com/x"}); err != nil {
		t.Fatal(err)
	}
	sink.waitFinished(t)
	output, err := compat.Vet(context.Background(), golang.ConfigurationForFile(filepath.Join(dir, "main.go"), nil))

	if err != nil || output != "" || !strings.Contains(sink.Stdout(), "hola desde go") {
		t.Errorf("vet = %q, %v, stdout = %q", output, err, sink.Stdout())
	}
	if _, err := compat.FormatSource("package main\nfunc main(){}\n"); err != nil {
		t.Errorf("FormatSource: %v", err)
	}
}
