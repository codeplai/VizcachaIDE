package gopls

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/lsp"
)

const uuidSource = "package main\n\nimport (\n\t\"fmt\"\n\n\t\"github.com/google/uuid\"\n)\n\nfunc main() {\n\tfmt.Println(len(uuid.NewString()))\n}\n"

// "go get" changes go.mod and go.sum outside the editor: gopls must hear about it, or it keeps
// saying the package cannot be imported (found in the multi-file experiment after M3). Needs the
// network the first time (the module proxy).
func TestRealGoplsSeesAPackageAddedWithGoGet(t *testing.T) {
	if _, _, err := NewFlavor(Config{}).Command(nil); err != nil {
		t.Skip("gopls is not installed")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tienda\n\ngo 1.21\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "main.go")
	if err := os.WriteFile(file, []byte(uuidSource), 0o600); err != nil {
		t.Fatal(err)
	}
	sink := newRecordingSink()
	server := New(sink, Config{}, lsp.Options{})
	t.Cleanup(func() { _ = server.Shutdown(context.Background()) })
	if err := server.OpenDocument(context.Background(), file, uuidSource); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, loadTimeout, func() bool { return cannotImport(sink.diagnosticsOf(file)) })

	get := exec.Command("go", "get", "github.com/google/uuid")
	get.Dir = dir
	if output, err := get.CombinedOutput(); err != nil {
		t.Skipf("go get failed (no network?): %v\n%s", err, output)
	}
	// The editor keeps asking (hover, inlay hints): those queries carry the news of go.mod.
	at := domain.SourceLocation{File: file, Line: 10, Column: 20}
	waitUntil(t, queryTimeout, func() bool {
		_, _ = server.Hover(context.Background(), at)
		return !cannotImport(sink.diagnosticsOf(file))
	})
}

func cannotImport(diagnostics []domain.Diagnostic) bool {
	for _, d := range diagnostics {
		if strings.Contains(d.Message, "could not import") || strings.Contains(d.Message, "no required module") {
			return true
		}
	}
	return false
}
