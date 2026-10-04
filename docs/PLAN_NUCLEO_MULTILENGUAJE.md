# Plan M0 · Núcleo multilenguaje de VizcachaIDE

> Prerrequisito de [PLAN_PYTHON.md](PLAN_PYTHON.md) y [PLAN_CPP.md](PLAN_CPP.md). El diseño y su
> justificación están en [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) y en el diagrama
> [arquitectura-multilenguaje.svg](arquitectura-multilenguaje.svg). Escrito el 2026-10-03 sobre la
> edición 2.0.0-rc1 (`wails/`).

## 0. Cómo usar este plan en un chat nuevo

1. Lee en este orden: este archivo; [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §2 a §4;
   [`wails/README.md`](../wails/README.md); [`docs/wails/PLAN_WAILS.md`](wails/PLAN_WAILS.md) §2
   (reglas) y §4 (prompt de los coders); [`docs/wails/UX_COPY.md`](wails/UX_COPY.md).
2. El código vive en `wails/` (Go 1.25, Wails v2.16, Svelte 5, CodeMirror 6). La edición PyQt
   (`vizcacha/`) no se toca.
3. Verificación obligatoria antes de cada commit:
   ```sh
   cd wails && go vet ./... && go test ./... && golangci-lint run
   cd frontend && npm run check && npm run test
   wails build
   ```
4. Reglas que aplican a todo: archivos de menos de 200 líneas, funciones de menos de 50, retorno
   temprano, nombres de dominio (nada de `utils`, `helpers`, `common`, `misc`), library-first, textos
   visibles sólo por i18n (`tools/po2json/ux_copy.json` → `go run ./tools/po2json`).
5. Commits locales en inglés, uno por track (N2 hace dos, ver §10), sin push salvo indicación
   expresa.
6. Rama: todo M0 se desarrolla en `m0-nucleo-multilenguaje` (creada desde `main`). Los worktrees
   de los coders parten de esa rama y el orquestador fusiona en ella; `main` sólo recibe M0 entero
   cuando pasa QA.
7. Modelos: el **orquestador** (N0, integración, CCR, fusión de textos y QA) es **Opus**; los
   **coders** de N1 a N5 son **Sonnet** (`Agent` con `model: "sonnet"` e `isolation: "worktree"`),
   cada uno con el prompt de §13.
8. Resultado esperado: **versión 2.1.0** con Go comportándose exactamente igual que en 2.0 y con el
   núcleo listo para recibir Python y C++.

## 1. Objetivo

Convertir el backend y el frontend de "un IDE de Go" en "un IDE con perfiles de lenguaje", sin añadir
ningún lenguaje todavía. Al terminar:

- El dominio y los puertos no nombran a Go (contrato v3).
- El código de protocolo (DAP, LSP, procesos, catálogo de errores) está en `internal/protocol` y lo
  usa el adaptador de Go a través de "flavors".
- `main.go` construye un `LanguageSupport` por lenguaje y un registro; el bridge resuelve el lenguaje
  por la extensión del archivo activo.
- El frontend lee los perfiles del backend y muestra u oculta botones y paneles por capacidad.

## 2. Alcance

Dentro: todo lo que hace falta para que un segundo lenguaje sea sólo "añadir una carpeta en
`adapters/` y un archivo en `main`". Fuera: Python, C++, cambios de diseño visual, cambios en la
edición PyQt, plugins externos.

## 3. Contrato v3 (un solo "Contract change request")

El contrato congelado en W0 cambia una sola vez. Los archivos del contrato son `internal/domain/*`,
`internal/app/ports.go`, `internal/bridge/events.go`, `frontend/src/lib/events.ts`,
`frontend/src/lib/domain.ts` y `frontend/src/lib/bridge/types.ts`. Los nombres de evento **no
cambian**; sólo cambian algunos payloads.

**Regla de nombres.** En el código de 2.0, "Language" a secas ya significa **idioma de la interfaz**:
`domain.LanguageEN/ES/Auto`, `Settings.Language`, `bridge.LanguageResolver`, `app.ResolveLanguage`,
`frontend/src/lib/language.ts` y las claves i18n `language.en`/`language.es`. Y `LanguageServer` /
`bridge.LanguageService` son el LSP. Para no mezclar conceptos, todo identificador nuevo que hable del
**lenguaje de programación** lleva `CodeLanguage` (tipo, constantes, campos, parámetros, stores y
claves i18n). `LanguageProfile`, `LanguageSupport` y `LanguageRegistry` se quedan así porque no
chocan con nada.

### 3.1 `internal/domain/code_language.go` (nuevo)

```go
type CodeLanguage string

const (
    CodeLanguageGo     CodeLanguage = "go"
    CodeLanguagePython CodeLanguage = "python"
    CodeLanguageCpp    CodeLanguage = "cpp"
)

// PackageAction is one verb of a language's package manager.
type PackageAction string

const (
    PackageInit   PackageAction = "init"
    PackageAdd    PackageAction = "add"
    PackageRemove PackageAction = "remove"
    PackageTidy   PackageAction = "tidy"
    PackageList   PackageAction = "list"
)

// IndentStyle is how the editor indents files of a language.
type IndentStyle struct {
    UseTabs bool `json:"useTabs"`
    Size    int  `json:"size"`
}

// Capabilities says which buttons and panels apply to a language. NewLanguageRegistry checks that
// they match the ports of the LanguageSupport (see 3.6), so they can never disagree.
type Capabilities struct {
    Build          bool            `json:"build"`          // Go, C++
    Console        bool            `json:"console"`        // Go (yaegi), Python (REPL)
    Format         bool            `json:"format"`         // gofmt, ruff, clang-format
    Check          bool            `json:"check"`          // go vet, ruff check
    PackageActions []PackageAction `json:"packageActions"` // empty = no manager
    ThreadsLabel   string          `json:"threadsLabel"`   // i18n key: "debug.goroutines" or "debug.threads"
}

type ToolRole string

const (
    RoleRuntime        ToolRole = "runtime"   // go, python
    RoleCompiler       ToolRole = "compiler"  // g++ / clang++
    RoleDebugAdapter   ToolRole = "debugAdapter"
    RoleLanguageServer ToolRole = "languageServer"
    RoleFormatter      ToolRole = "formatter"
)

// ToolSpec describes one tool a language needs and how to get it when it is missing.
type ToolSpec struct {
    ID             string   `json:"id"`      // "go", "dlv", "gopls", "python", "debugpy", "cxx", "lldb-dap", "clangd", "clang-format"
    Role           ToolRole `json:"role"`
    LabelKey       string   `json:"labelKey"`       // i18n key of the name shown in Settings
    MissingKey     string   `json:"missingKey"`     // i18n key of the "not found" notice
    InstallURL     string   `json:"installUrl"`     // "" when there is none
    InstallCommand string   `json:"installCommand"` // "" when there is none
    // ProvidedBy is the id of the tool that contains this one ("python" for debugpy, pylsp and
    // ruff): it has a status row in Settings but no path of its own, so PickExecutable rejects it.
    ProvidedBy string `json:"providedBy"` // "" for an executable of its own
}

type LanguageProfile struct {
    ID           CodeLanguage `json:"id"`
    NameKey      string       `json:"nameKey"`      // i18n key: "codeLanguage.go"...
    Extensions   []string     `json:"extensions"`   // lower case, with the dot
    Indent       IndentStyle  `json:"indent"`
    Capabilities Capabilities `json:"capabilities"`
    Tools        []ToolSpec   `json:"tools"`
}

// ToolStatus is what the IDE found for one tool (replaces ToolchainInfo).
type ToolStatus struct {
    ID           string       `json:"id"`
    CodeLanguage CodeLanguage `json:"codeLanguage"`
    Role         ToolRole     `json:"role"`
    Version  string     `json:"version"` // "" when missing
    Source   ToolSource `json:"source"`
    Path     string     `json:"path"`
}
```

### 3.2 `internal/domain/project.go`

- `GoModule` desaparece. Nace `ProjectContext{Root string; Kind ProjectKind; Name string}` con
  `ProjectKind` = `"gomod" | "folder" | "pyproject"`. Para Go, `Name` es el module path.
- `RunConfiguration` gana `CodeLanguage CodeLanguage` y `Echo bool` (true cuando el programa corre
  en una PTY, que ya devuelve como eco lo que se escribe; el frontend entonces no lo repite en
  Output; ver §7), y cambia `Module *GoModule` por `Project *ProjectContext`. `RunTarget` pasa a
  `"file" | "project"` (antes `"package"`).
- `GoTargetArgument` y `ExecutableName` se mudan a `adapters/golang` (son de Go).
- `ToolchainInfo` desaparece (tenía un campo `ToolSource` por herramienta); el tipo `ToolSource` se
  queda y cada `ToolStatus` lleva el suyo.

### 3.3 `internal/domain/debugging.go`

- `Goroutine` → `Thread{ThreadID int; Name string; Location *SourceLocation}`.
- `DebugState.Goroutines` → `Threads`; `CurrentGoroutine` → `CurrentThread`.
- `StopPanic` → `StopException` (valor JSON `"exception"`). La descripción sigue diciendo `panic:`
  en Go porque viene del texto de Delve.

### 3.4 `internal/domain/settings.go`

- `GoPath`, `DelvePath`, `GoplsPath` → `ToolPaths map[string]string` (clave = `ToolSpec.ID`).
- Nuevo `DefaultCodeLanguage CodeLanguage` (por defecto `"go"`) y
  `EnabledCodeLanguages []CodeLanguage` (vacío = todos; lo usa el asistente de primer arranque a
  partir de M1). `Settings.Language` sigue siendo el idioma de la interfaz.
- `FormatOnSave` se queda y se aplica sólo si `Capabilities.Format`.
- **Migración** en `adapters/settings/store.go`: al cargar un `settings.json` con las claves viejas,
  se copian a `ToolPaths["go"|"dlv"|"gopls"]` y se guarda el archivo nuevo. Test con un JSON de 2.0.

### 3.5 `internal/app/ports.go`

Se reemplaza `Toolchain` por `ProgramRunner` y nacen `CodeFormatter`, `CodeChecker` y
`PackageManager`. `Debugger`, `LanguageServer`, `ErrorExplainer`, `Console`, `SettingsStore`,
`FileWatcher`, `FileChangeSink` y `EventSink` no cambian de firma (sólo el tipo `DebugState` que
transportan).

**Convención para lo que un lenguaje no tiene:** cada capacidad opcional es un puerto aparte que vale
`nil` en `LanguageSupport` (`Console`, `Formatter`, `Checker`, `Packages`). La única excepción es
`Build`, que forma parte del ciclo de ejecución: es un método de `ProgramRunner` que devuelve
`ErrUnsupported` y el frontend sólo lo ofrece si `Capabilities.Build`. Los verbos de
`PackageManager` que falten también devuelven `ErrUnsupported`.

```go
// ProgramRunner runs programs of one language. Events: run:started, run:output, run:finished.
// Every runner receives the same *process.Supervisor (see 4.1), so only one program runs at a time
// in the whole IDE by construction, and Stop/IsRunning/WriteInput act on that single slot.
type ProgramRunner interface {
    // Configure decides what Run would run for a file: the file alone or its project.
    Configure(path string, programArgs []string) domain.RunConfiguration
    Run(ctx context.Context, config domain.RunConfiguration) error
    // Build compiles without running. Languages without a build step return ErrUnsupported.
    Build(ctx context.Context, config domain.RunConfiguration) error
    RunUntitled(ctx context.Context, path, source string, programArgs []string) (domain.RunConfiguration, error)
    Stop() error
    IsRunning() bool
    WriteInput(text string) error
    Tools(ctx context.Context) []domain.ToolStatus
    Environment() map[string]string
}

// CodeFormatter formats source text (go/format, ruff format, clang-format).
type CodeFormatter interface {
    Format(path, text string) (string, error)
}

// CodeChecker runs the language's checker (go vet, ruff check) without events; "" when clean.
type CodeChecker interface {
    Check(ctx context.Context, config domain.RunConfiguration) (string, error)
}

// PackageManager runs package commands through the shared process slot (they emit run events).
// Verbs a language lacks return ErrUnsupported; the profile lists the supported ones.
type PackageManager interface {
    Init(ctx context.Context, dir, name string) error
    Add(ctx context.Context, dir, pkg string) error
    Remove(ctx context.Context, dir, pkg string) error
    Tidy(ctx context.Context, dir string) error
    List(ctx context.Context, dir string) error
}
```

Errores nuevos en `app/errors.go`: `ErrUnsupported` ("this language does not support the action"),
`ErrUnknownCodeLanguage` y `ErrInconsistentProfile` (lo devuelve `NewLanguageRegistry`, §3.6). `ErrToolNotFound` se envuelve siempre como `fmt.Errorf("tool %q: %w", id, app.ErrToolNotFound)`
para que el frontend sepa qué herramienta falta (ver §7).

### 3.6 `internal/app/language_support.go` (nuevo)

```go
type LanguageSupport struct {
    Profile        domain.LanguageProfile
    Runner         ProgramRunner
    Debugger       Debugger
    LanguageServer LanguageServer
    Explainer      ErrorExplainer
    Console        Console        // nil when Capabilities.Console is false
    Formatter      CodeFormatter  // nil when Capabilities.Format is false
    Checker        CodeChecker    // nil when Capabilities.Check is false
    Packages       PackageManager // nil when PackageActions is empty
}

// LanguageRegistry is pure: it only maps extensions and ids to supports.
type LanguageRegistry struct{ /* supports []LanguageSupport; byExtension map[string]int */ }

// NewLanguageRegistry fails with ErrInconsistentProfile when a capability and its port disagree
// (Console, Format, Check, PackageActions), when two profiles claim the same extension, or when
// defaultID is not registered. The IDE then fails at start-up, never in the middle of a lesson.
func NewLanguageRegistry(defaultID domain.CodeLanguage, supports ...LanguageSupport) (*LanguageRegistry, error)
func (r *LanguageRegistry) Profiles() []domain.LanguageProfile
func (r *LanguageRegistry) ForPath(path string) (LanguageSupport, bool) // by lower-case extension
func (r *LanguageRegistry) ForID(id domain.CodeLanguage) (LanguageSupport, bool)
func (r *LanguageRegistry) Default() LanguageSupport
func (r *LanguageRegistry) All() []LanguageSupport

// UnavailableSupport is the support of a language whose adapter does not exist yet (Python and
// C++ in 2.1): its runner, debugger and language server answer ErrUnsupported and every optional
// port is nil. Its profile keeps the real capabilities so the menus exist, but NewLanguageRegistry
// skips the capability check for it.
func UnavailableSupport(profile domain.LanguageProfile) LanguageSupport
```

`app.ConfigurationForFile`, `FindGoModule`, `ParseModulePath`, `GoModFileName` y `go_commands.go` se
mudan a `adapters/golang`. `SplitProgramArguments` y `singleWord` (renombrado `SingleWordArgument`,
exportado) se quedan en `app` porque son neutrales.

### 3.7 Servicios del bridge: firmas antes y después

| Servicio | Antes (2.0) | Después (v3) |
|---|---|---|
| `RunService.Run(path, args)` | igual | resuelve el soporte por `path` |
| `RunService.RunUntitled(source, args)` | | `RunUntitled(path, source, args)`: la extensión del nombre sin título decide el lenguaje |
| `RunService.Vet(config)` | | `Check(config)` (usa `config.CodeLanguage`; `ErrUnsupported` si `Checker` es nil) |
| `RunService.Format(text)` | | `Format(path, text)` (`ErrUnsupported` si `Formatter` es nil) |
| `RunService.Toolchain()` | | se elimina; ver `CodeLanguagesService.Tools()` |
| `RunService.ModInit/ModGet/ModTidy` | | se eliminan; ver `PackagesService` |
| `RunService.Stop()`, `WriteInput` | igual | van al slot compartido de procesos (cualquier runner sirve; se usa el del soporte activo) |
| `DebugService.Start(path, bps, argsText)` | igual | resuelve por `path`; `ErrBusy` si hay otra sesión, comprobado y reservado bajo el mismo mutex (sin carrera entre comprobar y arrancar) |
| `DebugService.*` (pasos, variables, stop) | igual | sobre la sesión activa |
| `LanguageService.*` | igual | enrutan por `path` o `at.File`; `Shutdown` para todos |
| `AssistantService.Explain(raw, dir)` | | `Explain(codeLanguage, raw, dir)` |
| `AssistantService.ExplainDiagnostics(diags)` | igual | enruta por `diagnostic.Location.File`; sin archivo → lenguaje por defecto |
| `ConsoleService.Eval(code)`, `Reset()` | | `Eval(codeLanguage, code)`, `Reset(codeLanguage)`; `ErrUnsupported` si no hay consola |
| `PackagesService` (nuevo) | | `Init/Add/Remove/Tidy/List(codeLanguage, dir, arg)` |
| `CodeLanguagesService` (nuevo) | | `Profiles() []LanguageProfile`, `Tools() []ToolStatus`. No confundir con `LanguageService`, que sigue siendo el LSP |
| `SettingsService.PickExecutable(tool)` | devuelve `ToolchainInfo` | valida `tool` contra los `ToolSpec` del registro (rechaza los que tienen `ProvidedBy`); devuelve `[]ToolStatus` |
| `FilesService.SaveFileDialog(name, folder)` | añade `.go` | filtros con todas las extensiones de los perfiles; extensión por defecto = la del nombre sugerido |

Una sola ejecución y una sola depuración a la vez en todo el IDE, como hoy. La ejecución lo
garantiza por construcción el `Supervisor` compartido (§4.1); la depuración, `DebugService` con su
mutex.

### 3.8 Frontend: `domain.ts`, `types.ts`, `events.ts`

`domain.ts` y `types.ts` reflejan 3.1 a 3.7 (camelCase). `events.ts` no cambia de nombres; el
test Go que compara ambos archivos sigue pasando. `EventPayloads['debug:stopped']` transporta el
nuevo `DebugState`.

## 4. Paquete `internal/protocol` (nuevo)

Vive fuera de `adapters/` para que la regla "los adapters no se importan entre sí" siga intacta.
Importa sólo `app`, `domain` y librerías. Cada subpaquete expone un **Flavor**: la interfaz con lo
que cambia de un lenguaje a otro.

### 4.1 `protocol/process`: ejecutar y supervisar un proceso

Se extrae de `adapters/toolchain/runner.go`, `stream.go`, `process_other.go`, `process_windows.go`
e `interrupt_windows.go`.

```go
type Supervisor struct{ /* one slot: ErrBusy while a process is alive */ }

// SilentNotice is printed as stdout when nothing arrived after Delay (Go's "first build",
// C++'s "Compiling…").
type SilentNotice struct {
    Text  func() string
    Delay time.Duration
}

type Job struct {
    Config  domain.RunConfiguration // what run:started announces
    Command string
    Args    []string
    Dir     string
    Env     map[string]string
    Mode    Mode                    // Pipes or Terminal (PTY)
    Cleanup func()
    Notice  *SilentNotice           // nil when there is none
    // Stages: a job may have a compile stage before the program (C++). Each stage reports
    // its exit code; run:finished carries the last one.
    Then func(exitCode int) (*Job, bool)
}
func New(sink app.EventSink) *Supervisor
func (s *Supervisor) Start(ctx context.Context, job Job) error
func (s *Supervisor) Stop() error          // Ctrl+C semantics, then kill tree after ~2 s
func (s *Supervisor) IsRunning() bool
func (s *Supervisor) WriteInput(text string) error
```

**Un solo `Supervisor` para todo el IDE.** `main.go` crea una instancia y la inyecta en los
runners, en los package managers y en los `ReverseHandler` de DAP que arrancan el programa
depurado con `runInTerminal` (Python y C++, M1 y M2). Así "un programa a la vez" queda garantizado
por construcción: el segundo `Start` recibe `ErrBusy` sin importar de qué lenguaje venga, y el
bridge no tiene que recorrer runners preguntando `IsRunning()`, que tendría una carrera entre
comprobar y arrancar.

`Then` permite "compilar y luego ejecutar" sin que C++ reimplemente la supervisión. En modo
`Terminal` la entrada que el usuario escribe la devuelve la propia PTY como eco; el runner lo
indica con `RunConfiguration.Echo` y el frontend no la duplica (ver §7). En M0 Go sigue en modo
`Pipes` con `Echo: false`: la PTY sólo la estrenan Python y C++.

### 4.2 `protocol/pty`: prueba de concepto primero

Objetivo: un `exec.Cmd` corriendo en una pseudoterminal en Windows (ConPTY), macOS y Linux, con
lectura de salida, escritura de entrada, `Ctrl+C` por la PTY y cierre limpio.

- Candidatos (library-first): `github.com/creack/pty` en Unix; en Windows, un binding de ConPTY
  (`github.com/charmbracelet/x/conpty` o `github.com/UserExistsError/conpty`). El spike los compara
  y deja el elegido en `go.mod`.
- Criterio de éxito del spike: un programa C++ o Python que imprime sin salto de línea, lee una
  línea y luego se cae, muestra lo impreso **antes** de caer; y el panel Output no se ensucia con
  secuencias VT de ConPTY. Si aparecen, `protocol/process` las quita con un parser ANSI existente
  (library-first: candidato `github.com/charmbracelet/x/ansi`, función `Strip`), no con expresiones
  regulares propias; la librería elegida se añade a la lista permitida de §4.6.
- Si falla en Windows: `Mode: Pipes` para todos en 2.1 y se registra la PTY como mejora.

**Resultado del spike (N0, 2026-10-03, Windows 10).** Elegido `github.com/aymanbagabas/go-pty`
(ConPTY en Windows, `creack/pty` en Unix) con `github.com/charmbracelet/x/ansi` para limpiar.
Implementado en `internal/protocol/pty` (`Start`, `Read`, `Write`, `Interrupt`, `Wait`, `Close`,
`Clean`). Con una caída nativa real (`faulthandler._sigsegv`), por pipes se pierde lo impreso y por
la PTY se ve entero; `input()` funciona y la PTY devuelve el eco (`Echo: true` es correcto).
Hallazgos que N3 debe respetar:
- ConPTY **parte las líneas** al ancho de la terminal: `pty.Columns = 4096` lo evita (las rutas de
  los tracebacks no se cortan).
- ConPTY emite borrado de pantalla, título de ventana, cursor y colores; `pty.Clean` los quita
  todos. Python 3.13+ colorea los tracebacks en una terminal: `pty.Start` añade `NO_COLOR=1`,
  `PYTHON_COLORS=0` y `TERM=dumb`. `Clean` trabaja sobre cadenas completas: limpiar por líneas.
- **Ctrl+C escrito como ETX no llega** al programa en ConPTY. `Stop` en modo `Terminal` llama a
  `Interrupt` y, a los 2 s, mata el árbol como hoy. Go sigue en `Pipes`, así que no pierde nada.

### 4.3 `protocol/dap`

Se mueven desde `adapters/delve`: `client.go` (entero), el mapeo genérico de `mapping.go`
(variables, frames, threads, `localsReference`, `truncateValue`), la sesión (`session.go`,
`control.go`, `events.go`, `inspector.go`, `frames.go`, `variable_requests` equivalentes) y
`shutdown.go`. Quedan en `adapters/golang/delve`: `adapter.go` (la implementación de
`app.Debugger` que une sesión y flavor), `process.go`, `process_other.go` y `process_windows.go`
(arrancar `dlv dap`, leer la dirección TCP y el grupo de procesos por sistema), `launch.go`
(argumentos de launch, binario de depuración y su limpieza) y las peculiaridades (frames `subtle`,
línea de ruido, motivos de parada de Delve).

```go
type Transport interface {
    Open(ctx context.Context) (io.ReadWriteCloser, error) // TCP for Delve, stdio for lldb-dap and debugpy.adapter
    Close() error
}

// ReverseHandler answers requests the adapter sends to the client.
type ReverseHandler interface {
    RunInTerminal(args dap.RunInTerminalRequestArguments) (processID int, err error)
}

type Flavor interface {
    AdapterID() string                                       // "go", "python", "lldb"
    Launch(config domain.RunConfiguration, env map[string]string) (json.RawMessage, error)
    ExceptionFilters() []string                              // debugpy: ["uncaught"]; Delve, lldb: nil
    StopReason(event *dap.StoppedEvent) (domain.StopReason, string)
    KeepFrame(frame dap.StackFrame) bool                     // Delve drops "subtle" runtime frames
    KeepVariable(variable dap.Variable) bool                 // debugpy hides "special variables"
    IsLocalsScope(name string) bool
    Output(event *dap.OutputEvent) (text, category string, ok bool)
}

// SessionDeps groups what a session needs; a struct instead of seven positional parameters.
type SessionDeps struct {
    Sink      app.EventSink
    Book      *app.BreakpointBook
    Tracker   *app.ChangeTracker
    Transport Transport
    Flavor    Flavor
    Reverse   ReverseHandler       // nil = every reverse request is answered "unsupported"
    Texts     func(key string) string // backend i18n, injected: protocol knows keys, not languages
}

type Session struct{ /* generic: connect, initialize, launch, breakpoints, configurationDone, steps, runTo, variables, frames, finish */ }
func NewSession(deps SessionDeps) *Session
```

`NewClient` gana el parámetro `ReverseHandler`: una `dap.RunInTerminalRequest` se contesta con
`RunInTerminalResponse{ProcessId}`; cualquier otra petición inversa recibe un error "unsupported".
El Debugger de Go sigue sin usar `runInTerminal` (Delve no lo ofrece), así que su handler devuelve
error y nada cambia para Go.

### 4.4 `protocol/lsp`

Se mueven desde `adapters/gopls`: `connection.go`, `documents.go`, `positions.go`, `mapping.go`,
`messages.go` (los params genéricos), `symbols.go`, `signature.go`, `queries.go` y la máquina de
estados de `server.go` y `lifecycle.go`. Quedan en `adapters/golang/gopls`: `locator.go`,
`moduleRoot` e `initializeParams` propios de gopls.

```go
type Flavor interface {
    Command(env map[string]string) (executable string, args []string, err error) // locate + args
    RootOf(path string) string                                                   // module root or folder
    InitializationOptions() any                                                  // nil when none
    Configuration() any                                                          // sent as workspace/didChangeConfiguration after initialized; nil when none
    Environment() map[string]string
}

type Server struct{ /* lazy start on first OpenDocument; unavailable without the tool; idle shutdown */ }
func New(sink app.EventSink, flavor Flavor, options Options) *Server // Options{IdleTimeout: 5 * time.Minute}
```

Nuevo en M0: **apagado por inactividad**. Cuando no queda ningún documento abierto durante
`IdleTimeout`, el servidor hace `shutdown`/`exit` y vuelve a `stateIdle`; el siguiente
`OpenDocument` lo arranca otra vez. `lsp:status` pasa a `"unavailable"` sólo cuando falta la
herramienta, nunca por inactividad.

Es un comportamiento nuevo dentro de un hito que promete paridad con 2.0, así que va **en un commit
aparte y posterior** a la extracción: primero N2 mueve el código con los tests de 2.0 en verde
(paridad), y después añade el apagado con su test de reloj inyectado. Si QA encuentra algo, se sabe
si viene de mover o de lo nuevo, y el segundo commit se puede revertir solo.

### 4.4b `protocol/toollocator`

Se extrae de `adapters/toolchain/locator.go` y se generaliza: para una `ToolSpec` busca en este orden
la ruta configurada (`Settings.ToolPaths[id]`), las rutas empaquetadas que declara cada lenguaje
(`toolchain/<lenguaje>/...`, relativas al ejecutable del IDE) y el PATH. Devuelve un
`domain.ToolStatus` con `Source` (`configured`, `bundled`, `path`). Cada lenguaje le pasa sus
candidatos; la validación propia (la versión mínima de Python, `xcode-select` en macOS) la hace el
adaptador del lenguaje antes de aceptar el resultado.

### 4.5 `protocol/errorcatalog`

Se mueven `catalog.go`, `catalog_entry` y `explainer.go` con un cambio: el catálogo se carga desde
`[]byte` que entrega cada lenguaje (`NewExplainer(catalogJSON []byte, parse OutputParser)`), y el
parser se inyecta:

```go
// OutputParser turns raw tool output into diagnostics. identify returns the catalog id of a message.
type OutputParser func(rawOutput, workingDir string, identify func(message string) string) []domain.Diagnostic
```

`locations.go` se exporta como `ResolvePath`, `LocationFrom`, `NamedGroups`, `SplitLines` para que
los parsers de Python y C++ los reutilicen. El parser de Go y `data/catalog.json` (renombrado
`catalog.go.json`) van a `adapters/golang/errors`.

### 4.6 Reglas depguard nuevas (`.golangci.yml`)

```yaml
protocol-is-shared:
  list-mode: strict
  files: ["**/internal/protocol/**"]
  allow:
    - $gostd
    - github.com/codeplai/VizcachaIDE/wails/internal/domain
    - github.com/codeplai/VizcachaIDE/wails/internal/app
    - github.com/google/go-dap
    - go.lsp.dev
    - github.com/creack/pty      # or the ConPTY binding the spike chose
    - github.com/charmbracelet/x/ansi   # only if the spike needs the VT filter
  deny:
    - pkg: github.com/codeplai/VizcachaIDE/wails/internal/adapters
      desc: protocol never imports adapters
    - pkg: github.com/codeplai/VizcachaIDE/wails/internal/bridge
      desc: protocol never imports bridge
```

**La regla `adapters-are-independent` tiene que cambiar.** Hoy prohíbe a cualquier adaptador importar
`internal/adapters/...`. Pero tras la reagrupación, los subpaquetes de un lenguaje necesitan su raíz:
`adapters/golang/runner` usa `FindGoModule` (que §3.6 muda a `adapters/golang`); en M1
`adapters/python/runner` usa el locator y el entorno de `adapters/python`; en M2 `adapters/cpp/lldbdap`
reutiliza la etapa de compilación de `adapters/cpp/runner`. La regla nueva es: **dentro de un lenguaje
sí, entre lenguajes y hacia los neutrales no.** depguard v2 resuelve con la coincidencia más
específica, así que basta una regla por lenguaje:

```yaml
adapters-are-independent:         # neutral adapters: settings, filewatch, filesystem, windowstate
  files: ["**/internal/adapters/**", "!**/internal/adapters/golang/**",
          "!**/internal/adapters/python/**", "!**/internal/adapters/cpp/**"]
  deny:
    - pkg: github.com/codeplai/VizcachaIDE/wails/internal/bridge
      desc: adapters never import bridge
    - pkg: github.com/codeplai/VizcachaIDE/wails/internal/adapters
      desc: adapters never import other adapters
go-adapter-stays-in-go:           # M1 and M2 copy this rule for python/ and cpp/
  files: ["**/internal/adapters/golang/**"]
  allow:
    - github.com/codeplai/VizcachaIDE/wails/internal/adapters/golang
  deny:
    - pkg: github.com/codeplai/VizcachaIDE/wails/internal/bridge
      desc: adapters never import bridge
    - pkg: github.com/codeplai/VizcachaIDE/wails/internal/adapters
      desc: a language adapter imports only its own folder
```

**Comprobado en N0: depguard no aplica la coincidencia más específica** (el `deny` de
`internal/adapters` gana al `allow` más largo de `internal/adapters/golang`). Por eso la regla real
de `.golangci.yml` excluye `adapters/golang/**` de `adapters-are-independent` y, en
`go-adapter-stays-in-go`, **niega cada carpeta hermana por nombre** (python, cpp, settings,
filewatch, filesystem, windowstate y, hasta que N1 y N3 las muevan, toolchain, delve, gopls,
errorcatalog y console). Probado con paquetes temporales: `adapters/golang/x` puede importar
`adapters/golang` y `app`, y no `adapters/settings`. Al añadir una carpeta de adaptador hay que
sumarla a esas listas; M1 y M2 copian la regla. Los adapters sí pueden importar
`internal/protocol`. `internal/architecture_test.go` no cambia.

## 5. `adapters/golang`: reagrupación del adaptador de Go

| Hoy | En M0 | Qué cambia |
|---|---|---|
| `adapters/toolchain/*` | `adapters/golang/runner/` | implementa `ProgramRunner` sobre el `Supervisor` compartido (modo `Pipes`, `Echo: false`); `Configure` detecta go.mod; `Tools` devuelve tres `ToolStatus`. El mismo tipo implementa también `CodeChecker` (go vet) y `CodeFormatter` (go/format), y `support_go.go` lo asigna a los tres campos |
| `app/go_commands.go` + `RunGoCommand` | `adapters/golang/packages/` | implementa `PackageManager` sobre el `Supervisor` compartido: `Init`, `Add`, `Tidy`; `Remove` y `List` devuelven `ErrUnsupported` |
| `adapters/delve/*` | `adapters/golang/delve/` | sólo `adapter.go`, `process*.go`, `launch.go` y `flavor.go` (ver §4.3); el resto vino de `protocol/dap` |
| `adapters/gopls/*` | `adapters/golang/gopls/` | sólo `flavor.go` (localizador, moduleRoot, initialize) |
| `adapters/errorcatalog/*` | `adapters/golang/errors/` | `parser.go`, `data/catalog.go.json`, constructor que llama a `protocol/errorcatalog` |
| `adapters/console/*` | `adapters/golang/console/` | sin cambios |
| `adapters/toolchain/locator.go` | `protocol/toollocator/` | genérico (§4.4b): configurado → `toolchain/<lenguaje>/...` → PATH; cada lenguaje declara sus rutas empaquetadas |
| `adapters/{settings,filewatch,filesystem,windowstate}` | igual | neutrales |

`adapters/golang/profile.go` declara el `domain.LanguageProfile` de Go:

```go
var Profile = domain.LanguageProfile{
    ID: domain.CodeLanguageGo, NameKey: "codeLanguage.go", Extensions: []string{".go"},
    Indent: domain.IndentStyle{UseTabs: true, Size: 4},
    Capabilities: domain.Capabilities{Build: true, Console: true, Format: true, Check: true,
        PackageActions: []domain.PackageAction{domain.PackageInit, domain.PackageAdd, domain.PackageTidy},
        ThreadsLabel: "debug.goroutines"},
    Tools: []domain.ToolSpec{
        {ID: "go", Role: domain.RoleRuntime, LabelKey: "settings.toolGo", MissingKey: "errors.goNotFound", InstallURL: "https://go.dev/dl/"},
        {ID: "dlv", Role: domain.RoleDebugAdapter, LabelKey: "settings.toolDelve", MissingKey: "errors.delveNotFound", InstallCommand: "go install github.com/go-delve/delve/cmd/dlv@latest"},
        {ID: "gopls", Role: domain.RoleLanguageServer, LabelKey: "settings.toolGopls", MissingKey: "errors.goplsNotFound", InstallCommand: "go install golang.org/x/tools/gopls@latest"},
    },
}
```

Los tests de hoy se mueven con su código. Las fixtures de `errorcatalog/testdata/go_output` y de
`gopls/testdata` siguen valiendo.

## 6. Bridge y `main.go`

- Cada servicio recibe `*app.LanguageRegistry` en vez de un puerto concreto. `RunService` y
  `DebugService` guardan el soporte activo mientras hay algo corriendo.
- El paso repetido "buscar el soporte por ruta o lo devuelvo como error" vive en un archivo
  **nuevo**, `bridge/support_router.go`: `supportFor(path) (app.LanguageSupport, error)` y
  `supportOf(codeLanguage)`. No va en `bridge/language_resolver.go`, que resuelve el idioma de la
  interfaz (en/es) y no se toca.
- `main.go` se parte en `main.go` (Wails, ventana, `Supervisor` compartido, registro) y
  `support_go.go` con `newGoSupport(sink, store, texts, supervisor) (app.LanguageSupport, func(ctx))`.
  M1 añadirá `support_python.go` y M2 `support_cpp.go`. Sigue siendo el único sitio que crea
  adaptadores.
- En 2.1, Python y C++ se registran con `app.UnavailableSupport(profile)`. Sus perfiles provisionales
  (los de M1 §4.1 y M2 §4.1, sin `Tools`) viven en `support_unavailable.go` y desaparecen cuando
  llega cada adaptador.
- `shutdown` recorre `registry.All()` y para runner, debugger y LSP de cada uno.
- El `FirstBuildNotice` de Go se pasa como `Notice` del `Job`.

## 7. Frontend

| Pieza | Cambio |
|---|---|
| `lib/domain.ts`, `lib/bridge/types.ts`, `lib/bridge/wails.ts` | contrato v3 (§3) |
| `lib/stores/codeLanguages.ts` (nuevo; no confundir con `lib/language.ts`, que es el idioma de la interfaz) | `profiles` (de `bridge.codeLanguages.profiles()` al arrancar), `codeLanguageOf(path)` por extensión, `activeCodeLanguage` derivado de `activePath`, `capabilities` derivado, `tools` (reemplaza a `toolchain`), `toolStatus(id)` |
| `lib/editor/languageSupport.ts` (nuevo) | `languageExtensionsFor(path, profile)`: `go()` / `python()` / `cpp()` más `indentUnit` y `tabSize` del perfil. `createEditor` crea el `EditorState` de cada archivo con sus extensiones de lenguaje (ya hay un estado por archivo; no hace falta un compartimento) |
| `lib/editor/goCompletion.ts` | se renombra `lspCompletion.ts`; el código no cambia |
| `lib/panels/GoroutinesPanel.svelte` | se renombra `ThreadsPanel.svelte`; `DebugTab 'goroutines'` → `'threads'`; el rótulo usa `capabilities.threadsLabel` |
| `lib/stores/goModules.ts`, `shell/ModulesDialog.svelte` | se renombran `packages.ts` y `PackagesDialog.svelte`; los botones salen de `capabilities.packageActions`; llaman a `bridge.packages.*` con `activeCodeLanguage` |
| `lib/shell/SettingsTools.svelte` | un grupo por perfil, una fila por `ToolSpec`; `pickTool(toolId)`; las filas con `providedBy` muestran estado y versión pero no el botón "Elegir" |
| `lib/stores/saving.ts` | formato al guardar sólo si `capabilities.format`; `bridge.run.format(path, text)` |
| `lib/stores/untitled.ts` | `openUntitled(bridge, codeLanguage)`: plantilla y extensión por lenguaje (`editor/templates.ts`); el nombre sin título lleva la extensión del perfil |
| `lib/shell/FileMenu.svelte` | "Nuevo" abre un archivo del lenguaje por defecto; submenú "Nuevo archivo de…" con un ítem por perfil |
| `lib/stores/commands.ts` | `missingToolIn` lee `tool "<id>"` del mensaje del backend y busca el `ToolSpec` en los perfiles: aviso con `missingKey`, acción "Instalar" (`installUrl`) o "Copiar comando" (`installCommand`) y "Elegir en Ajustes". `runActiveFile` pasa `path` a `runUntitled`. `sendProgramInput` no duplica el texto cuando la configuración que llegó en `run:started` trae `echo: true` (contrato v3, §3.2) |
| `lib/stores/layout.ts`, `panels/BottomPanel.svelte` | la pestaña Consola sólo existe si `capabilities.console`; el botón Build sólo si `capabilities.build` |
| `lib/stores/run.ts` | `lastRunConfiguration.project` en vez de `.module` |
| `lib/bridge/mock*.ts` | `codeLanguages.profiles()` devuelve los tres perfiles (Python y C++ con sus capacidades) aunque sólo Go tenga escenarios; `tools()` por perfil; `packages`; `console.eval(codeLanguage, code)` |
| `lib/i18n` | claves nuevas (§8) |

## 8. Textos nuevos (`tools/po2json/ux_copy.json`)

| Clave | EN | ES |
|---|---|---|
| `codeLanguage.go` | Go | Go |
| `codeLanguage.python` | Python | Python |
| `codeLanguage.cpp` | C++ | C++ |
| `debug.threads` | Threads | Hilos |
| `debug.goroutines` | Goroutines | Goroutines |
| `files.newFileOf` | New {codeLanguage} file | Nuevo archivo de {codeLanguage} |
| `settings.toolsOf` | {codeLanguage} tools | Herramientas de {codeLanguage} |
| `settings.defaultCodeLanguage` | Language for new files | Lenguaje de los archivos nuevos |
| `errors.toolNotFound` | {tool} isn't available. Install it or choose it in Settings. | {tool} no está disponible. Instálalo o elígelo en Ajustes. |
| `errors.toolInstall` | Install | Instalar |
| `errors.toolCopyCommand` | Copy the command | Copiar el comando |
| `errors.unsupportedAction` | This isn't available for {codeLanguage}. | Esto no está disponible para {codeLanguage}. |
| `packages.title` | Packages | Paquetes |

El grupo `codeLanguage.*` es nuevo a propósito: `language.en` y `language.es` ya existen y son los
nombres de los idiomas del selector de Ajustes.

Las claves existentes `settings.toolGo`, `errors.goNotFound`, `errors.delveNotFound`,
**`errors.goplsNotFound`** y las del diálogo de módulos se conservan **con su texto actual** (los
perfiles las referencian). En particular, `errors.goplsNotFound` no se reescribe: el texto de 2.0
dice, con razón, que las sugerencias básicas siguen funcionando sin gopls.

**Dueño de `ux_copy.json` y de los `locales/*.json` generados:** N0 añade todas las claves de esta
tabla en el commit "Contract v3". Si un track necesita otra, la pide en su informe y el orquestador
la fusiona al integrar (como en M1 y M2); ningún track edita esos archivos, para evitar conflictos
seguros entre worktrees.

## 9. Tests

- Go: `app/language_support_test.go` (registro: extensión en mayúsculas, archivo sin extensión,
  lenguaje por defecto, `ErrInconsistentProfile` cuando una capacidad y su puerto no casan o dos
  perfiles reclaman la misma extensión, `UnavailableSupport` contesta `ErrUnsupported`);
  `protocol/process` (un slot compartido por dos runners, `Then` con dos etapas, notice silencioso,
  stop con árbol de procesos; y en modo PTY un programa que imprime, lee y cae); `protocol/dap` con
  un adaptador falso en memoria (`net.Pipe`) que cubre `runInTerminal`; `protocol/lsp` con el
  servidor falso de hoy más el apagado por inactividad con reloj inyectado; `protocol/errorcatalog`
  con un parser falso; `adapters/golang/*` con los tests movidos; `adapters/settings` migración;
  `bridge` con fakes del registro (enrutamiento, `ErrBusy` entre lenguajes, `ErrUnsupported`).
- Frontend: `codeLanguages.test.ts` (perfil activo, capacidades), `packages.test.ts`,
  `untitled.test.ts` (extensión por lenguaje), `commands.test.ts` (`missingToolIn` genérico), los
  tests existentes adaptados al contrato v3.
- CI: sin pasos nuevos; `golangci-lint` con las reglas de §4.6.

## 10. Tracks y orden

```
N0 (orquestador, secuencial) ──► N1 · N2 · N3 · N4 · N5 en paralelo ──► integración (orquestador) ──► QA
```

| Track | Dueño de | Entrega | Hecho cuando |
|---|---|---|---|
| **N0 · Contrato y spike** (secuencial) | `domain`, `app/ports.go`, `app/errors.go`, `domain.ts`, `types.ts`, `.golangci.yml`, `protocol/pty`, `go.mod`/`go.sum`, `tools/po2json/ux_copy.json` y los `locales/*.json` | spike de PTY con decisión escrita en el commit y su dependencia en `go.mod`; comprobación de la precedencia de depguard (§4.6); contrato v3 compilando en Go y TypeScript; todas las claves de §8; bridge y mock con las firmas nuevas aunque devuelvan datos de ejemplo | `go build ./...` y `npm run check` pasan; un commit "Contract v3" del que parten los demás |
| **N1 · DAP** | `protocol/dap`, `adapters/golang/delve` | extracción + `Flavor` + `ReverseHandler` + `SessionDeps` | tests de Delve de 2.0 en verde; depurar un programa Go con breakpoints, pasos, variables y goroutines igual que en 2.0 |
| **N2 · LSP** (dos commits) | `protocol/lsp`, `adapters/golang/gopls` | commit 1: extracción + `Flavor` (paridad); commit 2: apagado por inactividad | tras el commit 1, tests de gopls de 2.0 en verde y diagnósticos, completado, hover, definición y símbolos iguales que en 2.0; tras el 2, el test de inactividad con reloj inyectado |
| **N3 · Procesos y errores** | `protocol/process`, `protocol/errorcatalog`, `protocol/toollocator`, `adapters/golang/{runner,packages,errors,console,profile.go}` | `ProgramRunner` de Go, `PackageManager` de Go, `Explainer` de Go sobre el motor compartido | tests de toolchain y errorcatalog de 2.0 en verde; `go run`, stdin, stop, go mod, vet, gofmt y las 25 explicaciones iguales que en 2.0 |
| **N4 · Frontend** | `frontend/src/lib/*` | §7 completo contra el mock | `npm run check` y `npm run test` en verde; `npm run dev` muestra los tres perfiles en Ajustes y el menú Nuevo; Go se ve igual que en 2.0 |
| **N5 · App y bridge** | `app/language_support.go` (con `UnavailableSupport`), `bridge/*` (con `support_router.go`), `adapters/settings` (migración), `main.go`, `support_go.go`, `support_unavailable.go` | registro con su validación, `Supervisor` compartido, enrutamiento, servicios nuevos, migración, perfiles provisionales de Python y C++ | tests del bridge con fakes en verde; `wails build` arranca con el perfil de Go cuando N1 a N3 estén integrados |

Orden de integración: N3 → N1 → N2 → N5 → N4. Después, QA.

### 10.0 Estado (2026-10-03): N0–N5 integrados, falta QA

Todos los tracks están fusionados en `m0-nucleo-multilenguaje` y **todas las piezas transitorias
de la tabla 10.1 están retiradas** (puerto `app.Toolchain`, `ToolchainInfo` en Go y TS,
`app/go_project.go`, `app/go_commands.go`, los miembros `toolchain`/`mod*` del bridge TS, las
reglas depguard de las carpetas viejas). La validación neutral de argumentos es
`app.SingleWordArgument` / `app.ErrInvalidArgument`. Verificado: `go vet`, `go test ./...`,
`golangci-lint` (0 issues), `npm run check`, `npm run test` (251), `wails build` y arranque del
ejecutable. Queda **QA** (§11): la matriz de `docs/wails/QA_WAILS.md` §2 a mano, con capturas EN/ES.

Decisiones tomadas al integrar:
- `packages.remove` y `packages.list` usan los textos de M1 ("Uninstall"/"Desinstalar", "Show
  installed packages"/"Ver paquetes instalados") para no cambiarlos después.
- `@codemirror/lang-python` y `lang-cpp` no se añaden en M0: los traen P5 (M1) y C5 (M2); hasta
  entonces esos archivos se ven como texto plano con la sangría del perfil.
- El lenguaje por defecto del registro es siempre Go; `Settings.DefaultCodeLanguage` sólo decide el
  lenguaje de los archivos nuevos en el frontend. M1 puede leerlo al arrancar si hace falta.
- La UI de 2.0 no tiene botón Build: `capabilities.build` queda listo para cuando se añada.

### 10.1 Estado tras N0 (lo que queda transitorio y quién lo retira)

N0 dejó el contrato v3 compilando con todo 2.0 en verde. Para eso hay piezas marcadas
**`Transitional (M0)`** en el código (no `Deprecated`, porque staticcheck rompería el lint). Cada
track borra las suyas:

| Pieza transitoria | Dónde | La retira |
|---|---|---|
| Puerto `app.Toolchain` y `domain.ToolchainInfo` | `app/ports.go`, `domain/project.go` | N3 (el runner implementa `ProgramRunner`) y N5 (el bridge deja de usarlos) |
| `FindGoModule`, `ParseModulePath`, `GoModFileName`, `ConfigurationForFile`, `GoTargetArgument`, `GoExecutableName` | `app/go_project.go` | N3 los muda a `adapters/golang`. N1 sigue llamando `app.GoTargetArgument` en `launch.go`; el orquestador lo cambia al integrar (N3 se integra antes que N1) |
| `ModInitArguments`, `GetArguments`, `ModTidyArguments`, `singleWord` | `app/go_commands.go` | N3 (`adapters/golang/packages`); `SingleWordArgument` exportado se queda en `app` |
| Migración de `settings.json` de 2.0 (`goPath`, `delvePath`, `goplsPath` → `toolPaths`) | `adapters/settings` | N5. Hasta entonces, en la rama, un `settings.json` de 2.0 pierde las rutas configuradas |
| `run.toolchain`, `run.modInit/modGet/modTidy`, `ToolId = string`, `pickExecutable` que devuelve `ToolchainInfo` | `frontend/src/lib/bridge/types.ts` | N4 (usa `codeLanguages.tools()` y `packages.*`) y N5 (cambia el servicio Go) |
| `wailsCodeLanguages.ts`: perfiles fijos y herramientas convertidas desde `ToolchainInfo`; `packages` sobre `ModInit/ModGet/ModTidy` | `frontend/src/lib/bridge/` | N5 (crea `CodeLanguagesService` y `PackagesService`); `languageProfiles.ts` queda sólo para el mock |
| `runUntitled(path, …)`, `format(path, …)`, `console.eval(codeLanguage, …)` y `assistant.explain(codeLanguage, …)` ignoran el lenguaje en `wails.ts` | `frontend/src/lib/bridge/wails.ts` | N5 (las firmas Go nuevas) |
| `CONSOLE_LANGUAGE = 'go'` | `frontend/src/lib/stores/console.ts` | N4 (`activeCodeLanguage`) |
| Reglas depguard de las carpetas viejas (toolchain, delve, gopls, errorcatalog, console) | `.golangci.yml` | N1 y N3, al vaciar cada carpeta |

**Reparto fijado al lanzar N1–N5** (completa la tabla de §10):
- `main.go`: N1, N2 y N3 sólo tocan los imports y las líneas que construyen su adaptador; el resto
  (partirlo, `support_go.go`, `support_unavailable.go`) es de N5. El orquestador resuelve los choques.
- El pegamento Go↔TS (`frontend/src/lib/bridge/wails.ts`, `wailsCodeLanguages.ts` y
  `frontend/wailsjs/**`) es de **N5**, porque cambia cuando cambian las firmas Go; el resto de
  `frontend/src/lib` es de N4.
- Nadie borra una pieza transitoria que otro track aún usa: lo nuevo se crea al lado y el orquestador
  borra lo viejo al integrar. N3 mantiene el runner de Go cumpliendo también `app.Toolchain`; N5
  envuelve el `app.Toolchain` de 2.0 en los puertos nuevos dentro de `package main` hasta que se
  integra el runner de N3.

Además, N0 sincronizó `tools/po2json/ux_copy.json`: 96 claves existían sólo en los `locales/*.json`
generados (añadidas a mano por tracks anteriores) y `go run ./tools/po2json` las borraba. Ahora
están en la fuente con sus textos EN/ES exactos; regenerar ya no pierde nada.

## 11. Criterios de salida (2.1.0)

- La matriz de [`docs/wails/QA_WAILS.md`](wails/QA_WAILS.md) §2 se repite para Go sin regresiones,
  en Windows, con capturas EN/ES.
- `settings.json` de 2.0 se migra sin perder las rutas configuradas.
- Ajustes muestra "Herramientas de Go" con las tres filas; "Nuevo archivo de…" lista Go, Python y
  C++, y elegir Python o C++ abre un archivo con plantilla **pero** ejecutar muestra
  `errors.unsupportedAction` hasta M1 y M2 (los perfiles sin backend se registran con
  `app.UnavailableSupport`, §3.6 y §6; así el menú ya existe y los tests del enrutamiento se
  ejercitan).
- Tamaño de la variante `lite` sin cambios apreciables; `full` sigue incluyendo Go, Delve y gopls.
- `CHANGELOG.md` con la entrada 2.1.0 y las notas EN/ES en `docs/release/`.

## 12. Riesgos

| Riesgo | Mitigación |
|---|---|
| El spike de PTY falla en Windows | `Mode: Pipes` para todos y la PTY queda registrada; la API de `protocol/process` no cambia |
| Romper Go sin darse cuenta | los tests de 2.0 se mueven con el código y deben seguir en verde; QA de paridad al final |
| El refactor de delve/ y gopls/ cruza el límite de 200 líneas | partir por responsabilidad (transport, session, inspector, mapping) antes de mover |
| Tres servicios con el registro repiten el mismo `ForPath` + error | `bridge/support_router.go` (nuevo, §6): `supportFor(path) (LanguageSupport, error)` |
| El frontend queda a medias si N4 tarda | el mock cubre el contrato v3 desde N0; N4 no bloquea a los demás |
| Confundir el lenguaje de programación con el idioma de la interfaz | regla de nombres de §3 (`CodeLanguage*`, `codeLanguage.*`); revisión en la integración |
| Conflictos entre worktrees en `ux_copy.json`, `go.mod` o `.golangci.yml` | N0 es su único dueño; los tracks piden cambios en el informe y los fusiona el orquestador |
| depguard no aplica la coincidencia más específica | N0 lo comprueba antes del commit del contrato; alternativa: un `deny` por cada carpeta hermana |

## 13. Prompt común para los coders

```text
You are a coder on VizcachaIDE (wails/: Go 1.25 + Wails v2 + Svelte 5 + CodeMirror 6), a beginner IDE,
bilingual EN/ES. You work in an isolated git worktree, branched from m0-nucleo-multilenguaje, on track
<N?> of docs/PLAN_NUCLEO_MULTILENGUAJE.md.
First read that plan (sections 0, 3, 4, your track in 10), docs/EXTENSION_MULTILENGUAJE.md section 4,
wails/README.md and docs/wails/PLAN_WAILS.md sections 2 and 4.
- Edit ONLY your track's folders. The v3 contract (internal/domain, internal/app/ports.go,
  internal/bridge/events.go, frontend/src/lib/{events,domain}.ts, bridge/types.ts) is fixed by N0:
  request changes as a "Contract change request" in your report. go.mod, .golangci.yml,
  tools/po2json/ux_copy.json and the generated locales are owned by N0: list what you need in your
  report instead of editing them.
- Naming: "Language" alone means the UI language (en/es) and LanguageService is the LSP. Anything
  about the programming language is CodeLanguage (types, fields, params, stores, i18n keys
  codeLanguage.*).
- Go must behave exactly as in 2.0: move the existing tests with the code and keep them green.
- Library-first. Files under 200 lines, functions under 50, early return, domain names. Visible texts
  only through i18n.
- Verify: go vet ./... && go test ./... && golangci-lint run; cd frontend && npm run check && npm run test;
  wails build when your track affects it.
- When done: ONE local commit in English (N2: two, extraction then idle shutdown) ending with
  "Co-Authored-By: Claude <noreply@anthropic.com>", no push. Report: commit hash, files, how to test,
  limitations, CCRs and needed i18n keys or dependencies.
```
