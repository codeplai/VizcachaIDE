# VizcachaIDE (Wails variant)

A Go IDE for beginners, inspired by Thonny. English and Spanish from the first run.
Written in **Go + Wails v2** with a **Svelte 5 + TypeScript + CodeMirror 6** frontend.

The Python/PyQt version (folder `vizcacha/`) stays as the "classic" 1.x. This folder is the
rewrite. The plan is in [`docs/wails/PLAN_WAILS.md`](../docs/wails/PLAN_WAILS.md), the texts in
[`docs/wails/UX_COPY.md`](../docs/wails/UX_COPY.md) and the approved design in
[`docs/wails/prototype.html`](../docs/wails/prototype.html).

> Status: **W0 (foundations)**. The shell, theme, i18n, contracts and a mock backend are done.
> The Go services are stubs that return sample data; tracks G1-G4, F1, F2 of W1 fill them in.

## Requirements

- Go 1.25 or newer
- Node 24 and npm 11
- Wails CLI v2.16: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` (check with `wails doctor`)
- Windows: WebView2 (included in Windows 11). Linux: `libgtk-3-dev` and `libwebkit2gtk-4.1-dev`,
  and add `-tags webkit2_41` to every `go` and `wails` command.
- Optional: golangci-lint v2 (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)

## Run it

```sh
# Desktop app with hot reload (Go + frontend)
wails dev

# Frontend only, in the browser, with demo data and no Go
cd frontend
npm install
npm run dev
```

With `npm run dev` the app detects that `window.go` does not exist and uses the **mock bridge**
(`frontend/src/lib/bridge/mock.ts`). A strip at the top lets you switch between the three states of
the prototype (write, error explained, debug), the language and the theme. You can also open
`http://localhost:5173/?scenario=error&lang=es&theme=dark` to land on a state directly.

## Test, check and build

```sh
go vet ./... && go test ./...
golangci-lint run

cd frontend
npm run check      # svelte-check + eslint + prettier
npm run test       # vitest
npm run build

wails build        # produces build/bin/vizcacha.exe (or the app for your OS)
```

The same steps run in CI (`.github/workflows/wails.yml`) on Windows, macOS (arm64 and Intel) and Linux
whenever something under `wails/` changes.

## Translations

The catalogs are generated, do not edit the `locales/*.json` files by hand:

```sh
go run ./tools/po2json
```

It reads `tools/po2json/ux_copy.json` (semantic keys such as `actions.run`, with the exact texts of
`UX_COPY.md`) and the 1.0 gettext catalogs (`vizcacha/i18n/locale/{en,es}`, keys `legacy.*`), and writes:

- `frontend/src/lib/i18n/locales/{en,es}.json`: svelte-i18n, ICU MessageFormat.
- `internal/i18n/locales/{en,es}.json`: go-i18n v2, embedded in the binary.

The language follows the system (`es*` gives Spanish, anything else English) and can be switched from
the status bar. To add a text: add the key to `ux_copy.json`, run the tool and commit the result.

## Architecture

```
main.go                composition root (the only place that creates adapters and services)
internal/domain/       pure structs and rules, no project imports
internal/app/          ports (interfaces) and use cases, imports only domain
internal/adapters/     implementations of the ports; never import bridge or each other
internal/bridge/       thin services exposed to Wails + events.go (the only user of the Wails runtime)
internal/i18n/         go-i18n and the embedded catalogs
frontend/src/lib/      bridge (typed wrappers + mock), stores, shell, panels, editor, theme, i18n
```

Rules, checked in CI:

- Go: `.golangci.yml` (depguard) enforces the dependency rules; `internal/architecture_test.go` checks
  file length (under 200 lines) and the banned names `utils`, `helpers`, `common`, `misc`.
- Frontend: ESLint `no-restricted-imports` forbids importing `wailsjs` outside `src/lib/bridge`;
  `max-lines`, `max-depth` and `max-lines-per-function` enforce the size rules.
- The event names in `internal/bridge/events.go` and `frontend/src/lib/events.ts` must match
  (a Go test parses the `.ts`).

### Contracts frozen in W0

`internal/domain`, `internal/app/ports.go`, `internal/bridge/events.go`, `frontend/src/lib/events.ts`,
`frontend/src/lib/domain.ts` and the generated locale files are read-only for W1. Ask for changes with
a "Contract change request" in your report.

---

## Resumen en español

VizcachaIDE Wails es la reescritura en Go del IDE para principiantes. Usa Go + Wails v2 en el
backend y Svelte 5 + TypeScript + CodeMirror 6 en el frontend. La versión PyQt (`vizcacha/`) sigue
como 1.x "clásica" y no se toca.

- **Ejecutar con Go:** `wails dev` (ventana de escritorio con recarga en caliente).
- **Ejecutar sin Go:** `cd frontend && npm install && npm run dev`. Se activa el **mock** (datos de
  ejemplo) y aparece una barra superior para cambiar entre los tres estados del prototipo (escribir,
  error explicado, depurar), el idioma y el tema. También sirve `?scenario=error&lang=es&theme=dark`.
- **Pruebas:** `go vet ./... && go test ./...`, `golangci-lint run`, y en `frontend/`:
  `npm run check`, `npm run test`, `npm run build`.
- **Compilar:** `wails build` genera `build/bin/vizcacha.exe`.
- **Traducciones:** `go run ./tools/po2json` genera los JSON de ambos idiomas a partir de
  `tools/po2json/ux_copy.json` (textos de UX_COPY.md, claves semánticas) y de los `.po` de la 1.0
  (claves `legacy.*`). No se editan a mano.
- **Arquitectura:** `domain` (puro) ← `app` (puertos y casos de uso) ← `adapters` (implementaciones) y
  `bridge` (servicios finos para Wails, único que importa el runtime). `main.go` es el único sitio que
  crea adaptadores. Las reglas las verifican golangci-lint (depguard), ESLint y los tests.
- **Contratos congelados en W0:** `domain`, `app/ports.go`, `bridge/events.go`, `events.ts`,
  `domain.ts` y los JSON de i18n. Los cambios se piden como *Contract change request*.
