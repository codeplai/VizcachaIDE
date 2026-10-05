package packageindex

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestRecordFixtures downloads the answers the parser tests read from testdata. It only runs
// with VIZCACHA_RECORD_INDEX=1 (it needs the network); run it again when an index changes its
// pages, then run the parser tests to see what broke. pypi_search.html is written only when
// PyPI lets the request through: it sits behind a challenge for robots, which is recorded as
// pypi_challenge.html, so the fixture used by the tests is written by hand (pypi_handwritten.html).
func TestRecordFixtures(t *testing.T) {
	if os.Getenv("VIZCACHA_RECORD_INDEX") != "1" {
		t.Skip("set VIZCACHA_RECORD_INDEX=1 to record the fixtures from the live indexes")
	}
	index := New()
	endpoints := index.endpoints
	record := map[string]struct{ address, accept string }{
		"crates_rand.json":    {endpoints.Crates + "?per_page=10&q=rand", "application/json"},
		"pypi_numpy.json":     {endpoints.PyPIJSON + "numpy/json", "application/json"},
		"goproxy_uuid.json":   {endpoints.GoProxy + "github.com/google/uuid/@latest", "application/json"},
		"go_uuid.html":        {endpoints.GoSearch + "?m=package&q=uuid", "text/html"},
		"go_nothing.html":     {endpoints.GoSearch + "?m=package&q=zzqqxxnotapackage123", "text/html"},
		"pypi_challenge.html": {endpoints.PyPISearch + "?q=numpy", "text/html"},
	}
	for name, source := range record {
		body, status, err := index.get(context.Background(), source.address, source.accept)
		if err != nil || status != 200 {
			t.Errorf("%s: status %d, %v", name, status, err)
			continue
		}
		if name == "pypi_numpy.json" {
			body = trimPyPIJSON(t, body)
		}
		if name == "pypi_challenge.html" && containsSnippet(body) {
			name = "pypi_search.html"
		}
		if err := os.WriteFile(filepath.Join("testdata", name), body, 0o600); err != nil {
			t.Error(err)
		}
	}
}

// trimPyPIJSON keeps the three fields of "info" the parser reads: the whole answer is 3 MB.
func trimPyPIJSON(t *testing.T, body []byte) []byte {
	t.Helper()
	var full struct {
		Info struct {
			Name    string `json:"name"`
			Version string `json:"version"`
			Summary string `json:"summary"`
		} `json:"info"`
	}
	if err := json.Unmarshal(body, &full); err != nil {
		t.Fatal(err)
	}
	trimmed, err := json.MarshalIndent(full, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return trimmed
}

func containsSnippet(body []byte) bool {
	root, err := parseDocument(body)
	return err == nil && len(withClass(root, "package-snippet")) > 0
}
