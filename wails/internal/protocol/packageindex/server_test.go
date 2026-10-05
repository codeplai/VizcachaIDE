package packageindex

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

// fakeIndex answers every address of the indexes from testdata, or as the test says.
func fakeIndex(t *testing.T, routes map[string]func(http.ResponseWriter, *http.Request)) (*Index, *atomic.Int32) {
	t.Helper()
	var requests atomic.Int32
	mux := http.NewServeMux()
	for path, handler := range routes {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			if r.Header.Get("User-Agent") != UserAgent {
				t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
			}
			handler(w, r)
		})
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return NewWith(server.Client(), Endpoints{
		Crates:     server.URL + "/crates",
		PyPISearch: server.URL + "/search/",
		PyPIJSON:   server.URL + "/pypi/",
		PyPITop:    server.URL + "/top.json",
		GoSearch:   server.URL + "/gosearch",
		GoProxy:    server.URL + "/proxy/",
		GoPackage:  "https://pkg.go.dev/",
	}), &requests
}

func serve(body []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(body) }
}

func status(code int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "no", code) }
}

func TestEmptyQueryMakesNoRequest(t *testing.T) {
	index, requests := fakeIndex(t, nil)
	for _, query := range []string{"", "   "} {
		if found, err := index.Crates(context.Background(), query); found != nil || err != nil {
			t.Errorf("Crates(%q) = %v, %v", query, found, err)
		}
		if found, err := index.PyPI(context.Background(), query); found != nil || err != nil {
			t.Errorf("PyPI(%q) = %v, %v", query, found, err)
		}
		if found, err := index.Go(context.Background(), query); found != nil || err != nil {
			t.Errorf("Go(%q) = %v, %v", query, found, err)
		}
	}
	if requests.Load() != 0 {
		t.Errorf("%d requests", requests.Load())
	}
}

func TestCratesFromServer(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/crates": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("q") != "rand" || r.URL.Query().Get("per_page") != "10" {
				t.Errorf("query = %q", r.URL.RawQuery)
			}
			_, _ = w.Write(fixture(t, "crates_rand.json"))
		},
	})
	found, err := index.Crates(context.Background(), "  rand ")
	if err != nil || len(found) == 0 || found[0].Name != "rand" {
		t.Errorf("found %v, err %v", found, err)
	}
}

func TestCratesFailuresAreUnavailable(t *testing.T) {
	for name, handler := range map[string]func(http.ResponseWriter, *http.Request){
		"404":     status(http.StatusNotFound),
		"429":     status(http.StatusTooManyRequests),
		"garbage": serve([]byte("<html>")),
	} {
		index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){"/crates": handler})
		if _, err := index.Crates(context.Background(), "rand"); !errors.Is(err, app.ErrPackageIndexUnavailable) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestTimeoutIsUnavailable(t *testing.T) {
	index, _ := fakeIndex(t, map[string]func(http.ResponseWriter, *http.Request){
		"/crates": func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(5 * time.Second):
			}
		},
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := index.Crates(ctx, "rand"); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}

func TestNoServerIsUnavailable(t *testing.T) {
	index, _ := fakeIndex(t, nil)
	index.endpoints.Crates = "http://127.0.0.1:1/crates"
	if _, err := index.Crates(context.Background(), "rand"); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("err = %v", err)
	}
}
