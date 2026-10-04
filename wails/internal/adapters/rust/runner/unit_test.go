package runner

import (
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
)

func TestFilterShowsRenderedTextAndDropsTheRest(t *testing.T) {
	d := &diagnostics{binary: "hola"}
	tests := []struct {
		name, line, want string
		keep             bool
	}{
		{"rustc diagnostic", `{"$message_type":"diagnostic","message":"boom","rendered":"error: boom\n  --> a.rs:1:1\n\n"}`, "error: boom\n  --> a.rs:1:1\n\n", true},
		{"cargo diagnostic", `{"reason":"compiler-message","package_id":"x","message":{"message":"boom","rendered":"warning: x\n\n"}}`, "warning: x\n\n", true},
		{"cargo artifact", `{"reason":"compiler-artifact","target":{"name":"dep"},"executable":null}`, "", false},
		{"build finished", `{"reason":"build-finished","success":true}`, "", false},
		{"rustc without text", `{"$message_type":"artifact_notification","artifact":"x"}`, "", false},
		{"cargo progress", "   Compiling hola v0.1.0 (C:\\a b)", "   Compiling hola v0.1.0 (C:\\a b)", true},
		{"broken json", `{"reason": `, `{"reason": `, true},
	}
	for _, test := range tests {
		stream, got, keep := d.Filter("stdout", test.line)
		if got != test.want || keep != test.keep {
			t.Errorf("%s: got (%q, %v), want (%q, %v)", test.name, got, keep, test.want, test.keep)
		}
		// A diagnostic goes to stderr even when cargo printed its JSON on stdout.
		if keep && got != test.line && stream != "stderr" {
			t.Errorf("%s: stream = %s, want stderr", test.name, stream)
		}
	}
}

func TestFilterRemembersTheExecutableOfTheBinary(t *testing.T) {
	d := &diagnostics{binary: "hola"}
	d.Filter("stdout", `{"reason":"compiler-artifact","target":{"name":"serde"},"executable":null}`)
	d.Filter("stdout", `{"reason":"compiler-artifact","target":{"name":"otro"},"executable":"C:\\x\\otro.exe"}`)
	if d.built() != "" {
		t.Fatalf("built = %q, want none yet", d.built())
	}
	d.Filter("stdout", `{"reason":"compiler-artifact","target":{"name":"hola"},"executable":"C:\\x\\hola.exe"}`)
	if d.built() != `C:\x\hola.exe` {
		t.Fatalf("built = %q", d.built())
	}
	if text := d.text("{\"reason\":\"build-finished\"}\nhecho\n"); text != "hecho\n" {
		t.Fatalf("text = %q", text)
	}
}

func TestCrashLines(t *testing.T) {
	tests := []struct {
		goos string
		code int
		want string
	}{
		{"windows", 101, ""}, {"windows", 0, ""}, {"windows", 1, ""},
		{"windows", int(int32(-1073741819)), lineSegfault}, // 0xC0000005
		{"windows", int(int32(-1073740791)), lineAborted},  // 0xC0000409
		{"windows", 3, lineAborted},
		{"linux", 101, ""}, {"linux", -11, lineSegfault}, {"linux", -7, lineSegfault},
		{"linux", -6, lineAborted}, {"linux", 139, lineSegfault}, {"linux", 134, lineAborted}, {"linux", 2, ""},
	}
	for _, test := range tests {
		if got := crashLineFor(test.goos, test.code); got != test.want {
			t.Errorf("%s %d: %q, want %q", test.goos, test.code, got, test.want)
		}
	}
}

func TestLinkerFallbackOnlyWithoutTheSelfContainedMinGW(t *testing.T) {
	gnu := rust.Toolchain{Host: "x86_64-pc-windows-gnu", Sysroot: `C:\rust`}
	clang := func() string { return `C:\llvm\bin\clang.exe` }
	none := func() string { return "" }
	has := func(string) bool { return true }
	lacks := func(string) bool { return false }
	tests := []struct {
		name  string
		tc    rust.Toolchain
		goos  string
		exist func(string) bool
		clang func() string
		want  string
	}{
		{"own linker", gnu, "windows", has, clang, ""},
		{"no own linker, clang found", gnu, "windows", lacks, clang, `C:\llvm\bin\clang.exe`},
		{"no own linker, no clang", gnu, "windows", lacks, none, ""},
		{"linux", rust.Toolchain{Host: "x86_64-unknown-linux-gnu"}, "linux", lacks, clang, ""},
		{"msvc host", rust.Toolchain{Host: "x86_64-pc-windows-msvc"}, "windows", lacks, clang, ""},
	}
	for _, test := range tests {
		if got := chooseLinker(test.tc, test.goos, test.exist, test.clang); got != test.want {
			t.Errorf("%s: %q, want %q", test.name, got, test.want)
		}
	}
	if got := cargoLinkerVariable("x86_64-pc-windows-gnu"); got != "CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER" {
		t.Errorf("variable = %s", got)
	}
}

func TestVersionNumberOf(t *testing.T) {
	tests := map[string]string{
		"rustc 1.99.0 (b940084d7 2026-09-28)\n": "1.99.0",
		"cargo 1.99.0 (abc 2026-08-01)":         "1.99.0",
		"clippy 0.1.99 (b940084d7 2026-09-28)":  "0.1.99",
		"rustfmt 1.9.0-stable (b9 2026-09-28)":  "1.9.0",
		"  LLVM version 23.1.2\n  Optimized":    "23.1.2",
		"":                                      "",
	}
	for text, want := range tests {
		if got := versionNumberOf(text); got != want {
			t.Errorf("%q: %q, want %q", strings.TrimSpace(text), got, want)
		}
	}
}
