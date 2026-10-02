# VizcachaIDE Wails — plan de codificación

> Variante de VizcachaIDE escrita **en Go con Wails v2** y un frontend **Svelte 5 + CodeMirror 6**. Reemplaza la interfaz PyQt, que se considera poco agradable, sin perder nada de lo que ya hace la 1.0.
>
> - **Diseño de referencia:** prototipo interactivo (claude.ai) y [UX_COPY.md](UX_COPY.md).
> - **Orquestación:** Opus coordina e integra. Los coders son agentes **Sonnet**, cada uno en su propio worktree.
> - **Calidad:** skill *software-architecture* (Clean Architecture + DDD, library-first; archivos de menos de 200 líneas y funciones de menos de 50).

## 0. Decisiones tomadas

| Tema | Decisión |
|---|---|
| Backend | **Reescritura completa en Go.** Un solo binario, sin Python ni PyQt. El paquete deja de ser GPL por PyQt; quedan Go (BSD), Delve (MIT), gopls (BSD) y Wails (MIT). |
| Frontend | **Svelte 5 + TypeScript + Vite + CodeMirror 6.** |
| Ubicación | Carpeta **`wails/`** en este mismo repositorio. |
| Versión PyQt | Queda como **1.x "clásica"**, en mantenimiento (solo bugs y traducciones) hasta que Wails la iguale. Después se archiva. |
| Idiomas | Inglés y español desde el primer commit, con los mismos catálogos que la 1.0, convertidos. |
| Fuera de alcance | MicroPython, TinyGo y microcontroladores. |

## 1. Qué se reaprovecha de la 1.0

La 1.0 deja contratos, datos y pruebas que valen igual en Go:

| De la 1.0 (Python) | En la variante Wails |
|---|---|
| `vizcacha/domain/*` (dataclasses) | Mismos conceptos como structs en `internal/domain` (lenguaje ubicuo idéntico) |
| `application/ports.py` | Interfaces Go en `internal/app/ports.go` |
| Catálogo del Assistant (25 ids, regex, textos EN/ES) | Datos JSON embebidos (`//go:embed`), mismos ids |
| `tests/assistant/fixtures/go_output/*.txt` | Fixtures de los tests Go del parser |
| `tests/debugger/transcripts/*.json` (dlv real) | Fixtures de los tests del cliente DAP |
| `tests/language/transcripts/*` (gopls real) | Fixtures de los tests del cliente LSP |
| `vizcacha/i18n/locale/es/*.po` (303 textos) | Convertidos a JSON con un script (`wails/tools/po2json`) |
| `packaging/fetch_toolchain.py` + `versions.toml` | La variante "full" incluye la misma toolchain (Go 1.25.14, dlv 1.27.2, gopls 0.21.1) |
| `examples/` | Se reutilizan tal cual |

## 2. Arquitectura

```
wails/
├── main.go                    # composition root: crea adaptadores, servicios y wails.Run
├── wails.json
├── internal/
│   ├── domain/                # structs y reglas puras. Sin imports del proyecto
│   │   ├── diagnostics.go     # SourceLocation, SourceRange, Diagnostic, Severity
│   │   ├── debugging.go       # Breakpoint, Variable, StackFrame, Goroutine, DebugState
│   │   ├── explanation.go     # ErrorExplanation (id estable + placeholders)
│   │   ├── completion.go      # CompletionItem, SignatureHelp
│   │   ├── code_structure.go  # DocumentSymbol, SymbolKind
│   │   └── project.go         # GoModule, RunConfiguration
│   ├── app/                   # casos de uso + puertos (interfaces). Sólo importa domain
│   │   ├── ports.go           # Toolchain, Debugger, LanguageServer, ErrorExplainer, SettingsStore, EventSink
│   │   ├── run_program.go     # configuración desde el archivo activo (go.mod, argumentos)
│   │   ├── debug_session.go
│   │   └── explain_error.go
│   ├── adapters/              # implementaciones de los puertos
│   │   ├── toolchain/         # locator (config → incluida → PATH), runner (os/exec), modules, formatter (go/format)
│   │   ├── delve/             # cliente DAP con github.com/google/go-dap sobre TCP
│   │   ├── gopls/             # cliente LSP con go.lsp.dev/jsonrpc2 + go.lsp.dev/protocol por stdio
│   │   ├── errorcatalog/      # parser de salida de go + catálogo JSON embebido
│   │   └── settings/          # JSON en os.UserConfigDir()/VizcachaIDE/settings.json
│   ├── i18n/                  # go-i18n v2 + locales/{en,es}.json embebidos (textos del backend)
│   └── bridge/                # servicios expuestos a Wails (finos) + nombres de eventos
│       ├── events.go          # constantes de eventos (contrato con el frontend)
│       ├── run_service.go  debug_service.go  language_service.go
│       ├── assistant_service.go  files_service.go  settings_service.go
└── frontend/                  # Svelte 5 + TS + Vite
    ├── src/lib/bridge/        # wrappers tipados de wailsjs + mock para desarrollo sin Go
    ├── src/lib/events.ts      # espejo de bridge/events.go
    ├── src/lib/stores/        # run, debug, diagnostics, files, settings, layout (un store por dominio)
    ├── src/lib/editor/        # CodeMirror: lang-go, gutter de breakpoints, línea de depuración, lint, completado, tooltips
    ├── src/lib/panels/        # Files, Outline, Output, Problems, Assistant, Variables, CallStack
    ├── src/lib/shell/         # TitleBar, Rail, StatusBar, DebugToolbar, Layout, SettingsDialog, FirstRun
    ├── src/lib/theme/         # tokens CSS del prototipo (claro/oscuro) + Atkinson Hyperlegible (@fontsource, sin red)
    └── src/lib/i18n/          # svelte-i18n + locales/{en,es}.json (textos de la interfaz, ICU)
```

### 2.1 Reglas de dependencia (se verifican en CI)

- `domain` no importa nada del proyecto. `app` sólo importa `domain`.
- `adapters/*` importan `app` y `domain`, nunca `bridge` ni otro adaptador.
- `bridge` es el único paquete que importa `github.com/wailsapp/wails/v2/pkg/runtime`.
- `main.go` es el único sitio que instancia adaptadores (composition root).
- Se verifica con **golangci-lint + depguard**, el equivalente de import-linter en la 1.0.
- **Frontend:** los componentes no llaman a `wailsjs` directamente; pasan por `src/lib/bridge`. La lógica vive en los stores, no en los componentes. ESLint lo comprueba con `no-restricted-imports`.

### 2.2 Contrato de eventos (backend → frontend)

| Evento | Payload |
|---|---|
| `run:output` | `{ stream: "stdout" \| "stderr", text }` |
| `run:started` / `run:finished` | `RunConfiguration` / `{ exitCode, durationMs }` |
| `debug:stopped` | `DebugState` |
| `debug:variables` | `{ reference, variables }` (expansión asíncrona) |
| `debug:output` / `debug:terminated` | `{ text, category }` / `{ exitCode }`. `-1` significa detenido por el usuario |
| `lsp:diagnostics` | `{ path, diagnostics }` |
| `lsp:status` | `"starting" \| "ready" \| "unavailable"` |
| `assistant:explained` | `[{ diagnostic, explanation }]` (textos ya traducidos) |
| `settings:changed` | `Settings` |

Las llamadas frontend → backend son métodos de los servicios de `bridge`. Wails genera sus tipos TypeScript, así que el contrato queda tipado en los dos lados.

### 2.3 Decisiones *library-first*

| Necesidad | Librería | En lugar de… |
|---|---|---|
| Ventana y puente Go↔JS | Wails v2.16 | — |
| DAP (Delve) | `github.com/google/go-dap` | el framing propio de la 1.0 |
| LSP (gopls) | `go.lsp.dev/jsonrpc2` + `go.lsp.dev/protocol` | lsprotocol + JSON-RPC propio |
| gofmt | paquete estándar `go/format` | ejecutar el binario gofmt |
| i18n del backend | `github.com/nicksnyder/go-i18n/v2` | gettext |
| Editor | CodeMirror 6 (`@codemirror/lang-go`, `lint`, `autocomplete`, `search`) | — |
| i18n del frontend | `svelte-i18n` (ICU MessageFormat: plurales y placeholders) | — |
| Paneles redimensionables | `paneforge` (Svelte 5) | splitters propios |
| Fuentes sin red | `@fontsource/atkinson-hyperlegible-next` y `-mono` | Google Fonts |
| Tests | `go test` (+ `testify` si hace falta), `vitest` + `@testing-library/svelte`, `svelte-check` | — |
| Instalador Windows | NSIS integrado en `wails build -nsis` | Inno Setup |

### 2.4 Reglas de código para los coders

- **Go:**
  - Errores con `%w` y errores centinela (`errors.Is`).
  - `context.Context` en todo lo que lanza procesos.
  - Nada de `panic` fuera de `main`.
  - Archivos de menos de 200 líneas, funciones de menos de 50, anidamiento máximo de 3, early return.
- **TypeScript/Svelte:**
  - Funciones flecha.
  - Componentes de menos de 200 líneas.
  - Sin lógica de negocio en `.svelte`.
  - Tipos estrictos (`strict: true`).
- **Nombres de dominio:** prohibidos `utils`, `helpers`, `common` y `misc`.
- **Textos visibles:** siempre por i18n, con el texto de [UX_COPY.md](UX_COPY.md) tal cual.

## 3. Fases y tracks

```
W0 (1 coder Sonnet, bloqueante) ──► W1 (6 coders Sonnet en paralelo) ──► W2 (integración Opus + 2 coders) ──► W3 release
```

### W0 · Cimientos (secuencial)

1. `wails init -t svelte-ts` en `wails/`, con Svelte 5 y TypeScript estricto.
2. Estructura de carpetas de §2:
   - `domain` completo;
   - `app/ports.go` congelado;
   - `bridge/events.go` y `src/lib/events.ts`;
   - servicios de `bridge` con stubs que devuelven datos de ejemplo.
3. **Mock bridge** del frontend: `npm run dev` sin Go muestra los 3 estados del prototipo. Así los coders del frontend no dependen del backend.
4. **i18n:**
   - `tools/po2json` convierte los `.po` de la 1.0 en `internal/i18n/locales/*.json` y `frontend/src/lib/i18n/locales/*.json`.
   - Los textos nuevos salen de UX_COPY.md.
5. **Tema:**
   - Tokens CSS del prototipo, en claro y oscuro.
   - Atkinson Hyperlegible con @fontsource.
   - Shell con paneforge: barra de título, rail, panel lateral, editor, Assistant, Salida y barra de estado.
6. **Calidad:**
   - golangci-lint con depguard (reglas de §2.1);
   - ESLint + Prettier + svelte-check;
   - vitest;
   - workflow `.github/workflows/wails.yml`: go test, lint, frontend y `wails build` en Windows, macOS y Linux.

**Terminado cuando:** `wails build` genera un ejecutable que abre el shell con el diseño nuevo, cambia de idioma y pasa el CI.

### W1 · Seis coders en paralelo

| Track | Posee | Entrega |
|---|---|---|
| **G1 · Toolchain y ejecución** | `adapters/toolchain`, `app/run_program.go`, `bridge/run_service.go`, `bridge/files_service.go` | Herramientas (config → incluida → PATH), run/stop/build, Untitled, argumentos, `go.mod`, módulos, `go/format`, árbol de archivos, aviso de "primera vez" |
| **G2 · Depurador** | `adapters/delve`, `app/debug_session.go`, `bridge/debug_service.go` | Delve DAP con go-dap: breakpoints en vivo, pasos, continuar, ejecutar hasta el cursor, variables asíncronas, goroutines con ubicación y limpieza de binarios |
| **G3 · gopls** | `adapters/gopls`, `bridge/language_service.go` | Diagnósticos con rango, completado, hover, firma, definición, apariciones, símbolos. Degradación sin gopls |
| **G4 · Assistant** | `adapters/errorcatalog`, `app/explain_error.go`, `bridge/assistant_service.go`, `internal/i18n` | Parser de la salida de go y de panics, 25 ids con textos EN/ES, mensajes de error del IDE (UX_COPY) |
| **F1 · Editor** | `frontend/src/lib/editor`, `stores/diagnostics` | CodeMirror: lang-go, gutter de breakpoints, línea de depuración arena, lint ondulado, explicación en línea, completado, hover, buscar/reemplazar, temas |
| **F2 · Paneles y shell** | `frontend/src/lib/panels`, `shell`, `stores/{run,debug,files,settings,layout}` | Archivos, Outline, Salida (enlaces) y Problemas, Assistant (consejos, error, variables), tarjetas de variables con "acaba de cambiar", "Cómo llegaste aquí", barra de depuración, barra de estado, Configuración, asistente de primer arranque |

**Contratos compartidos** (`domain`, `ports.go`, `events.go`, `events.ts`, los JSON de i18n): son de sólo lectura en W1. Los cambios se piden como *Contract change request* al orquestador.

Las claves nuevas de i18n se proponen en el informe final; el orquestador las fusiona en W2.

### W2 · Integración, empaquetado y QA

- **Orquestador (Opus):** merge en el orden G1 → G4 → G2 → G3 → F1 → F2, CCRs, traducciones y pruebas de punta a punta.
- **Coder P (Sonnet), empaquetado:**
  - `wails build -nsis` (Windows), `.app` + `.dmg` (macOS) y AppImage (Linux);
  - variantes full y lite reutilizando `packaging/fetch_toolchain.py`;
  - firma pendiente de certificado.
- **Coder Q (Sonnet), QA:**
  - port de `packaging/qa` a la app Wails;
  - capturas EN/ES;
  - checklist de paridad con la 1.0 (cada función de la tabla de THONNY_COMPARISON que ya tenía la 1.0).

### W3 · Release

`v2.0.0-rc1` de la variante Wails. La 1.x PyQt sigue disponible como "clásica" hasta que la checklist de paridad esté completa.

## 4. Prompt común para los coders Sonnet

```text
Eres un coder del proyecto VizcachaIDE Wails (Go + Wails v2 + Svelte 5 + CodeMirror 6), un IDE de Go
para principiantes, bilingüe EN/ES. Trabajas en un git worktree aislado. Lee primero
docs/wails/PLAN_WAILS.md (§2 arquitectura, §2.1 reglas, §2.2 eventos, §2.4 código) y docs/wails/UX_COPY.md.
- Edita SÓLO las carpetas de tu track. Los contratos (domain, app/ports.go, bridge/events.go,
  frontend/src/lib/events.ts, locales) son de sólo lectura: los cambios se piden como
  "Contract change request" en tu informe.
- Textos visibles: por i18n, con el texto exacto de UX_COPY.md. Las claves nuevas van en tu informe con EN y ES.
- Library-first. Archivos de menos de 200 líneas, funciones de menos de 50, early return, nombres de dominio.
- Fuera de alcance: MicroPython y TinyGo.
- Verificación: go test ./..., golangci-lint run, npm run check, npm run test y wails build (si tu track lo afecta).
- Al terminar: UN commit local en español terminado en "Co-Authored-By: Claude Sonnet <noreply@anthropic.com>",
  sin push. Informe: hash del commit, archivos, cómo probarlo, limitaciones, CCRs y claves i18n nuevas.
```

## 5. Riesgos

| Riesgo | Mitigación |
|---|---|
| Reescribir todo es grande | Los contratos, datos y fixtures de la 1.0 se reutilizan. La 1.x sigue viva hasta igualarla. |
| CodeMirror y la integración con LSP | Sólo usamos extensiones oficiales (lint, autocomplete, tooltips). El backend traduce LSP a tipos de dominio simples. |
| WebView2 ausente en algún Windows 10 | El instalador NSIS de Wails incluye el bootstrapper de WebView2. |
| `macos-13` sin runners en GitHub | El CI usa `macos-14` y `macos-15-intel`. |
| Dos versiones en paralelo | La 1.x sólo recibe bugs y traducciones. Todo lo nuevo va a Wails. |
