package clangformat

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp/cpptest"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

type settingsStub struct{ paths map[string]string }

func (s settingsStub) Load() (domain.Settings, error) {
	settings := domain.DefaultSettings()
	settings.ToolPaths = s.paths
	return settings, nil
}
func (settingsStub) Save(domain.Settings) error { return nil }

func realTool(t *testing.T) *Tool {
	t.Helper()
	bin := cpptest.LLVMBin(t)
	return New(cpp.NewLocator(cpp.Options{
		Settings: settingsStub{paths: map[string]string{cpp.ToolClangFormat: filepath.Join(bin, cpptest.Exe("clang-format"))}},
		AppDir:   t.TempDir(), BaseEnvironment: []string{"PATH="},
	}))
}

const messy = "#include <iostream>\nint main(){\nif(true){\nstd::cout<<\"hi\";\n}\nreturn 0;}\n"

func TestBadlyIndentedCodeGetsFourSpaces(t *testing.T) {
	tool := realTool(t)
	out, err := tool.Format(filepath.Join(t.TempDir(), "main.cpp"), messy)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\n    if (true) {\n        std::cout << \"hi\";\n    }\n") {
		t.Errorf("not formatted with 4 spaces:\n%s", out)
	}
}

func TestClangFormatFileOfTheFolderWins(t *testing.T) {
	tool := realTool(t)
	folder := t.TempDir()
	if err := os.WriteFile(filepath.Join(folder, ".clang-format"), []byte("BasedOnStyle: LLVM\nIndentWidth: 2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := tool.Format(filepath.Join(folder, "main.cpp"), messy)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\n  if (true) {\n    std::cout << \"hi\";\n  }\n") {
		t.Errorf("not formatted with 2 spaces:\n%s", out)
	}
}

func TestSyntaxErrorsDoNotFail(t *testing.T) {
	out, err := realTool(t).Format(filepath.Join(t.TempDir(), "main.cpp"), "int main( {\nreturn 0\n")
	if err != nil || out == "" {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestUntitledFileHasNoFolder(t *testing.T) {
	out, err := realTool(t).Format(filepath.Join(t.TempDir(), "gone", "untitled.cpp"), messy)
	if err != nil || !strings.Contains(out, "    if (true)") {
		t.Errorf("out = %q, err = %v", out, err)
	}
}

func TestMissingClangFormatIsMissingTool(t *testing.T) {
	tool := New(cpp.NewLocator(cpp.Options{
		Settings: settingsStub{}, AppDir: t.TempDir(), BaseEnvironment: []string{"PATH="},
		Probe: func(context.Context, string) (string, error) { return "", errors.New("none") },
	}))
	_, err := tool.Format("main.cpp", messy)
	if !errors.Is(err, app.ErrToolNotFound) || err.Error() != app.MissingTool("clang-format").Error() {
		t.Errorf("err = %v", err)
	}
}
