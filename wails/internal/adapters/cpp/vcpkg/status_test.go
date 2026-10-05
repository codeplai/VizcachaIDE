package vcpkg

import (
	"path/filepath"
	"testing"
)

func TestTargetNameReadsTheExecutable(t *testing.T) {
	dir := t.TempDir()
	if got := targetName(dir); got != "main" {
		t.Errorf("no file: %q", got)
	}
	mustWrite(t, filepath.Join(dir, "CMakeLists.txt"), "project(demo)\n")
	if got := targetName(dir); got != "demo" {
		t.Errorf("project: %q", got)
	}
	mustWrite(t, filepath.Join(dir, "CMakeLists.txt"), template)
	if got := targetName(dir); got != "nandu" {
		t.Errorf("add_executable: %q", got)
	}
}

func TestInstalledVersionsReadTheStatusFile(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "build", "vcpkg_installed", "vcpkg", "status"),
		"Package: vcpkg-cmake\nVersion: 2024-01-01\nArchitecture: x64-mingw-static\nMulti-Arch: same\nStatus: install ok installed\n\n"+
			"Package: fmt\nVersion: 12.2.0\nPort-Version: 1\nArchitecture: x64-mingw-static\nStatus: install ok installed\n\n"+
			"Package: fmt\nFeature: extra\nArchitecture: x64-mingw-static\nStatus: install ok installed\n\n"+
			"Package: broken\nVersion: 1\nStatus: purge ok not-installed\n")
	versions := installedVersions(dir)
	if versions["fmt"] != "12.2.0" || versions["vcpkg-cmake"] != "2024-01-01" || versions["broken"] != "" {
		t.Errorf("versions = %v", versions)
	}
	if len(installedVersions(t.TempDir())) != 0 {
		t.Error("no status file must give no versions")
	}
}
