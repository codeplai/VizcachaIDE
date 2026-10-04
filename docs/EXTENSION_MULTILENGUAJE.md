# VizcachaIDE — Extensión a Python y C++: factibilidad y planteamiento

> Estudio de arquitectura hecho el 2026-10-03 leyendo el código de la edición 2.0
> (`wails/`: Go + Wails v2, Svelte 5, CodeMirror 6), no el README. Diagrama:
> [arquitectura-multilenguaje.svg](arquitectura-multilenguaje.svg). La edición clásica 1.x
> (`vizcacha/`, PyQt5) queda fuera: está en mantenimiento y no se toca.
>
> Planes de ejecución, uno por chat: [PLAN_NUCLEO_MULTILENGUAJE.md](PLAN_NUCLEO_MULTILENGUAJE.md)
> (M0, prerrequisito), [PLAN_PYTHON.md](PLAN_PYTHON.md) (M1) y [PLAN_CPP.md](PLAN_CPP.md) (M2).

## 1. Respuesta corta

**Es factible, con riesgo moderado y sin cambiar de tecnología.** La arquitectura hexagonal de la
2.0 ya separa dominio, puertos, adaptadores y bridge, y los dos puertos difíciles están modelados
sobre protocolos que no son de Go:

1. `Debugger` calca el Debug Adapter Protocol (DAP) y `LanguageServer` calca el Language Server
   Protocol (LSP), en [`internal/app/ports.go`](../wails/internal/app/ports.go). Python y C++ tienen
   adaptadores DAP y servidores LSP libres y maduros: debugpy, lldb-dap o GDB, python-lsp-server,
   clangd.
2. El cliente DAP ([`delve/client.go`](../wails/internal/adapters/delve/client.go)) y la conexión LSP
   ([`gopls/connection.go`](../wails/internal/adapters/gopls/connection.go)) no contienen nada de
   Delve ni de gopls. Se reutilizan tal cual.
3. El motor del Assistant (catálogo de expresiones regulares con textos EN/ES, placeholders e ids
   estables) es neutral. Sólo el parser de la salida y el JSON de datos son de Go.
4. El frontend consume structs del dominio, no de Go. CodeMirror tiene paquetes oficiales
   `@codemirror/lang-python` y `@codemirror/lang-cpp` con la misma licencia MIT que `lang-go`.
5. El empaquetado (localizador configurado → empaquetado → PATH, versiones y sha256 fijados) se
   extiende con una carpeta por lenguaje.

Lo que no sirve es todo lo que dice "Go" por nombre: el inventario está en la sección 3 y es menos
de lo que parece, unas 2.000 líneas de las 7.000 de `internal/` y 16 archivos del frontend.

**El cambio obligatorio es uno:** construir primero un núcleo multilenguaje (fase M0), que es
refactorización sin funcionalidad visible, y recién después añadir Python (M1) y C++ (M2). Meter
Python "al lado" de Go sin ese núcleo duplicaría toolchain, bridge y stores, y rompería la regla
de que los adaptadores no se importan entre sí.

**La condición de usar siempre compiladores libres se cumple** con CPython (licencia PSF),
debugpy, python-lsp-server y ruff (MIT), LLVM (Apache-2.0) o GCC y GDB (GPL). Tabla en la sección 5.

## 2. Lo que la arquitectura actual ya tiene a favor

| Pieza | Dónde | Por qué sirve para Python y C++ |
|---|---|---|
| Puerto `Debugger` | `app/ports.go` | `Start`, `SetBreakpoints`, `StepOver`, `StepInto`, `StepOut`, `Resume`, `RunTo`, `RequestVariables`, `FrameVariables` son las peticiones DAP `launch`, `setBreakpoints`, `next`, `stepIn`, `stepOut`, `continue`, `variables` y `scopes`. |
| Cliente DAP | `adapters/delve/client.go` | Secuencias, eventos, errores y cierre con `github.com/google/go-dap`. Cero Delve. |
| Mapeo DAP | `adapters/delve/mapping.go` | Frames, ámbito `Locals`, variables, hilos y motivos de parada. Sólo el filtro de frames `subtle` y una cadena de ruido son de Delve. |
| Puerto `LanguageServer` | `app/ports.go` | `didOpen`/`didChange`/`didClose`, completion, hover, definition, signatureHelp, documentHighlight y documentSymbol: métodos LSP estándar que pylsp y clangd implementan. |
| Conexión LSP | `adapters/gopls/connection.go`, `documents.go`, `positions.go`, `mapping.go` | JSON-RPC por stdio con `go.lsp.dev/jsonrpc2`, documentos abiertos, posiciones UTF-16 y conversión a `domain.Diagnostic`. Sólo `initialize`, el localizador y `moduleRoot` son de gopls. |
| `EventSink` y eventos | `app/ports.go`, `bridge/events.go` | `run:*`, `debug:*`, `lsp:*`, `assistant:*` son neutrales. Sólo el campo `Goroutines` de `DebugState` lleva nombre de Go. |
| Motor del catálogo | `adapters/errorcatalog/catalog.go`, `explainer.go` | Regex con grupos nombrados, `{placeholders}`, textos EN/ES, ids estables, validación en tests. |
| Casos de uso | `app/explain_error.go`, `debug_session.go`, `run_program.go` | `ExplainError`, `BreakpointBook`, `ChangeTracker` y `SplitProgramArguments` no saben de Go. |
| Paneles del frontend | `lib/panels/*` | Variables, Calls, CallStack, Problems, Assistant y Output consumen `domain.ts`. `goCompletion.ts` ya es genérico salvo por el nombre. |
| Localizador y empaquetado | `adapters/toolchain/locator.go`, `packaging/versions.toml`, `fetch_toolchain.py` | El patrón "configurado → empaquetado → PATH" y "versión y sha256 fijados" se repite por lenguaje. |
| Reglas en CI | `.golangci.yml` (depguard), `architecture_test.go`, ESLint | Protegen el refactor: cualquier import cruzado o archivo largo rompe el build. |

## 3. Lo que está atado a Go (inventario)

### Backend

| Archivo | Qué es de Go | Qué pasa en M0 |
|---|---|---|
| `domain/project.go` | `GoModule`, `RunConfiguration.Module`, `GoTargetArgument`, `ToolchainInfo` con `GoVersion`, `DelveVersion`, `GoplsVersion` | `ProjectContext` genérico (go.mod, carpeta, pyproject), `RunConfiguration.Language`, `[]ToolStatus` por rol |
| `domain/debugging.go` | `Goroutine`, `DebugState.Goroutines`, `StopPanic` | `Thread`, `Threads`, `StopException`. En Go la UI sigue diciendo "goroutine" y "panic" vía i18n |
| `domain/settings.go` | `GoPath`, `DelvePath`, `GoplsPath` | `ToolPaths map[rol]ruta`, `DefaultLanguage`. `FormatOnSave` se queda y aplica por capacidad |
| `app/ports.go` | `Toolchain` con `RunGoCommand`, `Vet`, `FormatSource` (gofmt) | `ProgramRunner` con `Configure`, `Run`, `Build`, `RunUntitled`, `Stop`, `WriteInput`, `Check`, `Format`, `Tools`. `PackageManager` aparte y opcional |
| `app/run_program.go`, `app/go_commands.go` | Detección de go.mod, argumentos de `go mod` | `Configure` vive en el runner de cada lenguaje; `go_commands` se muda a `adapters/golang` |
| `adapters/toolchain` (838 LOC) | Todo | → `adapters/golang/runner` |
| `adapters/delve` (1.277 LOC) | `launch.go`, parte de `session.go` | `client.go` y el mapeo genérico → `protocol/dap`; lo demás → `adapters/golang/delve` |
| `adapters/gopls` (1.257 LOC) | `lifecycle.go` (initialize), `locator.go`, `moduleRoot` | `connection`, `documents`, `positions`, `mapping` → `protocol/lsp`; lo demás → `adapters/golang/gopls` |
| `adapters/errorcatalog` (406 LOC) | `parser.go`, `data/catalog.json` | Motor → `protocol/errorcatalog`; parser y datos → `adapters/golang/errors` |
| `adapters/console` (295 LOC) | yaegi | → `adapters/golang/console`. El puerto `Console` pasa a ser opcional |
| `bridge/run_service.go` | `ModInit`, `ModGet`, `ModTidy`, `Vet`, `Format` | `PackagesService` nuevo; `Run`, `Build`, `Check` y `Format` resuelven el lenguaje por la ruta |
| `bridge/settings_service.go`, `files_dialogs.go` | `ToolId` go/dlv/gopls; `.go` como extensión por defecto | `ToolPaths` por rol; extensión según el perfil |
| `main.go` | Un toolchain, un depurador, un LSP | Tres `LanguageSupport` y un `LanguageRegistry` |

### Frontend (16 archivos sin contar tests)

`domain.ts`, `bridge/types.ts`, `bridge/wails.ts`, `bridge/mock*.ts`, `editor/extensions.ts`
(`go()` y sangría con tabs), `editor/goCompletion.ts` (sólo el nombre), `panels/GoroutinesPanel.svelte`,
`panels/TraceSection.svelte`, `shell/SettingsTools.svelte`, `shell/FirstRunWizard.svelte`,
`stores/commands.ts`, `stores/goModules.ts`, `stores/layout.ts`, `stores/saving.ts`,
`stores/settings.ts`, y las claves de i18n que nombran Go.

### Empaquetado

`packaging/versions.toml`, `packaging/fetch_toolchain.py`, `packaging/go_tools.py`,
`wails/packaging/build_release.py`, `wails/packaging/smoke_test.py` y el instalador NSIS.

## 4. Planteamiento nuevo: núcleo neutral + un perfil por lenguaje

### 4.1 La idea: `LanguageSupport`

Un lenguaje es un **perfil** (datos puros, en el dominio) más un **paquete de puertos** (en app),
que `main.go` arma una vez por lenguaje y registra. Nada más en el código sabe cuántos lenguajes
hay.

Boceto: la versión vinculante del contrato, con todos los campos, está en
[PLAN_NUCLEO_MULTILENGUAJE.md](PLAN_NUCLEO_MULTILENGUAJE.md) §3. Allí se fija también la regla de
nombres: "Language" a secas ya es el idioma de la interfaz (en/es), así que el lenguaje de
programación se llama `CodeLanguage`.

```go
// internal/domain/code_language.go
type CodeLanguage string // "go", "python", "cpp"

type LanguageProfile struct {
    ID           CodeLanguage
    Name         string       // "Go", "Python", "C++"
    Extensions   []string     // [".go"], [".py"], [".cpp", ".cc", ".cxx", ".h", ".hpp"]
    Indent       IndentStyle  // tabs | 4 espacios
    Capabilities Capabilities // Build, Console, Packages, Format, Check, ThreadsLabel
    Tools        []ToolRole   // runtime/compilador, debugAdapter, languageServer, formatter
}

// internal/app/language_support.go
type LanguageSupport struct {
    Profile        domain.LanguageProfile
    Runner         ProgramRunner
    Debugger       Debugger
    LanguageServer LanguageServer
    Explainer      ErrorExplainer
    Console        Console        // nil = sin consola (C++)
    Formatter      CodeFormatter  // nil = sin formateador
    Checker        CodeChecker    // nil = sin checker (C++)
    Packages       PackageManager // nil = sin gestor (C++)
}

type LanguageRegistry interface {
    Profiles() []domain.LanguageProfile
    ForPath(path string) (LanguageSupport, bool)
    ForID(id domain.CodeLanguage) (LanguageSupport, bool)
}
```

Las capacidades se leen en el frontend desde `codeLanguages.profiles()` y deciden qué se muestra:
botón Build, panel Consola, diálogo de paquetes, formato al guardar, y el rótulo del panel de hilos
("Goroutines" en Go, "Hilos" en los otros).

### 4.2 Enrutamiento: la regla de Thonny

Se ejecuta y se depura **el archivo activo**, y el lenguaje lo decide su extensión. Así funcionan
las carpetas mixtas sin configuración y la UI no necesita un selector de lenguaje por proyecto.
Cada servicio del bridge hace `registry.ForPath(path)` y delega. Reglas que se mantienen:

- Una sola ejecución o depuración a la vez (`ErrBusy`), como hoy.
- Un servidor LSP por lenguaje, arrancado al abrir el primer archivo de ese lenguaje y apagado
  tras unos minutos sin archivos abiertos. Esto importa por memoria: gopls ronda los 200 MB,
  clangd entre 100 y 300 MB, pylsp unos 50 MB.
- Los archivos sin título llevan el lenguaje elegido en el menú Nuevo (plantilla "hola mundo" por
  lenguaje), y `DefaultCodeLanguage` en Settings decide el que sale por defecto.

### 4.3 Paquete `internal/protocol`: el código compartido sale de `adapters/`

La regla depguard "los adapters nunca importan otros adapters" es valiosa y se conserva entre
lenguajes (dentro de la carpeta de un mismo lenguaje sí hay imports; ver M0 §4.6). Por eso el
código común no va en un adapter sino en un paquete nuevo, `internal/protocol`, que sólo importa
`app`, `domain` y librerías:

- `protocol/dap`: `Client` (hoy `delve/client.go`), atención a las **peticiones inversas** del
  adaptador (`runInTerminal`, ver 4.4), y el mapeo genérico de pila, ámbitos, variables e hilos.
- `protocol/lsp`: conexión stdio, documentos abiertos, posiciones UTF-16, mapeo de diagnósticos,
  símbolos, completado y firma.
- `protocol/errorcatalog`: el motor de hoy, con el parser inyectado por lenguaje (`OutputParser`).
- `protocol/pty`: ejecución en pseudoterminal (4.4).

Nueva regla depguard: `protocol` no importa `adapters` ni `bridge`. El `architecture_test.go`
(archivos de menos de 200 líneas, nombres prohibidos) se aplica sin cambios.

### 4.4 Hallazgo que condiciona el diseño: ejecutar en una pseudoterminal

Hoy el programa del alumno corre con tuberías (`os/exec` con `StdoutPipe`). Con Go no se nota
porque `fmt.Print` escribe sin búfer. Con C++ y Python sí, porque `std::cout` y `print` usan búfer
completo cuando la salida es una tubería. El aviso antes de `cin` sí aparece, porque `cin` vacía
`cout` antes de leer; el problema está en lo demás:

- **Salida perdida al caer el programa.** Si el alumno imprime y después el programa se cae con un
  fallo de segmentación, el búfer nunca se vacía y no se ve nada de lo impreso: "mi cout no funciona".
- **Salida tardía al depurar.** Paso a paso, la línea del `cout` se ejecuta y en el panel no aparece
  nada hasta que el búfer se llena o el programa termina.
- **Orden de stdout y stderr.** stderr no tiene búfer, así que un mensaje de error aparece antes que
  lo impreso justo antes.
- **`input()` al depurar Python.** `python -u` arregla los tres puntos anteriores en Python, pero
  debugpy en modo consola interna no acepta entrada por teclado.

La solución estándar es ejecutar el programa bajo una **pseudoterminal**: ConPTY en Windows
(disponible desde Windows 10 1809, que ya es el mínimo del IDE) y `pty` en macOS y Linux. Hay
librerías Go para ambos (`github.com/creack/pty` para Unix y bindings de ConPTY para Windows);
hay que hacer una prueba de concepto antes de M0 para elegirlas. Con PTY, además, los colores ANSI
y el `input()` se comportan como en una terminal, y el panel Output ya entiende ANSI.

En depuración, lldb-dap y debugpy soportan la petición inversa `runInTerminal`: el adaptador le pide
al IDE que lance el programa, el IDE lo lanza en su PTY y responde con el pid. El cliente DAP de hoy
ignora las peticiones que vienen del servidor, así que `protocol/dap` tiene que atenderlas. Para
Python hay una alternativa igual de válida: lanzar
`python -m debugpy --listen 127.0.0.1:<puerto> --wait-for-client archivo.py` bajo la PTY y hacer
`attach`.

### 4.5 El perfil de cada lenguaje, rol por rol

| Rol | Go (hoy) | Python (M1) | C++ (M2) |
|---|---|---|---|
| Ejecutar | `go run` | `python -u archivo.py` | compilar con `g++`/`clang++ -g -O0 -std=c++17 -Wall -Wextra` a un binario temporal y ejecutarlo |
| Build | `go build` | no aplica (botón oculto) | compilar sin ejecutar |
| Proyecto | go.mod → paquete | la carpeta; detecta `.venv` si existe | la carpeta: si hay varios `.cpp`, se compilan todos. Sin CMake en la primera versión |
| Depurar (DAP) | `dlv dap` | debugpy | lldb-dap (LLVM) o `gdb -i=dap` (GDB 14 o más) |
| Inteligencia (LSP) | gopls | python-lsp-server (jedi + pyflakes), `python -m pylsp` | clangd; `compile_flags.txt` opcional para includes extra |
| Formato | gofmt (`go/format`) | `ruff format` | `clang-format` |
| Check (hoy `go vet`) | `go vet` | `ruff check` | los warnings del propio compilador |
| Consola | yaegi | REPL con el módulo `code` de la stdlib, JSON por stdio | ninguna (cling pesa cientos de MB; panel oculto) |
| Paquetes | `go mod init/get/tidy` | `pip install/uninstall/list` en el intérprete empaquetado | ninguno |
| Errores | parser de la salida de go, 25 entradas | tracebacks: `NameError`, `IndentationError`, `TabError`, `SyntaxError`, `TypeError`, `ZeroDivisionError`, `IndexError`, `KeyError`, `AttributeError`, `ModuleNotFoundError`, `ValueError`, `RecursionError`, `UnboundLocalError`, `FileNotFoundError`… | formato GNU `archivo:línea:col: error: …` de GCC y Clang (el catálogo admite varios patrones por entrada): `was not declared in this scope` / `use of undeclared identifier`, `expected ';'`, `no matching function`, `cannot convert`, `control reaches end of non-void function`, errores del enlazador (`undefined reference`) y caídas en ejecución (SIGSEGV, código 139 o 0xC0000005) |
| Hilos | goroutines | threads (normalmente uno) | threads |
| Editor | `lang-go`, tabs | `lang-python`, 4 espacios | `lang-cpp`, 4 espacios |
| Nuevo archivo | plantilla hola mundo | plantilla hola mundo | plantilla con `#include <iostream>` |

Sobre C++: GCC y Clang imprimen el **mismo formato** de diagnóstico, así que hay un solo parser; las
palabras de cada mensaje cambian y por eso el catálogo lleva patrones de los dos. El adaptador
detecta la familia por el nombre del ejecutable (`g++` o `clang++`) y usa el adaptador DAP
correspondiente.

### 4.6 Qué cambia en el frontend

- `domain.ts` y `bridge/types.ts`: espejo del contrato v3 (`LanguageProfile`, `Thread`,
  `ToolStatus[]`, `RunConfiguration.language`); `events.ts` no cambia.
- `stores/language.ts` nuevo: perfiles cargados al arrancar, lenguaje del archivo activo,
  capacidades derivadas.
- `editor/extensions.ts`: un `Compartment` de lenguaje (`go()`, `python()`, `cpp()`) y la sangría
  del perfil; `goCompletion.ts` pasa a llamarse `lspCompletion.ts` sin cambiar de código.
- `GoroutinesPanel` → `ThreadsPanel` con rótulo por lenguaje; `goModules.ts` y `ModulesDialog` →
  `packages.ts` y `PackagesDialog` con verbos por lenguaje.
- `SettingsTools` agrupa las herramientas por lenguaje; `FirstRunWizard` pregunta qué lenguajes se
  van a usar y revisa sólo esas herramientas.
- `saving.ts`: formato al guardar según la capacidad; `saveFileDialog` sugiere la extensión del perfil.
- Mock bridge: perfiles y escenarios demo de Python y C++, para desarrollar la UI sin herramientas.

### 4.7 Empaquetado

- Layout: `toolchain/go/`, `toolchain/python/`, `toolchain/cpp/` junto al ejecutable; el localizador
  busca en `toolchain/<lenguaje>/bin`.
- `versions.toml` gana secciones `[python]` (python-build-standalone por plataforma, con sha256, y
  wheels fijados de debugpy, python-lsp-server y ruff) y `[cpp]` (LLVM o GCC por plataforma, clangd
  y clang-format). `fetch_toolchain.py` se generaliza con un manifiesto por lenguaje.
- Variantes: `lite` (sólo el IDE, usa lo instalado), `full-go`, `full-python`, `full-cpp` y `full`.
- Tamaños, orden de magnitud a validar: Go ≈ 90 MB (hoy); Python ≈ 50 a 70 MB; C++ en Windows
  ≈ 150 a 300 MB según la distribución (llvm-mingw o WinLibs); `full` con los tres ≈ 300 a 450 MB.
- CI: una prueba de humo por lenguaje (ejecutar hola mundo, parar en un breakpoint, recibir un
  diagnóstico del LSP) en los tres sistemas.

## 5. Herramientas libres y licencias

| Herramienta | Rol | Licencia | Nota |
|---|---|---|---|
| Go, Delve, gopls | hoy | BSD-3, MIT, BSD-3 | ya empaquetadas |
| CPython (python-build-standalone) | runtime Python | PSF | builds portables para Windows, macOS y Linux; es lo que usa `uv` |
| debugpy | DAP Python | MIT | de Microsoft, mantenido con VS Code |
| python-lsp-server (jedi, pyflakes) | LSP Python | MIT | puro Python, se instala con pip en el intérprete empaquetado |
| ruff | formato y check Python | MIT | un solo binario, muy rápido |
| LLVM: clang++, lldb-dap, clangd, clang-format | compilador, DAP, LSP y formato C++ | Apache-2.0 con excepciones LLVM | un solo proveedor para los tres sistemas; en Windows, llvm-mingw |
| GCC (MinGW-w64 / WinLibs) + GDB 14 o más | compilador y DAP C++ alternativos | GPLv3 con excepción de runtime; GDB GPLv3 | lo que usan muchos cursos; `gdb -i=dap` exige un GDB compilado con soporte de Python |
| @codemirror/lang-python, lang-cpp | editor | MIT | igual que `lang-go` |

Sobre la GPL: empaquetar GCC o GDB como programas aparte, con sus licencias en `toolchain/licenses`
como ya se hace con Delve y gopls, no cambia la licencia MIT del IDE (es agregación, no enlace). Los
programas de los alumnos tampoco se ven afectados gracias a la excepción de runtime de GCC.

## 6. Qué cambia y qué se conserva

| Capa | Cambia en M0 | Se conserva |
|---|---|---|
| Dominio | `LanguageProfile`, `Thread`, `StopException`, `ProjectContext`, `ToolStatus`, `Settings.ToolPaths` | `Diagnostic`, `ErrorExplanation`, `CompletionItem`, `DocumentSymbol`, `DebugState`, `Variable`, `StackFrame`, `ConsoleResult`, `FileNode` |
| App | `ProgramRunner` reemplaza a `Toolchain`; `PackageManager`; `LanguageRegistry` | `Debugger`, `LanguageServer`, `ErrorExplainer`, `Console`, `EventSink`, los casos de uso |
| Protocolo | nace `internal/protocol` con lo extraído de delve y gopls, el motor del catálogo y la PTY | las librerías: go-dap, jsonrpc2, protocol |
| Adapters | se reagrupan en `adapters/golang`; nacen `python` y `cpp` | el comportamiento 2.0 de Go, con sus tests |
| Bridge | resolución por ruta; `PackagesService`, `LanguagesService`; `ToolPaths` | nombres de servicios y de eventos |
| Frontend | compartimento de lenguaje, `stores/language.ts`, capacidades, paneles renombrados | shell, paneles, stores de run/debug/diagnostics, i18n, temas |
| Empaquetado | manifiesto por lenguaje, variantes | la cadena de build, NSIS, firmas sha256, CI |

El contrato congelado en W0 (`domain`, `ports.go`, `events.go`, `domain.ts`) cambia de versión. Hay
que presentarlo como un "Contract change request" formal, actualizar el mock del frontend y sus
tests en el mismo cambio, y publicarlo como 2.1 antes de que exista Python.

## 7. Fases

### M0 · Núcleo multilenguaje (secuencial, bloqueante)

1. Prueba de concepto de PTY en los tres sistemas (dos o tres días; decide la librería).
2. Contrato v3: dominio y puertos; espejo en `domain.ts`; mock actualizado.
3. Extracción de `protocol/dap`, `protocol/lsp`, `protocol/errorcatalog`, `protocol/pty`; reglas
   depguard nuevas.
4. `LanguageRegistry`, `LanguageSupport`, reagrupación en `adapters/golang`, resolución por ruta en
   el bridge, `main.go` con un solo perfil (Go).
5. Frontend: perfiles, compartimento, capacidades, paneles renombrados.

Criterio de salida: la 2.0 se comporta igual para Go (misma matriz QA), todos los tests verdes,
y el mock ya expone tres perfiles aunque sólo uno tenga backend. Se publica como **2.1**.

### M1 · Python (tracks en paralelo, como en W1)

| Track | Entrega |
|---|---|
| P1 Ejecutor | `python -u` bajo PTY, carpeta como proyecto, `.venv`, REPL con `code`, `pip` como `PackageManager` |
| P2 Depurador | debugpy por `runInTerminal` o `attach`; `justMyCode` activado para ocultar la stdlib |
| P3 Inteligencia | pylsp por `protocol/lsp`; ruff para formato y check |
| P4 Assistant | parser de tracebacks y `catalog.python.json` con unas 25 entradas EN/ES, con casos reales de alumnos |
| P5 Empaquetado | python-build-standalone + wheels fijados, variante `full-python`, asistente de primer arranque, prueba de humo |

### M2 · C++ (tracks en paralelo)

| Track | Entrega |
|---|---|
| C1 Compilador | familias g++ y clang++, compilar → ejecutar bajo PTY, carpeta con varios `.cpp` |
| C2 Depurador | lldb-dap y `gdb -i=dap` por `runInTerminal`; paradas por SIGSEGV como `StopException` |
| C3 Inteligencia | clangd y clang-format |
| C4 Assistant | parser GNU, errores del enlazador, caídas, `catalog.cpp.json` con patrones de GCC y Clang |
| C5 Empaquetado | LLVM por plataforma (o WinLibs en Windows), variante `full-cpp`, prueba de humo |

### Rust (planificado como M3)

Rust se planifica como el milestone **M3** en [PLAN_RUST.md](PLAN_RUST.md), después de M2 y con la
misma arquitectura (adaptador `adapters/rust/`, sin tocar el núcleo): rustc y cargo vía rustup (en
Windows sólo el toolchain GNU, libre), rust-analyzer, clippy, rustfmt y el mismo `lldb-dap` de C++. No se
empaqueta (pesa más de 600 MB). La fase "Didáctica y release" de abajo pasa a ser M4.

### M4 · Didáctica y release

Plantillas y ejemplos por lenguaje, documentación EN/ES por lenguaje, página web, matriz QA de
tres sistemas por tres lenguajes, release **3.0**.

### Por qué Python antes que C++

Python tiene el toolchain más simple (no compila, no enlaza, no hay familias de compilador), su
adaptador DAP es el más usado del mundo, y es el lenguaje **más distinto** de Go: sin build, con
REPL nativo, con pip. Si el diseño de perfiles aguanta Python, aguanta C++. C++ concentra lo más
duro del empaquetado (tamaño en Windows, dos familias) y de la explicación de errores (enlazador,
comportamiento indefinido, caídas), y conviene abordarlo con el núcleo ya probado.

### Tamaño aproximado del trabajo

| Fase | Backend Go | Frontend | Otros |
|---|---|---|---|
| M0 | ≈ 1.500 líneas movidas, ≈ 600 nuevas | ≈ 400 líneas | reglas de lint, tests |
| M1 Python | ≈ 1.200 líneas | ≈ 150 | script REPL de ≈ 80 líneas, catálogo de 25 entradas, empaquetado |
| M2 C++ | ≈ 1.100 líneas | ≈ 100 | catálogo, empaquetado |

Son órdenes de magnitud para comparar fases, no compromisos.

## 8. Riesgos y decisiones que corresponden a los profesores

1. **Familia de C++.** Recomendación: el adaptador soporta las dos familias; la variante `full`
   empaqueta LLVM (un solo proveedor, Apache-2.0, lldb-dap y clangd incluidos) y la `lite` usa el
   GCC que ya tengan los laboratorios. Si el curso exige los mensajes de GCC, se invierte.
2. **LSP de Python.** pylsp (puro Python, ligero) frente a pyright o basedpyright (mejor análisis,
   pero necesita Node o un paquete más pesado). Recomendación: pylsp en la primera versión.
3. **PTY.** Es el riesgo técnico principal de M0 y por eso va primero como prueba de concepto. Sin
   PTY se puede publicar con tuberías, pero la salida perdida en las caídas y la salida tardía al
   depurar confundirán a los principiantes de C++; si la prueba falla en Windows, se publica con
   tuberías y la PTY queda como mejora registrada.
4. **GDB como DAP.** Su modo DAP está escrito en Python; hay que comprobar que las distribuciones
   de Windows (WinLibs, MSYS2) lo incluyan. lldb-dap no tiene ese problema.
5. **Memoria.** Tres servidores LSP a la vez pueden pasar de 500 MB. El arranque perezoso y el
   apagado por inactividad son parte de M0, no un extra.
6. **Contrato congelado.** Romperlo es inevitable; hacerlo una sola vez (v3) y antes de M1.
7. **Calidad de los catálogos.** Las 25 entradas de Go salieron de errores reales. Para Python y C++
   hace falta un corpus de errores de alumnos; los profesores pueden recogerlo este ciclo.
8. **Windows.** Las distribuciones MinGW disparan falsos positivos de antivirus y el instalador ya
   avisa de SmartScreen; el tamaño de `full-cpp` sube el tiempo de descarga en aulas.
9. **macOS.** El clang de Apple no se puede redistribuir; se empaqueta el LLVM oficial o la `lite`
   usa las Command Line Tools instaladas.
10. **pip en el intérprete empaquetado.** Funciona porque la instalación es por usuario (así lo
    hace Thonny). No se propone un venv por proyecto en la primera versión.
11. **Comportamiento indefinido en C++.** Un índice fuera de rango no da error; `-fsanitize=address`
    lo explicaría, pero no existe en MinGW. Queda para M3 en macOS y Linux.

## 9. Lo que no se propone

- Cambiar Wails o Svelte, ni volver a PyQt.
- Un sistema de plugins externo: los perfiles viven en el repositorio y se registran en `main.go`.
- CMake, venv por proyecto, cling o gestores de paquetes de C++ en la primera versión.
- Tocar la edición 1.x.
