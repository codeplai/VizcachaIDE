package rust

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/mod/semver"
)

// MinimumStable is the stable Rust current when M3 started (docs/PLAN_RUST.md section 3.2 point
// 7): there is no fixed floor, only "the stable channel"; an older rustc gets a non-blocking
// advice to run "rustup update stable".
const MinimumStable = "1.99.0"

// Release channels of rustc.
const (
	ChannelStable  = "stable"
	ChannelBeta    = "beta"
	ChannelNightly = "nightly"
)

// Toolchain is what "rustc -vV" and "rustc --print sysroot" say about the Rust in use.
type Toolchain struct {
	Rustc   string // path of rustc (a rustup proxy, usually)
	Version string // "1.99.0"
	Channel string // stable, beta or nightly
	Host    string // "x86_64-pc-windows-gnu"
	Sysroot string // the toolchain folder: lib/rustlib/etc has LLDB's formatters
}

// parseVerbose reads "rustc -vV": "rustc 1.99.0 (b940084d7 2026-09-28)", "host: ...",
// "release: 1.99.0" (or "1.100.0-beta.3", "1.101.0-nightly").
func parseVerbose(output string) (Toolchain, error) {
	var found Toolchain
	for _, line := range strings.Split(strings.ReplaceAll(output, "\r\n", "\n"), "\n") {
		name, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		switch name {
		case "host":
			found.Host = strings.TrimSpace(value)
		case "release":
			found.Version, found.Channel = splitRelease(strings.TrimSpace(value))
		}
	}
	if found.Version == "" || found.Host == "" {
		return Toolchain{}, fmt.Errorf("unexpected rustc -vV output: %q", output)
	}
	return found, nil
}

func splitRelease(release string) (version, channel string) {
	version, suffix, _ := strings.Cut(release, "-")
	switch {
	case strings.HasPrefix(suffix, ChannelBeta):
		return version, ChannelBeta
	case strings.HasPrefix(suffix, ChannelNightly):
		return version, ChannelNightly
	}
	return version, ChannelStable
}

// Edition is the edition a loose .rs file compiles with: 2024 from rustc 1.85 on, 2021 before.
// A Cargo project uses the edition of its Cargo.toml.
func (t Toolchain) Edition() string {
	if semver.Compare("v"+t.Version, "v1.85.0") >= 0 {
		return "2024"
	}
	return "2021"
}

// Advice returns the i18n keys of the non-blocking warnings about this toolchain, for goos:
// a Visual Studio (msvc) host on Windows, a channel other than stable, a stable older than
// MinimumStable.
func (t Toolchain) Advice(goos string) []string {
	advice := []string{}
	if goos == "windows" && strings.HasSuffix(t.Host, "-msvc") {
		advice = append(advice, "errors.rustMsvcHost")
	}
	if t.Channel != ChannelStable {
		advice = append(advice, "errors.rustNotStable")
	} else if semver.Compare("v"+t.Version, "v"+MinimumStable) < 0 {
		advice = append(advice, "errors.rustTooOld")
	}
	return advice
}

// SelfContainedLinker reports whether the toolchain brings its own MinGW linker (the
// x86_64-pc-windows-gnu toolchain of rustup does, in lib/rustlib/<host>/bin/self-contained). Only
// without it does the IDE pass llvm-mingw's clang as the linker (docs/PLAN_RUST.md section 3.2
// point 2: a fallback, never the default).
func (t Toolchain) SelfContainedLinker(exists func(path string) bool) bool {
	if !strings.HasSuffix(t.Host, "-windows-gnu") {
		return true // macOS and Linux link with the system cc; MSVC is not supported
	}
	return exists(t.SelfContainedGCC())
}

// SelfContainedGCC is the MinGW linker the Windows GNU toolchain brings, "" for other hosts.
// The IDE passes it explicitly: rustc otherwise runs the first x86_64-w64-mingw32-gcc on PATH,
// and llvm-mingw's (a clang wrapper) cannot find libgcc_eh (found in M3 R2).
func (t Toolchain) SelfContainedGCC() string {
	if !strings.HasSuffix(t.Host, "-windows-gnu") {
		return ""
	}
	return filepath.Join(t.Sysroot, "lib", "rustlib", t.Host, "bin", "self-contained", "x86_64-w64-mingw32-gcc.exe")
}

// FormattersDir is the folder of rustc's LLDB formatters (lldb_lookup.py, lldb_providers.py).
func (t Toolchain) FormattersDir() string {
	return filepath.Join(t.Sysroot, "lib", "rustlib", "etc")
}

// Environment is what every Rust process runs with (docs/PLAN_RUST.md sections 4.2 and 3.2
// point 6): plain output, the backtrace that locates a panic in the student's code, and no
// incremental build to save disk in target/.
func Environment(base []string) []string {
	fixed := map[string]string{
		"CARGO_TERM_COLOR": "never", "NO_COLOR": "1", "RUST_BACKTRACE": "1", "CARGO_INCREMENTAL": "0",
	}
	env := make([]string, 0, len(base)+len(fixed))
	for _, entry := range base {
		name, _, _ := strings.Cut(entry, "=")
		if _, replaced := fixed[strings.ToUpper(name)]; !replaced {
			env = append(env, entry)
		}
	}
	for _, name := range []string{"CARGO_TERM_COLOR", "NO_COLOR", "RUST_BACKTRACE", "CARGO_INCREMENTAL"} {
		env = append(env, name+"="+fixed[name])
	}
	return env
}
