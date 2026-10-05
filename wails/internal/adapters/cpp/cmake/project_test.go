package cmake

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTargetName(t *testing.T) {
	cases := map[string]string{
		"Ñandú":         "nandu",
		"Mi Programa":   "mi_programa",
		"  hola--mundo": "hola_mundo",
		"2da tarea":     "app_2da_tarea",
		"":              "app",
		"¿?":            "app",
		"test":          "app_test",
		"Año_2024":      "ano_2024",
		"日本語":           "app",
	}
	for folder, want := range cases {
		if got := TargetName(folder); got != want {
			t.Errorf("TargetName(%q) = %q, want %q", folder, got, want)
		}
	}
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestFindProjectGoesUpToTheTopButNotAboveStop(t *testing.T) {
	top := t.TempDir()
	inner := filepath.Join(top, "libs", "util")
	write(t, filepath.Join(top, ListsFile), "project(top)\nadd_executable(principal main.cpp)\n")
	write(t, filepath.Join(inner, ListsFile), "project(util)\n")
	write(t, filepath.Join(inner, "util.cpp"), "")

	project, ok := FindProject(filepath.Join(inner, "util.cpp"), "")
	if !ok || project.Root != top || project.Target != "principal" {
		t.Fatalf("project = %+v, %v, want the top folder and its first add_executable", project, ok)
	}
	project, ok = FindProject(filepath.Join(inner, "util.cpp"), filepath.Join(top, "libs"))
	if !ok || project.Root != inner || project.Target != "util" {
		t.Fatalf("project = %+v, %v, want to stop at libs", project, ok)
	}
	if _, ok := FindProject(filepath.Join(inner, "util.cpp"), filepath.Join(inner, "elsewhere")); !ok {
		t.Fatal("a stop folder that does not hold the file only limits the search to its folder")
	}
	if _, ok := FindProject(t.TempDir(), ""); ok {
		t.Fatal("an empty folder has no project")
	}
}

func TestGenerateWritesTheTemplatesAndNeverOverwrites(t *testing.T) {
	folder := filepath.Join(t.TempDir(), "Ñandú")
	write(t, filepath.Join(folder, "main.cpp"), "int main() {}\n")
	write(t, filepath.Join(folder, ".gitignore"), "*.o\n")

	project, err := Generate(folder, "es")
	if err != nil || project.Root != folder || project.Target != "nandu" {
		t.Fatalf("Generate = %+v, %v", project, err)
	}
	lists := read(t, filepath.Join(folder, ListsFile))
	for _, want := range []string{"project(nandu LANGUAGES CXX)", "add_executable(nandu ${SOURCES})", "set(CMAKE_CXX_EXTENSIONS OFF)", "Todo .cpp de esta carpeta", LibrariesBegin + "\n" + LibrariesEnd + "\n"} {
		if !strings.Contains(lists, want) {
			t.Errorf("CMakeLists.txt lacks %q:\n%s", want, lists)
		}
	}
	if got := read(t, filepath.Join(folder, IgnoreFile)); got != "*.o\n" {
		t.Errorf(".gitignore was overwritten: %q", got)
	}
	if got := read(t, filepath.Join(folder, ManifestFile)); !strings.Contains(got, `"name": "nandu"`) || !strings.Contains(got, `"dependencies": []`) {
		t.Errorf("vcpkg.json = %q", got)
	}
	if got := read(t, filepath.Join(folder, PresetsFile)); !strings.Contains(got, `$env{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake`) || !strings.Contains(got, `"binaryDir": "${sourceDir}/build"`) {
		t.Errorf("CMakePresets.json = %q", got)
	}

	write(t, filepath.Join(folder, ListsFile), "mine\n")
	if _, err := Generate(folder, "en"); err != nil || read(t, filepath.Join(folder, ListsFile)) != "mine\n" {
		t.Fatalf("a second Generate must keep the file: %v", err)
	}
}

func TestFilesInEnglishAndTheMarkersDoNotChange(t *testing.T) {
	english, spanish := Files("Mi Programa", "en"), Files("Mi Programa", "es")
	if !strings.Contains(english[ListsFile], "Every .cpp in this folder") || strings.Contains(spanish[ListsFile], "Every .cpp") {
		t.Fatalf("languages: %q / %q", english[ListsFile], spanish[ListsFile])
	}
	for _, text := range []string{english[ListsFile], spanish[ListsFile]} {
		if !strings.Contains(text, "# VizcachaIDE libraries (vcpkg) · begin") || !strings.Contains(text, "# VizcachaIDE libraries (vcpkg) · end") {
			t.Errorf("markers missing in %q", text)
		}
	}
	if !strings.Contains(english[ManifestFile], `"name": "mi-programa"`) {
		t.Errorf("vcpkg names allow only a-z, 0-9 and -: %s", english[ManifestFile])
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
