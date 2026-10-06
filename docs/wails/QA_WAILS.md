# QA — VizcachaIDE Wails 2.0.0-rc1

## 2.5.0 (M4, C++ with CMake and vcpkg): QA, 2026-10-05

Same environment and harness, branch `feature/cpp-cmake-vcpkg` (plan: `docs/PLAN_CPP_CMAKE.md`),
with CMake 4.4.4, Ninja 1.13.2 and a vcpkg snapshot (2026-09-26, no `.git`) plus its seed in
`wails/.toolchain-dev`. Also covers the work merged after 2.4.0: New project, package search, Close
folder and the integrated terminal.

- **Result: 116/118 steps pass.** Go, settings, first run (34/34), Python (14/14), C++ (16/16), the new
  `cpp-cmake` phase (`steps-cpp-cmake.mjs`, 6 steps × EN/ES: New project C++ with CMake, `fmt` searched
  and installed from the Packages dialog, `fmt::print` runs with no Problems, a breakpoint in the
  debugger, an old folder without `CMakeLists.txt` gets one and runs, removing `fmt` cleans
  `vcpkg.json` and the block), New project (12/12) and the terminal (12/12).
- **The 2 failures are `rust-inlay` (EN and ES) and are not a regression of this release:** the same
  step fails on main (70a4576), on 1d37c61 and on 2.4.0 itself (4ee9077), where it passed on
  2026-10-04, so something in this machine's environment changed. rust-analyzer still gives the
  `Vec<i32>` hint when asked directly (`TestRealRustAnalyzerGivesTheTypeOfAVec` passes); the
  editor side is to be investigated separately.
- **Fixed during QA:** the empty Problems tab said "Go checks your code…" in every language.
- **Known limitations:** the first install of a library needs internet and compiles it (about 3 min
  cold for `fmt`, 14 s from the binary cache); vcpkg extracts PowerShell 7 (247 MB) once per user;
  with an antivirus hooking every process (Bitdefender's `bdhkm64.dll` here) a stack overflow is
  reported as an access violation (0xC0000005); the full-cpp installer was not rebuilt (only the
  fetch of CMake, Ninja, vcpkg and the seed was run: 50 + 21 + 109 MB).
- Evidence: [qa-2.5/](qa-2.5/).

## 2.4.0 (M3, Rust): QA, 2026-10-04

Same environment and harness, branch `m3-rust`, with Rust 1.99.0 stable (`x86_64-pc-windows-gnu`,
clippy, rustfmt, rust-analyzer) installed by rustup in `wails/.toolchain-dev` and lldb-dap 23.1.2 from
llvm-mingw.

- **Result: 80/80 steps pass.** The full run gave 78/80: the two Go `editor-intel` steps opened
  `qa/args/main.go` instead of `qa/complete/main.go` (a harness bug: `openByName` took the first
  `main.go` of the tree; it now takes a folder, using the tree's new `data-path`), and the EN and ES
  phases were rerun, 14/14 each. Go, Python and C++ have no regressions, and the new Rust phase
  (`steps-rust.mjs`, 9 steps × EN/ES) passes: run a loose file with `read_line`, the moved value
  (E0382) and an index panic explained in the IDE language, live problems from rust-analyzer in a
  Cargo project, the `Vec<i32>` inlay hint, *More → Build*, the debugger (variables `n`/`resultado`/`i`)
  with keyboard input, and rustfmt on save.
- **Fixed during QA:**
  - rust-analyzer started on a loose file never loaded a Cargo project opened later: one
    rust-analyzer per Cargo workspace plus one for loose files (`analyzer.Router`); every loose file
    is a detached file (`lsp.PulledConfiguration`).
  - Inlay hints asked once while rust-analyzer was still indexing came back empty: the editor asks
    again with backoff, and when diagnostics or the server status change.
  - LLDB 23 read the payload of Rust enums at the wrong offset on `*-gnu` (`Some(7)` → `Some(440)`):
    `lldbdap/data/vizcacha_rust_enums.py` patches the formatters; the hidden `iter` of a `for` loop is
    no longer a variable.
  - An llvm-mingw on PATH shadowed rustc's own MinGW linker (`-lgcc_eh` not found): the linker is
    pinned.
- **Known limitations:** Rust is not bundled (rustup); rust-analyzer needs several seconds and a few
  hundred MB per project; borrow-checker errors of a loose file appear after a run (clippy), not
  live. macOS and Linux are covered only by CI.
- **Multi-file experiment** (`qa.mjs --only multi`, asked by the user, 24/24 after the fixes): in
  each language a 4-file project (main → a second file → a third one, and a fourth that uses an
  external library: Go `github.com/google/uuid`, Python `humanize`, header-only `nlohmann/json`,
  Rust `rand`) runs, Ctrl+click opens the definition in another file, hover documents the library's
  function, a breakpoint in the third file shows `main → second → third` with its variables, and an
  error in the third file is listed in Problems and opens that file. It found four bugs, all fixed:
  breakpoints of other files were not sent to the debugger; gopls (and rust-analyzer) never heard of
  `go.mod`/`Cargo.toml` changes made by the package commands (`lsp.ManifestFiles`); `rand` did not
  build with rustup's GNU toolchain (its dlltool needs an assembler it lacks: LLVM's `llvm-dlltool`
  is passed, and `RS-DLLTOOL` explains a missing one); LLDB's disassembly pseudo-paths
  (`tienda.exe\`main`) counted as user code, so the standard library frames below `main` showed.
  Known gaps: with only rustup (lite, no llvm-mingw) `rand` cannot build; C++ Ctrl+click reaches the
  declaration in the header (clangd runs without a background index).
- After those fixes the regression run gave 77/80: the Go `modules` step (EN and ES) opened
  `qa/loop/main.go` because the reload right after the harness restart still showed the previous
  folder (the step now retries until the tree is `qa/mod`; both phases then passed 14/14), and the
  C++ crash card was missing once in ES (not reproduced: the C++ ES phase passed 8/8 when run again).
- Evidence: [qa-2.4/](qa-2.4/).

## 2.3.0 (M2, C++): QA, 2026-10-04

Same environment and harness, branch `m2-cpp`, with llvm-mingw 20260922 (clang++, lldb-dap, clangd
and clang-format 23.1.2) from `wails/.toolchain-dev`; the adapter tests also use WinLibs GCC 16.2.0.

- **Result: 62/62 steps pass.** Go (34 steps) and Python (12) have no regressions, and the new C++
  phase (`steps-cpp.mjs`, 8 steps × EN/ES) passes: compile and run with `std::cin` in a terminal, the
  undeclared name explained in the IDE language, a null-pointer crash ending with "Segmentation
  fault" and explained, live problems from clangd, *More → Build* leaving `main.exe` without running
  it, the debugger (breakpoint, variables `n`/`resultado`/`i`, no C runtime frames), keyboard input
  **while debugging** and clang-format on save.
- **Fixed during QA** (all found by the new steps):
  - A crash line has no file, so the Assistant explained it with the Go catalog (no card text):
    `ExplainDiagnostics` now takes the language of the last run for diagnostics without a file.
  - Two explain requests could answer out of order and the older one replaced the newer cards; the
    frontend now sends them one at a time.
  - A crash ended with "Your program didn't run: there is 1 problem" instead of `run.crashed`.
  - The More menu failed for C++ (`packageActions` was `null`), so *Build* was missing.
  - A reopened file (reloaded window) got a second `didOpen` that clangd ignores, so its problems
    and the "ready" status never came back; the last diagnostics and the status are sent again.
  - clangd writes clang's messages capitalised and with " (fix available)": the C++ catalog now
    reads them.
  - What a debugged program printed vanished from Output when the session ended.
- **Known limitations:** the NSIS installers of `full-cpp` and `full` were not built on this
  machine, only the portable zips (smoke-tested: `full-cpp` 107 MB before lldb's Python was kept,
  `full` 247 MB). macOS and Linux use the system compiler and are built only by CI. Without
  lldb-dap (GCC alone) C++ runs but does not debug.
- Evidence: [qa-2.3/](qa-2.3/).

## 2.2.0 (M1, Python): QA, 2026-10-04

Same environment and harness, branch `m1-python`, with Python 3.12.14 (debugpy 1.8.22,
python-lsp-server 1.15.0 + pyflakes 3.2.0, ruff 0.16.10) from the development venv.

- **Result: 46/46 steps pass.** The Go parity matrix (34 steps: EN, ES, settings persistence, first
  run) has no regressions, and the new Python phase (`steps-python.mjs`, 6 steps × EN/ES) passes:
  run with `input()` in a terminal, the NameError explained in the IDE language, live problems from
  pyflakes, the debugger (breakpoint, variables `n`/`resultado`/`i`, no internal variables), keyboard
  input **while debugging** (`input()` reads what is typed in Output), and the console (`>>>`,
  `2 + 2`, a remembered `x`).
- **Fixed during QA** (all found by the new steps):
  - A debug session that never pauses stayed on "Starting the debugger" and hid the program's
    output; it is now "Debugging <file> · running", with a "Running" chip, no step toolbar and no
    empty "Variables in ()".
  - Tracebacks were not explained after a run in a terminal: a PTY merges stderr into stdout, so the
    Assistant now reads the whole output when the run had `echo: true`.
  - pyflakes (live) messages were not in the Python catalog (only ruff's wording).
  - ConPTY writes a prompt's trailing space as a cursor move: `input("Nombre: ")` showed
    `Nombre:Ana`. `pty.Clean` turns cursor-forward into spaces.
  - The status bar said "gopls ..." and "Go 1.25.5" for Python files: it now says "Code helper ..."
    and shows the active language's runtime ("Python 3.12.14").
- **Known limitations:** `lsp:status` carries no language yet, so with Go and Python files open the
  status shows the last server that reported. The NSIS installer of `full-python` was not built on
  this machine (LZMA of 223 MB was too slow); the portable zip (68.2 MB) was built and smoke-tested.
  macOS and Linux packages are built only by CI.
- Evidence: [qa-2.2/](qa-2.2/).

## 2.1.0 (M0, multi-language core): parity QA, 2026-10-03

Same environment and harness as below (Windows 10, `wails dev`, Go 1.25.5 / dlv 1.27.2 /
gopls 0.23.0, Edge 154 headless), on branch `m0-nucleo-multilenguaje` after N0–N5 were integrated.

- **Result: 34/34 steps pass** in English, Spanish, settings persistence and first run. Go behaves
  as in 2.0 on top of the new core (registry, `internal/protocol`, shared process supervisor):
  run, stdin, arguments, Stop, explained errors and panics, gopls (underline, hover, definition,
  completion), gofmt on save, the close dialog, modules, editor extras, the real debugger (variables,
  Next line, "just changed", goroutines, Run to here), the 1024 px layout and the first-run wizard.
- **Q2 closed:** settings persistence after a restart is now confirmed end to end (Spanish, dark
  theme, 18 px survive the restart; "auto" resolves to the system language). The harness writes
  2.0-style settings (`goPath`…), so the 2.0 → 2.1 migration also runs in every session.
- **Fixed in the harness:**
  - Edge 154 headless left a reloaded tab with `visibilityState: hidden`, so no frames were
    painted and every wait or click after a reload hung. 2.0 (`main`) showed the same failure, so it
    was not a regression. `ui.mjs` `reload` now brings the tab to the front and enables focus
    emulation.
  - `format-save` failed in the second language because both phases share the project and the file
    was already formatted. The step now restores the unformatted file first.
- **Fixed in the app:** Settings → Tools showed empty "Python tools" / "C++ tools" headings,
  because those profiles have no tools in 2.1. Groups without tools are now hidden.
- Evidence: EN/ES screenshots in [qa-2.1/](qa-2.1/). Python and C++ only exist as profiles in
  2.1: new files get their template, and running one answers "not available".

## Summary (2.0.0-rc1)

Parity QA against 1.0, done on **2026-10-01 on Windows 10** with the **real backend**: `wails dev`, Go 1.25.5, Delve 1.27.2 and gopls 0.21.1, driven through headless Edge with CDP.

- **Works the same as or better than 1.0:** run, stdin, arguments, Stop, explained errors (compilation and panics), gopls underlining, completion, hover, go to definition, gofmt on save, closing with unsaved changes, the real debugger, first run, automatic language and the modules dialog.
- **Fixed during QA:**
  - Keyboard input (stdin) was missing in Output.
  - Program arguments were hidden in Settings; they are now in the title bar.
  - The Assistant did not explain gopls live diagnostics.
  - A panic message arrived split across two lines.
  - "Run to here" was missing.
  - There was a bug in the messages sent to gopls.
  - The Assistant showed cards from other files; it now shows only the active file and the last run.
- **Fixed after QA:** each example now lives in its own folder, so gopls no longer reports "main redeclared". See §3 for what is still pending.

## 1. Environment

| Item | Value |
|---|---|
| OS | Windows 10 Pro 19045 x64 |
| Backend | `wails dev` (Wails 2.16.0), real Go 1.25.5 / dlv 1.27.2 / gopls 0.21.1 |
| Driver | Edge headless + Chrome DevTools Protocol (`wails/packaging/qa/e2e/`) |
| Data | Copy of `examples/` in `%TEMP%\wq\proj`; settings in `%APPDATA%\VizcachaIDE` (removed after the run) |

## 2. Parity with 1.0

✅ same · ✅+ better · ⚠️ partial · ❌ missing · ⏳ not verified end-to-end

| 1.0 feature | Wails | Evidence |
|---|---|---|
| Run (F5), output, "finished" line | ✅ | E2E `hello`: "✓ Finished in 0.7 s" |
| Keyboard input (stdin) | ✅ (fixed in QA) | E2E `stdin`: "Hola, Ana. Tienes 30 años." |
| Program arguments with quotes | ✅+ (title bar) | E2E `args`: `[one, two words, three]` |
| Stop a running program | ✅ | E2E `stop` |
| Run unsaved file | ✅ | E2E first-run: blank file ran |
| Modules (`go.mod`, `go run .`) | ✅ | E2E `modules` (2nd run): two-file module printed "module ok" |
| Go Modules dialog | ✅ | E2E `modules-dialog` (ES) |
| Assistant: compile errors explained, EN/ES | ✅ | E2E `error-flow`: "Variable "count" is never used" / "La variable "count" nunca se usa" |
| Assistant: runtime panics | ✅ (panic line fixed) | E2E `panic`: P-NIL-MAP card |
| Clickable `file.go:L:C` in Output | ✅ | E2E: link → line 5, column 2 |
| Live diagnostics (gopls) with exact range | ✅+ | E2E: underline exactly on `count`; inline explanation |
| Completion, hover, go to definition | ✅ | E2E `editor-intel` |
| gofmt on save | ✅ | E2E `format-save` |
| Close tab with unsaved changes | ✅ | E2E dialog "Save changes to main.go before closing?" |
| Find/replace, go to line, comment, zoom, Outline | ✅ | E2E `editor-extras` (2nd run): find, goto, comment toggled, zoom 14→15 px, Outline `add`, `main` |
| Debugger: breakpoints, steps, variables, call stack | ✅+ | E2E screenshots: paused, "Next line", "just changed", "How you got here", Goroutines |
| Run to cursor | ✅ (added in QA) | unit test + toolbar |
| Themes, font size, language | ⚠️ | auto language ✅ (backend resolves `es`); persistence after restart not confirmed by E2E (selector step failed) |
| First-run wizard | ✅+ (new) | E2E `firstrun`: language → "Go 1.25.5 está listo" → hello ran |
| Bundled Go toolchain | ✅ | `wails/packaging` install QA (full installer) |
| Installers (Windows) | ✅+ | 57.5 MB full / 6.4 MB lite (1.0: 75 / 25 MB), no GPL |
| 1024 px window | ✅ | E2E `layout-1024` (2nd run): no overflow; debug toolbar 473 px inside a 584 px editor |

## 3. Open issues

| # | Severity | Issue | Next step |
|---|---|---|---|
| Q1 | ~~High~~ **Fixed** | Top-level examples shared one folder (`package main` ×6), so gopls reported "main redeclared". | Fixed: each example now lives in `examples/<name>/<name>.go`; `gopls check` reports nothing. |
| Q2 | Medium | Settings persistence after restart not confirmed end-to-end (the settings JSON store has unit tests). | Repeat `settings-steps.mjs` once the menu selector is fixed. |
| Q3 | Low | The E2E harness is timing-sensitive under load (CDP `protocolTimeout`, 30 s waits): a step can fail in one run and pass in the next while the app behaves correctly (checked with the failure screenshots). | Raise the harness timeouts; run the steps in smaller groups. |
| Q4 | Low | "Replace all" does not ask for confirmation (CodeMirror's search panel replaces directly). | Custom search panel if wanted. |

## 4. How to repeat

```
cd wails && wails dev                        # backend + http://localhost:34115
cd wails/packaging/qa/e2e && npm ci
node qa.mjs --lang en                        # and --lang es; results in %TEMP%\wq\results.json
```

macOS and Linux were not executed (no machines here). The CI workflows cover build and unit tests on the 3 OS.
