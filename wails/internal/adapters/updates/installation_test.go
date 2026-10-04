package updates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAssetNamesFollowTheReleaseScript(t *testing.T) {
	cases := []struct {
		install Installation
		want    string
	}{
		{Installation{OS: "windows", Arch: "amd64", Variant: "full", Installed: true}, "VizcachaIDE-2.3.0-windows-amd64-full-setup.exe"},
		{Installation{OS: "windows", Arch: "amd64", Variant: "lite"}, "VizcachaIDE-2.3.0-windows-amd64-lite-portable.zip"},
		{Installation{OS: "darwin", Arch: "arm64", Variant: "full-go"}, "VizcachaIDE-2.3.0-macos-arm64-full-go.dmg"},
		{Installation{OS: "linux", Arch: "amd64", Variant: "full-python", AppImage: true}, "VizcachaIDE-2.3.0-linux-amd64-full-python.AppImage"},
		{Installation{OS: "linux", Arch: "amd64", Variant: "lite"}, "VizcachaIDE-2.3.0-linux-amd64-lite.tar.gz"},
	}
	for _, tc := range cases {
		if got := tc.install.AssetName("2.3.0"); got != tc.want {
			t.Errorf("%+v: %s, want %s", tc.install, got, tc.want)
		}
	}
	if got := (Installation{OS: "darwin", Arch: "arm64"}).SumsName(); got != "SHA256SUMS-darwin-arm64.txt" {
		t.Errorf("sums = %s", got)
	}
}

func TestDetectInstallationReadsTheBundledToolchain(t *testing.T) {
	dir := t.TempDir()
	if got := DetectInstallation("windows", "amd64", dir); got.Variant != "lite" || got.Installed {
		t.Fatalf("empty folder: %+v", got)
	}
	for _, sub := range []string{"toolchain/go", "toolchain/python"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "uninstall.exe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := DetectInstallation("windows", "amd64", dir); got.Variant != "full" || !got.Installed {
		t.Fatalf("full installed: %+v", got)
	}
	if got := DetectInstallation("darwin", "arm64", dir); got.Installed {
		t.Error("only Windows copies install themselves")
	}
}
