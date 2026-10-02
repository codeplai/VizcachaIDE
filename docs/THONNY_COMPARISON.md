# VizcachaIDE vs Thonny — Gap analysis

> Reference: Thonny 5.0.0 (April 2026). VizcachaIDE at commit `ba14318`. See [ARCHITECTURE.md](ARCHITECTURE.md) for the code details.
>
> **Scope:** VizcachaIDE will be published worldwide with **English and Spanish** support. **Out of scope**: MicroPython, TinyGo and microcontrollers. The executable plan, split into phases and agents, is in [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md).
>
> **Status legend:** ✅ has it · 🟡 partial / UI only · 🔴 simulated · ❌ missing
> **Priority:** P0 = essential to be a "Thonny for Go" · P1 = important · P2 = nice to have · P3 = niche

## 1. Summary

Thonny does not stand out for having many features. It stands out for **making execution visible** to a beginner: a step-by-step debugger, live variables, a frame per call, an assistant that explains errors, and an all-inclusive installer.

What VizcachaIDE already has is the **shell**: the editor, execution with `go run`, and the variables and stack panels. It lacks the **engine**, because the debugger is a simulation. Until there is a real debugger built on Delve, the value proposition against Thonny does not exist. That is priority number one.

## 2. Comparison table

### 2.1 Editor

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| Syntax highlighting | ✅ (custom regex) | Works; optionally use the Pygments Go lexer, which is already a dependency | — |
| Live syntax error highlighting (unclosed quotes and parentheses) | ❌ | `gopls` diagnostics or `gofmt -e` in the background | P1 |
| Parenthesis matching | ❌ (the README advertises it, but it is not implemented) | `QTextEdit.ExtraSelection` on the pair | P2 |
| Highlight occurrences of a name, and locals vs globals | ❌ | `gopls` `textDocument/documentHighlight` | P2 |
| Smart autocompletion (Jedi) | 🟡 (static lists, ~10 packages) | **gopls** `textDocument/completion` | P1 |
| Call-tips (parameters when typing `(`) | ❌ | `gopls` `signatureHelp` | P2 |
| Go to definition (Ctrl+click) | ❌ | `gopls` `definition` | P2 |
| Indent or dedent a block, comment or uncomment | ❌ | Simple editor actions | P2 |
| Auto-indent | ✅ (fixed 4 spaces, also after `:`) | Go uses **tabs**: integrate `gofmt` and honor `tab_size` | P1 |
| Format code | ❌ | `gofmt` / `goimports` on save | **P0** (idiomatic in Go) |
| Find / Replace, Go to line | ❌ | Non-modal `QDialog` + Ctrl+F / Ctrl+H / Ctrl+G | P1 |
| Multiple tabs | ✅ | — | — |
| Recent files | 🟡 (only the last one) | MRU list in `QSettings` | P2 |
| Automatic reload on external changes | ❌ | `QFileSystemWatcher` | P2 |
| Run without saving ("Untitled") | ❌ (forces saving) | Write to a temp file and run `go run` on it | P1 |
| Line-length guides, Ctrl +/- zoom | ❌ | Simple | P3 |
| Print | ❌ | `QPrintDialog` | P3 |

### 2.2 Execution and Shell

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| Run (F5) | ✅ `go run` | Extend to package or module (`go run .`) | — |
| stdin input in the console | ✅ | — | — |
| Stop / Interrupt | 🟡 (`kill` only) | Send Ctrl+C / SIGINT before killing the process | P2 |
| **Interactive Shell / REPL** | ❌ | Go has no official REPL: embed **yaegi** (Go interpreter) or **gomacro** | P1 (very Thonny, but hard) |
| Program arguments | ❌ | "Program arguments" field passed as `go run file.go -- args` | P1 |
| Run in the system terminal / open a terminal with PATH | ❌ | Launch `cmd` or `wt` with the configured Go environment | P2 |
| ANSI colors and `\r` (progress bars) | ❌ | ANSI parser in `ConsoleWidget` | P2 |
| Links from errors to the line | ❌ | Parse `file.go:L:C` in the compiler's stderr or in the panic and make it clickable | **P0** |
| Plotter for printed numbers | ❌ | View with `QtCharts` / pyqtgraph | P3 |
| Build | ✅ (`go build`, but it blocks the UI) | Move it to `QProcess` | P1 |
| Tests | ❌ (Thonny does not have them either) | `go test` with a results panel: Go-specific added value | P2 |

### 2.3 Debugger — the core of Thonny

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| Real line-by-line debugging | 🔴 **simulated** | **Delve**: `dlv dap` (Debug Adapter Protocol) or `dlv debug --headless --api-version=2` with JSON-RPC | **P0** |
| Step over / into / out | 🔴 (random line) | Delve's `next` / `step` / `stepOut` commands | **P0** |
| Breakpoints | 🟡 (drawn; the debugger ignores them) | `setBreakpoints` (DAP) | **P0** |
| Resume and Run to cursor | ❌ | `continue` + temporary breakpoint | P1 |
| Current line highlight | ✅ (UI) | Feed it with Delve's real `stopped` event | **P0** |
| Variables panel | 🟡 (UI ready, fake data) | `scopes` / `variables` (DAP), with lazy expansion of structs, slices and maps | **P0** |
| Stack panel | 🟡 (UI ready, fake data) | `stackTrace`; clicking a frame jumps to the code | **P0** |
| **Frames as nested windows** (visible recursion) | ❌ | Feasible with `stackTrace` data and `variables` per frame | P1 — **differentiator** |
| **Step-by-step expression evaluation** ("nicer debug") | ❌ | Very hard in a compiled language. Alternative: instrument the AST with `go/ast` and emit traces, or use **yaegi** as a "teaching mode" engine | P2 — research |
| Step back | ❌ | `rr` + Delve (Linux only); or record the snapshots of each step and navigate them in the UI | P3 |
| Heap / reference model (name → id → value) | ❌ | Show pointer addresses and pointers inside slices and maps: very pedagogical in Go (pointers, slices sharing an array) | P2 — differentiator |
| Object inspector | ❌ | Panel with details of the selected value: `len` and `cap` of a slice, fields of a struct | P2 |
| Goroutines | ❌ (Python has none) | Delve's goroutines view: **Go-specific added value** | P1 |

### 2.4 Help for beginners

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| **Assistant**: explains errors in plain language | ❌ | Catalog of frequent Go compiler errors (`declared and not used`, `imported and not used`, `missing return`, `cannot use x (type) as …`, `index out of range` and `nil map` panics) with a **bilingual EN/ES** explanation (the original Go message is shown too) | **P0** — differentiator |
| Static warnings (Pylint, MyPy) | ❌ | `go vet` + `staticcheck`, shown in the Assistant | P1 |
| "Module shadowing" warning | ❌ | Go equivalent: warn about a `package` other than `main`, a missing `func main` or a missing `go.mod` | P2 |
| Outline view | ❌ | `gopls documentSymbol` or `go/ast` | P2 |
| TODO, Notes, built-in Help views | ❌ | Simple; the help can link to *A Tour of Go* | P3 |

### 2.5 Packages, interpreters and environment

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| **Python bundled in the installer** | ❌ (requires installing Go and Delve by hand) | Package a portable **Go toolchain + dlv + gopls** inside the installer | **P0** for beginners |
| Graphical package manager (pip) | ❌ | UI for `go mod init`, `go get`, `go mod tidy` and search on pkg.go.dev | P1 |
| Interpreter selector and venvs | 🟡 (configurable Go path) | Go version selector (multiple toolchains, `GOTOOLCHAIN`) in the status bar | P2 |
| Multi-file project | ❌ (loose single file only) | Open a folder, detect `go.mod` and run `go run .` | P1 |
| Remote over SSH | ❌ | `GOOS`/`GOARCH` + `scp` + remote execution | P3 |
| Files view (local and remote) | ❌ | `QTreeView` + `QFileSystemModel` | P1 |

### 2.6 Appearance, i18n and distribution

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| Light and dark themes | 🟡 (saved, **not applied**) | Apply the palette and the highlighter colors from `QSettings` | P1 (low effort) |
| Configurable fonts | ✅ | — | — |
| Simple / regular / expert modes | ❌ | Hide menus and panels according to the chosen level | P2 |
| HiDPI scaling | ❌ | `Qt.AA_EnableHighDpiScaling` | P2 |
| **English/Spanish internationalization** | ❌ (UI in English only) | gettext + Babel; automatic detection of the system language and a selector; errors, help, README and installer in EN/ES | **P0** (worldwide release) |
| First-run dialog | ❌ | Language, theme, and detection or installation of Go and Delve | P1 |
| Installer and portable version | ❌ | PyInstaller/Nuitka + Inno Setup; portable zip | **P0** |
| Plugins | ❌ | `vizcacha.plugins` entry points with `load_plugin(workbench)` | P3 |

### 2.7 Classroom and research

| Thonny feature | VizcachaIDE | Go equivalent / how to do it | Priority |
|---|---|---|---|
| Action logs and export | ❌ | Optional JSONL log | P3 |
| Replayer | ❌ | Replay those logs | P3 |
| `defaults.ini` for classrooms | ❌ | `QSettings` with a system-wide defaults file | P3 |

## 3. Suggested roadmap

> The breakdown by phases, contracts and parallel tracks is in [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md).

**Phase 1: make what is advertised work (P0)**
1. Real debugger with `dlv dap`: breakpoints, stepping, variables, stack and current line. The existing Qt signals already fit.
2. Clickable compiler errors and panics, plus an **Assistant** with bilingual EN/ES explanations of the ~25 most common errors.
3. `gofmt` on save and tab indentation.
4. Windows installer with Go, Delve and gopls included.

**Phase 2: practical parity with Thonny (P1)**
5. gopls over LSP: real autocompletion, live diagnostics and hover.
6. Apply the already-saved themes and options, bilingual EN/ES interface and a first-run wizard.
7. Projects with `go.mod`, Files view and a `go get` / `go mod tidy` interface.
8. Frames in nested windows and a goroutines view.
9. Program arguments, Find/Replace, Run to cursor and `go vet`/`staticcheck`.
10. REPL with yaegi.

**Phase 3: Go-specific differentiators (P2+)**
11. Memory visualizer: pointers, slices sharing an array, `len` and `cap`.
12. `go test` panel.
13. Teaching mode for expression evaluation (yaegi or AST instrumentation).
14. Simple/expert modes, plugins and classroom logs.

## 4. Sources on Thonny

- https://thonny.org · https://thonny.org/blog/
- Official CHANGELOG: https://github.com/thonny/thonny/blob/master/CHANGELOG.rst
- Release notes 3.1 to 5.0: https://newreleases.io/project/github/thonny/thonny
- https://pypi.org/project/thonny/ · https://en.wikipedia.org/wiki/Thonny
- Wiki: Plugins, User-action-logs and DeploymentOptions at https://github.com/thonny/thonny/wiki
- Annamaa, A. (2015). *Thonny, a Python IDE for learning programming*. ACM Koli Calling. https://dl.acm.org/doi/10.1145/2828959.2828969
