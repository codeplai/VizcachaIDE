package packages

import (
	"strings"
	"testing"
)

func TestScaffoldCrateNames(t *testing.T) {
	cases := map[string]string{
		"hola":         `name = "hola"`,
		"Mi Programa":  `name = "mi_programa"`,
		"Ñandú":        `name = "nandu"`,
		"2048":         `name = "app_2048"`,
		"Hola Mundo 2": `name = "hola_mundo_2"`,
	}
	for name, want := range cases {
		files, main := Scaffold{}.Scaffold(name)
		if main != "src/main.rs" {
			t.Fatalf("main file = %q", main)
		}
		manifest := files["Cargo.toml"]
		for _, line := range []string{want, `version = "0.1.0"`, `edition = "2021"`, "[dependencies]"} {
			if !strings.Contains(manifest, line) {
				t.Errorf("%q: Cargo.toml lacks %s: %q", name, line, manifest)
			}
		}
	}
}

func TestScaffoldFiles(t *testing.T) {
	files, _ := Scaffold{}.Scaffold("hola")
	if files[".gitignore"] != "/target\n" {
		t.Errorf(".gitignore = %q", files[".gitignore"])
	}
	if !strings.Contains(files["src/main.rs"], "read_line") || !strings.Contains(files["src/main.rs"], `println!("Hola, {}", nombre.trim())`) {
		t.Errorf("main.rs does not greet: %q", files["src/main.rs"])
	}
	if len(files) != 3 {
		t.Errorf("files = %d, want 3", len(files))
	}
}
