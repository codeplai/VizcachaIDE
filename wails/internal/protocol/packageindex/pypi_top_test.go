package packageindex

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

const topListJSON = `{"last_update":"2026-10-01","rows":[{"download_count":9,"project":"oldest-supported-numpy"},{"download_count":8,"project":"numpy"},{"download_count":7,"project":"requests"},{"download_count":6,"project":"numpydoc"},{"download_count":5,"project":"json_numpy"}]}`

func TestNamesThatStartWithTheQueryComeFirst(t *testing.T) {
	names := []string{"oldest-supported-numpy", "numpy", "requests", "numpydoc", "json_numpy"}
	want := []string{"numpy", "numpydoc", "oldest-supported-numpy", "json_numpy"}
	if got := matchNames(names, "NumPy"); !reflect.DeepEqual(got, want) {
		t.Errorf("matchNames = %v, want %v", got, want)
	}
	if got := matchNames(names, "json-numpy"); !reflect.DeepEqual(got, []string{"json_numpy"}) {
		t.Errorf("- and _ must match: %v", got)
	}
}

func TestPartialPythonNamesUseThePopularList(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/top.json":        serve([]byte(topListJSON)),
		"/pypi/numpy/json": serve([]byte(`{"info":{"name":"numpy","version":"2.5.3","summary":"Arrays"}}`)),
		"/search/":         status(http.StatusForbidden), // the JavaScript challenge, today
	})
	found, err := index.PyPI(context.Background(), "nump")
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 4 || found[0].Name != "numpy" || found[0].Version != "2.5.3" || found[0].Description != "Arrays" {
		t.Fatalf("found = %+v", found)
	}
	if found[1].Name != "numpydoc" || found[1].URL != "https://pypi.org/project/numpydoc/" {
		t.Errorf("a project whose details failed keeps its name: %+v", found[1])
	}
}

func TestThePopularListIsCachedOnDiskAndUsedWhenTheNetworkFails(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){"/top.json": serve([]byte(topListJSON))})
	index.cacheDir = t.TempDir()
	if _, err := index.topNames(context.Background()); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-30 * 24 * time.Hour)
	if err := os.Chtimes(filepath.Join(index.cacheDir, "top-pypi-packages.json"), stale, stale); err != nil {
		t.Fatal(err)
	}
	offline := NewWith(http.DefaultClient, Endpoints{PyPITop: "http://127.0.0.1:1/top.json"})
	offline.cacheDir = index.cacheDir
	names, err := offline.topNames(context.Background())
	if err != nil || len(names) != 5 {
		t.Errorf("stale cache = %v, %v", names, err)
	}
}
