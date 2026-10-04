# Plan M2 · C++ en VizcachaIDE

> Requiere [PLAN_NUCLEO_MULTILENGUAJE.md](PLAN_NUCLEO_MULTILENGUAJE.md) (M0) terminado: contrato v3,
> `internal/protocol`, `LanguageRegistry` y el frontend por perfiles. [PLAN_PYTHON.md](PLAN_PYTHON.md)
> (M1) no es obligatorio, pero si ya existe sirve de segundo ejemplo de perfil. Diseño general en
> [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §4.5 y en el diagrama
> [arquitectura-multilenguaje.svg](arquitectura-multilenguaje.svg). Escrito el 2026-10-03.

## 0. Cómo usar este plan en un chat nuevo

1. Comprueba que M0 está integrado: existe `wails/internal/protocol/`, `wails/support_go.go` y
   `app.LanguageRegistry`. Si no, ejecuta primero el plan M0.
2. Lee: este archivo; [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §4 y §8;
   [`wails/README.md`](../wails/README.md); [`docs/wails/PLAN_WAILS.md`](wails/PLAN_WAILS.md) §2 y §4;
   [`docs/wails/UX_COPY.md`](wails/UX_COPY.md); el adaptador de Go en `wails/internal/adapters/golang/`
   y, si existe, el de Python en `adapters/python/`.
3. Para desarrollar y probar hace falta en la máquina un compilador (`g++` o `clang++`), `lldb-dap`,
   `clangd` y `clang-format`. En la máquina de desarrollo están en `wails/.toolchain-dev/` (ignorado por
   git): **llvm-mingw 20260922** (LLVM 23.1.2, con clang++, lldb-dap, clangd y clang-format) y
   **WinLibs GCC 16.2.0** (`mingw64/bin/g++.exe`, GCC real: el `g++.exe` de llvm-mingw es clang). Los
   tests los encuentran con `VIZCACHA_TEST_LLVM_BIN` y `VIZCACHA_TEST_GCC_BIN` (carpetas `bin`) a
   través de `adapters/cpp/cpptest`, y se saltan solos si faltan.
4. Verificación antes de cada commit:
   ```sh
   cd wails && go vet ./... && go test ./... && golangci-lint run
   cd frontend && npm run check && npm run test
   wails build
   ```
5. Reglas de siempre: archivos de menos de 200 líneas, funciones de menos de 50, retorno temprano,
   nombres de dominio, library-first, textos sólo por i18n. Commits locales en inglés, sin push salvo
   indicación.
6. Rama `m2-cpp` desde `main`. Orquestador **Opus** (C0, integración, CCR, textos, QA); coders
   **Sonnet** en worktrees (C1–C6), con el prompt de §12 y el reparto de M0/M1 (`go.mod`,
   `.golangci.yml`, `ux_copy.json` y locales son del orquestador).
7. Resultado esperado: **versión 2.3.0** con C++ como lenguaje, la variante `full-cpp` para Windows y
   las actualizaciones automáticas (ya en `main`).

## 1. Objetivo

Que un principiante escriba, compile, ejecute, entienda sus errores (de compilación, de enlace y
caídas en ejecución) y depure un programa C++ paso a paso, con compiladores libres y sin instalar
nada más que VizcachaIDE en Windows.

## 2. Alcance

Dentro:

- Compilar y ejecutar un `.cpp` (con entrada por teclado), o una carpeta con varios `.cpp` como
  proyecto; Build sin ejecutar; argumentos; archivo sin título.
- Dos familias de compilador con un solo adaptador: GCC (`g++`) y LLVM (`clang++`).
- Depurar con `lldb-dap`: breakpoints, pasos, continuar, ejecutar hasta aquí, variables, pila, vista
  de llamadas, hilos, parada en caída (SIGSEGV, división entre cero, abort).
- Inteligencia de código con `clangd`; formato con `clang-format`.
- Catálogo de unas 25 errores de C++ explicados en EN/ES: compilador (GCC y Clang), enlazador y
  caídas en ejecución.
- Empaquetado `full-cpp` para Windows con LLVM (llvm-mingw); en macOS y Linux, uso del compilador
  del sistema con guía en el primer arranque.

Fuera (primera versión): CMake o Makefiles, bibliotecas externas y gestores de paquetes (vcpkg,
Conan), consola interactiva (cling pesa cientos de MB), sanitizers, C como lenguaje aparte (un `.c`
se puede compilar con el mismo adaptador en una versión posterior), GDB como depurador salvo lo
que diga el spike C2b.

## 3. Decisiones (library-first) y licencias

| Necesidad | Decisión | Licencia | Por qué |
|---|---|---|---|
| Compilador | **clang++** (LLVM) o **g++** (GCC), detectados; el adaptador soporta las dos familias | Apache-2.0 con excepciones LLVM; GPLv3 con excepción de runtime | los cursos usan g++; LLVM da un solo proveedor para compilador, depurador, LSP y formato |
| Depurador | **lldb-dap** (viene con LLVM; en macOS con las Command Line Tools de Xcode 16 o más) | Apache-2.0 | DAP nativo; `runInTerminal` para que `cin` y la salida funcionen como en una terminal |
| Depurador alternativo | `gdb -i=dap` (GDB 14 o más, compilado con Python) | GPLv3 | para equipos con GCC y sin LLDB; sólo si el spike C2b lo valida |
| Inteligencia | **clangd** | Apache-2.0 | funciona con los dos compiladores; `--query-driver` para los headers de g++ |
| Formato | **clang-format** | Apache-2.0 | estándar de facto; respeta `.clang-format` si el profesor lo pone |
| Windows | **llvm-mingw** (clang++, lld, lldb, lldb-dap, clang-format, libc++) podado a x86_64 | Apache-2.0 + MinGW-w64 (licencias permisivas) | un solo zip con todo; sin instalador; alternativa WinLibs (GCC + GDB) |
| Editor | `@codemirror/lang-cpp` | MIT | igual que `lang-go` |

Sobre la GPL: si se empaqueta GCC o GDB van como programas aparte con sus licencias en
`toolchain/licenses/`; el IDE sigue siendo MIT y los programas de los alumnos no se ven afectados.

## 3.1 Decisiones de la revisión (2026-10-04, antes de ejecutar)

Contrastado con el código de M0 y M1. C0 hace los cambios de contrato antes de lanzar los tracks:

1. **Línea de caída.** El Supervisor no puede añadir texto tras la última etapa. `process.Job` gana
   `Finished func(exitCode int) string`: si devuelve texto, se emite como `stderr` antes de
   `run:finished` (la línea `Segmentation fault`… de §4.3 la produce el runner con ella).
2. **Comprobador.** C++ tiene `CodeChecker`: `<cxx> -fsyntax-only <flags> <fuentes>` con los mismos
   avisos que la compilación (`-Wall -Wextra`). `Capabilities.Check` = true. Así los avisos llegan a
   Problemas explicados tras una ejecución correcta, igual que `go vet` y `ruff check`. El parser de
   §4.7 ya entiende esa salida.
3. **Estado del ayudante de código por lenguaje.** El evento `lsp:status` pasa a llevar
   `{codeLanguage, status}` (contrato). `app.EventSink.LanguageServerStatus(codeLanguage, status)`,
   `lsp.Options.CodeLanguage`, y en el frontend un estado por lenguaje: la barra muestra el del
   archivo activo. Cierra la limitación conocida de M1.
4. **Sin GDB en M2.** Se quita el spike C2b y `gdbdap/`. Con GCC y sin `lldb-dap` (copia `lite` con
   Dev-C++ o Code::Blocks) se compila y ejecuta pero no se depura; el aviso `errors.lldbDapNotFound`
   lo dice.
5. **Teclado al depurar.** `Capabilities.DebugInput` = true: lldb-dap con `runInTerminal` arranca el
   programa en el Supervisor (PTY) con `Job.Events` → `debug:output`, como debugpy. C0 lo comprueba
   en Windows con la versión de llvm-mingw elegida (riesgo §11). **Comprobado en C0 (2026-10-04,
   llvm-mingw 20260922 / LLVM 23.1.2):** lldb-dap envía `runInTerminal` con un lanzador
   `lldb-dap.exe --comm-file \\.\pipe\lldb-dap-run-in-terminal-comm-<n> --launch-target <exe>` y
   para en el breakpoint, con y sin
   `runInTerminal`. **Requisito: compilar con `-static`**; sin él el `.exe` no arranca fuera del IDE
   ni al depurar (`0xc0000135`, falta `libc++.dll`), confirmado por el spike.
6. **APIs reales:** `dap.StdioTransport`, `dap.NewSession(SessionDeps)`, `ReverseHandler`,
   `lsp.New(sink, flavor, lsp.Options{Name: "clangd", LanguageID: "cpp", ...})`,
   `errorcatalog.NewExplainer`, `toollocator.Tool`, `process.Job{Then, Events, Mode}`,
   `process.PrepareTree/KillTree`. `lldbdap/` puede importar `adapters/cpp/runner` para compilar.
7. **depguard:** regla `cpp-adapter-stays-in-cpp` (un `deny` por carpeta hermana) y exclusión de
   `adapters/cpp/**` en `adapters-are-independent`.
8. **Actualizaciones:** `updates.DetectInstallation` reconoce `toolchain/cpp`: variantes `full-cpp`
   y `full` (Go + Python + C++). Una copia 2.2 `full` (Go + Python) se actualiza a la nueva `full`.
9. **LLDB compartido (pensando en Rust, M3).** El flavor de lldb-dap (launch, `runInTerminal`, motivos
   de parada, filtro de frames del sistema, descripciones de caída) vive en `protocol/dap/lldb`, no en
   `adapters/cpp`: los adaptadores no pueden importarse entre sí y Rust usará el mismo depurador.
   `adapters/cpp/lldbdap` sólo añade la etapa de compilación y lo propio de C++.
10. **QA:** el arnés E2E gana una fase C++ (compilar y ejecutar con `cin`, error explicado, caída
   explicada, depurar con variables y teclado, aviso de `-Wall` en Problemas) en EN/ES.

## 4. Diseño: `wails/internal/adapters/cpp/`

```
adapters/cpp/
├── profile.go        LanguageProfile de C++
├── locator.go        compilador, lldb-dap, clangd, clang-format: configurado → empaquetado → PATH
├── family.go         GCC o Clang según el ejecutable; flags por familia y sistema
├── runner/           ProgramRunner: compilar (etapa 1) → ejecutar (etapa 2) con protocol/process
├── lldbdap/          Flavor DAP + Transport stdio + ReverseHandler (runInTerminal)
├── gdbdap/           (sólo si C2b lo valida) Flavor DAP para gdb -i=dap
├── clangd/           Flavor LSP
├── clangformat/      Format
└── errors/           OutputParser de diagnósticos GNU, enlazador y caídas + data/catalog.cpp.json
```

### 4.1 Perfil

```go
var Profile = domain.LanguageProfile{
    ID: domain.CodeLanguageCpp, NameKey: "codeLanguage.cpp",
    Extensions: []string{".cpp", ".cc", ".cxx", ".c++", ".h", ".hpp", ".hh"},
    Indent: domain.IndentStyle{UseTabs: false, Size: 4},
    Capabilities: domain.Capabilities{Build: true, Console: false, Format: true, Check: true, DebugInput: true,
        PackageActions: nil, ThreadsLabel: "debug.threads"},  // Console y Packages: nil
    Tools: []domain.ToolSpec{
        {ID: "cxx", Role: domain.RoleCompiler, LabelKey: "settings.toolCxx", MissingKey: "errors.cxxNotFound", InstallURL: "https://winlibs.com/"},
        {ID: "lldb-dap", Role: domain.RoleDebugAdapter, LabelKey: "settings.toolLldbDap", MissingKey: "errors.lldbDapNotFound"},
        {ID: "clangd", Role: domain.RoleLanguageServer, LabelKey: "settings.toolClangd", MissingKey: "errors.clangdNotFound", InstallURL: "https://clangd.llvm.org/installation"},
        {ID: "clang-format", Role: domain.RoleFormatter, LabelKey: "settings.toolClangFormat", MissingKey: "errors.clangFormatNotFound"},
    },
}
```

`Check` = `-fsyntax-only` con los mismos avisos (§3.1 punto 2): tras una ejecución correcta los avisos
de `-Wall -Wextra` llegan a Problemas explicados. Los `InstallURL` cambian por sistema en `support_cpp.go`
(macOS: `xcode-select --install` como `InstallCommand`; Linux: `sudo apt install g++ lldb clangd clang-format`).

### 4.2 Localizador y familia (`locator.go`, `family.go`)

- Compilador: `ToolPaths["cxx"]` → empaquetado `toolchain/cpp/bin/clang++(.exe)` → PATH en el orden
  `g++`, `clang++`, `c++`. La familia se deduce del nombre del ejecutable (`g++`, `x86_64-w64-mingw32-g++`
  → GCC; `clang++`, `c++` en macOS → Clang) y se confirma con `--version` (contiene `clang` o `g++`).
- macOS: antes de invocar `clang++`, `xcode-select -p` debe responder; si no, el compilador cuenta
  como ausente (invocarlo abriría el diálogo de instalación de las Command Line Tools).
- `lldb-dap`, `clangd`, `clang-format`: configurado → `toolchain/cpp/bin/` → junto al compilador
  (misma carpeta `bin`) → PATH. En macOS, `lldb-dap` se busca también con `xcrun -f lldb-dap`.
- Flags por familia y sistema (`family.go`):

  | | GCC | Clang |
  |---|---|---|
  | comunes | `-std=c++17 -g -O0 -Wall -Wextra -fdiagnostics-color=never` | igual |
  | Windows | `-static-libgcc -static-libstdc++` (el `.exe` corre sin DLL en el PATH) | `-static` con llvm-mingw (libc++ y libunwind dentro) |
  | macOS | | `-fno-color-diagnostics` ya cubierto; nada más |
  | Linux | | |

  El estándar es configurable en `Settings` más adelante; en esta versión, C++17 fijo.

### 4.3 Ejecutor (`runner/`)

- `Configure(path, args)`: si la carpeta tiene más de un archivo fuente (`.cpp`, `.cc`, `.cxx`),
  `Mode: project`, `Target: carpeta`, `Project: {Root: carpeta, Kind: "folder", Name: nombre}`; si
  no, `Mode: file`. Un `.h` o `.hpp` siempre resuelve a su carpeta como proyecto (si la carpeta no
  tiene fuentes, `ErrNoSources` con el aviso `errors.cppHeaderOnly`).
- `Run`: un `Job` de dos etapas con `Then`, en el `Supervisor` compartido (M0 §4.1):
  1. **Compilar**: `<cxx> <flags> -o <salida> <fuentes…>` con `Dir: carpeta`. La salida va a
     `<UserCacheDir>/VizcachaIDE/build/<hash de la carpeta>/<nombre>(.exe)` para no ensuciar la
     carpeta del alumno. El stderr del compilador se emite como `stderr` (el Assistant lo lee). Si
     el código de salida no es 0, `run:finished` con ese código y no hay segunda etapa.
     `Notice` = `run.compiling` tras 2 s de silencio.
  2. **Ejecutar**: el binario con `Mode: Terminal` (`Echo: true`) y los argumentos del programa.
- `Build`: sólo la etapa 1, con la salida junto al fuente (`main.exe` / `main`), como `go build`.
- `RunUntitled`: escribe `main.cpp` en `<temp>/vizcacha_cpp_<pid>/` y ejecuta; `Cleanup` borra la
  carpeta.
- **Caídas**: cuando el programa termina por señal o por código de excepción de Windows, el runner
  escribe una última línea en `stderr` con el texto que mostraría una terminal (vía
  `Job.Finished`, §3.1 punto 1) y el Assistant la reconoce:

  | Causa | Unix (`ProcessState` con señal) | Windows (código de salida) | Línea emitida |
  |---|---|---|---|
  | acceso inválido | SIGSEGV, SIGBUS | `0xC0000005` (3221225477) | `Segmentation fault` |
  | desbordamiento de pila | SIGSEGV con pila agotada (no distinguible) | `0xC00000FD` (3221225725) | `Stack overflow` |
  | división entera entre cero | SIGFPE | `0xC0000094` (3221225620) | `Floating point exception` |
  | `abort()` / `terminate` | SIGABRT | `0xC0000409`, `3` | `Aborted` (el mensaje `terminate called after throwing…` ya viene en stderr) |

- `CodeChecker` (§3.1 punto 2): `<cxx> -fsyntax-only <flags> <fuentes>`, sin eventos y fuera del
  Supervisor; devuelve la salida del compilador ("" si no hay avisos). El formato lo da
  `clangformat/` como `app.CodeFormatter` (§4.6).

### 4.4 Depurador (`lldbdap/`)

- Transport: stdio sobre `lldb-dap`. `Flavor.AdapterID()` = `"lldb"`.
- Antes del `launch`, el adaptador compila con la etapa 1 del runner (misma salida en caché). Un
  error de compilación termina la sesión con `debug:terminated` y la salida del compilador por
  `debug:output` como `stderr`, para que el Assistant explique.
- `Launch`:
  ```json
  {"request":"launch","program":"<binario>","cwd":"<carpeta>","args":[...],"env":{...},
   "stopOnEntry":false,"runInTerminal":true,"initCommands":["settings set target.disable-aslr false"]}
  ```
  Con `runInTerminal: true`, lldb-dap envía la petición inversa `runInTerminal` con el comando de su
  lanzador; el `ReverseHandler` lo arranca con `protocol/process` en modo PTY y devuelve el pid. Si la
  PTY no existe (M0 la dejó en `Pipes`), `runInTerminal: false`: la salida llega por eventos `output`
  y no hay teclado (aviso `run.debugStdin`).
- `StopReason`: `breakpoint`, `step`, `exception` (con la descripción de LLDB: `EXC_BAD_ACCESS`,
  `signal SIGSEGV`, `Exception 0xc0000005`) → `StopException`; la descripción se traduce a la misma
  frase de §4.3 (`Segmentation fault`…) más el texto original.
- `KeepFrame`: oculta frames sin `Source.Path` o cuya ruta no está en la carpeta del proyecto
  (`libc`, `libstdc++`, `ntdll`), salvo el primero si no queda ninguno. `IsLocalsScope`: `Locals`.
  `KeepVariable`: todos; para `std::string` y contenedores, el valor mostrado es el `value` que
  lldb-dap resume con sus formateadores (libc++ completos; libstdc++ parciales, ver riesgos).
- `Output`: `stdout`/`stderr` del programa; el resto a `console`.

### 4.5 Spike C2b: GDB como DAP (descartado en M2, §3.1 punto 4)

Dos días como máximo: con un WinLibs (GCC + GDB 14 o más) comprobar que `gdb -i=dap` arranca (su
modo DAP está escrito en Python y necesita un GDB compilado con Python), que `launch` + breakpoints
+ `next` + `variables` funcionan y cómo entrega la salida del programa. Si funciona, `gdbdap/` es un
segundo `Flavor` con el mismo `protocol/dap`; si no, se documenta que la variante `lite` con GCC
depura sólo si hay `lldb-dap` en el PATH.

### 4.6 Inteligencia y formato (`clangd/`, `clangformat/`)

- `Command`: `clangd --background-index=false --header-insertion=never --completion-style=detailed
  --log=error` y, cuando el compilador es GCC, `--query-driver=<ruta de g++>` para que clangd
  encuentre sus headers. `RootOf(path)` = carpeta del archivo.
- `InitializationOptions()`: `{"fallbackFlags": ["-std=c++17", "-Wall"]}`; si existe
  `compile_flags.txt` o `compile_commands.json` en la carpeta, clangd lo usa solo.
- El mapeo genérico de `protocol/lsp` sirve; grabar fixtures de clangd real (`testdata/clangd/*.json`)
  para diagnósticos con rango, completado con `detail`, hover (markdown), firma, símbolos (clangd
  anida métodos dentro de clases).
- `Format`: `clang-format --assume-filename=<nombre> --style=file --fallback-style="{BasedOnStyle: LLVM, IndentWidth: 4}"`
  con el texto por stdin. Un `.clang-format` en la carpeta del profesor manda.

### 4.7 Errores (`errors/`)

Parser (`parser.go`, con `protocol/errorcatalog.{SplitLines,LocationFrom,NamedGroups,ResolvePath}`):

- Diagnóstico GNU, común a GCC y Clang:
  `^(?P<path>(?:[A-Za-z]:)?[^:\r\n]+?):(?P<line>\d+):(?P<column>\d+): (?P<kind>error|warning|note|fatal error): (?P<message>.*)$`.
  `error` y `fatal error` → `SeverityError`; `warning` → `SeverityWarning`; `note` se anexa al
  `RawText` del diagnóstico anterior. Las líneas de contexto (`In function 'int main()':`,
  `   12 | código`, `      |    ^~~~`, la línea de código y el `^` de Clang, `N error(s) generated.`)
  se anexan al `RawText` o se ignoran. `Source` = `"compiler"`.
- Enlazador: GNU ld `…: undefined reference to \`(?P<symbol>[^']+)'` y
  `.../ld: … (.text+0x…)`; lld `ld.lld: error: undefined symbol: (?P<symbol>.+)` seguido de
  `>>> referenced by …`. `Source` = `"linker"`, sin columna, con el archivo si aparece
  (`main.cpp:(.text+0x1a)`). `collect2: error: ld returned 1 exit status` es ruido.
- Caídas: la línea emitida por el runner (§4.3) y `terminate called after throwing an instance of
  '(?P<type>[\w:]+)'` + `  what():  (?P<what>.*)`. `Source` = `"runtime"`, sin ubicación (el
  depurador es quien la da).
- Rutas: GCC en Windows imprime `C:\carpeta\main.cpp` o `main.cpp`; `ResolvePath` ya las une a la
  carpeta de trabajo.

Catálogo `data/catalog.cpp.json`; cada entrada lleva patrones para **las dos familias**. Ids
propuestos, unas 25 entradas:

| Id | GCC | Clang |
|---|---|---|
| `CPP-UNDECLARED` | `'(?P<name>\w+)' was not declared in this scope` | `use of undeclared identifier '(?P<name>\w+)'` |
| `CPP-COUT-UNDECLARED` | `'(?P<name>cout\|cin\|endl\|string\|vector)' was not declared` (sin `#include` o sin `std::`) | `use of undeclared identifier '(?P<name>cout\|cin\|endl)'`; `no member named '(?P<name>\w+)' in namespace 'std'` |
| `CPP-EXPECTED-SEMICOLON` | `expected ';' (before\|after) …` | `expected ';' after …` |
| `CPP-EXPECTED-BRACE` | `expected '}' at end of input` / `expected unqualified-id` | `expected '}'` / `expected expression` |
| `CPP-EXPECTED-PRIMARY` | `expected primary-expression before '(?P<token>[^']+)'` | `expected expression` |
| `CPP-NO-MATCHING-FUNCTION` | `no matching function for call to '(?P<call>[^']+)'` | `no matching function for call to '(?P<call>[^']+)'` |
| `CPP-TOO-FEW-ARGS` | `too few arguments to function '(?P<function>[^']+)'` | `too few arguments to function call` |
| `CPP-TOO-MANY-ARGS` | `too many arguments to function '(?P<function>[^']+)'` | `too many arguments to function call` |
| `CPP-CANNOT-CONVERT` | `cannot convert '(?P<from>[^']+)' to '(?P<to>[^']+)'` | `cannot initialize … of type '(?P<to>[^']+)' with … of type '(?P<from>[^']+)'` |
| `CPP-INVALID-OPERANDS` | `invalid operands of types '(?P<left>[^']+)' and '(?P<right>[^']+)' to binary 'operator(?P<op>[^']+)'` | `invalid operands to binary expression \('(?P<left>[^']+)' and '(?P<right>[^']+)'\)` |
| `CPP-NO-MEMBER` | `'(?P<type>[^']+)' has no member named '(?P<name>\w+)'` | `no member named '(?P<name>\w+)' in '(?P<type>[^']+)'` |
| `CPP-REDECLARED` | `redeclaration of '(?P<name>[^']+)'` / `redefinition of '(?P<name>[^']+)'` | `redefinition of '(?P<name>\w+)'` |
| `CPP-MISSING-RETURN` | `control reaches end of non-void function` (aviso) | `non-void function does not return a value` |
| `CPP-NO-SUCH-FILE` | `(?P<file>[^:]+): No such file or directory` (fatal error) | `'(?P<file>[^']+)' file not found` |
| `CPP-UNUSED-VARIABLE` | `unused variable '(?P<name>\w+)'` | `unused variable '(?P<name>\w+)'` |
| `CPP-UNINITIALIZED` | `'(?P<name>\w+)' (may be\|is) used uninitialized` | `variable '(?P<name>\w+)' is uninitialized when used here` |
| `CPP-SIGN-COMPARE` | `comparison of integer expressions of different signedness` | `comparison of integers of different signs` |
| `CPP-ASSIGN-IN-CONDITION` | `suggest parentheses around assignment used as truth value` | `using the result of an assignment as a condition without parentheses` |
| `CPP-ARRAY-BOUNDS` | `array subscript (?P<index>\d+) is above array bounds` | `array index (?P<index>\d+) is past the end of the array` |
| `CPP-STRING-COMPARE` | `comparison with string literal results in unspecified behavior` | `result of comparison against a string literal is unspecified` |
| `CPP-UNDEFINED-REFERENCE` | `undefined reference to \`(?P<symbol>[^']+)'` | `undefined symbol: (?P<symbol>.+)` |
| `CPP-UNDEFINED-MAIN` | `undefined reference to \`WinMain'` / `` `main' `` | `undefined symbol: main` |
| `CPP-SEGFAULT` | `^Segmentation fault` | igual (lo emite el runner) |
| `CPP-STACK-OVERFLOW` | `^Stack overflow` | igual |
| `CPP-DIVIDE-ZERO` | `^Floating point exception` | igual |
| `CPP-TERMINATE` | `terminate called after throwing an instance of '(?P<type>[\w:]+)'` | `libc\+\+abi: terminating due to uncaught exception of type (?P<type>[\w:]+)` |

Como en Go, el texto EN/ES dice qué pasó, por qué y cómo se arregla, y el mensaje original se
conserva. Pedir a los profesores errores reales de alumnos antes de redactar.

### 4.8 `support_cpp.go` (en `wails/`, paquete `main`)

```go
func newCppSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts) (app.LanguageSupport, func(ctx context.Context))
```

Recibe también el `Supervisor` compartido. `Console` y `Packages` son `nil`; `Formatter` es
`clangformat/` y `Checker` el `-fsyntax-only` del runner. `main.go` lo añade al registro después de Go (y de Python si existe) en lugar
del `app.UnavailableSupport` de 2.1, y borra el perfil provisional de C++ de
`support_unavailable.go`. `.golangci.yml` gana la regla `cpp-adapter-stays-in-cpp` (copia de la de
Go de M0 §4.6): `lldbdap/` puede importar `adapters/cpp/runner` para la etapa de compilación.

## 5. Frontend

| Pieza | Cambio |
|---|---|
| `package.json` | `@codemirror/lang-cpp` |
| `lib/editor/languageSupport.ts` | caso `cpp`: `cpp()`, 4 espacios; los `.h`/`.hpp` también |
| `lib/editor/templates.ts` | plantilla con `#include <iostream>`, `using namespace std;` y `cout << "Hola, C++" << endl;` (decisión didáctica: los cursos de introducción usan `using namespace std`) y plantilla en blanco |
| `lib/bridge/mock*.ts` | escenarios de C++: ejecución correcta, error de compilación (`CPP-UNDECLARED` explicado), caída (`Segmentation fault` explicada), depuración con frames de `factorial(int n)` |
| `lib/panels/OutputPanel.svelte` | los enlaces a `archivo:línea:col` ya existen para Go; comprobar que las rutas con `\` de Windows se reconocen |
| `lib/panels/BottomPanel.svelte` | sin pestaña Consola para C++ (capacidad apagada en M0; verificar) |
| `lib/shell/DebugToolbar.svelte` | el botón Build aparece (capacidad `build`) |
| `lib/shell/FirstRunWizard.svelte` | paso de herramientas: en macOS enseña `xcode-select --install`; en Linux el comando `apt`; en Windows ofrece la variante `full-cpp` o WinLibs |
| `lib/shell/AboutDialog.svelte` | créditos: LLVM (clang, lldb, clangd, clang-format), MinGW-w64, GCC si se empaqueta |

## 6. Textos nuevos (`tools/po2json/ux_copy.json`)

| Clave | EN | ES |
|---|---|---|
| `settings.toolCxx` | C++ compiler (g++ or clang++) | Compilador de C++ (g++ o clang++) |
| `settings.toolLldbDap` | Debugger (lldb-dap) | Depurador (lldb-dap) |
| `settings.toolClangd` | Code helper (clangd) | Ayudante de código (clangd) |
| `settings.toolClangFormat` | Formatter (clang-format) | Formateador (clang-format) |
| `errors.cxxNotFound` | No C++ compiler was found. Install one or choose it in Settings. | No se encontró un compilador de C++. Instala uno o elígelo en Ajustes. |
| `errors.lldbDapNotFound` | The debugger (lldb-dap) isn't available. Running still works. | El depurador (lldb-dap) no está disponible. Ejecutar sigue funcionando. |
| `errors.clangdNotFound` | The code helper (clangd) isn't available. Suggestions and live problems are off. | El ayudante de código (clangd) no está disponible. No habrá sugerencias ni problemas en vivo. |
| `errors.clangFormatNotFound` | clang-format isn't available, so the file was saved as it is. | clang-format no está disponible, así que el archivo se guardó tal cual. |
| `errors.cppHeaderOnly` | This is a header. Open a .cpp file of the same folder to run it. | Esto es un encabezado. Abre un archivo .cpp de la misma carpeta para ejecutarlo. |
| `errors.cppXcodeTools` | On macOS the compiler comes with the Command Line Tools. Run the command below in Terminal. | En macOS el compilador viene con las Command Line Tools. Ejecuta el comando de abajo en Terminal. |
| `run.compiling` | Compiling… | Compilando… |
| `run.compileFailed` | The program didn't compile. See the Assistant. | El programa no compiló. Mira el Assistant. |
| `run.crashed` | The program crashed. Debug it with F6 to see the line. | El programa se cayó. Depúralo con F6 para ver la línea. |
| `firstRun.cppWindows` | The full installer already includes a C++ compiler. | El instalador completo ya incluye un compilador de C++. |

## 7. Empaquetado

- Sólo Windows tiene `full-cpp` en esta versión. macOS usa las Command Line Tools (gratuitas, con
  clang, lldb-dap desde Xcode 16 y clangd) y Linux los paquetes de la distribución; el asistente de
  primer arranque lo explica. Empaquetar LLVM para macOS y Linux queda como mejora.
- `packaging/versions.toml` gana:
  ```toml
  [cpp.windows]
  distribution = "llvm-mingw"
  release = "YYYYMMDD"                        # tag de github.com/mstorsjo/llvm-mingw
  llvm_version = "xx.y.z"
  url = "https://github.com/mstorsjo/llvm-mingw/releases/download/{release}/llvm-mingw-{release}-ucrt-x86_64.zip"
  sha256 = "…"
  # Se conservan sólo estas carpetas y binarios (el zip trae cuatro targets):
  keep = ["bin/clang++.exe", "bin/clang.exe", "bin/ld.lld.exe", "bin/lld.exe", "bin/lldb-dap.exe",
          "bin/lldb.exe", "bin/clangd.exe", "bin/clang-format.exe", "bin/*.dll",
          "x86_64-w64-mingw32/", "lib/clang/", "include/"]
  ```
  Si la release de llvm-mingw no trae `clangd`, se añade desde
  `https://github.com/clangd/clangd/releases` (zip de Windows, Apache-2.0) con su propio sha256.
- `packaging/fetch_cpp.py`: descarga, verifica, extrae, **poda** según `keep` (quitar i686, armv7,
  aarch64 reduce el tamaño a menos de la mitad), copia licencias a `toolchain/licenses/` y escribe
  `VERSIONS.txt`. Layout: `toolchain/cpp/bin/…`.
- `wails/packaging/build_release.py`: variantes `full-cpp` y `full` (Go + Python + C++). NSIS sin
  cambios.
- Prueba de humo: `smoke_test.py --cxx <bundled>` compila y ejecuta un `hola.cpp` desde una carpeta
  con espacios y acentos en el nombre, y comprueba que `lldb-dap --version` responde.
- Tamaño: medir; objetivo menos de 200 MB comprimido tras la poda. Si no se consigue, `full` (los
  tres lenguajes) se publica aparte de `full-cpp` y el sitio web lo explica.
- Antivirus: los binarios de MinGW provocan falsos positivos a veces; `SHA256SUMS` y las notas de
  versión ya cubren cómo verificar la descarga.

## 8. Tests

- Unitarios Go: `family_test` (detección por nombre y por `--version`, flags por sistema),
  `locator_test` (orden, `xcode-select` falso), `runner_test` (Configure con uno o varios `.cpp`,
  header, caché de salida, dos etapas con `Then`, línea de caída por código de salida),
  `errors/parser_test` con `testdata/cpp_output/{gcc,clang}/*.txt` (error simple, con notas y caret,
  varios errores, enlazador GNU y lld, `terminate`, rutas de Windows), `catalog_test` (cada id tiene
  fixture GCC **y** Clang, placeholders válidos), `lldbdap/flavor_test` (launch JSON, filtro de
  frames, motivos de parada, descripciones de caída), `clangd/flavor_test` (query-driver con GCC,
  fallbackFlags).
- Integración (se saltan sin compilador o sin lldb-dap): compilar y ejecutar `hola.cpp` con `cin`;
  un programa que imprime y luego se cae muestra lo impreso y la línea `Segmentation fault`;
  depurar `factorial.cpp` hasta un breakpoint, un paso y variables; un error de compilación llega
  por `debug:output` y termina la sesión; clangd da un diagnóstico de identificador no declarado;
  `clang-format` de un archivo mal sangrado.
- Frontend: escenarios mock de C++, plantilla, botón Build visible, pestaña Consola ausente.
- QA manual (sección "C++" en `docs/wails/QA_WAILS.md`): la lista de §10.

## 9. Tracks y orden

```
C0 (orquestador) ──► C1 · C2 (+C2b) · C3 · C4 · C5 · C6 en paralelo ──► integración (support_cpp.go) ──► QA
```

| Track | Dueño de | Entrega | Hecho cuando |
|---|---|---|---|
| **C0 · Preparación y contrato** (orquestador) | §3.1 (`Job.Finished`, `lsp:status` por lenguaje, depguard, detección de variantes del actualizador, claves de §6), `adapters/cpp/profile.go`, `locator.go`, `family.go`, `cpptest`, fixtures | contrato compilando; perfil, localizador y familia con tests; `testdata/cpp_output/{gcc,clang}` grabado de GCC 16.2 y clang 23.1 reales; `runInTerminal` de lldb-dap probado en Windows | los demás tracks compilan contra el locator |
| **C1 · Compilar y ejecutar** | `adapters/cpp/runner` | §4.3 | `hola.cpp` con `cin` por la PTY; carpeta con dos `.cpp`; Build junto al fuente; caída → línea `Segmentation fault` |
| **C2 · Depurador** | `adapters/cpp/lldbdap` | §4.4 (sin GDB, §3.1 punto 4) | breakpoints, pasos, variables con "acaba de cambiar", pila sin frames del sistema, parada en SIGSEGV con la línea |
| **C3 · Inteligencia y formato** | `adapters/cpp/clangd`, `clangformat` | §4.6 | diagnósticos en vivo con g++ y con clang++ (query-driver), completado, hover, definición, símbolos anidados; formato al guardar |
| **C4 · Assistant** | `adapters/cpp/errors` | §4.7 | 25 ids con fixtures de las dos familias y textos EN/ES; enlazador y caídas explicados |
| **C5 · Frontend** | `frontend/src/lib/*` (§5) | editor, plantillas, mock, primer arranque, créditos | `npm run dev` muestra los cuatro escenarios de C++ sin backend |
| **C6 · Empaquetado** | `packaging/fetch_cpp.py`, `versions.toml`, `wails/packaging/*` | §7 | `full-cpp` compila, ejecuta y depura `hola.cpp` en una máquina limpia de Windows; tamaño medido y anotado |

Integración: C0 → C1 → C4 → C2 → C3 → C5 → C6.

## 10. Criterios de salida

Con `full-cpp` en una máquina Windows sin compilador, con `lite` en Linux con `g++` + `lldb-dap`
+ `clangd`, y con `lite` en macOS con Command Line Tools, en EN y ES:

1. Nuevo archivo de C++ → plantilla → F5 compila e imprime `Hola, C++`; la primera compilación
   muestra `Compilando…` si tarda.
2. Un programa con `cin` lee lo escrito en Output; el aviso aparece antes de escribir.
3. Quitar el `;` → el Assistant explica `CPP-EXPECTED-SEMICOLON` con la línea subrayada.
4. Usar `cout` sin `#include <iostream>` → `CPP-COUT-UNDECLARED` sugiere el include.
5. Llamar a una función declarada y no definida → `CPP-UNDEFINED-REFERENCE` explica el enlazador.
6. Escribir fuera de un arreglo hasta caer → lo impreso antes se ve, y la línea `Segmentation fault`
   se explica con la sugerencia de depurar; F6 para en la línea culpable con `StopException`.
7. Breakpoint en `factorial` → F6 para ahí; F7/F8/F9; Variables marca cambios; Llamadas muestra
   `factorial(n=3)`; Hilos muestra uno; una `std::string` se ve legible.
8. `std::vec` ofrece `vector`; hover sobre `push_back`; F12 salta a una función propia; un
   identificador mal escrito se subraya mientras se escribe.
9. Guardar con formato activado reacomoda llaves y sangría.
10. Una carpeta con `main.cpp` y `util.cpp` se compila junta; Build deja `main.exe` junto al fuente.
11. `-Wall` muestra una variable sin usar como aviso en Problemas tras ejecutar.
12. Un `.cpp`, un `.py` y un `.go` abiertos: cada uno ejecuta con su lenguaje; la pestaña Consola
    desaparece al pasar al `.cpp`.
13. Tests Go y frontend en verde; `golangci-lint` limpio; CI en los tres sistemas; `full-cpp` con
    tamaño anotado en las notas de versión.

## 11. Riesgos

| Riesgo | Mitigación |
|---|---|
| `runInTerminal` de lldb-dap no funciona en Windows en la versión empaquetada | probarlo en C0 con la release elegida; si falla, `runInTerminal: false` (salida por eventos, sin teclado) y anotar la versión mínima |
| LLDB con binarios de GCC (`lite` + WinLibs) | LLDB lee DWARF de GCC; los formateadores de libstdc++ son parciales (una `std::string` puede verse como estructura). Documentar y, si C2b funciona, ofrecer GDB |
| El `.exe` no arranca fuera del IDE por DLL ausentes | `-static-libgcc -static-libstdc++` (GCC) o `-static` (llvm-mingw); test de humo lo comprueba |
| Carpetas con espacios o acentos | nunca pasar comandos por shell; `exec.Command` con argumentos separados; test de humo con una ruta así |
| Diálogo de Xcode al invocar `clang++` sin Command Line Tools | `xcode-select -p` antes (§4.2) |
| clangd no encuentra los headers de g++ | `--query-driver` con la ruta exacta del compilador |
| Tamaño de `full-cpp` | poda de targets en `fetch_cpp.py`; variante aparte si supera el objetivo |
| Comportamiento indefinido que no se cae (índice fuera de rango sin error) | fuera de alcance; sanitizers en macOS y Linux como mejora futura |
| Antivirus marca MinGW | `SHA256SUMS`, código público y notas de versión, como ya se hace |
| Alumnos con Dev-C++ o Code::Blocks instalados | el localizador acepta su `g++` del PATH y la familia GCC lo cubre |

## 12. Prompt común para los coders

```text
You are a coder on VizcachaIDE (wails/: Go 1.25 + Wails v2 + Svelte 5 + CodeMirror 6), a beginner IDE,
bilingual EN/ES, that has a multi-language core (docs/PLAN_NUCLEO_MULTILENGUAJE.md). You work in an
isolated git worktree on track <C?> of docs/PLAN_CPP.md. First read that plan (sections 0, 3, 4 and
your track in 9), docs/EXTENSION_MULTILENGUAJE.md section 4, wails/README.md, docs/wails/PLAN_WAILS.md
sections 2 and 4, and the Go adapter in wails/internal/adapters/golang (and adapters/python if present)
as reference implementations.
- Edit ONLY your track's folders. The v3 contract (internal/domain, internal/app/ports.go,
  internal/bridge/events.go, frontend/src/lib/{events,domain}.ts, bridge/types.ts) and internal/protocol
  are read-only: request changes as a "Contract change request" in your report.
- Use internal/protocol (process, dap, lsp, errorcatalog, toollocator): never copy its code.
- Only free tools: GCC or LLVM (clang++, lldb-dap, clangd, clang-format), GDB. Both compiler families
  must work through the same adapter; every catalog entry needs a GCC and a Clang pattern. Versions live
  in packaging/versions.toml, never in code.
- Never build command lines through a shell: exec.Command with separate arguments (paths with spaces
  and accents are the norm in student folders).
- Library-first. Files under 200 lines, functions under 50, early return, domain names. Visible texts
  only through i18n (tools/po2json/ux_copy.json, then go run ./tools/po2json). Compiler messages are
  never translated.
- Tests that need a compiler, lldb-dap, clangd or clang-format must skip cleanly when they are missing.
- Verify: go vet ./... && go test ./... && golangci-lint run; cd frontend && npm run check && npm run test;
  wails build when your track affects it.
- When done: ONE local commit in English ending with "Co-Authored-By: Claude <noreply@anthropic.com>",
  no push. Report: commit hash, files, how to test, limitations, CCRs and new i18n keys.
```
