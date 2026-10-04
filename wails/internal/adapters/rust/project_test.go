package rust_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// tree writes files (slash paths) under a temp folder and returns it.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, text := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestLooseFileHasNoProject(t *testing.T) {
	root := tree(t, map[string]string{"hola.rs": "fn main() {}"})
	if _, ok, err := rust.FindProject(filepath.Join(root, "hola.rs")); ok || err != nil {
		t.Errorf("ok = %v, err = %v", ok, err)
	}
}

func TestCrateWithSeveralPrograms(t *testing.T) {
	root := tree(t, map[string]string{
		"Cargo.toml":           "[package]\nname = \"juego\"\nedition = \"2021\"\n",
		"src/main.rs":          "fn main() {}",
		"src/util.rs":          "",
		"src/bin/extra.rs":     "fn main() {}",
		"src/bin/otro/main.rs": "fn main() {}",
	})
	project, ok, err := rust.FindProject(filepath.Join(root, "src", "util.rs"))
	if !ok || err != nil || project.Name != "juego" || project.Edition != "2021" || project.Workspace != root {
		t.Fatalf("project = %+v, %v, %v", project, ok, err)
	}
	if context := project.Context(); context.Kind != domain.ProjectCargo || context.Root != root {
		t.Errorf("context = %+v", context)
	}
	cases := map[string]string{
		filepath.Join(root, "src", "util.rs"):             "juego",
		filepath.Join(root, "src", "bin", "extra.rs"):     "extra",
		filepath.Join(root, "src", "bin", "otro", "x.rs"): "otro",
	}
	for file, want := range cases {
		if binary, ok := project.BinaryFor(file); !ok || binary.Name != want {
			t.Errorf("BinaryFor(%s) = %+v, %v; want %s", file, binary, ok, want)
		}
	}
}

func TestLibraryHasNoProgram(t *testing.T) {
	root := tree(t, map[string]string{"Cargo.toml": "[package]\nname = \"util\"\n", "src/lib.rs": ""})
	project, ok, _ := rust.FindProject(filepath.Join(root, "src", "lib.rs"))
	if _, hasMain := project.BinaryFor(filepath.Join(root, "src", "lib.rs")); !ok || hasMain || project.Edition != "2015" {
		t.Errorf("project = %+v", project)
	}
}

func TestWorkspaceMembersAndTheVirtualRoot(t *testing.T) {
	root := tree(t, map[string]string{
		"Cargo.toml":             "[workspace]\nmembers = [\"crates/*\"]\n",
		"crates/app/Cargo.toml":  "[package]\nname = \"app\"\nedition = \"2024\"\n",
		"crates/app/src/main.rs": "fn main() {}",
		"crates/util/Cargo.toml": "[package]\nname = \"util\"\n",
		"crates/util/src/lib.rs": "",
		"notas.rs":               "",
	})
	member, ok, err := rust.FindProject(filepath.Join(root, "crates", "app", "src", "main.rs"))
	if !ok || err != nil || member.Name != "app" || member.Workspace != root || member.Virtual {
		t.Fatalf("member = %+v, %v, %v", member, ok, err)
	}
	virtual, ok, err := rust.FindProject(filepath.Join(root, "notas.rs"))
	if !ok || err != nil || !virtual.Virtual || len(virtual.Members) != 2 {
		t.Fatalf("virtual = %+v, %v, %v", virtual, ok, err)
	}
}

func TestBrokenManifestIsAnError(t *testing.T) {
	root := tree(t, map[string]string{"Cargo.toml": "[package\nname =", "src/main.rs": ""})
	if _, _, err := rust.FindProject(filepath.Join(root, "src", "main.rs")); err == nil {
		t.Error("no error for a broken Cargo.toml")
	}
}
