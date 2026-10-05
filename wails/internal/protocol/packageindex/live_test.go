package packageindex

import (
	"context"
	"os"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// TestLive asks the real indexes. It only runs with VIZCACHA_LIVE_INDEX=1 (it needs the
// network, and PyPI's search page may answer with a challenge: then only the exact name shows).
func TestLive(t *testing.T) {
	if os.Getenv("VIZCACHA_LIVE_INDEX") != "1" {
		t.Skip("set VIZCACHA_LIVE_INDEX=1 to search the live indexes")
	}
	index := New()
	searches := map[string]func(context.Context, string) ([]domain.PackageInfo, error){
		"numpy": index.PyPI, "nump": index.PyPI, "rand": index.Crates, "github.com/google/uuid": index.Go, "uuid": index.Go,
	}
	for query, search := range searches {
		found, err := search(context.Background(), query)
		if err != nil || len(found) == 0 {
			t.Errorf("%q: %v, %v", query, found, err)
			continue
		}
		t.Logf("%q: %d results, first %+v", query, len(found), found[0])
	}
}
