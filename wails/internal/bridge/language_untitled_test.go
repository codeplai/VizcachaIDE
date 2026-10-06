package bridge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUntitledFileIsAnAbsolutePathInItsOwnFolder(t *testing.T) {
	service := NewLanguageService(nil)
	first, err := service.UntitledFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.UntitledFile("main2.go")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(first) || filepath.Base(first) != "main.go" {
		t.Fatalf("first = %q", first)
	}
	if filepath.Dir(first) == filepath.Dir(second) {
		t.Fatalf("two unsaved files share a folder: %q and %q", first, second)
	}
	if info, err := os.Stat(first); err != nil || info.IsDir() {
		t.Fatalf("%q was not created: %v", first, err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(first)); _ = os.RemoveAll(filepath.Dir(second)) })
}

func TestUntitledFileRefusesPaths(t *testing.T) {
	service := NewLanguageService(nil)
	for _, name := range []string{"", "../main.go", `sub\main.go`, "sub/main.go"} {
		if _, err := service.UntitledFile(name); err == nil {
			t.Errorf("%q was accepted", name)
		}
	}
}
