package bridge

import (
	"context"
	"errors"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// fakeSearch finds one package named like the query; "down" cannot be searched.
type fakeSearch struct{}

func (fakeSearch) Search(_ context.Context, query string) ([]domain.PackageInfo, error) {
	switch query {
	case "down":
		return nil, app.ErrPackageIndexUnavailable
	case "":
		return nil, nil
	}
	return []domain.PackageInfo{{Name: query, Version: "1.0"}}, nil
}

func TestPackagesSearch(t *testing.T) {
	service := NewPackagesService(newTestLanguages(t).registry)
	found, err := service.Search("go", "uuid")
	if err != nil || len(found) != 1 || found[0].Name != "uuid" {
		t.Errorf("found %v, err %v", found, err)
	}
	if found, err := service.Search("go", ""); err != nil || found == nil || len(found) != 0 {
		t.Errorf("an empty query must give an empty, non-nil list: %v, %v", found, err)
	}
	if _, err := service.Search("go", "down"); !errors.Is(err, app.ErrPackageIndexUnavailable) {
		t.Errorf("an index that is down: %v", err)
	}
	if _, err := service.Search("python", "x"); !errors.Is(err, app.ErrUnsupported) {
		t.Errorf("a language without an index: %v", err)
	}
	if _, err := service.Search("cobol", "x"); !errors.Is(err, app.ErrUnknownCodeLanguage) {
		t.Errorf("an unknown language: %v", err)
	}
}
