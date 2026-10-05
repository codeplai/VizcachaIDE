package cpp

import (
	"strings"
	"testing"
)

func TestScaffoldFiles(t *testing.T) {
	for _, name := range []string{"hola", "Mi Programa", "Ñandú"} {
		files, main := Scaffold{}.Scaffold(name)
		if main != "main.cpp" || len(files) != 2 {
			t.Fatalf("%q: main = %q, files = %d", name, main, len(files))
		}
		if files["compile_flags.txt"] != "-std=c++17\n-Wall\n-Wextra\n" {
			t.Errorf("compile_flags.txt = %q", files["compile_flags.txt"])
		}
		for _, part := range []string{"<iostream>", "<string>", "std::getline", "Hola, "} {
			if !strings.Contains(files["main.cpp"], part) {
				t.Errorf("main.cpp lacks %s", part)
			}
		}
	}
}
