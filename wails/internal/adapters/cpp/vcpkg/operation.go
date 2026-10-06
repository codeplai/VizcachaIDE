package vcpkg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Busy reports whether an operation is running.
func (m *Manager) Busy() bool { return m.busy.Load() }

// Stop cancels the operation in progress, if any.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Manager) slowNotice() string {
	if m.options.SlowNotice != nil {
		return m.options.SlowNotice()
	}
	return defaultSlowNotice
}

// start runs an operation in the background, reporting it with run events, one at a time.
func (m *Manager) start(ctx context.Context, root, title string, operation func(context.Context, func(string)) error) error {
	if !m.busy.CompareAndSwap(false, true) {
		return app.ErrBusy
	}
	ctx, cancel := context.WithCancel(ctx)
	m.mu.Lock()
	m.cancel = cancel
	m.mu.Unlock()
	events := m.options.Events
	events.RunStarted(domain.RunConfiguration{
		CodeLanguage: domain.CodeLanguageCpp, Target: title, WorkingDir: root,
		Mode: domain.RunProject, ProgramArgs: []string{},
	})
	go func() {
		started := time.Now()
		defer m.busy.Store(false)
		defer cancel()
		err := operation(ctx, func(text string) { events.RunOutput("stdout", text) })
		code := 0
		if err != nil {
			code = 1
			events.RunOutput("stderr", err.Error()+"\n")
		}
		events.RunFinished(code, time.Since(started).Milliseconds())
	}()
	return nil
}

// projectRoot finds the CMake project that holds dir (a folder or a file of it).
func projectRoot(dir string) (string, error) {
	current := dir
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		current = filepath.Dir(dir)
	}
	for {
		if isFile(filepath.Join(current, CMakeFile)) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", ErrNeedsProject
		}
		current = parent
	}
}

// snapshot remembers vcpkg.json and CMakeLists.txt and returns the function that puts them back.
func snapshot(root string) (func(), error) {
	names := []string{ManifestFile, CMakeFile}
	saved := map[string][]byte{}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			saved[name] = data
		}
	}
	return func() {
		for _, name := range names {
			path := filepath.Join(root, name)
			if data, ok := saved[name]; ok {
				_ = os.WriteFile(path, data, 0o644)
			} else {
				_ = os.Remove(path)
			}
		}
	}, nil
}
