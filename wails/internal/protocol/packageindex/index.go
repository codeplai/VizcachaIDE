// Package packageindex searches the package indexes of the languages by name: crates.io (an
// official JSON API), PyPI (its search page, read as HTML, plus the exact-name JSON) and
// pkg.go.dev (its search page, plus the module proxy for the exact name). The adapters of each
// language call one method and keep nothing else.
//
// Every failure to read an index (network down, timeout, blocked, a page that no longer looks
// like it used to) wraps app.ErrPackageIndexUnavailable; "found nothing" is an empty list.
package packageindex

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const (
	// MaxResults is the most packages one search returns.
	MaxResults = 10
	// Timeout is how long one search waits for each answer.
	Timeout = 8 * time.Second
	// UserAgent identifies the IDE; crates.io requires a descriptive one.
	UserAgent = "VizcachaIDE/2 (https://github.com/codeplai/VizcachaIDE)"

	maxBody = 4 << 20
)

// Endpoints are the base addresses of the indexes. Tests point them at a fake server.
type Endpoints struct {
	Crates     string // https://crates.io/api/v1/crates
	PyPISearch string // https://pypi.org/search/
	PyPIJSON   string // https://pypi.org/pypi/
	GoSearch   string // https://pkg.go.dev/search
	GoProxy    string // https://proxy.golang.org/
	GoPackage  string // https://pkg.go.dev/ (the page of a package)
}

// DefaultEndpoints are the real indexes.
func DefaultEndpoints() Endpoints {
	return Endpoints{
		Crates:     "https://crates.io/api/v1/crates",
		PyPISearch: "https://pypi.org/search/",
		PyPIJSON:   "https://pypi.org/pypi/",
		GoSearch:   "https://pkg.go.dev/search",
		GoProxy:    "https://proxy.golang.org/",
		GoPackage:  "https://pkg.go.dev/",
	}
}

// Index searches the indexes. The zero value is not usable: create it with New or NewWith.
type Index struct {
	client    *http.Client
	endpoints Endpoints
}

// New creates an Index that reads the real indexes.
func New() *Index { return NewWith(&http.Client{}, DefaultEndpoints()) }

// NewWith creates an Index with its own client and addresses (tests).
func NewWith(client *http.Client, endpoints Endpoints) *Index {
	return &Index{client: client, endpoints: endpoints}
}

// get reads one address. It returns the body and the HTTP status; only a request that could not
// be made or read at all is an error (wrapping app.ErrPackageIndexUnavailable).
func (i *Index) get(ctx context.Context, address, accept string) ([]byte, int, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, http.NoBody)
	if err != nil {
		return nil, 0, unavailable(err)
	}
	request.Header.Set("User-Agent", UserAgent)
	request.Header.Set("Accept", accept)
	response, err := i.client.Do(request)
	if err != nil {
		return nil, 0, unavailable(err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody))
	if err != nil {
		return nil, 0, unavailable(err)
	}
	return body, response.StatusCode, nil
}

// getOK is get for answers that must be a 200.
func (i *Index) getOK(ctx context.Context, address, accept string) ([]byte, error) {
	body, status, err := i.get(ctx, address, accept)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, unavailable(fmt.Errorf("%s answered %d", address, status))
	}
	return body, nil
}

func unavailable(cause error) error {
	return fmt.Errorf("%w: %w", app.ErrPackageIndexUnavailable, cause)
}

// clean trims a query and tells whether there is anything to search.
func clean(query string) (string, bool) {
	query = strings.TrimSpace(query)
	return query, query != ""
}

// oneLine collapses all the white space of a text into single spaces.
func oneLine(text string) string { return strings.Join(strings.Fields(text), " ") }
