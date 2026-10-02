# VizcachaIDE Wails: coding plan

> A variant of VizcachaIDE written **in Go with Wails v2** and a **Svelte 5 + CodeMirror 6** frontend. It replaces the PyQt interface, which is considered unappealing, without losing anything 1.0 already does.
>
> - **Reference design:** interactive prototype (claude.ai) and [UX_COPY.md](UX_COPY.md).
> - **Orchestration:** Opus coordinates and integrates. The coders are **Sonnet** agents, each in its own worktree.
> - **Quality:** the *software-architecture* skill (Clean Architecture + DDD, library-first; files under 200 lines and functions under 50).

## 0. Decisions made

| Topic | Decision |
|---|---|
| Backend | **Full rewrite in Go.** A single binary, no Python or PyQt. The package is no longer GPL because of PyQt; what remains is Go (BSD), Delve (MIT), gopls (BSD) and Wails (MIT). |
| Frontend | **Svelte 5 + TypeScript + Vite + CodeMirror 6.** |
| Location | The **`wails/`** folder in this same repository. |
| PyQt version | Stays as **1.x "classic"**, in maintenance (bugs and translations only) until Wails matches it. Then it is archived. |
| Languages | English and Spanish from the first commit, with the same catalogs as 1.0, converted. |
| Out of scope | MicroPython, TinyGo and microcontrollers. |

## 1. What is reused from 1.0

1.0 leaves contracts, data and tests that are just as valid in Go:

| From 1.0 (Python) | In the Wails variant |
|---|---|
| `vizcacha/domain/*` (dataclasses) | Same concepts as structs in `internal/domain` (identical ubiquitous language) |
| `application/ports.py` | Go interfaces in `internal/app/ports.go` |
| Assistant catalog (25 ids, regex, EN/ES texts) | Embedded JSON data (`//go:embed`), same ids |
| `tests/assistant/fixtures/go_output/*.txt` | Fixtures for the Go parser tests |
| `tests/debugger/transcripts/*.json` (real dlv) | Fixtures for the DAP client tests |
| `tests/language/transcripts/*` (real gopls) | Fixtures for the LSP client tests |
| `vizcacha/i18n/locale/es/*.po` (303 strings) | Converted to JSON with a script (`wails/tools/po2json`) |
| `packaging/fetch_toolchain.py` + `versions.toml` | The "full" variant includes the same toolchain (Go 1.25.14, dlv 1.27.2, gopls 0.21.1) |
| `examples/` | Reused as is |

## 2. Architecture

```
wails/
├── main.go                    # composition root: creates adapters, services and wails.Run
├── wails.json
├── internal/
│   ├── domain/                # structs and pure rules. No project imports
│   │   ├── diagnostics.go     # SourceLocation, SourceRange, Diagnostic, Severity
│   │   ├── debugging.go       # Breakpoint, Variable, StackFrame, Goroutine, DebugState
│   │   ├── explanation.go     # ErrorExplanation (stable id + placeholders)
│   │   ├── completion.go      # CompletionItem, SignatureHelp
│   │   ├── code_structure.go  # DocumentSymbol, SymbolKind
│   │   └── project.go         # GoModule, RunConfiguration
│   ├── app/                   # use cases + ports (interfaces). Only imports domain
│   │   ├── ports.go           # Toolchain, Debugger, LanguageServer, ErrorExplainer, SettingsStore, EventSink
│   │   ├── run_program.go     # configuration from the active file (go.mod, arguments)
│   │   ├── debug_session.go
│   │   └── explain_error.go
│   ├── adapters/              # implementations of the ports
│   │   ├── toolchain/         # locator (config → bundled → PATH), runner (os/exec), modules, formatter (go/format)
│   │   ├── delve/             # DAP client with github.com/google/go-dap over TCP
│   │   ├── gopls/             # LSP client with go.lsp.dev/jsonrpc2 + go.lsp.dev/protocol over stdio
│   │   ├── errorcatalog/      # go output parser + embedded JSON catalog
│   │   └── settings/          # JSON in os.UserConfigDir()/VizcachaIDE/settings.json
│   ├── i18n/                  # go-i18n v2 + embedded locales/{en,es}.json (backend texts)
│   └── bridge/                # services exposed to Wails (thin) + event names
│       ├── events.go          # event constants (contract with the frontend)
│       ├── run_service.go  debug_service.go  language_service.go
│       ├── assistant_service.go  files_service.go  settings_service.go
└── frontend/                  # Svelte 5 + TS + Vite
    ├── src/lib/bridge/        # typed wailsjs wrappers + mock for development without Go
    ├── src/lib/events.ts      # mirror of bridge/events.go
    ├── src/lib/stores/        # run, debug, diagnostics, files, settings, layout (one store per domain)
    ├── src/lib/editor/        # CodeMirror: lang-go, breakpoint gutter, debug line, lint, completion, tooltips
    ├── src/lib/panels/        # Files, Outline, Output, Problems, Assistant, Variables, CallStack
    ├── src/lib/shell/         # TitleBar, Rail, StatusBar, DebugToolbar, Layout, SettingsDialog, FirstRun
    ├── src/lib/theme/         # CSS tokens from the prototype (light/dark) + Atkinson Hyperlegible (@fontsource, no network)
    └── src/lib/i18n/          # svelte-i18n + locales/{en,es}.json (interface texts, ICU)
```

### 2.1 Dependency rules (checked in CI)

- `domain` imports nothing from the project. `app` only imports `domain`.
- `adapters/*` import `app` and `domain`, never `bridge` or another adapter.
- `bridge` is the only package that imports `github.com/wailsapp/wails/v2/pkg/runtime`.
- `main.go` is the only place that instantiates adapters (composition root).
- Checked with **golangci-lint + depguard**, the equivalent of import-linter in 1.0.
- **Frontend:** components do not call `wailsjs` directly; they go through `src/lib/bridge`. Logic lives in the stores, not in the components. ESLint checks this with `no-restricted-imports`.

### 2.2 Event contract (backend → frontend)

| Event | Payload |
|---|---|
| `run:output` | `{ stream: "stdout" \| "stderr", text }` |
| `run:started` / `run:finished` | `RunConfiguration` / `{ exitCode, durationMs }` |
| `debug:stopped` | `DebugState` |
| `debug:variables` | `{ reference, variables }` (asynchronous expansion) |
| `debug:output` / `debug:terminated` | `{ text, category }` / `{ exitCode }`. `-1` means stopped by the user |
| `lsp:diagnostics` | `{ path, diagnostics }` |
| `lsp:status` | `"starting" \| "ready" \| "unavailable"` |
| `assistant:explained` | `[{ diagnostic, explanation }]` (texts already translated) |
| `settings:changed` | `Settings` |

Frontend → backend calls are methods of the `bridge` services. Wails generates their TypeScript types, so the contract is typed on both sides.

### 2.3 *Library-first* decisions

| Need | Library | Instead of… |
|---|---|---|
| Window and Go↔JS bridge | Wails v2.16 | — |
| DAP (Delve) | `github.com/google/go-dap` | 1.0's own framing |
| LSP (gopls) | `go.lsp.dev/jsonrpc2` + `go.lsp.dev/protocol` | lsprotocol + custom JSON-RPC |
| gofmt | standard `go/format` package | running the gofmt binary |
| Backend i18n | `github.com/nicksnyder/go-i18n/v2` | gettext |
| Editor | CodeMirror 6 (`@codemirror/lang-go`, `lint`, `autocomplete`, `search`) | — |
| Frontend i18n | `svelte-i18n` (ICU MessageFormat: plurals and placeholders) | — |
| Resizable panels | `paneforge` (Svelte 5) | custom splitters |
| Offline fonts | `@fontsource/atkinson-hyperlegible-next` and `-mono` | Google Fonts |
| Tests | `go test` (+ `testify` if needed), `vitest` + `@testing-library/svelte`, `svelte-check` | — |
| Windows installer | NSIS built into `wails build -nsis` | Inno Setup |

### 2.4 Code rules for the coders

- **Go:**
  - Errors wrapped with `%w` and sentinel errors (`errors.Is`).
  - `context.Context` on everything that launches processes.
  - No `panic` outside `main`.
  - Files under 200 lines, functions under 50, maximum nesting of 3, early return.
- **TypeScript/Svelte:**
  - Arrow functions.
  - Components under 200 lines.
  - No business logic in `.svelte` files.
  - Strict types (`strict: true`).
- **Domain names:** `utils`, `helpers`, `common` and `misc` are forbidden.
- **Visible texts:** always through i18n, using the text from [UX_COPY.md](UX_COPY.md) verbatim.

## 3. Phases and tracks

```
W0 (1 Sonnet coder, blocking) ──► W1 (6 Sonnet coders in parallel) ──► W2 (Opus integration + 2 coders) ──► W3 release
```

### W0 · Foundations (sequential)

1. `wails init -t svelte-ts` in `wails/`, with Svelte 5 and strict TypeScript.
2. Folder structure from §2:
   - complete `domain`;
   - frozen `app/ports.go`;
   - `bridge/events.go` and `src/lib/events.ts`;
   - `bridge` services with stubs that return sample data.
3. Frontend **mock bridge**: `npm run dev` without Go shows the 3 prototype states. This way the frontend coders do not depend on the backend.
4. **i18n:**
   - `tools/po2json` converts the 1.0 `.po` files into `internal/i18n/locales/*.json` and `frontend/src/lib/i18n/locales/*.json`.
   - New texts come from UX_COPY.md.
5. **Theme:**
   - CSS tokens from the prototype, in light and dark.
   - Atkinson Hyperlegible via @fontsource.
   - Shell with paneforge: title bar, rail, side panel, editor, Assistant, Output and status bar.
6. **Quality:**
   - golangci-lint with depguard (rules from §2.1);
   - ESLint + Prettier + svelte-check;
   - vitest;
   - workflow `.github/workflows/wails.yml`: go test, lint, frontend and `wails build` on Windows, macOS and Linux.

**Done when:** `wails build` produces an executable that opens the shell with the new design, switches language and passes CI.

### W1 · Six coders in parallel

| Track | Owns | Delivers |
|---|---|---|
| **G1 · Toolchain and execution** | `adapters/toolchain`, `app/run_program.go`, `bridge/run_service.go`, `bridge/files_service.go` | Tools (config → bundled → PATH), run/stop/build, Untitled, arguments, `go.mod`, modules, `go/format`, file tree, "first time" notice |
| **G2 · Debugger** | `adapters/delve`, `app/debug_session.go`, `bridge/debug_service.go` | Delve DAP with go-dap: live breakpoints, stepping, continue, run to cursor, asynchronous variables, goroutines with location, and binary cleanup |
| **G3 · gopls** | `adapters/gopls`, `bridge/language_service.go` | Diagnostics with range, completion, hover, signature, definition, occurrences, symbols. Graceful degradation without gopls |
| **G4 · Assistant** | `adapters/errorcatalog`, `app/explain_error.go`, `bridge/assistant_service.go`, `internal/i18n` | Parser for go output and panics, 25 ids with EN/ES texts, IDE error messages (UX_COPY) |
| **F1 · Editor** | `frontend/src/lib/editor`, `stores/diagnostics` | CodeMirror: lang-go, breakpoint gutter, sand-colored debug line, wavy lint underline, inline explanation, completion, hover, find/replace, themes |
| **F2 · Panels and shell** | `frontend/src/lib/panels`, `shell`, `stores/{run,debug,files,settings,layout}` | Files, Outline, Output (links) and Problems, Assistant (tips, error, variables), variable cards with "just changed", "How you got here", debug bar, status bar, Settings, first-run wizard |

**Shared contracts** (`domain`, `ports.go`, `events.go`, `events.ts`, the i18n JSON files): read-only in W1. Changes are requested from the orchestrator as a *Contract change request*.

New i18n keys are proposed in the final report; the orchestrator merges them in W2.

### W2 · Integration, packaging and QA

- **Orchestrator (Opus):** merges in the order G1 → G4 → G2 → G3 → F1 → F2, CCRs, translations and end-to-end tests.
- **Coder P (Sonnet), packaging:**
  - `wails build -nsis` (Windows), `.app` + `.dmg` (macOS) and AppImage (Linux);
  - full and lite variants reusing `packaging/fetch_toolchain.py`;
  - signing pending a certificate.
- **Coder Q (Sonnet), QA:**
  - port of `packaging/qa` to the Wails app;
  - EN/ES screenshots;
  - parity checklist with 1.0 (every feature in the [THONNY_COMPARISON](../THONNY_COMPARISON.md) table that 1.0 already had).

### W3 · Release

`v2.0.0-rc1` of the Wails variant. The PyQt 1.x remains available as "classic" until the parity checklist is complete.

## 4. Common prompt for the Sonnet coders

```text
You are a coder on the VizcachaIDE Wails project (Go + Wails v2 + Svelte 5 + CodeMirror 6), a Go IDE
for beginners, bilingual EN/ES. You work in an isolated git worktree. First read
docs/wails/PLAN_WAILS.md (§2 architecture, §2.1 rules, §2.2 events, §2.4 code) and docs/wails/UX_COPY.md.
- Edit ONLY your track's folders. The contracts (domain, app/ports.go, bridge/events.go,
  frontend/src/lib/events.ts, locales) are read-only: request changes as a
  "Contract change request" in your report.
- Visible texts: through i18n, using the exact text from UX_COPY.md. New keys go in your report with EN and ES.
- Library-first. Files under 200 lines, functions under 50, early return, domain names.
- Out of scope: MicroPython and TinyGo.
- Verification: go test ./..., golangci-lint run, npm run check, npm run test and wails build (if your track affects it).
- When done: ONE local commit in English ending with "Co-Authored-By: Claude Sonnet <noreply@anthropic.com>",
  no push. Report: commit hash, files, how to test it, limitations, CCRs and new i18n keys.
```

## 5. Risks

| Risk | Mitigation |
|---|---|
| Rewriting everything is big | 1.0's contracts, data and fixtures are reused. 1.x stays alive until matched. |
| CodeMirror and LSP integration | We only use official extensions (lint, autocomplete, tooltips). The backend translates LSP into simple domain types. |
| WebView2 missing on some Windows 10 machines | The Wails NSIS installer includes the WebView2 bootstrapper. |
| `macos-13` has no runners on GitHub | CI uses `macos-14` and `macos-15-intel`. |
| Two versions in parallel | 1.x only gets bug fixes and translations. Everything new goes to Wails. |
