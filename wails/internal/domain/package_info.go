package domain

// PackageInfo is one result of searching a package index (PyPI, crates.io, pkg.go.dev): what
// the student needs to choose which package to install. Name is what the package manager takes
// ("requests", "serde", "github.com/google/uuid"); Version and Description may be empty.
type PackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	URL         string `json:"url"`
}
