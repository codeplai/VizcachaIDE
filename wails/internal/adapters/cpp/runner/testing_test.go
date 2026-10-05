package runner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

const waitLimit = 90 * time.Second

// testSink records the run events. The other EventSink methods are never called here.
type testSink struct {
	app.EventSink
	mu       sync.Mutex
	stdout   strings.Builder
	stderr   strings.Builder
	started  []domain.RunConfiguration
	finished chan int
}

func newTestSink() *testSink { return &testSink{finished: make(chan int, 4)} }

func (s *testSink) RunOutput(stream, text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if stream == "stderr" {
		s.stderr.WriteString(text)
		return
	}
	s.stdout.WriteString(text)
}

func (s *testSink) RunStarted(config domain.RunConfiguration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = append(s.started, config)
}

func (s *testSink) RunFinished(exitCode int, _ int64) { s.finished <- exitCode }

func (s *testSink) text() (stdout, stderr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stdout.String(), s.stderr.String()
}

func (s *testSink) waitOutput(t *testing.T, text string) {
	t.Helper()
	deadline := time.Now().Add(waitLimit)
	for time.Now().Before(deadline) {
		if out, errText := s.text(); strings.Contains(out+errText, text) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	out, errText := s.text()
	t.Fatalf("output never contained %q: stdout=%q stderr=%q", text, out, errText)
}

func (s *testSink) waitFinished(t *testing.T) int {
	t.Helper()
	select {
	case code := <-s.finished:
		return code
	case <-time.After(waitLimit):
		t.Fatal("the run did not finish in time")
		return -1
	}
}

// settings are the tool paths chosen by the user: the compiler and, when cmakeBin is given, the
// cmake and ninja of that folder.
type settings struct{ path, cmakeBin string }

func (s settings) Load() (domain.Settings, error) {
	loaded := domain.DefaultSettings()
	loaded.ToolPaths = map[string]string{cpp.ToolCompiler: s.path}
	if s.cmakeBin != "" {
		loaded.ToolPaths[cpp.ToolCMake] = filepath.Join(s.cmakeBin, cpptest.Exe("cmake"))
		loaded.ToolPaths[cpp.ToolNinja] = filepath.Join(s.cmakeBin, cpptest.Exe("ninja"))
	}
	return loaded, nil
}
func (settings) Save(domain.Settings) error { return nil }

// compilers are the real compilers to test with: both families, skipping the missing ones.
func compilers(t *testing.T) map[string]string {
	t.Helper()
	found := map[string]string{}
	t.Run("probe", func(t *testing.T) {
		found["clang"] = filepath.Join(cpptest.LLVMBin(t), cpptest.Exe("clang++"))
	})
	t.Run("probe-gcc", func(t *testing.T) {
		found["gcc"] = filepath.Join(cpptest.GCCBin(t), cpptest.Exe("g++"))
	})
	return found
}

// eachCompiler runs the test with a runner for every available compiler.
func eachCompiler(t *testing.T, test func(t *testing.T, r *Runner, sink *testSink)) {
	t.Helper()
	cmakeBin := cpptest.CMakeBin(t)
	for name, path := range compilers(t) {
		t.Run(name, func(t *testing.T) {
			sink := newTestSink()
			r := New(process.New(sink), Options{
				Settings: settings{path: path, cmakeBin: cmakeBin}, AppDir: t.TempDir(), CacheDir: t.TempDir(),
				CompilingNotice: func() string { return "Compiling..." },
			})
			test(t, r, sink)
		})
	}
}

func newSupervisor() *process.Supervisor { return process.New(newTestSink()) }

// writeSource writes main.cpp in a folder with spaces and an accent, like a student's.
func writeSource(t *testing.T, source string) string {
	t.Helper()
	folder := filepath.Join(t.TempDir(), "mis programas ñandú")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(folder, "main.cpp")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
