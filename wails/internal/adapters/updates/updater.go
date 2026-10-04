package updates

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// DefaultAPIURL lists the releases of the repository (newest first).
const DefaultAPIURL = "https://api.github.com/repos/codeplai/VizcachaIDE/releases?per_page=30"

const progressEvery = 200 * time.Millisecond

// Options configures an Updater. Current, CacheDir and Installation are required.
type Options struct {
	Current      string // the running version ("2.2.0")
	APIURL       string // DefaultAPIURL when empty (tests point it to a fake server)
	Client       *http.Client
	CacheDir     string // where the downloaded file is kept
	Installation Installation
	Sink         app.UpdateSink
	LastCheck    string             // Settings.LastUpdateCheck, shown until the first check
	Launch       func(string) error // starts the downloaded installer, detached
	Quit         func()             // closes the IDE after the installer started
	Reveal       func(string) error // shows the downloaded file in its folder
}

// Updater implements app.Updater. It is safe for concurrent use.
type Updater struct {
	options  Options
	mu       sync.Mutex
	state    domain.UpdateState
	offer    *offer
	file     string
	lastEmit time.Time
}

var _ app.Updater = (*Updater)(nil)

// New creates the updater; nothing touches the network until Check.
func New(options Options) *Updater {
	if options.APIURL == "" {
		options.APIURL = DefaultAPIURL
	}
	if options.Client == nil {
		options.Client = &http.Client{Timeout: 30 * time.Minute}
	}
	state := domain.UpdateState{
		Status: domain.UpdateIdle, Current: options.Current, CheckedAt: options.LastCheck,
		Installs: options.Installation.SelfInstalls(),
	}
	return &Updater{options: options, state: state}
}

// State implements app.Updater.
func (u *Updater) State() domain.UpdateState {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.state
}

// Check implements app.Updater.
func (u *Updater) Check(ctx context.Context) (domain.UpdateState, error) {
	u.change(func(s *domain.UpdateState) { s.Status, s.Error = domain.UpdateChecking, "" })
	releases, err := fetchReleases(ctx, u.options.Client, u.options.APIURL)
	if err != nil {
		return u.fail(err), err
	}
	found := newestOffer(releases, u.options.Current, u.options.Installation)
	u.mu.Lock()
	u.offer = found
	u.mu.Unlock()
	return u.change(func(s *domain.UpdateState) {
		s.CheckedAt = time.Now().UTC().Format(time.RFC3339)
		if found == nil {
			s.Status, s.Latest = domain.UpdateUpToDate, nil
			return
		}
		release := found.release
		s.Status, s.Latest, s.TotalBytes, s.DownloadedBytes = domain.UpdateAvailable, &release, release.AssetSize, 0
	}), nil
}

// Download implements app.Updater.
func (u *Updater) Download(ctx context.Context) error {
	u.mu.Lock()
	found, busy := u.offer, u.state.Status == domain.UpdateDownloading
	u.mu.Unlock()
	if found == nil {
		return app.ErrNoUpdate
	}
	if busy {
		return nil
	}
	u.change(func(s *domain.UpdateState) { s.Status, s.DownloadedBytes, s.Error = domain.UpdateDownloading, 0, "" })
	path, err := u.fetchVerified(ctx, found)
	if err != nil {
		u.fail(err)
		return err
	}
	u.mu.Lock()
	u.file = path
	u.mu.Unlock()
	removeOthers(u.options.CacheDir, path)
	u.change(func(s *domain.UpdateState) { s.Status, s.DownloadedBytes = domain.UpdateReady, s.TotalBytes })
	return nil
}

// Install implements app.Updater.
func (u *Updater) Install() error {
	u.mu.Lock()
	ready, path := u.state.Status == domain.UpdateReady, u.file
	u.mu.Unlock()
	if !ready || path == "" {
		return app.ErrNoUpdate
	}
	if !u.options.Installation.SelfInstalls() {
		return u.options.Reveal(path)
	}
	if err := u.options.Launch(path); err != nil {
		u.fail(err)
		return err
	}
	u.options.Quit()
	return nil
}

func (u *Updater) progress(done, total int64) {
	u.mu.Lock()
	due := time.Since(u.lastEmit) >= progressEvery || (total > 0 && done >= total)
	u.mu.Unlock()
	if !due {
		return
	}
	u.change(func(s *domain.UpdateState) {
		s.DownloadedBytes = done
		if total > 0 {
			s.TotalBytes = total
		}
	})
}

// change applies edit to the state, reports it and returns the new state.
func (u *Updater) change(edit func(*domain.UpdateState)) domain.UpdateState {
	u.mu.Lock()
	edit(&u.state)
	state := u.state
	u.lastEmit = time.Now()
	u.mu.Unlock()
	if u.options.Sink != nil {
		u.options.Sink.UpdateState(state)
	}
	return state
}

func (u *Updater) fail(err error) domain.UpdateState {
	return u.change(func(s *domain.UpdateState) { s.Status, s.Error = domain.UpdateFailed, err.Error() })
}
