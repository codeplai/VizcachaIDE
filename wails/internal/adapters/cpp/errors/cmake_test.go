package errors

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const projectFolder = `D:\Users\ana\Ñandú`

// parseCMake reads CMake output the way the app does, with the project folder as working folder.
func parseCMake(t *testing.T, output string) []domain.Diagnostic {
	t.Helper()
	return newExplainer(t).Parse(output, projectFolder)
}

func TestCMakeErrorPointsAtTheLineOfCMakeLists(t *testing.T) {
	output := "-- The CXX compiler identification is Clang 23.1.2\n" +
		"CMake Error at CMakeLists.txt:12 (find_package):\n" +
		"  Could not find a package configuration file provided by \"fmt\" with any\n" +
		"  of the following names:\n\n    fmtConfig.cmake\n\n" +
		"Call Stack (most recent call first):\n  CMakeLists.txt:9 (include)\n\n" +
		"-- Configuring incomplete, errors occurred!\n"
	diagnostics := parseCMake(t, output)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	found := diagnostics[0]
	if found.Location == nil || found.Location.Line != 12 || !strings.HasSuffix(found.Location.File, "CMakeLists.txt") || !strings.Contains(found.Location.File, "Ñandú") {
		t.Fatalf("location = %+v", found.Location)
	}
	if found.Severity != domain.SeverityError || found.Source != "cmake" || !strings.HasPrefix(found.Message, "Could not find a package configuration file") {
		t.Fatalf("diagnostic = %+v", found)
	}
	if found.Code != "CPP-VCPKG-PACKAGE-NOT-FOUND" || !strings.Contains(found.RawText, "of the following names") {
		t.Fatalf("code = %q, raw = %q", found.Code, found.RawText)
	}
}

func TestCMakeNoSourcesNamesTheTarget(t *testing.T) {
	diagnostics := parseCMake(t, "CMake Error at CMakeLists.txt:8 (add_executable):\n  No SOURCES given to target: nandu\n\n")
	if len(diagnostics) != 1 || diagnostics[0].Code != "CPP-CMAKE-NO-SOURCES" || diagnostics[0].Location.Line != 8 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	for _, language := range []string{"en", "es"} {
		explanation := newExplainer(t).Explain(diagnostics[0], language)
		if explanation == nil || !strings.Contains(explanation.Title, `"nandu"`) || unfilled.MatchString(explanation.Title+explanation.Body+explanation.FixHint) {
			t.Errorf("%s: explanation = %+v", language, explanation)
		}
	}
}

func TestSymbolDefinedTwiceSuggestsSeparateFolders(t *testing.T) {
	lld := "ld.lld: error: duplicate symbol: main\n>>> defined at " + projectFolder + `\otra.cpp:3` + "\n>>> otra.cpp.obj:(main)\n>>> defined at " + projectFolder + `\main.cpp:5` + "\n"
	gnu := "ld: CMakeFiles/x.dir/b.cpp.obj: in function `main':\n" + strings.ReplaceAll(projectFolder, `\`, "/") + "/b.cpp:3:(.text+0x0): multiple definition of `main'; CMakeFiles/x.dir/a.cpp.obj:" + strings.ReplaceAll(projectFolder, `\`, "/") + "/a.cpp:3:(.text+0x0): first defined here\n"
	for name, output := range map[string]string{"lld": lld, "gnu": gnu} {
		diagnostics := parseCMake(t, output)
		found := withCode(diagnostics, "CPP-CMAKE-MULTIPLE-MAIN")
		if found == nil || found.Location == nil || !strings.Contains(found.Location.File, "Ñandú") {
			t.Errorf("%s: diagnostics = %+v", name, diagnostics)
			continue
		}
		explanation := newExplainer(t).Explain(*found, "es")
		if explanation == nil || !strings.Contains(explanation.FixHint, "carpeta") {
			t.Errorf("%s: explanation = %+v", name, explanation)
		}
	}
}

func TestLongPathWarningAndMissingToolsAreExplained(t *testing.T) {
	long := "CMake Warning in D:/x/build/CMakeFiles/nandu.dir:\n  The object file directory\n\n    D:/x/build/CMakeFiles/nandu.dir/./\n\n  has 200 characters.\n"
	diagnostics := parseCMake(t, long)
	if len(diagnostics) != 1 || diagnostics[0].Severity != domain.SeverityWarning || diagnostics[0].Code != "CPP-CMAKE-LONG-PATH" || diagnostics[0].Location != nil {
		t.Fatalf("long path = %+v", diagnostics)
	}
	missing := "CMake Error: CMake was unable to find a build program corresponding to \"Ninja\".  CMAKE_MAKE_PROGRAM is not set.\n"
	diagnostics = parseCMake(t, missing)
	if len(diagnostics) != 1 || diagnostics[0].Code != "CPP-CMAKE-TOOLS-MISSING" {
		t.Fatalf("missing tool = %+v", diagnostics)
	}
}

func TestCMakeDoesNotDisturbCompilerDiagnostics(t *testing.T) {
	output := "CMake Error at CMakeLists.txt:3 (foo):\n  Unknown CMake command \"foo\".\n\n" + projectFolder + `\main.cpp:4:5: error: use of undeclared identifier 'x'` + "\n"
	diagnostics := parseCMake(t, output)
	if got := codes(diagnostics); len(got) != 2 || got[0] != "CPP-CMAKE-ERROR" || got[1] != "CPP-UNDECLARED" {
		t.Fatalf("codes = %v", got)
	}
}
