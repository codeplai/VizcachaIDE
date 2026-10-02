# VizcachaIDE — Parallel development plan

> Goal: go from the current prototype ([ARCHITECTURE.md](ARCHITECTURE.md)) to a **public 1.0 release, bilingual (English/Spanish), with installers for Windows, macOS and Linux**. It covers the P0/P1 gaps from [THONNY_COMPARISON.md](THONNY_COMPARISON.md).
>
> The work is split into **independent tracks** that AI agents (or people) can develop in parallel, each in its own *git worktree*.

![Target architecture](arquitectura-objetivo.svg)

## 0. Decisions made

| Topic | Decision |
|---|---|
| Qt binding | **PyQt5** (kept). See the license risk in §8 |
| Platforms | Windows, macOS (Intel + Apple Silicon) and Linux |
| Languages | **English and Spanish** in the interface, messages, error explanations and documentation |
| Language selection | Automatic based on the system (`es*` → Spanish, anything else → English). Can be changed on first run and in Settings |
| Out of scope | **MicroPython, TinyGo and microcontrollers in general** |
| Architecture style | Clean Architecture + DDD, library-first |

## 1. Target architecture

```
vizcacha/
├── domain/                 # Pure Python. Importing Qt or subprocess is forbidden.
│   ├── debugging.py        # Breakpoint, StackFrame, Variable, Goroutine, DebugState
│   ├── diagnostics.py      # Diagnostic, Severity, SourceLocation
│   ├── explanations.py     # ErrorExplanation (stable ids, translatable texts)
│   ├── completion.py       # CompletionItem, SignatureHelp
│   └── project.py          # GoModule, RunConfiguration
├── application/            # Use cases + ports. Imports only domain.
│   ├── ports.py            # Contracts (typing.Protocol), frozen in phase 0
│   ├── run_program.py
│   ├── debug_session.py
│   ├── explain_error.py
│   └── format_on_save.py
├── infrastructure/         # Concrete adapters that implement the ports.
│   ├── go_toolchain/       # GoEnvironment, GoProcessRunner (QProcess), GoFormatter
│   ├── delve_dap/          # DelveDapDebugger: DAP client over QTcpSocket
│   ├── gopls_lsp/          # GoplsLanguageServer: JSON-RPC over QProcess + lsprotocol
│   ├── error_catalog/      # GoOutputParser + ErrorCatalog (regex → ErrorExplanation)
│   └── settings/           # QSettingsRepository (keeps the current keys)
├── i18n/
│   ├── translator.py       # install_language(), _() and ngettext()
│   ├── babel.cfg
│   └── locale/{en,es}/LC_MESSAGES/vizcacha.po
├── ui/                     # Everything that is Qt.
│   ├── app.py              # Composition root: creates adapters and injects them
│   ├── workbench.py        # Workbench: registry of actions, menus, views and panels
│   ├── main_window.py      # Layout + Workbench only (no logic)
│   ├── editor/             # CodeEditor, LineNumberArea, GoSyntaxHighlighter
│   └── features/           # Each feature is an internal plugin with register(workbench)
│       ├── run/  debugger/  assistant/  language/  editor/  project/
└── main.py
```

**Dependency rules** (checked by an architecture test in CI):

- `domain` → imports nothing from the project or from Qt.
- `application` → only `domain`.
- `infrastructure` → `domain`, `application.ports` and Qt *Core* (QProcess/QTcpSocket); never widgets.
- `ui` → everything above. It is the only place with `QtWidgets`.
- `ui/app.py` is the only place that instantiates concrete adapters (composition root).

**Code rules** (*software-architecture* skill):

- Files under 200 lines and functions under 50; maximum nesting of 3; *early return*.
- Domain names. `utils.py`, `helpers.py` and `common.py` are forbidden.
- Typed exceptions (`GoToolchainNotFoundError`, `DebugAdapterError`…), never a silent `except Exception:`.
- Library-first: before writing custom infrastructure, look for a maintained package.

### 1.1 Contracts, frozen in phase 0

The **source of truth** is the code: [`vizcacha/application/ports.py`](../vizcacha/application/ports.py)
(ports and mandatory Qt signals), [`vizcacha/domain/`](../vizcacha/domain/) (domain objects),
[`vizcacha/application/errors.py`](../vizcacha/application/errors.py) (typed exceptions) and
[`vizcacha/ui/workbench.py`](../vizcacha/ui/workbench.py) (API for features). Summary:

| Port | Methods | Mandatory Qt signals |
|---|---|---|
| `GoToolchainPort` | `environment`, `run`, `build`, `stop`, `is_running`, `write_input`, `format_source` | `output_received(str)`, `error_received(str)`, `execution_finished(int)` |
| `DebuggerPort` | `start(config, breakpoints)`, `set_breakpoints`, `step_over/into/out`, `resume`, `run_to`, `variables(reference)`, `stop`, `is_active` | `stopped(DebugState)`, `output(str, category)`, `terminated(int)` |
| `LanguageServerPort` | `open/change/close_document`, `completion`, `hover`, `definition`, `signature_help`, `shutdown` | `diagnostics_published(Path, list[Diagnostic])` |
| `ErrorExplainerPort` | `parse(raw_output, working_dir)`, `explain(diagnostic)` | — |
| `SettingsRepository` | `get(key, default)`, `set(key, value)` | — |

`tests/test_contracts.py` verifies that every registered adapter fulfills its port and exposes those signals.

## 2. Library-first decisions

| Need | Solution | Rationale |
|---|---|---|
| Debugger | **`dlv dap`** (Delve with the Debug Adapter Protocol) | It is the standard protocol (used by VS Code) and Delve ships with it. More stable than the JSON-RPC v2 API |
| DAP client | Minimal custom implementation (~150 LOC) over `QTcpSocket` | There is no mature DAP client in Python; the framing (`Content-Length` + JSON) is trivial |
| LSP | **gopls** + **`lsprotocol`** types | Official Microsoft types; JSON-RPC transport over stdio with QProcess |
| i18n | **gettext + Babel** (`pybabel extract/init/update/compile`) | Works in `domain`, which has no Qt (the error texts live there). `.po` files are the standard understood by Weblate, POEditor and Crowdin |
| Formatting | `gofmt` (and `goimports` if available) | — |
| Static analysis | `go vet` (+ optional `staticcheck`) | — |
| Tests | `pytest`, `pytest-qt` | — |
| Python lint/format | `ruff` | — |
| Architecture test | `import-linter` | Enforces the dependency rules between layers |
| Packaging | PyInstaller + Inno Setup (Windows), `create-dmg` (macOS), `appimagetool` (Linux) | — |
| CI/CD | GitHub Actions with a matrix of `windows-latest`, `macos-14`, `macos-13`, `ubuntu-22.04` | — |

## 3. EN/ES internationalization (cross-cutting)

### 3.1 Interface
- Every visible text is written **in English as the key** and goes through `_()`: `_("Run")`, `_("Step over")`.
- `i18n/translator.py`:
  - `install_language(code)` loads the `vizcacha` catalog with `gettext.translation(..., fallback=True)`.
  - `detect_language()` takes the `general/language` key from `QSettings` or, if absent, `QLocale.system().name()`: `es*` → `es`, anything else → `en`.
  - Changing the language requires a restart. It is simpler and more robust than retranslating on the fly.
- Plurals with `ngettext`. Dates and numbers with `QLocale`.
- Keyboard shortcuts are not translated.

### 3.2 Compiler and runtime errors
Go **always** emits its errors in English. VizcachaIDE adds a bilingual explanatory layer:

```
raw go output ──► GoOutputParser ──► Diagnostic(file, line, col, raw_text)
                                              │
                                ErrorCatalog (regex → stable id)
                                              │
                    ErrorExplanation(id="E-UNUSED-VAR",
                                     title=_("Variable declared but never used"),
                                     body=_("Go does not allow unused variables… {name}"),
                                     fix_hint=_("Use the variable or delete it, or assign it to _"))
```

- The original, untranslated message is **always** shown too, so it can be searched on the web.
- Stable ids (`E-UNUSED-VAR`, `E-UNUSED-IMPORT`, `E-MISSING-RETURN`, `E-UNDEFINED`, `E-TYPE-MISMATCH`, `E-NO-MAIN`, `P-INDEX-RANGE`, `P-NIL-MAP`, `P-NIL-POINTER`, `P-DEADLOCK`, …). The ids are used to link to the help and for tests.
- Initial catalog: **~25 errors**, the most frequent among beginners.
- When an error is not in the catalog, the original text is shown with a "Search this error" / "Buscar este error" link.

### 3.3 Documentation and release
- `README.md` in English and `README.es.md` in Spanish, with cross-links.
- Release notes in EN and ES.
- Built-in help (`ui/features/help`) in `docs/help/{en,es}/*.md`.

### 3.4 Rule for parallel work
Each track **only wraps its strings in `_()`** and does not touch the `.po` files. Track G (phase 2) runs `pybabel extract/update` and translates everything. This avoids merge conflicts in the catalogs.

## 4. Phases

```
Phase 0 (1 agent) ──► Phase 1 (6 agents in parallel) ──► Phase 2 (3 agents in parallel) ──► Phase 3 (release)
foundations            A B C D E F                         G H I                             QA + publishing
```

### Phase 0: foundations (sequential, blocking)

**Goal**: prepare the ground so that 6 agents can work without stepping on each other.

1. Move `core/` and `gui/` to `vizcacha/` as described in §1, **without changing the visible behavior**.
2. Create `domain/*.py` and `application/ports.py` with the contracts from §1.1.
3. Create `ui/workbench.py`, which exposes:
   - `add_action(menu, action, toolbar=False)`;
   - `add_panel(id, widget, area)`;
   - `add_status_widget(widget)`;
   - `events` (signal bus: `file_opened`, `file_saved`, `run_requested`, `diagnostics_changed`, `navigate_to(SourceLocation)`).

   Features register with `register(workbench, services)`. `main_window.py` is reduced to layout.
4. Create `ui/app.py` as the composition root.
5. Create `i18n/translator.py` + `babel.cfg` and wrap the existing strings in `_()`, without translating them yet.
6. **Remove the simulated debugger** (`_simulate_step`) and leave a `NullDebugger` that warns "Debugger not available yet".
7. Minimal implementation of `GoEnvironment` to unify the logic duplicated between [core/runner.py](../core/runner.py) and `build_code()` in [gui/main_window.py](../gui/main_window.py).
8. Set up `pyproject.toml` (dependencies, ruff, pytest, import-linter), `tests/` with smoke tests and `.github/workflows/ci.yml` with the 3-OS matrix.
9. Remove `Pygments` from the dependencies, because it is unused.

**Done when**: `python -m vizcacha` opens the app and Run works as it does today; CI is green on the 3 OSes; `import-linter` passes; the contracts are documented.

#### Phase 0 result (source of truth for the tracks)

**Done** (branch `fase-0-cimientos`):
- Layered `vizcacha/` package. `core/` and `gui/` removed. `python -m vizcacha` and `python main.py` work.
- Simulated debugger removed: `NullDebugger` warns that it is not available yet.
- Asynchronous `GoToolchain` for run **and build** (build no longer freezes the UI). `GoEnvironment` is a single class and `format_source` uses gofmt.
- Bilingual interface: 113 texts wrapped in `_()`, **Spanish catalog 100% translated** and automatic language detection.
- 43 tests (pytest + pytest-qt, with real integration with Go). `ruff` and 4 `import-linter` contracts are green. CI for 3 OSes is written but has not run yet: it needs network access to GitHub.

**File ownership** (replaces the table below where they differ):

| Track | May edit |
|---|---|
| A · Debugger | `infrastructure/delve_dap/` (new), `application/debug_session.py` (new), `ui/features/debugger/`, `tests/debugger/` |
| B · Assistant | `infrastructure/error_catalog/` (new), `application/explain_error.py` (new), `ui/features/assistant/` (new), **`ui/widgets/console.py`** (clickable links), `examples/errors/`, `tests/assistant/` |
| C · gopls | `infrastructure/gopls_lsp/` (new), `infrastructure/static_completion/`, `ui/features/language/` (new), `tests/language/` |
| D · Editor | `ui/editor/`, `ui/features/editor/` (includes the Editor and Appearance pages), `ui/features/files/`, `tests/editor/` |
| E · Project/toolchain | `infrastructure/go_toolchain/`, `application/run_program.py` (new), `ui/features/run/` (includes the Environment page), `ui/features/project/` (new), `tests/toolchain/` |
| F · Packaging | `packaging/` (new), `.github/workflows/release.yml` (new) |
| Integrator (me) | `domain/`, `application/ports.py`, `application/errors.py`, `application/settings_keys.py`, `ui/workbench.py`, `ui/services.py`, `ui/main_window.py`, `ui/settings_dialog.py`, `i18n/`, `pyproject.toml`, `tests/conftest.py`, `tests/test_contracts.py` |

In `ui/app.py`, each track may only add **its import line and its line in `FEATURES`**. A and C may also change **their service** in `build_services`: A the `debugger` and C the `language_server`, an optional field that already exists in `Services`.

**Extension points**, so that nobody has to edit someone else's code:
- `CodeEditor.set_selection_layer(name, selections)`: independent highlight layers, for example `debug_line`, `diagnostics`, `search` or `brackets`.
- `CodeEditor.completion_provider`: replaced by C with gopls.
- `TabbedEditor.editor_created(CodeEditor)`: to hook in *event filters*, tooltips, Ctrl+click or extra keys.
- `workbench.events`: `navigate_to`, `process_output`, `program_started`, `program_finished`, `diagnostics_changed` and `settings_changed`.
- `workbench.add_action / add_panel / add_status_widget / add_settings_page / add_close_guard`.

### Phase 1: six parallel tracks

| Track | Owns (may edit only this) | Delivers |
|---|---|---|
| **A · Delve debugger** | `infrastructure/delve_dap/`, `application/debug_session.py`, `ui/features/debugger/`, `tests/debugger/` | Real debugging: breakpoints, stepping, resume, run to cursor, variables, stack, goroutines |
| **B · EN/ES Assistant** | `infrastructure/error_catalog/`, `application/explain_error.py`, `ui/features/assistant/`, `tests/assistant/` | Parser + catalog of ~25 errors + Assistant panel + clickable links |
| **C · gopls** | `infrastructure/gopls_lsp/`, `ui/features/language/`, `tests/language/` | Real autocompletion, live diagnostics, hover, go to definition, call-tips, outline |
| **D · Editor** | `ui/editor/`, `ui/features/editor/`, `tests/editor/` | gofmt on save, tabs, find/replace, go to line, braces, block comment, recent files, reload, zoom, applied themes |
| **E · Project and toolchain** | `infrastructure/go_toolchain/`, `application/run_program.py`, `ui/features/run/`, `ui/features/project/`, `tests/toolchain/` | `go.mod` projects, `go run .`, program arguments, async build, "Untitled", Files view, graphical `go mod` |
| **F · Packaging** | `packaging/`, `.github/workflows/release.yml` | Installers for the 3 OSes, with and without bundled Go |

**Shared read-only** files for all tracks: `domain/`, `application/ports.py`, `ui/workbench.py` and `ui/app.py`. If a track needs to change a contract, it **does not edit it**: it describes the change in its PR as a "Contract change request" and the integrator resolves it.

The only exception is `ui/app.py`: each track may add **one line** registering its feature. The merge order resolves the trivial conflict.

**Merge order**: E → A → C → B → D → F. After each merge the full CI runs and the remaining worktrees are rebased.

#### Phase 1 result (integrated into `main`, local only)

The 6 tracks were integrated into `main` in this order: E → D → B → A → F → C (`1a3be66`).
- **Quality:** 334 tests pass, including real integration tests with Go, Delve and gopls. ruff and the 4 import-linter contracts are clean.
- **End-to-end check:** running an example with an error shows the gopls underline, the Outline, the clickable link in the console and the Assistant's explanation.

| Track | Delivered |
|---|---|
| A | Real debugger with `dlv dap`: breakpoints (also live), stepping, continue, run to cursor, lazy variables, stack and goroutines |
| B | Catalog of 25 errors with `examples/errors/<ID>/`, Assistant panel, console links and a draft `docs/i18n/assistant.es.po` |
| C | gopls: completion with a static fallback, wavy diagnostics, hover, Ctrl+click, call-tips, occurrences and Outline |
| D | 4 editor themes and 2 console themes, Go tabs, gofmt on save, find/replace, go to line, comment, zoom, recent files and external reload |
| E | Bundled toolchain → PATH, `go.mod` modules, program arguments, running Untitled, real Stop, Files panel and Go Modules dialog |
| F | PyInstaller lite (69 MB) and full (307 MB with Go 1.25.14, dlv 1.27.2 and gopls 0.21.1) verified on Windows; Inno Setup, dmg, AppImage and `release.yml` |

**Pending after phase 1** (distributed across phase 2):
- **Translation (G):** of 287 texts, 112 are translated, 86 are in track B's draft and 89 are still pending.
- **Proposed and deferred contracts:**
  - `LanguageServerPort.document_highlights` and `document_symbols` with types in `domain`, and `Diagnostic.end` to underline ranges.
  - Asynchronous `DebuggerPort.variables` and the location of goroutines.
  - `QFileOpenEvent` on macOS.
  - Register the Files and Outline panels from startup.
- **UX (H):** the Assistant panel is small, in the bottom-right corner. The default panel layout needs a review, especially in simple mode.
- **Untested for lack of network or OS:** GitHub CI and `release.yml`, dmg on macOS, AppImage on Linux and the Inno Setup installer (`iscc` is not installed).

### Phase 2: integration and second wave (parallel)

| Track | Delivers |
|---|---|
| **G · Translation and i18n** | `pybabel extract/update`, **complete** Spanish translation, English review, language selector in Settings, **first-run wizard** (language, theme, Go/dlv/gopls detection with an "Install" button), `README.es.md`, help in `docs/help/{en,es}` |
| **H · Teaching features** | Frames as nested windows (visible recursion), memory visualizer (pointers, `len`/`cap`, slices sharing an array), simple/regular/expert interface modes |
| **I · Extras** | `go test` panel, ANSI colors and `\r` in the console, experimental REPL with **yaegi** |

### Phase 3: 1.0 release
- Manual QA on Windows, macOS and Linux with `examples/`, plus new examples of typical errors (one for each catalog id).
- Code signing (Authenticode on Windows; notarization on macOS if there is an Apple Developer account).
- GitHub Release with installers, EN/ES notes and screenshots in both languages.

## 5. Agent prompts (ready to copy)

All agents are launched with `isolation: "worktree"` on the branch resulting from phase 0. Each prompt starts with this **common block**:

```text
COMMON CONTEXT
Project: VizcachaIDE, an IDE for Go beginners inspired by Thonny. Python 3.10+ and PyQt5.
Read docs/DEVELOPMENT_PLAN.md (§1 architecture, §1.1 contracts, §3 i18n) before starting.
Rules:
- Edit ONLY the folders owned by your track. domain/, application/ports.py and ui/workbench.py are
  read-only. If you need to change a contract, describe it in your final report as a
  "Contract change request" and do not edit it.
- In ui/app.py you may add a single line to register your feature.
- Every user-visible text goes inside _() / ngettext(), written in English. Do not edit the .po files.
- Files under 200 lines, functions under 50, early return, domain names
  (utils/helpers/common are forbidden), typed exceptions.
- Prefer maintained libraries over custom code.
- Out of scope: MicroPython, TinyGo, microcontrollers.
Done when: your tests in tests/<your_track>/ pass, `ruff check` and `lint-imports` are clean,
the app starts, and your final report includes: files created, how to test it by hand in 3 steps,
known limitations and contract change requests (if any).
```

### Track A · Delve debugger
```text
Implement DelveDapDebugger (infrastructure/delve_dap/), which fulfills DebuggerPort, using `dlv dap`.
- Launch `dlv dap --listen=127.0.0.1:0` with QProcess. Read the port from the output and connect with
  QTcpSocket. DAP framing: Content-Length header + JSON.
- Sequence: initialize → launch (mode "debug", program = folder or file from RunConfiguration,
  args, env from GoToolchainPort.environment()) → setBreakpoints → configurationDone.
- Map the stopped/output/terminated events to Qt signals. On stopped, request threads, stackTrace and
  scopes/variables and emit a domain DebugState.
- Variables with lazy expansion (variablesReference). Long strings truncated.
- Operations: step_over=next, step_into=stepIn, step_out=stepOut, resume=continue,
  run_to = temporary breakpoint + continue.
- UI in ui/features/debugger/: reuse the logic of the variables and stack widgets in ui/,
  add a goroutines view and make a click on a frame emit workbench.events.navigate_to.
  The program's output (stdout/stderr) goes to the existing console.
- dlv path: SettingsRepository "env/delve_path" or, if empty, the PATH. If it does not exist, raise
  DebugAdapterNotFoundError and show a translatable message with installation instructions.
- Tests: DAP framing and parsing with recorded transcripts (without dlv); one integration test marked
  @pytest.mark.requires_dlv that debugs examples/functions.go.
```

### Track B · EN/ES Assistant
```text
Implement error explanation for beginners.
- infrastructure/error_catalog/go_output_parser.py: converts the raw output of go build/run/vet and
  panics (including the goroutine trace) into list[Diagnostic] with file/line/col and raw_text.
  Supports relative and absolute paths, Windows and POSIX.
- infrastructure/error_catalog/catalog.py + entries in per-category modules (compile_errors.py,
  runtime_panics.py, vet_warnings.py): regex → ErrorExplanation with a stable id
  (E-UNUSED-VAR, E-UNUSED-IMPORT, E-MISSING-RETURN, E-UNDEFINED, E-TYPE-MISMATCH, E-NO-MAIN,
  E-PACKAGE-NOT-MAIN, E-SYNTAX-UNEXPECTED, E-MISSING-BRACE, E-ASSIGN-MISMATCH, E-NOT-ENOUGH-ARGS,
  E-TOO-MANY-ARGS, E-NO-NEW-VARS, E-UNEXPORTED, E-IMPORT-NOT-FOUND, P-INDEX-RANGE, P-NIL-MAP,
  P-NIL-POINTER, P-DEADLOCK, P-DIVIDE-ZERO, P-SLICE-BOUNDS, P-TYPE-ASSERTION, V-PRINTF-ARGS,
  V-UNREACHABLE, V-SHADOW…). title/body/fix_hint are written in English with _() and support
  named placeholders ({name}, {type}).
- application/explain_error.py: use case that receives the output or the gopls diagnostics and
  returns (Diagnostic, ErrorExplanation | None) pairs.
- ui/features/assistant/: panel that opens by itself when there is an error. It shows the title,
  explanation, suggestion and the ORIGINAL UNTRANSLATED TEXT, with a "Search this error" button. It also
  turns the `file.go:12:5` references in the console into links that emit workbench.events.navigate_to.
- Create examples/errors/<id>.go with a minimal program that triggers each error.
- Tests: each example in examples/errors produces exactly the expected id (output recorded in
  fixtures, without requiring go); parser with Windows and POSIX paths.
```

### Track C · gopls
```text
Implement GoplsLanguageServer (infrastructure/gopls_lsp/), which fulfills LanguageServerPort.
- QProcess `gopls serve` over stdio. JSON-RPC with Content-Length. Use the `lsprotocol` types
  and cattrs for (de)serialization.
- initialize with the project's rootUri (or the file's folder), didOpen/didChange (with a 300 ms
  debounce), completion, hover, definition, signatureHelp, documentHighlight, documentSymbol, and the
  publishDiagnostics notification → diagnostics_published signal.
- ui/features/language/: replaces the static autocompletion (keep GoAnalyzer as a fallback
  if gopls is missing). Wavy underline for diagnostics, hover tooltip, Ctrl+click to go to
  definition, call-tip when typing "(", occurrence highlighting, Outline panel.
- Diagnostics are also published on workbench.events.diagnostics_changed so that the
  Assistant (track B) can consume them.
- If gopls does not exist: silent degradation to the fallback and a single translatable notice in the status bar.
- Tests: framing and message mapping with recorded transcripts; integration with
  @pytest.mark.requires_gopls.
```

### Track D · Editor
```text
Improve the editor (ui/editor/ and ui/features/editor/).
- Apply ALL the options that are currently saved and ignored: editor_theme (Light, Dark, Solarized
  Light/Dark: define the themes as data, with syntax and editor colors), console_theme,
  tab_size, auto_indent, show_line_numbers, show_status_bar, background_color/text_color.
  Keep the existing QSettings keys.
- Indentation with a real TAB (idiomatic in Go) and visual width = tab_size. Remove the "indent after :".
- gofmt on save, through GoToolchainPort.format_source. Preserve the cursor position.
  If it fails, it does not block saving and leaves the error in the status bar.
- Non-modal Find/Replace (Ctrl+F / Ctrl+H, F3), Go to line (Ctrl+G), brace matching,
  comment/uncomment (Ctrl+/), indent or dedent the selection, zoom (Ctrl+= / Ctrl+-).
- Recent files (maximum 10) in the File menu. QFileSystemWatcher that reloads or asks if the
  file changed outside the IDE.
- Response to workbench.events.navigate_to: open the file and move the cursor.
- pytest-qt tests for find/replace, commenting, matching and theme application.
```

### Track E · Project and toolchain
```text
Implement GoToolchainPort (infrastructure/go_toolchain/) and the project experience.
- GoEnvironment: the single place that resolves go_path, GOPATH, GOROOT and the extra variables
  (current env/* keys) and locates go, dlv and gopls (configured → bundle next to the executable → PATH).
- GoProcessRunner with QProcess for run and build, both ASYNCHRONOUS (today build blocks the UI). Stop
  first sends an interrupt and kills the process after 2 s.
- RunConfiguration: single file (`go run file.go`) or module (folder with go.mod → `go run .`),
  with program arguments ("Program arguments" field in the toolbar).
- Run "Untitled" tabs by writing them to a temporary directory.
- format_source (gofmt via stdin) and vet (go vet -json → Diagnostic).
- ui/features/run/: Run, Stop and Build actions with shortcuts F5, Shift+F5 and Ctrl+B.
- ui/features/project/: "Open folder…", Files view (filtered QFileSystemModel), modules dialog:
  go mod init, go get <pkg>[@version], go mod tidy, with output in the console.
- Tests: command and environment construction with a mocked GoEnvironment; integration
  @pytest.mark.requires_go that runs examples/hello.go.
```

### Track F · Packaging
```text
Create the cross-platform distribution (packaging/ and .github/workflows/release.yml).
- packaging/fetch_toolchain.py: downloads pinned versions (file packaging/versions.toml) of Go,
  and builds or downloads dlv and gopls for the target OS and architecture, verifying the sha256.
- PyInstaller (packaging/vizcacha.spec): includes vizcacha/i18n/locale/*.mo, logo and examples.
  "full" variant (with the toolchain in ./toolchain) and "lite" (uses the system Go).
- Windows: Inno Setup (per-user installer, shortcuts, optional .go association) + portable
  zip. macOS: .app + create-dmg, x86_64 and arm64 builds (verify PyQt5 wheels on arm64;
  if there are none, x86_64 only via Rosetta and document it). Linux: AppImage.
- The installer offers EN/ES language (Inno Setup supports it natively).
- release.yml: when a v* tag is pushed, builds on the OS matrix, uploads artifacts and creates a
  draft GitHub Release.
- Document the license notice in packaging/README.md: PyQt5 is GPL, so the distributed binary
  falls under GPLv3. It must appear in the installer and in "About".
```

## 6. Integration (you or an integrator agent)

After each phase 1 merge:
1. `git rebase` the pending worktrees onto `main`.
2. Full CI (3 OSes) + `lint-imports`.
3. Resolve the *contract change requests* (only the integrator edits `ports.py` and `domain/`).
4. Manual smoke test: open `examples/functions.go` → Run → Debug with a breakpoint → trigger an error from `examples/errors/` → see the Assistant.

## 7. Coverage matrix vs Thonny

| P0/P1 gap (THONNY_COMPARISON.md) | Track |
|---|---|
| Real debugger, breakpoints, variables, stack | A |
| Goroutines | A |
| Assistant and clickable errors (EN/ES) | B |
| Smart autocompletion, live diagnostics, call-tips, go to definition | C |
| gofmt, find/replace, applied themes | D |
| go.mod projects, program arguments, async build, Files view, module manager | E |
| Installer with bundled Go | F |
| EN/ES interface, first run | Foundation in phase 0, translation in G |
| Nested frames, memory visualizer, interface modes | H |
| REPL, go test | I |

## 8. Risks and mitigations

| Risk | Mitigation |
|---|---|
| **PyQt5 is GPL** and the repository is MIT: the distributed binary falls under the GPL | Document it in the README, the installer and "About". The isolated `ui/` layer allows migrating to PySide6 (LGPL) later with bounded changes |
| PyQt5 without arm64 wheels for macOS | Verify it in CI in track F; if there are none, build only x86_64 (Rosetta 2) |
| Heavy "full" installer (~250–300 MB with Go) | Also offer the "lite" variant |
| Protocol changes in dlv/gopls | Versions pinned in `packaging/versions.toml` + tests with recorded transcripts |
| Conflicts between agents | Disjoint folders, frozen contracts, `.po` only in track G, fixed merge order |
| Poor-quality translations | Human review of the Spanish `.po` before the release; glossary of terms (breakpoint = "punto de interrupción", slice = "slice") in `docs/i18n-glosario.md` |
