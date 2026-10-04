package runner

import (
	"context"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

func TestToolsReportTheToolchainWithVersions(t *testing.T) {
	r, _ := newRunner(t)
	statuses := r.Tools(context.Background())
	if len(statuses) != len(rust.Profile.Tools) {
		t.Fatalf("%d statuses for %d tools", len(statuses), len(rust.Profile.Tools))
	}
	number := regexp.MustCompile(`^\d+\.\d+`)
	found := map[string]domain.ToolStatus{}
	for _, status := range statuses {
		found[status.ID] = status
	}
	for _, id := range []string{rust.ToolRustc, rust.ToolCargo, rust.ToolRustAnalyzer, rust.ToolClippy, rust.ToolRustfmt} {
		status := found[id]
		if status.Source == domain.ToolMissing || !number.MatchString(status.Version) || status.CodeLanguage != domain.CodeLanguageRust {
			t.Errorf("%s = %+v, want it found with a version", id, status)
		}
	}
	if found[rust.ToolRustc].Version != found[rust.ToolCargo].Version {
		t.Errorf("rustc %s and cargo %s ship together", found[rust.ToolRustc].Version, found[rust.ToolCargo].Version)
	}
}

func TestMissingToolsAreNamed(t *testing.T) {
	empty := t.TempDir()
	r := New(process.New(newTestSink()), Options{
		BaseEnvironment: []string{"PATH=" + empty, "CARGO_HOME=" + empty}, AppDir: empty, CacheDir: empty,
	})
	path := looseFile(t, hello)
	if err := r.Run(context.Background(), r.Configure(path, nil)); err == nil || err.Error() != `tool "rustc": tool not found` {
		t.Fatalf("loose file: err = %v, want the missing rustc", err)
	}
	root := t.TempDir()
	crate(t, root, "c", map[string]string{"src/main.rs": hello})
	err := r.Run(context.Background(), r.Configure(filepath.Join(root, "src", "main.rs"), nil))
	if err == nil || err.Error() != `tool "cargo": tool not found` {
		t.Fatalf("crate: err = %v, want the missing cargo", err)
	}
	for _, status := range r.Tools(context.Background()) {
		if status.Source != domain.ToolMissing || status.Version != "" {
			t.Errorf("%+v, want missing", status)
		}
	}
}
