package vcpkg

import (
	"path/filepath"
	"strings"
	"testing"
)

const fmtUsage = `The package fmt provides CMake targets:

    find_package(fmt CONFIG REQUIRED)
    target_link_libraries(main PRIVATE fmt::fmt)

    # Or use the header-only version
    find_package(fmt CONFIG REQUIRED)
    target_link_libraries(main PRIVATE fmt::fmt-header-only)
`

func TestUsageKeepsTheFirstRunAndUsesTheProjectTarget(t *testing.T) {
	got := parseUsage(fmtUsage, "nandu")
	want := "find_package(fmt CONFIG REQUIRED)|target_link_libraries(nandu PRIVATE fmt::fmt)"
	if strings.Join(got, "|") != want {
		t.Errorf("got %v", got)
	}
}

func TestUsageDropsProseAndLaterBlocks(t *testing.T) {
	text := "sqlite3 provides pkgconfig bindings.\nsqlite3 provides CMake targets:\n\n" +
		"    find_package(unofficial-sqlite3 CONFIG REQUIRED)\n    target_link_libraries(main PRIVATE unofficial::sqlite3::sqlite3)\n\n" +
		"The package can be configured:\n\n    set(X_IMPLICIT OFF)\n"
	got := parseUsage(text, "app")
	if len(got) != 2 || got[1] != "target_link_libraries(app PRIVATE unofficial::sqlite3::sqlite3)" {
		t.Errorf("got %v", got)
	}
}

func TestUsageJoinsMultiLineCommandsAndRenamesAnyExampleTarget(t *testing.T) {
	text := "  find_package(Boost REQUIRED COMPONENTS asio)\n  target_link_libraries(<your target> PRIVATE\n      Boost::asio\n      Boost::system)\n  target_include_directories(main PRIVATE ${X})\n"
	got := parseUsage(text, "juego")
	if len(got) != 3 || got[1] != "target_link_libraries(juego PRIVATE Boost::asio Boost::system)" || got[2] != "target_include_directories(juego PRIVATE ${X})" {
		t.Errorf("got %v", got)
	}
}

func TestUsageIsReadFromTheInstalledShareFolder(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(InstalledDirectory(dir, "x64-linux"), "share", "fmt", "usage"), fmtUsage)
	got := UsageLines(dir, "x64-linux", "fmt", "nandu")
	if len(got) != 2 || got[0] != "find_package(fmt CONFIG REQUIRED)" {
		t.Errorf("got %v", got)
	}
}

func TestHeaderOnlyPortsFallBackToFindPath(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "build", "vcpkg_installed", "vcpkg", "info", "stb_2024-01-01_x64-linux.list"),
		"x64-linux/\nx64-linux/include/\nx64-linux/include/stb/detail/internal.h\nx64-linux/include/stb_image.h\n"+
			"x64-linux/include/stb_image_write.h\nx64-linux/share/stb/copyright\n")
	got := UsageLines(dir, "x64-linux", "stb", "nandu")
	want := []string{`find_path(STB_INCLUDE_DIRS "stb_image.h")`, "target_include_directories(nandu PRIVATE ${STB_INCLUDE_DIRS})"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("got %v", got)
	}
	mustWrite(t, filepath.Join(dir, "build", "vcpkg_installed", "vcpkg", "info", "nlohmann-json_3.11_x64-linux.list"),
		"x64-linux/include/nlohmann/json.hpp\n")
	if got := UsageLines(dir, "x64-linux", "nlohmann-json", "nandu"); got[0] != `find_path(NLOHMANN_JSON_INCLUDE_DIRS "nlohmann/json.hpp")` {
		t.Errorf("hyphen port: %v", got)
	}
	if got := UsageLines(dir, "x64-linux", "unknown", "nandu"); got != nil {
		t.Errorf("unknown port: %v", got)
	}
}
