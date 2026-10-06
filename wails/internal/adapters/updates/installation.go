// Package updates keeps the IDE up to date from its GitHub releases (tags wails-v*): it finds the
// file of this installation, downloads it, verifies its SHA-256 against the release's
// SHA256SUMS file and installs it (wails/packaging/build_release.py names the files).
package updates

import (
	"fmt"
	"os"
	"path/filepath"
)

const product = "VizcachaIDE"

// Installation is what this copy of the IDE is, which decides the file to download.
type Installation struct {
	OS      string // runtime.GOOS: windows, darwin, linux
	Arch    string // runtime.GOARCH: amd64, arm64
	Variant string // lite, full-go, full-python, full-cpp or full
	// Installed is true for the Windows installer's copy (uninstall.exe next to the executable).
	Installed bool
	// AppImage is true when Linux runs the IDE from an AppImage.
	AppImage bool
}

// DetectInstallation looks at the folder of the executable: the bundled toolchain gives the
// variant (all variants share the executable) and uninstall.exe tells an installed copy.
func DetectInstallation(goos, goarch, exeDir string) Installation {
	return Installation{
		OS: goos, Arch: goarch, Variant: variantOf(exeDir),
		Installed: goos == "windows" && exists(filepath.Join(exeDir, "uninstall.exe")),
		AppImage:  goos == "linux" && os.Getenv("APPIMAGE") != "",
	}
}

// variantOf names the variant by the bundled toolchains: none is lite, one is full-<language>,
// more than one is full (all languages; a 2.2 "full" with Go and Python updates to it).
func variantOf(exeDir string) string {
	var bundled []string
	for _, part := range []string{"go", "python", "cpp", "rust"} {
		if exists(filepath.Join(exeDir, "toolchain", part)) {
			bundled = append(bundled, part)
		}
	}
	switch len(bundled) {
	case 0:
		return "lite"
	case 1:
		return "full-" + bundled[0]
	default:
		return "full"
	}
}

// AssetName is the release file for this installation, as build_release.py names it.
func (i Installation) AssetName(version string) string {
	switch i.OS {
	case "windows":
		base := fmt.Sprintf("%s-%s-windows-%s-%s", product, version, i.Arch, i.Variant)
		if i.Installed {
			return base + "-setup.exe"
		}
		return base + "-portable.zip"
	case "darwin":
		return fmt.Sprintf("%s-%s-macos-%s-%s.dmg", product, version, i.Arch, i.Variant)
	default:
		base := fmt.Sprintf("%s-%s-linux-%s-%s", product, version, i.Arch, i.Variant)
		if i.AppImage {
			return base + ".AppImage"
		}
		return base + ".tar.gz"
	}
}

// SumsName is the checksum file of the release for this system and architecture.
func (i Installation) SumsName() string {
	return fmt.Sprintf("SHA256SUMS-%s-%s.txt", i.OS, i.Arch)
}

// SelfInstalls is true when Install can run the new installer (installed Windows copy).
func (i Installation) SelfInstalls() bool { return i.Installed }

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
