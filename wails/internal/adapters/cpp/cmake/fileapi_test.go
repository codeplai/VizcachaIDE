package cmake

import (
	"path/filepath"
	"testing"
)

// writeReply writes the File API reply CMake gives for two programs and a library.
func writeReply(t *testing.T, buildDir, source string) {
	t.Helper()
	reply := filepath.Join(apiFolder(buildDir), "reply")
	write(t, filepath.Join(reply, "index-2026-10-05T10-00-00-0000.json"), `{"objects":[{"kind":"codemodel","jsonFile":"codemodel-v2-aa.json"}]}`)
	write(t, filepath.Join(reply, "codemodel-v2-aa.json"), `{"paths":{"source":"`+filepath.ToSlash(source)+`"},"configurations":[{"targets":[
		{"name":"calc","jsonFile":"target-calc.json"},{"name":"juego","jsonFile":"target-juego.json"},{"name":"util","jsonFile":"target-util.json"}]}]}`)
	write(t, filepath.Join(reply, "target-calc.json"), `{"name":"calc","type":"EXECUTABLE","artifacts":[{"path":"calc.exe"}],"sources":[{"path":"calc.cpp"},{"path":"suma.cpp"}]}`)
	write(t, filepath.Join(reply, "target-juego.json"), `{"name":"juego","type":"EXECUTABLE","artifacts":[{"path":"juego.exe"}],"sources":[{"path":"juego/main.cpp"}]}`)
	write(t, filepath.Join(reply, "target-util.json"), `{"name":"util","type":"STATIC_LIBRARY","artifacts":[{"path":"libutil.a"}],"sources":[{"path":"util.cpp"}]}`)
}

func TestExecutablesAndChoose(t *testing.T) {
	source, build := t.TempDir(), filepath.Join(t.TempDir(), "build")
	writeReply(t, build, source)
	executables, err := Executables(build)
	if err != nil || len(executables) != 2 {
		t.Fatalf("executables = %+v, %v, want the two programs and not the library", executables, err)
	}
	if executables[1].Path != filepath.Join(build, "juego.exe") || executables[1].Sources[0] != filepath.Join(source, "juego", "main.cpp") {
		t.Fatalf("juego = %+v", executables[1])
	}
	cases := []struct{ active, target, want string }{
		{filepath.Join(source, "juego", "main.cpp"), "calc", "juego"}, // the active file wins
		{filepath.Join(source, "SUMA.cpp"), "juego", "calc"},          // by sources, whatever the case
		{filepath.Join(source, "otro.cpp"), "juego", "juego"},         // else the project's target
		{"", "nada", "calc"}, // else the first
	}
	for _, c := range cases {
		if chosen, ok := Choose(executables, c.active, c.target); !ok || chosen.Name != c.want {
			t.Errorf("Choose(%q, %q) = %q, want %q", c.active, c.target, chosen.Name, c.want)
		}
	}
	if _, ok := Choose(nil, "", ""); ok {
		t.Error("no executables, nothing to choose")
	}
}

func TestExecutablesWithoutReplyIsNoExecutable(t *testing.T) {
	if _, err := Executables(t.TempDir()); err == nil {
		t.Fatal("a build folder CMake never answered for has no executables")
	}
}

func TestQueryIsWrittenBeforeConfigure(t *testing.T) {
	build := t.TempDir()
	if err := writeQuery(build); err != nil {
		t.Fatal(err)
	}
	if !isFile(filepath.Join(apiFolder(build), "query", "codemodel-v2")) {
		t.Fatal("the codemodel-v2 query is missing")
	}
}
