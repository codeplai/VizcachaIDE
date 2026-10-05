package packages

import (
	"context"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/packageindex"
)

// Search is the adapter of app.PackageSearch that looks packages up in crates.io.
type Search struct{ index *packageindex.Index }

var _ app.PackageSearch = (*Search)(nil)

// NewSearch creates the search over a package index client.
func NewSearch(index *packageindex.Index) *Search { return &Search{index: index} }

// Search implements app.PackageSearch.
func (s *Search) Search(ctx context.Context, query string) ([]domain.PackageInfo, error) {
	return s.index.Crates(ctx, query)
}
