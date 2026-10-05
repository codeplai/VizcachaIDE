package packageindex

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func TestPyPIExactFirstThenListed(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/pypi/numpy/json": serve(fixture(t, "pypi_numpy.json")),
		"/search/":         serve(fixture(t, "pypi_handwritten.html")),
	})
	found, err := index.PyPI(context.Background(), "numpy")
	if err != nil || len(found) != 3 {
		t.Fatalf("found %v, err %v", found, err)
	}
	if found[0].Name != "numpy" || found[0].Version != "2.5.3" {
		t.Errorf("the exact match must come first with its own version: %+v", found[0])
	}
	if found[1].Name != "numpy-financial" {
		t.Errorf("the duplicate of numpy was not dropped: %+v", found)
	}
}

func TestPyPIBlockedPageStillGivesExactName(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/pypi/numpy/json": serve(fixture(t, "pypi_numpy.json")),
		"/search/":         serve(fixture(t, "pypi_challenge.html")),
	})
	found, err := index.PyPI(context.Background(), "numpy")
	if err != nil || len(found) != 1 || found[0].Name != "numpy" {
		t.Errorf("found %v, err %v", found, err)
	}
}

func TestPyPIBlockedPageWithoutExactNameIsUnavailable(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/pypi/nump/json": status(http.StatusNotFound),
		"/search/":        serve(fixture(t, "pypi_challenge.html")),
	})
	if _, err := index.PyPI(context.Background(), "nump"); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}

func TestPyPINothingFound(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/pypi/zzqqxx/json": status(http.StatusNotFound),
		"/search/":          serve(fixture(t, "pypi_nothing.html")),
	})
	found, err := index.PyPI(context.Background(), "zzqqxx")
	if err != nil || len(found) != 0 {
		t.Errorf("found %v, err %v", found, err)
	}
}

func TestGoExactFirstWithSynopsis(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/proxy/github.com/google/uuid/@latest": serve(fixture(t, "goproxy_uuid.json")),
		"/gosearch":                             serve(fixture(t, "go_uuid.html")),
	})
	found, err := index.Go(context.Background(), "github.com/google/uuid")
	if err != nil || len(found) == 0 || len(found) > MaxResults {
		t.Fatalf("found %v, err %v", found, err)
	}
	if found[0].Name != "github.com/google/uuid" || found[0].Version != "v1.6.0" || found[0].Description == "" {
		t.Errorf("first = %+v", found[0])
	}
	for _, other := range found[1:] {
		if other.Name == found[0].Name {
			t.Error("the exact module is listed twice")
		}
	}
}

func TestGoBlockedPageStillGivesExactModule(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/proxy/github.com/google/uuid/@latest": serve(fixture(t, "goproxy_uuid.json")),
		"/gosearch":                             status(http.StatusForbidden),
	})
	found, err := index.Go(context.Background(), "github.com/google/uuid")
	if err != nil || len(found) != 1 || found[0].Version != "v1.6.0" {
		t.Errorf("found %v, err %v", found, err)
	}
}

func TestGoGarbageIsUnavailable(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/gosearch": serve([]byte("<html><body>oops</body></html>")),
	})
	if _, err := index.Go(context.Background(), "uuid"); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}
