# QA — VizcachaIDE Wails 2.0.0-rc1

## Resumen (español)

QA de paridad con la 1.0, hecho el **2026-10-01 en Windows 10** con el **backend real**: `wails dev`, Go 1.25.5, Delve 1.27.2 y gopls 0.21.1, recorrido por Edge sin pantalla con CDP.

- **Funciona igual o mejor que la 1.0:** ejecutar, stdin, argumentos, Stop, errores explicados (compilación y panics), subrayado de gopls, completado, hover, ir a la definición, gofmt al guardar, cerrar con cambios, depurador real, primer arranque, idioma automático y diálogo de módulos.
- **Corregido durante el QA:**
  - Faltaba la entrada de teclado (stdin) en la Salida.
  - Los argumentos del programa estaban escondidos en Configuración; ahora están en la barra de título.
  - El Asistente no explicaba los diagnósticos en vivo de gopls.
  - El mensaje de un panic llegaba partido en dos líneas.
  - Faltaba "Ejecutar hasta aquí".
  - Había un fallo en los mensajes a gopls.
  - El Asistente mostraba tarjetas de otros archivos; ahora muestra solo el archivo activo y la última ejecución.
- **Corregido después del QA:** los ejemplos ahora viven cada uno en su carpeta, así que gopls ya no marca "main redeclared". Ver §3 para lo pendiente.

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
| Modules (`go.mod`, `go run .`) | ⏳ | backend integration tests pass; E2E step timed out |
| Go Modules dialog | ✅ | E2E `modules-dialog` (ES) |
| Assistant: compile errors explained, EN/ES | ✅ | E2E `error-flow`: "Variable "count" is never used" / "La variable "count" nunca se usa" |
| Assistant: runtime panics | ✅ (panic line fixed) | E2E `panic`: P-NIL-MAP card |
| Clickable `file.go:L:C` in Output | ✅ | E2E: link → line 5, column 2 |
| Live diagnostics (gopls) with exact range | ✅+ | E2E: underline exactly on `count`; inline explanation |
| Completion, hover, go to definition | ✅ | E2E `editor-intel` |
| gofmt on save | ✅ | E2E `format-save` |
| Close tab with unsaved changes | ✅ | E2E dialog "Save changes to main.go before closing?" |
| Find/replace, go to line, comment, zoom, Outline | ⏳ | unit tests pass; E2E step timed out |
| Debugger: breakpoints, steps, variables, call stack | ✅+ | E2E screenshots: paused, "Next line", "just changed", "How you got here", Goroutines |
| Run to cursor | ✅ (added in QA) | unit test + toolbar |
| Themes, font size, language | ⚠️ | auto language ✅ (backend resolves `es`); persistence after restart not confirmed by E2E (selector step failed) |
| First-run wizard | ✅+ (new) | E2E `firstrun`: language → "Go 1.25.5 está listo" → hello ran |
| Bundled Go toolchain | ✅ | `wails/packaging` install QA (full installer) |
| Installers (Windows) | ✅+ | 57.5 MB full / 6.4 MB lite (1.0: 75 / 25 MB), no GPL |
| 1024 px window | ⏳ | E2E step timed out |

## 3. Open issues

| # | Severity | Issue | Next step |
|---|---|---|---|
| Q1 | ~~High~~ **Fixed** | Top-level examples shared one folder (`package main` ×6), so gopls reported "main redeclared". | Fixed: each example now lives in `examples/<name>/<name>.go`; `gopls check` reports nothing. |
| Q2 | Medium | Settings persistence after restart not confirmed end-to-end (the settings JSON store has unit tests). | Repeat `settings-steps.mjs` once the menu selector is fixed. |
| Q3 | Low | E2E steps `modules`, `editor-extras` and `layout-1024` hit 30 s timeouts in the harness. | Re-run with longer waits; features have unit/integration tests. |
| Q4 | Low | "Replace all" does not ask for confirmation (CodeMirror's search panel replaces directly). | Custom search panel if wanted. |

## 4. How to repeat

```
cd wails && wails dev                        # backend + http://localhost:34115
cd wails/packaging/qa/e2e && npm ci
node qa.mjs --lang en                        # and --lang es; results in %TEMP%\wq\results.json
```

macOS and Linux were not executed (no machines here). The CI workflows cover build and unit tests on the 3 OS.
