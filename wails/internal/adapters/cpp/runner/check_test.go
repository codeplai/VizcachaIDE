package runner

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestCheckReportsWarningsAndErrorsAndNothingWhenClean(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		warn := writeSource(t, "int main() {\n    int sin_usar = 3;\n    return 0;\n}\n")
		output, err := r.Check(context.Background(), r.Configure(warn, nil))
		if err != nil || !strings.Contains(output, "sin_usar") || !strings.Contains(output, "warning") {
			t.Fatalf("output = %q, err = %v, want the unused variable warning", output, err)
		}
		broken := writeSource(t, "int main() { return totl; }\n")
		if output, err = r.Check(context.Background(), r.Configure(broken, nil)); err != nil || !strings.Contains(output, "totl") {
			t.Fatalf("output = %q, err = %v, want the error as output", output, err)
		}
		clean := writeSource(t, "int main() { return 0; }\n")
		if output, err = r.Check(context.Background(), r.Configure(clean, nil)); err != nil || output != "" {
			t.Fatalf("output = %q, err = %v, want nothing", output, err)
		}
		if len(sink.started) != 0 || r.IsRunning() {
			t.Fatal("Check must not use the run slot")
		}
	})
}

func TestCheckOfAFolderThatIsGoneIsEmpty(t *testing.T) {
	config := domain.RunConfiguration{WorkingDir: filepath.Join(t.TempDir(), "gone"), Mode: domain.RunFile}
	if output, err := bareRunner().Check(context.Background(), config); output != "" || err != nil {
		t.Fatalf("output = %q, err = %v", output, err)
	}
}

func TestToolsReportVersions(t *testing.T) {
	eachCompiler(t, func(t *testing.T, r *Runner, _ *testSink) {
		tools := r.Tools(context.Background())
		if len(tools) != 7 || tools[0].ID != "cxx" || tools[0].Version == "" || strings.Contains(tools[0].Version, " ") {
			t.Fatalf("tools = %+v", tools)
		}
		t.Logf("%+v", tools)
	})
	t.Run("llvm tools", func(t *testing.T) {
		bin := cpptest.LLVMBin(t)
		settings := settings{path: filepath.Join(bin, cpptest.Exe("clang++")), cmakeBin: cpptest.CMakeBin(t)}
		r := New(newSupervisor(), Options{Settings: settings, AppDir: t.TempDir()})
		for _, tool := range r.Tools(context.Background()) {
			if tool.ID == "vcpkg" {
				continue // found by the vcpkg package, not here
			}
			if tool.Source == domain.ToolMissing || tool.Version == "" {
				t.Errorf("%s = %+v, want found with a version", tool.ID, tool)
			}
		}
	})
}

func TestToolsMissing(t *testing.T) {
	r := New(newSupervisor(), Options{AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="}})
	tools := r.Tools(context.Background())
	if len(tools) != 7 {
		t.Fatalf("tools = %+v", tools)
	}
	for _, tool := range tools {
		if tool.Source != domain.ToolMissing || tool.Version != "" {
			t.Errorf("%s = %+v, want missing", tool.ID, tool)
		}
	}
}

func TestVersionNumberOf(t *testing.T) {
	cases := map[string]string{
		"clang version 23.1.2 (https://github.com/llvm/llvm-project 1.2.3)\nTarget: x": "23.1.2",
		"g++ (Ubuntu 13.2.0-23ubuntu4) 13.2.0":                                         "13.2.0",
		"clangd version 23.1.2 (x)":                                                    "23.1.2",
		"nothing":                                                                      "",
	}
	for text, want := range cases {
		if got := versionNumberOf(text); got != want {
			t.Errorf("versionNumberOf(%q) = %q, want %q", text, got, want)
		}
	}
}
