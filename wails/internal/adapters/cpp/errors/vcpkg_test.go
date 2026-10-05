package errors

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// Real output of vcpkg 2026-09 inside "cmake -S" (see docs/PLAN_CPP_CMAKE.md section 8), and the
// CMake error of a find_package without the library.
const (
	missingPortOutput = `D:\vcpkg\ports\nonexistentport: error: nonexistentport does not exist
`
	downloadFailedOutput = `Building fmt:x64-mingw-static@12.2.0#1...
Downloading https://github.com/fmtlib/fmt/commit/588b3a0.patch?full_index=1 -> fmt-backport-4813.patch
error: curl operation failed with error code 7 (Could not connect to server).
error: Not a transient network error, won't retry download from https://github.com/fmtlib/fmt/commit/588b3a0.patch
note: If you are using a proxy, please ensure your proxy settings are correct.
CMake Error at scripts/cmake/vcpkg_download_distfile.cmake:134 (message):
  Download failed, halting portfile.
Call Stack (most recent call first):
  scripts/cmake/vcpkg_download_distfile.cmake:156 (z_vcpkg_download_distfile)


error: building fmt:x64-mingw-static failed with: BUILD_FAILED
See https://learn.microsoft.com/vcpkg/troubleshoot/build-failures for more information.
`
	packageNotFoundOutput = `CMake Error at CMakeLists.txt:14 (find_package):
  Could not find a package configuration file provided by "spdlog" with any
  of the following names:

    spdlogConfig.cmake
    spdlog-config.cmake
`
	hostToolOutput = `CMake Error at scripts/cmake/vcpkg_find_acquire_program.cmake:50 (message):
  Could not find perl. Please install it.
`
)

func diagnosticsOf(t *testing.T, output string) []domain.Diagnostic {
	t.Helper()
	explainer := newExplainer(t)
	return explainer.Parse(output, `C:\proyectos\Ñandú`)
}

func TestVcpkgMissingPortIsExplained(t *testing.T) {
	diagnostics := diagnosticsOf(t, missingPortOutput)
	if len(diagnostics) != 1 || diagnostics[0].Code != "CPP-VCPKG-PORT-MISSING" || diagnostics[0].Source != "vcpkg" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	explanation := newExplainer(t).Explain(diagnostics[0], "es")
	if explanation.Placeholders["port"] != "nonexistentport" || !strings.Contains(explanation.Title, `"nonexistentport"`) {
		t.Errorf("explanation = %+v", explanation)
	}
}

func TestVcpkgDownloadFailureAndBuildFailureAreTwoDiagnostics(t *testing.T) {
	diagnostics := diagnosticsOf(t, downloadFailedOutput)
	var codes []string
	for _, item := range diagnostics {
		codes = append(codes, item.Code)
	}
	want := "CPP-VCPKG-NO-INTERNET,CPP-VCPKG-NO-INTERNET,CPP-VCPKG-BUILD-FAILED"
	if got := strings.Join(codes, ","); got != want {
		t.Fatalf("codes = %s, want %s", got, want)
	}
	last := diagnostics[len(diagnostics)-1]
	explanation := newExplainer(t).Explain(last, "en")
	if explanation.Placeholders["port"] != "fmt" || !strings.Contains(explanation.Title, `"fmt"`) {
		t.Errorf("explanation = %+v", explanation)
	}
}

func TestCMakeCannotFindAPackageSuggestsInstallingIt(t *testing.T) {
	diagnostics := diagnosticsOf(t, packageNotFoundOutput)
	if len(diagnostics) != 1 || diagnostics[0].Code != "CPP-VCPKG-PACKAGE-NOT-FOUND" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	location := diagnostics[0].Location
	if location == nil || location.Line != 14 || !strings.HasSuffix(location.File, "CMakeLists.txt") {
		t.Errorf("location = %+v", location)
	}
	for language, want := range map[string]string{"en": "Packages", "es": "Paquetes"} {
		explanation := newExplainer(t).Explain(diagnostics[0], language)
		if explanation.Placeholders["name"] != "spdlog" || !strings.Contains(explanation.FixHint, want) || !strings.Contains(explanation.FixHint, `"spdlog"`) {
			t.Errorf("%s: %+v", language, explanation)
		}
	}
}

func TestMissingHostToolIsExplained(t *testing.T) {
	diagnostics := diagnosticsOf(t, hostToolOutput)
	if len(diagnostics) != 1 || diagnostics[0].Code != "CPP-VCPKG-HOST-TOOL" {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	if explanation := newExplainer(t).Explain(diagnostics[0], "en"); explanation.Placeholders["tool"] != "perl" {
		t.Errorf("explanation = %+v", explanation)
	}
}

func TestOtherErrorLinesAreNotTurnedIntoLibraryDiagnostics(t *testing.T) {
	// "CMake Error at …" lines are the CMake handler's (cmake_test.go), not the libraries'.
	output := "error: something else entirely\n"
	if diagnostics := diagnosticsOf(t, output); len(diagnostics) != 0 {
		t.Errorf("diagnostics = %+v", diagnostics)
	}
}

func TestEveryLibraryEntryHasBothLanguagesAndNoUnfilledPlaceholders(t *testing.T) {
	explainer := newExplainer(t)
	samples := map[string]string{
		"CPP-VCPKG-NO-INTERNET":       "curl operation failed with error code 6 (Couldn't resolve host name).",
		"CPP-VCPKG-PORT-MISSING":      "boost-foo does not exist",
		"CPP-VCPKG-BUILD-FAILED":      "building zlib:x64-linux failed with: BUILD_FAILED",
		"CPP-VCPKG-PACKAGE-NOT-FOUND": `Could not find a package configuration file provided by "fmt" with any`,
		"CPP-VCPKG-HOST-TOOL":         "Could not find Python3 (missing: Python3_EXECUTABLE)",
	}
	for id, message := range samples {
		for _, language := range []string{"en", "es"} {
			got := explainer.Explain(domain.Diagnostic{Message: message}, language)
			if got == nil || got.ExplanationID != id || unfilled.MatchString(got.Title+got.Body+got.FixHint) {
				t.Errorf("%s/%s: %+v", id, language, got)
			}
		}
	}
}
