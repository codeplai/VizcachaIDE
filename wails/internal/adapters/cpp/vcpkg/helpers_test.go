package vcpkg

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakeRoot makes a folder that looks like a vcpkg root.
func fakeRoot(t *testing.T, goos string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "vcpkg")
	if err := os.MkdirAll(filepath.Join(root, "ports"), 0o755); err != nil {
		t.Fatal(err)
	}
	executable := "vcpkg"
	if goos == "windows" {
		executable += ".exe"
	}
	for _, name := range []string{".vcpkg-root", executable} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

type settingsWith string

func (s settingsWith) Load() (domain.Settings, error) {
	return domain.Settings{ToolPaths: map[string]string{ToolID: string(s)}}, nil
}
func (settingsWith) Save(domain.Settings) error { return nil }

func readText(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// recordedEvents collects the run events and signals the end of the run.
type recordedEvents struct {
	mu       sync.Mutex
	output   []string
	started  []domain.RunConfiguration
	finished chan int
}

func newEvents() *recordedEvents { return &recordedEvents{finished: make(chan int, 1)} }

func (r *recordedEvents) RunStarted(config domain.RunConfiguration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.started = append(r.started, config)
}

func (r *recordedEvents) RunOutput(_, text string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.output = append(r.output, text)
}

func (r *recordedEvents) RunFinished(exitCode int, _ int64) { r.finished <- exitCode }

func (r *recordedEvents) wait(t *testing.T) int {
	t.Helper()
	select {
	case code := <-r.finished:
		return code
	case <-time.After(5 * time.Second):
		t.Fatal("the operation did not finish")
		return -1
	}
}

func (r *recordedEvents) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.output = nil
	r.finished = make(chan int, 1)
}

func (r *recordedEvents) text() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return strings.Join(r.output, "")
}
