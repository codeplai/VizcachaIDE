package clippy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust/rusttest"
)

// withoutClippy is a toolchain with cargo and rustc but no clippy component: a CARGO_HOME whose
// bin has copies of the rustup proxies for those two only.
func withoutClippy(t *testing.T) *Tool {
	t.Helper()
	real := rusttest.Environment(t)
	var cargoHome, rustupHome string
	for _, entry := range real {
		name, value, _ := strings.Cut(entry, "=")
		switch strings.ToUpper(name) {
		case "CARGO_HOME":
			cargoHome = value
		case "RUSTUP_HOME":
			rustupHome = value
		}
	}
	home := t.TempDir()
	for _, tool := range []string{"cargo", "rustc"} {
		content, err := os.ReadFile(filepath.Join(cargoHome, "bin", rusttest.Exe(tool)))
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(home, "bin", rusttest.Exe(tool)), string(content))
	}
	env := []string{"CARGO_HOME=" + home, "RUSTUP_HOME=" + rustupHome, "PATH=" + filepath.Join(home, "bin")}
	if root := os.Getenv("SystemRoot"); root != "" {
		env = append(env, "SystemRoot="+root)
	}
	return New(rust.NewLocator(rust.Options{AppDir: t.TempDir(), BaseEnvironment: env}))
}
