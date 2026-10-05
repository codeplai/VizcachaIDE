package vcpkg

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const template = `cmake_minimum_required(VERSION 3.25)
project(nandu LANGUAGES CXX)
add_executable(nandu main.cpp)

` + LibrariesBegin + `
` + LibrariesEnd + `
`

func TestAddDependencyKeepsKeyOrderAndTheRestOfTheManifest(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "vcpkg.json"), `{
  "$schema": "https://example/x.json",
  "name": "ñandú",
  "version": "0.1.0",
  "dependencies": [
    "zlib",
    { "name": "boost-asio", "platform": "windows" }
  ],
  "overrides": [{"name": "zlib", "version": "1.2"}]
}
`)
	added, err := AddDependency(dir, "fmt")
	if err != nil || !added {
		t.Fatalf("add: %v %v", added, err)
	}
	if again, _ := AddDependency(dir, "fmt"); again {
		t.Error("a second add must report no change")
	}
	if again, _ := AddDependency(dir, "boost-asio"); again {
		t.Error("an object dependency must count as present")
	}
	want := `{
  "$schema": "https://example/x.json",
  "name": "ñandú",
  "version": "0.1.0",
  "dependencies": [
    "zlib",
    {
      "name": "boost-asio",
      "platform": "windows"
    },
    "fmt"
  ],
  "overrides": [
    {
      "name": "zlib",
      "version": "1.2"
    }
  ]
}
`
	if got := readText(t, filepath.Join(dir, "vcpkg.json")); got != want {
		t.Errorf("manifest =\n%s\nwant\n%s", got, want)
	}
	if removed, err := RemoveDependency(dir, "boost-asio"); err != nil || !removed {
		t.Fatalf("remove: %v %v", removed, err)
	}
	names, err := Dependencies(dir)
	if err != nil || strings.Join(names, ",") != "zlib,fmt" {
		t.Errorf("dependencies = %v, %v", names, err)
	}
	if removed, _ := RemoveDependency(dir, "nothing"); removed {
		t.Error("removing an absent port must report no change")
	}
}

func TestAddDependencyOnTheScaffoldAndWithoutManifest(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Mi Programa")
	mustWrite(t, filepath.Join(dir, "vcpkg.json"), `{ "name": "nandu", "version": "0.1.0", "dependencies": [] }`)
	if _, err := AddDependency(dir, "fmt"); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, filepath.Join(dir, "vcpkg.json")); !strings.Contains(got, "\"dependencies\": [\n    \"fmt\"\n  ]") {
		t.Errorf("manifest = %s", got)
	}
	empty := filepath.Join(t.TempDir(), "Ñandú Nuevo")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Dependencies(empty); !errors.Is(err, ErrNoManifest) {
		t.Errorf("err = %v", err)
	}
	if _, err := AddDependency(empty, "fmt"); err != nil {
		t.Fatal(err)
	}
	got := readText(t, filepath.Join(empty, "vcpkg.json"))
	if !strings.Contains(got, `"name": "nandu-nuevo"`) {
		t.Errorf("manifest = %s", got)
	}
}

func TestLibraryBlockGroupsLinesPerPort(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "CMakeLists.txt"), template)
	fmtLines := []string{"find_package(fmt CONFIG REQUIRED)", "target_link_libraries(nandu PRIVATE fmt::fmt)"}
	if err := SetLibraryLines(dir, "fmt", fmtLines); err != nil {
		t.Fatal(err)
	}
	if err := SetLibraryLines(dir, "zlib", []string{"find_package(ZLIB REQUIRED)"}); err != nil {
		t.Fatal(err)
	}
	want := "cmake_minimum_required(VERSION 3.25)\nproject(nandu LANGUAGES CXX)\nadd_executable(nandu main.cpp)\n\n" +
		LibrariesBegin + "\n# fmt\n" + strings.Join(fmtLines, "\n") + "\n# zlib\nfind_package(ZLIB REQUIRED)\n" + LibrariesEnd + "\n"
	if got := readText(t, filepath.Join(dir, "CMakeLists.txt")); got != want {
		t.Errorf("CMakeLists =\n%s\nwant\n%s", got, want)
	}
	if err := SetLibraryLines(dir, "fmt", []string{"find_package(fmt CONFIG REQUIRED)"}); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, filepath.Join(dir, "CMakeLists.txt")); strings.Count(got, "# fmt") != 1 || strings.Contains(got, "target_link_libraries") {
		t.Errorf("set must replace the group of the port:\n%s", got)
	}
	if err := RemoveLibraryLines(dir, "fmt"); err != nil {
		t.Fatal(err)
	}
	if err := RemoveLibraryLines(dir, "zlib"); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, filepath.Join(dir, "CMakeLists.txt")); got != template {
		t.Errorf("after removing everything:\n%s", got)
	}
}

func TestLibraryBlockIsAppendedWhenTheMarkersAreMissingAndCRLFSurvives(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "CMakeLists.txt"), "project(x)\r\nadd_executable(x main.cpp)\r\n")
	if err := RemoveLibraryLines(dir, "fmt"); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, filepath.Join(dir, "CMakeLists.txt")); got != "project(x)\r\nadd_executable(x main.cpp)\r\n" {
		t.Errorf("removing from a file without block changed it: %q", got)
	}
	if err := SetLibraryLines(dir, "fmt", []string{"find_package(fmt CONFIG REQUIRED)"}); err != nil {
		t.Fatal(err)
	}
	want := "project(x)\r\nadd_executable(x main.cpp)\r\n" + LibrariesBegin + "\r\n# fmt\r\nfind_package(fmt CONFIG REQUIRED)\r\n" + LibrariesEnd + "\r\n"
	if got := readText(t, filepath.Join(dir, "CMakeLists.txt")); got != want {
		t.Errorf("got %q\nwant %q", got, want)
	}
	if err := SetLibraryLines(t.TempDir(), "fmt", nil); !errors.Is(err, ErrNoCMakeLists) {
		t.Errorf("err = %v", err)
	}
}

func TestStudentLinesInsideTheBlockAreKept(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "CMakeLists.txt"), LibrariesBegin+"\nfind_package(Threads REQUIRED)\n"+LibrariesEnd+"\n")
	if err := SetLibraryLines(dir, "fmt", []string{"find_package(fmt CONFIG REQUIRED)"}); err != nil {
		t.Fatal(err)
	}
	if err := RemoveLibraryLines(dir, "fmt"); err != nil {
		t.Fatal(err)
	}
	if got := readText(t, filepath.Join(dir, "CMakeLists.txt")); !strings.Contains(got, "find_package(Threads REQUIRED)") {
		t.Errorf("the student's line was lost:\n%s", got)
	}
}
