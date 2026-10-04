# QA — VizcachaIDE Wails 2.0.0-rc1

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
