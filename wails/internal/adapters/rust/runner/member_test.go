package runner

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTheChosenMemberRunsItsMainProgram(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"Cargo.toml":      "[workspace]\nmembers = [\"app\", \"cli\"]\n",
		"app/Cargo.toml":  "[package]\nname = \"app\"\n",
		"app/src/main.rs": "fn main() {}",
		"cli/Cargo.toml":  "[package]\nname = \"cli\"\n",
		"cli/src/main.rs": "fn main() {}",
		"notas.rs":        "",
	}
	for name, text := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runner := &Runner{}
	config, err := runner.ConfigureMember(filepath.Join(root, "notas.rs"), "cli", []string{"x"})
	if err != nil {
		t.Fatal(err)
	}
	if config.Target != filepath.Join(root, "cli", "src", "main.rs") || config.Project == nil || config.Project.Name != "cli" {
		t.Errorf("config = %+v", config)
	}
	if _, err := runner.ConfigureMember(filepath.Join(root, "notas.rs"), "otro", nil); !errors.Is(err, ErrChooseMember) {
		t.Errorf("unknown member: %v", err)
	}
}
