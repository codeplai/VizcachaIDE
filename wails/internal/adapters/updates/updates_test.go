package updates

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var installed = Installation{OS: "windows", Arch: "amd64", Variant: "full-python", Installed: true}

const setupName = "VizcachaIDE-2.3.0-windows-amd64-full-python-setup.exe"

type recordingSink struct {
	mu     sync.Mutex
	states []domain.UpdateState
}

func (r *recordingSink) UpdateState(state domain.UpdateState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states = append(r.states, state)
}

// fakeGitHub serves a release list, one installer and its checksum file. sum overrides the
// published checksum (to test a corrupt download); downloads counts installer requests.
func fakeGitHub(t *testing.T, payload []byte, sum string, downloads *atomic.Int32) *httptest.Server {
	t.Helper()
	if sum == "" {
		digest := sha256.Sum256(payload)
		sum = hex.EncodeToString(digest[:])
	}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	asset := func(name string, size int) githubAsset {
		return githubAsset{Name: name, Size: int64(size), URL: server.URL + "/files/" + name}
	}
	releases := []githubRelease{
		{Tag: "v1.9.0", Assets: []githubAsset{asset(setupName, len(payload))}}, // classic edition: ignored
		{Tag: "wails-v2.4.0-rc1", Prerelease: true, Assets: []githubAsset{asset(setupName, 1)}},
		{Tag: "wails-v2.5.0", Draft: true},
		{Tag: "wails-v2.3.0", Body: "Notas", PageURL: "https://example.org/2.3.0",
			Assets: []githubAsset{asset(setupName, len(payload)), asset("SHA256SUMS-windows-amd64.txt", 80)}},
		{Tag: "wails-v2.1.0", Assets: []githubAsset{asset(setupName, 1)}},
	}
	mux.HandleFunc("/releases", func(w http.ResponseWriter, _ *http.Request) { _ = json.NewEncoder(w).Encode(releases) })
	mux.HandleFunc("/files/"+setupName, func(w http.ResponseWriter, _ *http.Request) {
		downloads.Add(1)
		_, _ = w.Write(payload)
	})
	mux.HandleFunc("/files/SHA256SUMS-windows-amd64.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(sum + "  " + setupName + "\nffff  other.zip\n"))
	})
	t.Cleanup(server.Close)
	return server
}

func newTestUpdater(t *testing.T, server *httptest.Server, sink *recordingSink, launched *string) *Updater {
	t.Helper()
	return New(Options{
		Current: "2.2.0", APIURL: server.URL + "/releases", CacheDir: t.TempDir(), Installation: installed, Sink: sink,
		Launch: func(path string) error { *launched = path; return nil },
		Quit:   func() {}, Reveal: func(string) error { return nil },
	})
}

func TestCheckDownloadAndInstallTheNewestRelease(t *testing.T) {
	var downloads atomic.Int32
	payload := []byte("installer bytes")
	server := fakeGitHub(t, payload, "", &downloads)
	sink, launched := &recordingSink{}, ""
	updater := newTestUpdater(t, server, sink, &launched)

	state, err := updater.Check(context.Background())
	if err != nil || state.Status != domain.UpdateAvailable || state.Latest.Version != "2.3.0" || state.Latest.Asset != setupName {
		t.Fatalf("check: %+v, %v", state, err)
	}
	if err := updater.Download(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := updater.State(); got.Status != domain.UpdateReady || !got.Installs || got.DownloadedBytes != int64(len(payload)) {
		t.Fatalf("after download: %+v", got)
	}
	if err := updater.Install(); err != nil || filepath.Base(launched) != setupName {
		t.Fatalf("install launched %q, %v", launched, err)
	}
	if len(sink.states) < 4 || sink.states[0].Status != domain.UpdateChecking {
		t.Errorf("states reported: %d", len(sink.states))
	}
}

func TestACorruptDownloadFailsAndIsDeleted(t *testing.T) {
	var downloads atomic.Int32
	server := fakeGitHub(t, []byte("installer bytes"), "0000", &downloads)
	updater := newTestUpdater(t, server, &recordingSink{}, new(string))
	_, _ = updater.Check(context.Background())

	err := updater.Download(context.Background())
	if !errors.Is(err, errChecksum) || updater.State().Status != domain.UpdateFailed {
		t.Fatalf("download: %v, %+v", err, updater.State())
	}
	if _, statErr := os.Stat(filepath.Join(updater.options.CacheDir, setupName)); statErr == nil {
		t.Error("the corrupt file was kept")
	}
	if !errors.Is(updater.Install(), app.ErrNoUpdate) {
		t.Error("a failed download must not be installable")
	}
}

func TestAVerifiedFileFromAnEarlierSessionIsNotDownloadedAgain(t *testing.T) {
	var downloads atomic.Int32
	server := fakeGitHub(t, []byte("installer bytes"), "", &downloads)
	cache := t.TempDir()
	for range 2 {
		updater := New(Options{Current: "2.2.0", APIURL: server.URL + "/releases", CacheDir: cache, Installation: installed})
		_, _ = updater.Check(context.Background())
		if err := updater.Download(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if downloads.Load() != 1 {
		t.Errorf("installer downloaded %d times, want 1", downloads.Load())
	}
}

func TestUpToDateWhenNothingNewerHasAFileForThisInstallation(t *testing.T) {
	var downloads atomic.Int32
	server := fakeGitHub(t, []byte("x"), "", &downloads)
	updater := New(Options{Current: "2.3.0", APIURL: server.URL + "/releases", CacheDir: t.TempDir(), Installation: installed})
	if state, _ := updater.Check(context.Background()); state.Status != domain.UpdateUpToDate || state.CheckedAt == "" {
		t.Fatalf("state = %+v", state)
	}
	portable := installed
	portable.Variant = "lite" // the release has no lite file
	updater = New(Options{Current: "2.2.0", APIURL: server.URL + "/releases", CacheDir: t.TempDir(), Installation: portable})
	if state, _ := updater.Check(context.Background()); state.Status != domain.UpdateUpToDate {
		t.Fatalf("lite: %+v", state)
	}
	if !errors.Is(updater.Download(context.Background()), app.ErrNoUpdate) {
		t.Error("Download without an offer must answer ErrNoUpdate")
	}
}
