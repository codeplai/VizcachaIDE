package gopls

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func sameFolder(a, b string) bool {
	first, err := os.Stat(a)
	if err != nil {
		return false
	}
	second, err := os.Stat(b)
	return err == nil && os.SameFile(first, second)
}

func TestModuleRootIsTheNearestGoMod(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "cmd", "app")
	if err := os.MkdirAll(nested, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(nested, "main.go")
	if got := moduleRoot(file); !sameFolder(got, root) {
		t.Errorf("moduleRoot = %q, want %q", got, root)
	}
	loose := filepath.Join(t.TempDir(), "solo.go")
	if got := moduleRoot(loose); !sameFolder(got, filepath.Dir(loose)) {
		t.Errorf("without go.mod the root is the folder, got %q", got)
	}
}

func TestMissingGoplsIsReportedAsMissingTool(t *testing.T) {
	flavor := NewFlavor(Config{Executable: func() string { return filepath.Join(t.TempDir(), "no-gopls") }})
	_, _, err := flavor.Command(nil)
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("gopls").Error() {
		t.Errorf("err = %v", err)
	}
}
