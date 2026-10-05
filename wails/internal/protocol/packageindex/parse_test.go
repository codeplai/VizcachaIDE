package packageindex

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestParseCrates(t *testing.T) {
	found, err := parseCrates(fixture(t, "crates_rand.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(found) == 0 || len(found) > MaxResults {
		t.Fatalf("%d results", len(found))
	}
	first := found[0]
	if first.Name != "rand" || first.Version == "" || first.Description == "" {
		t.Errorf("first = %+v", first)
	}
	if first.URL != "https://crates.io/crates/rand" {
		t.Errorf("URL = %q", first.URL)
	}
}

func TestParseCratesGarbage(t *testing.T) {
	if _, err := parseCrates([]byte("<html>not json</html>")); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}

func TestParsePyPIPage(t *testing.T) {
	found, err := parsePyPIPage(fixture(t, "pypi_handwritten.html"))
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 3 {
		t.Fatalf("%d results: %+v", len(found), found)
	}
	if found[0].Name != "numpy" || found[0].Version != "2.1.0" || found[0].URL != "https://pypi.org/project/numpy/" {
		t.Errorf("first = %+v", found[0])
	}
	if found[1].Description != "Simple financial functions for NumPy" {
		t.Errorf("white space not collapsed: %q", found[1].Description)
	}
	if found[2].Description != "" {
		t.Errorf("third = %+v", found[2])
	}
}

func TestParsePyPIPageWithoutResults(t *testing.T) {
	found, err := parsePyPIPage(fixture(t, "pypi_nothing.html"))
	if err != nil || len(found) != 0 {
		t.Errorf("found %v, err %v", found, err)
	}
}

func TestParsePyPIChallengeIsUnavailable(t *testing.T) {
	if _, err := parsePyPIPage(fixture(t, "pypi_challenge.html")); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}

func TestParseGoPage(t *testing.T) {
	found, err := parseGoPage(fixture(t, "go_uuid.html"), "https://pkg.go.dev/")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != MaxResults {
		t.Fatalf("%d results, want the first %d", len(found), MaxResults)
	}
	first := found[0]
	if first.Name != "github.com/google/uuid" || first.Version != "v1.6.0" {
		t.Errorf("first = %+v", first)
	}
	if first.Description != "Package uuid generates and inspects UUIDs." {
		t.Errorf("description = %q", first.Description)
	}
	if first.URL != "https://pkg.go.dev/github.com/google/uuid" {
		t.Errorf("URL = %q", first.URL)
	}
}

func TestParseGoPageWithoutResults(t *testing.T) {
	found, err := parseGoPage(fixture(t, "go_nothing.html"), "https://pkg.go.dev/")
	if err != nil || len(found) != 0 {
		t.Errorf("found %v, err %v", found, err)
	}
}

func TestParseGoPageUnexpected(t *testing.T) {
	_, err := parseGoPage([]byte("<html><body><h1>Maintenance</h1></body></html>"), "https://pkg.go.dev/")
	if !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}

func TestEscapeModule(t *testing.T) {
	if got := escapeModule("github.com/Azure/azure-sdk"); got != "github.com/!azure/azure-sdk" {
		t.Errorf("escapeModule = %q", got)
	}
}
