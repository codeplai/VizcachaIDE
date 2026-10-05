package vcpkg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakePorts writes ports/<name>/vcpkg.json for each entry (the JSON body without braces).
func fakePorts(t *testing.T, root string, ports map[string]string) {
	t.Helper()
	for name, body := range ports {
		mustWrite(t, filepath.Join(root, "ports", name, "vcpkg.json"), "{"+body+"}")
	}
}

func newSearch(t *testing.T, goos string) (*Search, string, string) {
	t.Helper()
	root, cache := fakeRoot(t, goos), t.TempDir()
	locator := NewLocator(LocatorOptions{AppDir: t.TempDir(), BaseEnvironment: []string{"VCPKG_ROOT=" + root}, CacheDir: cache, GOOS: goos, GOARCH: "amd64"})
	return NewSearch(locator), root, cache
}

func names(found []domain.PackageInfo) []string {
	list := []string{}
	for _, item := range found {
		list = append(list, item.Name)
	}
	return list
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index] != b[index] {
			return false
		}
	}
	return true
}

func TestSearchRanksLikePyPI(t *testing.T) {
	search, root, _ := newSearch(t, "linux")
	fakePorts(t, root, map[string]string{
		"fmt":         `"name":"fmt","version":"12.2.0","description":"{fmt} formatting library"`,
		"fmt-extra":   `"name":"fmt-extra","version-semver":"1.0.0","description":"extras"`,
		"spdlog":      `"name":"spdlog","version":"1.15.0","description":["Fast logger","uses fmt inside"]`,
		"libfmt-bind": `"name":"libfmt-bind","version-date":"2025-01-01","description":"bindings"`,
		"zlib":        `"name":"zlib","version-string":"1.3","description":"compression"`,
		"imgui":       `"name":"imgui","version":"1.91","description":"GUI"`,
	})
	found, err := search.Search(context.Background(), "  FMT ")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"fmt", "fmt-extra", "libfmt-bind", "spdlog"}; !equal(names(found), want) {
		t.Errorf("order = %v, want %v", names(found), want)
	}
	if found[0].Version != "12.2.0" || found[0].Description == "" {
		t.Errorf("first = %+v", found[0])
	}
	if found[2].Version != "2025-01-01" || found[3].Description != "Fast logger uses fmt inside" {
		t.Errorf("other spellings: %+v %+v", found[2], found[3])
	}
	if got, _ := search.Search(context.Background(), "zlib"); len(got) != 1 || got[0].Version != "1.3" {
		t.Errorf("version-string: %+v", got)
	}
}

func TestSearchLimitsTheResultsAndAnswersEmptyForAnEmptyQuery(t *testing.T) {
	search, root, _ := newSearch(t, "linux")
	many := map[string]string{}
	for _, name := range []string{"a1", "a2", "a3", "a4", "a5", "a6", "a7", "a8", "a9", "a10", "a11", "a12"} {
		many[name] = `"name":"` + name + `","version":"1"`
	}
	fakePorts(t, root, many)
	if found, _ := search.Search(context.Background(), "a"); len(found) != 10 {
		t.Errorf("%d results, want 10", len(found))
	}
	found, err := search.Search(context.Background(), "   ")
	if err != nil || found == nil || len(found) != 0 {
		t.Errorf("empty query: %v, %v", found, err)
	}
}

func TestSearchHidesPortsThatClearlyDoNotSupportThisSystem(t *testing.T) {
	search, root, _ := newSearch(t, "windows")
	fakePorts(t, root, map[string]string{
		"onlylinux": `"name":"onlylinux","version":"1","supports":"linux | osx"`,
		"notwin":    `"name":"notwin","version":"1","supports":"!windows"`,
		"mingwok":   `"name":"mingwok","version":"1","supports":"windows & !uwp & !arm"`,
		"weird":     `"name":"weird","version":"1","supports":"((("`,
		"unknown":   `"name":"unknown","version":"1","supports":"somefuturething"`,
	})
	found, _ := search.Search(context.Background(), "")
	if len(found) != 0 {
		t.Fatal("empty query must not list")
	}
	var all []string
	for _, query := range []string{"onlylinux", "notwin", "mingwok", "weird", "unknown"} {
		got, _ := search.Search(context.Background(), query)
		all = append(all, names(got)...)
	}
	if want := []string{"mingwok", "weird", "unknown"}; !equal(all, want) {
		t.Errorf("visible = %v, want %v", all, want)
	}
}

func TestSearchKeepsTheIndexOnDiskAndReusesIt(t *testing.T) {
	search, root, cache := newSearch(t, "linux")
	fakePorts(t, root, map[string]string{"fmt": `"name":"fmt","version":"1"`})
	if _, err := search.Search(context.Background(), "fmt"); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(cache, "vcpkg", "index-*.json"))
	if len(files) != 1 {
		t.Fatalf("index files = %v", files)
	}
	// A second Search object reads the file, not the ports: delete the port to prove it.
	if err := os.Remove(filepath.Join(root, "ports", "fmt", "vcpkg.json")); err != nil {
		t.Fatal(err)
	}
	again := NewSearch(search.locator)
	found, err := again.Search(context.Background(), "fmt")
	if err != nil || len(found) != 1 {
		t.Fatalf("%v %v", found, err)
	}
}

func TestSearchWithoutVcpkgIsAnUnavailableIndex(t *testing.T) {
	locator := NewLocator(LocatorOptions{AppDir: t.TempDir(), BaseEnvironment: []string{}, CacheDir: t.TempDir(), GOOS: "linux"})
	_, err := NewSearch(locator).Search(context.Background(), "fmt")
	if !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}

func TestSupportsExpressions(t *testing.T) {
	linux := platform{goos: "linux", goarch: "amd64"}
	cases := map[string]bool{
		"linux":                true,
		"!linux":               false,
		"windows | osx":        false,
		"(linux | osx) & x64":  true,
		"linux & !x64":         false,
		"!(windows & !static)": true,
		"windows &":            true, // unreadable: kept
		"linux linux":          true,
	}
	for expression, want := range cases {
		if got := linux.supports(expression); got != want {
			t.Errorf("%q = %v, want %v", expression, got, want)
		}
	}
}
