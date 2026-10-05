package cpp

import (
	"strings"
	"testing"
)

func TestScaffoldFiles(t *testing.T) {
	for _, name := range []string{"hola", "Mi Programa", "Ñandú"} {
		files, main := Scaffold{}.Scaffold(name)
		if main != "main.cpp" {
			t.Fatalf("%q: main = %q", name, main)
		}
		for _, file := range []string{"CMakeLists.txt", "CMakePresets.json", "vcpkg.json", ".gitignore", ".clang-format", "main.cpp"} {
			if files[file] == "" {
				t.Errorf("%q: %s is missing", name, file)
			}
		}
		if _, found := files["compile_flags.txt"]; found || len(files) != 6 {
			t.Errorf("%q: files = %d, compile_flags.txt must not exist", name, len(files))
		}
		if files[".clang-format"] != "BasedOnStyle: LLVM\nIndentWidth: 4\n" {
			t.Errorf(".clang-format = %q", files[".clang-format"])
		}
		for _, part := range []string{"<iostream>", "<string>", "std::getline", "Hola, "} {
			if !strings.Contains(files["main.cpp"], part) {
				t.Errorf("main.cpp lacks %s", part)
			}
		}
	}
}

func TestScaffoldTargetAndLanguage(t *testing.T) {
	files, _ := Scaffold{}.Scaffold("Ñandú")
	if !strings.Contains(files["CMakeLists.txt"], "add_executable(nandu ") || !strings.Contains(files["vcpkg.json"], `"nandu"`) {
		t.Errorf("target is not nandu:\n%s", files["CMakeLists.txt"])
	}
	spanish, _ := Scaffold{Language: func() string { return "es" }}.Scaffold("x")
	english, _ := Scaffold{}.Scaffold("x")
	if spanish["CMakeLists.txt"] == english["CMakeLists.txt"] {
		t.Error("the comments of CMakeLists.txt do not follow the language")
	}
}
