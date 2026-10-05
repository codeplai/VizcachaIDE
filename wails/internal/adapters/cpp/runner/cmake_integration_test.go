package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cpperrors "github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/errors"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const (
	greetingHeader = "#pragma once\nvoid saludar();\n"
	greetingSource = "#include <iostream>\n#include \"saludo.h\"\nvoid saludar() {\n    std::cout << \"¿Cómo estás, Ñandú?\" << std::endl;\n}\n"
	mainSource     = "#include \"saludo.h\"\nint main() {\n    saludar();\n    return 0;\n}\n"
)

// studentFolder writes a student's folder: Ñandú with main.cpp calling a function of saludo.cpp.
func studentFolder(t *testing.T) (folder string) {
	t.Helper()
	folder = filepath.Join(t.TempDir(), "Ñandú")
	for name, text := range map[string]string{"main.cpp": mainSource, "saludo.cpp": greetingSource, "saludo.h": greetingHeader} {
		if err := os.MkdirAll(folder, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(folder, name), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return folder
}

// Every C++ project is a CMake project: the whole path of docs/PLAN_CPP_CMAKE.md section 4.1 with
// the real tools, in a folder with an accent and two sources.
func TestCMakeProjectRunDebugBuildAndProblems(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		folder := studentFolder(t)
		ctx := context.Background()
		config := r.Configure(filepath.Join(folder, "main.cpp"), nil)
		if config.Mode != domain.RunProject || config.Target != folder {
			t.Fatalf("config = %+v, want the CMake project at %s", config, folder)
		}

		t.Run("run prints accents", func(t *testing.T) {
			if err := r.Run(ctx, config); err != nil {
				t.Fatal(err)
			}
			if code := sink.waitFinished(t); code != 0 {
				t.Fatalf("exit code %d, stdout %q stderr %q", code, firstText(sink), errText(sink))
			}
			out, _ := sink.text()
			if !strings.Contains(out, "¿Cómo estás, Ñandú?") || strings.Contains(out, "[1/") || strings.Contains(out, "-- ") {
				t.Fatalf("stdout = %q, want the program's text without the build's noise", out)
			}
		})

		t.Run("debug build has debug information", func(t *testing.T) {
			exe, output, err := r.CompileForDebug(ctx, config)
			if err != nil {
				t.Fatalf("err = %v, output = %q", err, output)
			}
			data, err := os.ReadFile(exe)
			if err != nil || !strings.Contains(string(data), ".debug_info") {
				t.Fatalf("exe %s: %v, want DWARF sections", exe, err)
			}
		})

		t.Run("build puts the program beside the code", func(t *testing.T) {
			if err := r.Build(ctx, config); err != nil {
				t.Fatal(err)
			}
			if code := sink.waitFinished(t); code != 0 {
				t.Fatalf("exit code %d, stderr %q", code, errText(sink))
			}
			if _, err := os.Stat(filepath.Join(folder, executableName("nandu"))); err != nil {
				t.Fatalf("no program beside the code: %v", err)
			}
		})

		t.Run("a syntax error is a Problem on the right file and line", func(t *testing.T) {
			broken := strings.Replace(greetingSource, "std::endl;", "std::endl", 1)
			if err := os.WriteFile(filepath.Join(folder, "saludo.cpp"), []byte(broken), 0o600); err != nil {
				t.Fatal(err)
			}
			output, err := r.Check(ctx, config)
			if err != nil {
				t.Fatal(err)
			}
			found := problemIn(t, output, folder, "saludo.cpp")
			if found.Location.Line != 4 || found.Severity != domain.SeverityError {
				t.Fatalf("problem = %+v, want an error on line 4\n%s", found, output)
			}
		})

		t.Run("an error in CMakeLists.txt is a Problem on CMakeLists.txt", func(t *testing.T) {
			lists := filepath.Join(folder, "CMakeLists.txt")
			text, err := os.ReadFile(lists)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(lists, append(text, []byte("find_package(NoExisteEstaLibreria REQUIRED)\n")...), 0o600); err != nil {
				t.Fatal(err)
			}
			output, err := r.Check(ctx, config)
			if err != nil {
				t.Fatal(err)
			}
			found := problemIn(t, output, folder, "CMakeLists.txt")
			if want := strings.Count(string(text), "\n") + 1; found.Location.Line != want {
				t.Fatalf("problem = %+v, want CMakeLists.txt line %d\n%s", found, want, output)
			}
			if strings.Contains(output, "-- ") {
				t.Fatalf("output keeps CMake's status lines: %q", output)
			}
		})
	})
}

// problemIn parses output like the app does and returns the first problem in the file.
func problemIn(t *testing.T, output, folder, name string) domain.Diagnostic {
	t.Helper()
	explainer, err := cpperrors.NewExplainer()
	if err != nil {
		t.Fatal(err)
	}
	for _, diagnostic := range explainer.Parse(output, folder) {
		if diagnostic.Location != nil && strings.EqualFold(filepath.Base(diagnostic.Location.File), name) &&
			strings.EqualFold(filepath.Clean(filepath.Dir(diagnostic.Location.File)), filepath.Clean(folder)) {
			return diagnostic
		}
	}
	t.Fatalf("no problem in %s:\n%s", name, output)
	return domain.Diagnostic{}
}

func firstText(sink *testSink) string {
	out, _ := sink.text()
	return out
}

// Two programs in one folder cannot link: the error says main is defined twice and points at a file.
func TestTwoMainsInOneFolderAreExplained(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, _ *testSink) {
		folder := studentFolder(t)
		if err := os.WriteFile(filepath.Join(folder, "otro.cpp"), []byte("int main() { return 1; }\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		output, err := r.Check(context.Background(), r.Configure(filepath.Join(folder, "main.cpp"), nil))
		if err != nil {
			t.Fatal(err)
		}
		explainer, err := cpperrors.NewExplainer()
		if err != nil {
			t.Fatal(err)
		}
		for _, diagnostic := range explainer.Parse(output, folder) {
			if diagnostic.Code == "CPP-CMAKE-MULTIPLE-MAIN" {
				return
			}
		}
		t.Fatalf("no CPP-CMAKE-MULTIPLE-MAIN in:\n%s", output)
	})
}
