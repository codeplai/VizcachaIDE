package updates

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/mod/semver"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// tagPrefix marks the releases of this edition; the repository also publishes the classic
// PyQt edition (v1.x) and those must never be offered here.
const tagPrefix = "wails-v"

type githubAsset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	URL  string `json:"browser_download_url"`
}

type githubRelease struct {
	Tag         string        `json:"tag_name"`
	Body        string        `json:"body"`
	PageURL     string        `json:"html_url"`
	Draft       bool          `json:"draft"`
	Prerelease  bool          `json:"prerelease"`
	PublishedAt string        `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

// offer is a newer release with the URLs of this installation's file and checksums.
type offer struct {
	release domain.Release
	fileURL string
	sumsURL string
}

// fetchReleases reads the latest releases of the repository (GitHub REST API).
func fetchReleases(ctx context.Context, client *http.Client, apiURL string) ([]githubRelease, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("asking for releases: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("asking for releases: %s", response.Status)
	}
	var releases []githubRelease
	if err := json.NewDecoder(response.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("reading releases: %w", err)
	}
	return releases, nil
}

// newestOffer returns the highest published wails-v* release newer than current that has a
// file for this installation and a checksum file, or nil.
func newestOffer(releases []githubRelease, current string, install Installation) *offer {
	var best *offer
	bestVersion := "v" + strings.TrimPrefix(current, "v")
	for _, release := range releases {
		version, ok := releaseVersion(release)
		if !ok || semver.Compare(version, bestVersion) <= 0 {
			continue
		}
		if candidate := offerFor(release, version, install); candidate != nil {
			best, bestVersion = candidate, version
		}
	}
	return best
}

func releaseVersion(release githubRelease) (string, bool) {
	if release.Draft || release.Prerelease || !strings.HasPrefix(release.Tag, tagPrefix) {
		return "", false
	}
	version := "v" + strings.TrimPrefix(release.Tag, tagPrefix)
	return version, semver.IsValid(version) && semver.Prerelease(version) == ""
}

func offerFor(release githubRelease, version string, install Installation) *offer {
	plain := strings.TrimPrefix(version, "v")
	file, sums := findAsset(release, install.AssetName(plain)), findAsset(release, install.SumsName())
	if file == nil || sums == nil {
		return nil
	}
	return &offer{
		release: domain.Release{
			Version: plain, Notes: release.Body, NotesURL: release.PageURL,
			PublishedAt: release.PublishedAt, Asset: file.Name, AssetSize: file.Size,
		},
		fileURL: file.URL,
		sumsURL: sums.URL,
	}
}

func findAsset(release githubRelease, name string) *githubAsset {
	for index := range release.Assets {
		if release.Assets[index].Name == name {
			return &release.Assets[index]
		}
	}
	return nil
}
