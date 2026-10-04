package runner

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang"
)

const vetProgram = `package main

import "fmt"

func main() {
	fmt.Printf("%d %d\n", 1)
}
`

func TestVetReportsWarningsWithoutTouchingTheRunSlot(t *testing.T) {
	tc, sink := newTestRunner(t)
	dir := writeFiles(t, map[string]string{"main.go": vetProgram})
	config := golang.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)

	output, err := tc.Check(context.Background(), config)

	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "reads arg #2") {
		t.Errorf("vet output = %q, want the Printf warning", output)
	}
	if tc.IsRunning() || sink.startedCount() != 0 {
		t.Error("vet must not use the run slot or emit run events")
	}
}

func TestVetIsSilentWhenTheCodeIsFine(t *testing.T) {
	tc, _ := newTestRunner(t)
	dir := writeFiles(t, map[string]string{"main.go": helloProgram})
	config := golang.ConfigurationForFile(filepath.Join(dir, "main.go"), nil)

	output, err := tc.Check(context.Background(), config)

	if err != nil || output != "" {
		t.Errorf("Vet = %q, %v; want no output", output, err)
	}
}

func TestVetSkipsAFolderThatIsGone(t *testing.T) {
	tc, _ := newTestRunner(t)
	config := golang.ConfigurationForFile(filepath.Join(t.TempDir(), "gone", "main.go"), nil)

	output, err := tc.Check(context.Background(), config)

	if err != nil || output != "" {
		t.Errorf("Vet = %q, %v; want empty and no error", output, err)
	}
}
