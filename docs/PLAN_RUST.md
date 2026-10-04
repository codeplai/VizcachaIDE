# Plan M3 · Rust en VizcachaIDE

> Requiere [PLAN_NUCLEO_MULTILENGUAJE.md](PLAN_NUCLEO_MULTILENGUAJE.md) (M0) y [PLAN_CPP.md](PLAN_CPP.md)
> (M2) integrados: M2 deja `protocol/dap/lldb` (el depurador que Rust reutiliza), `process.Job.Finished`
> y el evento `lsp:status` por lenguaje. [PLAN_PYTHON.md](PLAN_PYTHON.md) (M1) es el ejemplo de
> adaptador con paquetes y comprobador. Diseño general en
> [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §4.5 y en el diagrama
> [arquitectura-multilenguaje.svg](arquitectura-multilenguaje.svg). Escrito el 2026-10-04.

## 0. Cómo usar este plan en un chat nuevo

1. Comprueba que M2 está integrado en `main`: existen `wails/internal/adapters/cpp/`,
   `wails/internal/protocol/dap/lldb/`, `wails/support_cpp.go`, `process.Job.Finished` y
   `lsp.Options.CodeLanguage`. Si no, ejecuta primero el plan M2. Sin esos tres cambios de contrato, R0
   no puede empezar.
2. Lee: este archivo; [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §4 y §8;
   [`wails/README.md`](../wails/README.md); [`docs/wails/PLAN_WAILS.md`](wails/PLAN_WAILS.md) §2 y §4;
   [`docs/wails/UX_COPY.md`](wails/UX_COPY.md); [PLAN_CPP.md](PLAN_CPP.md) §3.1 y §4 (runner de dos
   etapas, flavor de LLDB, parser de diagnósticos); el adaptador de Python en
   `wails/internal/adapters/python/` (perfil con `ProvidedBy`, paquetes, comprobador) y el de Go en
   `adapters/golang/`.
3. Para desarrollar y probar hace falta **rustup** con el toolchain estable y los componentes `clippy`,
   `rustfmt` y `rust-analyzer` (`rustup component add clippy rustfmt rust-analyzer`), más `lldb-dap`
   (en la máquina de desarrollo viene en `wails/.toolchain-dev/` con llvm-mingw, igual que para M2). En
   Windows el toolchain es `stable-x86_64-pc-windows-gnu` (§3 y §4.2). Los tests lo encuentran con
   `VIZCACHA_TEST_CARGO_HOME` (la carpeta que contiene `bin/`) y `VIZCACHA_TEST_LLVM_BIN` a través de
   `adapters/rust/rusttest`, y se saltan solos si faltan. Hace falta red la primera vez (descarga de
   crates para el test de `cargo add`).
4. Verificación antes de cada commit:
   ```sh
   cd wails && go vet ./... && go test ./... && golangci-lint run
   cd frontend && npm run check && npm run test
   wails build
   ```
5. Reglas de siempre: archivos de menos de 200 líneas, funciones de menos de 50, retorno temprano,
   nombres de dominio, library-first, textos sólo por i18n. Commits locales en inglés, sin push salvo
   indicación.
6. Rama `m3-rust` desde `main` (con M2 dentro). Orquestador **Opus** (R0, integración, CCR, textos, QA);
   coders **Sonnet** en worktrees (R1–R5), con el prompt de §12 y el reparto de M0/M1/M2 (`go.mod`,
   `.golangci.yml`, `ux_copy.json` y locales son del orquestador; nadie borra una pieza transitoria que
   otro track usa).
7. Resultado esperado: **versión 2.4.0** con Rust como cuarto lenguaje. No hay variante nueva: Rust
   funciona en `lite` y en las demás con rustup instalado (§7).

## 1. Objetivo

Que un principiante escriba, compile, ejecute, entienda sus errores (del compilador, del préstamo de
valores y los `panic!` en ejecución) y depure un programa Rust paso a paso, en Windows, macOS y Linux,
con herramientas libres. Rust es el lenguaje cuyos errores más necesitan explicación: el compilador es
muy bueno pero habla de propiedad, préstamos y tiempos de vida, y un alumno de primer curso no los
conoce todavía. Ahí el Assistant bilingüe aporta más que en ningún otro lenguaje.

## 2. Alcance

Dentro:

- Ejecutar un `.rs` suelto con `rustc` (con entrada por teclado), o un proyecto Cargo con `cargo` cuando
  se encuentra un `Cargo.toml` hacia arriba; Build sin ejecutar; argumentos; archivo sin título.
- Depurar con `lldb-dap` (el mismo que C++): breakpoints, pasos, continuar, ejecutar hasta aquí,
  variables, pila, vista de llamadas, hilos, parada en `panic!`.
- Inteligencia de código con `rust-analyzer`; formato con `rustfmt`; avisos de `clippy` en Problemas
  después de ejecutar.
- Paquetes con Cargo: crear proyecto (`cargo init`), añadir, quitar y listar dependencias. Sin Tidy.
- Catálogo de unas 30 entradas de errores explicados en EN/ES (compilador y `panic` en ejecución) más
  cinco avisos, con el mensaje original sin traducir.
- Detección de rustup y guía de instalación en el primer arranque y en el aviso de herramienta ausente.

Fuera (primera versión): Rust embebido en el instalador (§7), WebAssembly y otros `--target`, ayuda
con runtimes asíncronos (`tokio`, `async-std`) más allá de explicar sus errores como cualquier otro,
workspaces de varios crates más allá de detectarlos (se ejecuta el crate del archivo abierto; no hay
selector de miembros), `build.rs` y macros procedurales como tema didáctico, `cargo test` y `cargo
bench`, consola interactiva (no hay REPL ligero: `evcxr` pesa cientos de MB), perfil `--release`,
`cargo publish`, documentación (`cargo doc`).

## 3. Decisiones (library-first) y licencias

La arquitectura no cambia: Rust es un adaptador nuevo `wails/internal/adapters/rust/` más
`wails/support_rust.go`. Los únicos añadidos al contrato son `domain.CodeLanguageRust = "rust"`,
`domain.ProjectCargo ProjectKind = "cargo"` (valor `"cargo"`) y la clave i18n `codeLanguage.rust`.
Extensión: `.rs`.

| Necesidad | Decisión | Licencia | Por qué |
|---|---|---|---|
| Compilador y gestor | **rustc + cargo**, instalados por **rustup** | MIT o Apache-2.0 | es la única distribución oficial; `cargo` ya trae `add`, `remove`, `tree` y `init` |
| Inteligencia | **rust-analyzer** (componente de rustup) | MIT o Apache-2.0 | el servidor LSP oficial del proyecto; diagnósticos, completado, hover, definición, símbolos, inlay hints |
| Comprobador | **clippy** (`cargo clippy` en proyectos, `clippy-driver` en archivos sueltos) | MIT o Apache-2.0 | los consejos de clippy son lo más parecido a `go vet` y `ruff check`; ver abajo |
| Formato | **rustfmt** | MIT o Apache-2.0 | el estándar; respeta `rustfmt.toml` |
| Depurador | **lldb-dap** (LLVM), el mismo de C++ | Apache-2.0 con excepciones LLVM | DAP nativo, `runInTerminal`; su flavor genérico ya vive en `protocol/dap/lldb` (PLAN_CPP.md §3.1 punto 9). `rust-gdb` se descarta como en M2 |
| Windows | sólo el toolchain **x86_64-pc-windows-gnu** | los de arriba + MinGW-w64 (permisiva) | el target MSVC necesita el enlazador de Visual Studio, que no es libre (regla del proyecto: compiladores siempre libres) |
| Editor | `@codemirror/lang-rust` | MIT | igual que `lang-go` |

**Windows y rustup.** `rustup-init.exe` instala por defecto el host `x86_64-pc-windows-msvc` y pide las
Visual Studio Build Tools. VizcachaIDE guía a lo contrario: instalar con `rustup-init.exe
--default-host x86_64-pc-windows-gnu --default-toolchain stable -y`, o, si el alumno ya tiene rustup
con MSVC, `rustup default stable-x86_64-pc-windows-gnu` (descarga el toolchain y lo deja como el
predeterminado). El toolchain GNU trae el componente `rust-mingw` con las bibliotecas y objetos de arranque
de MinGW-w64, pero el enlace llama a un `gcc` como enlazador: si el spike R0 comprueba que no basta
con lo que trae rustup, el localizador usa el `clang` de **llvm-mingw** (el que empaqueta `full-cpp`,
PLAN_CPP.md §7, o el que haya en el PATH) con `-C linker=<ruta>` en `rustc` y
`CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER` en `cargo`. Así el binario de LLVM sirve a C++ y a Rust, y
el DWARF que genera el target GNU lo lee LLDB sin PDB (otra razón para no usar MSVC: LLDB lee PDB sólo
en parte).

**Comprobador de archivos sueltos.** Se decide así:

- Proyecto (hay `Cargo.toml`): `cargo clippy --message-format=json --quiet --manifest-path <toml>`.
  Es lo que clippy soporta de verdad: resuelve dependencias y el `edition` del manifiesto. Va después de
  la ejecución, así no compite con ella por el bloqueo de `target/`.
- Archivo suelto: `clippy-driver --edition <e> --crate-type bin --emit=metadata --out-dir <caché>
  --error-format=json <archivo>`. `clippy-driver` es el binario que viene con el componente clippy (el
  proxy de rustup lo expone) y ejecuta rustc con los lints de clippy sin necesitar manifiesto; con
  `--emit=metadata` no enlaza, así que es rápido y no exige enlazador. Da los avisos de rustc más los de
  clippy en un solo paso. Alternativa descartada: crear un `Cargo.toml` temporal (lento y ensucia).
- Sin clippy instalado: el mismo comando con `rustc` (sólo avisos del compilador: variables sin usar,
  `mut` sobrante) y un aviso `errors.clippyNotFound`. Un spike de R3 verifica que `clippy-driver` por el
  proxy de rustup funciona fuera de Cargo en los tres sistemas.

Sobre las licencias: todo el toolchain de Rust es MIT o Apache-2.0 y nada se redistribuye en esta
versión (§7). LLVM sigue en `toolchain/licenses/` cuando se empaqueta `full-cpp` o `full`.

## 3.1 Decisiones de la revisión (2026-10-04, antes de ejecutar)

Contrastado con el código de M0, M1 y el plan de M2. R0 hace estos cambios antes de lanzar los tracks:

1. **Sin cambios de arquitectura.** El perfil, el `toollocator`, `process.Job{Then, Finished, Events}`,
   `dap.NewSession`, `lsp.New(sink, flavor, lsp.Options{Name: "rust-analyzer", LanguageID: "rust",
   CodeLanguage: "rust", ...})` y `errorcatalog.NewExplainer` son los de M2. Sólo el contrato de la
   cabecera de §3.
2. **Un escalón por herramienta, no una sola.** Rust no es como Python (un intérprete con módulos): el
   compilador, Cargo, rust-analyzer y lldb-dap son ejecutables distintos. `rustc` y `cargo` vienen
   juntos, pero se localizan por separado. clippy, rustfmt y rust-analyzer son componentes de rustup que
   aparecen como proxies en la misma carpeta.
3. **Diagnósticos en JSON.** La etapa de compilación pide `--error-format=json` (`rustc`) o
   `--message-format=json` (`cargo`). El parser de §4.7 entiende ese JSON (código estable, rango exacto
   y notas hijas) y el texto de `panic`. Para que el panel Output siga mostrando el texto de siempre, el
   runner convierte cada diagnóstico JSON en su campo `rendered` antes de emitirlo. Eso necesita un
   añadido pequeño al contrato: `process.Job.OutputFilter func(stream, line string) (string, bool)` (R0;
   transforma o descarta una línea). Si R0 prefiere no tocar `process`, la alternativa es compilar con
   el formato de texto y parsear la cabecera `error[E0382]:` y la línea `--> ruta:línea:columna`, que
   también lleva el código; el JSON se usaría sólo en el comprobador. Decisión de R0 tras un día de
   prueba; el catálogo no cambia.
4. **Línea de caída.** Un `panic!` sale con código 101 en todos los sistemas y su texto ya está en
   stderr: no hace falta línea extra. `Job.Finished` (PLAN_CPP.md §3.1 punto 1) sólo cubre las caídas
   que no pasan por `panic`: `unsafe` con acceso inválido (`Segmentation fault`) y `abort` sin mensaje
   (`Aborted`). El desbordamiento de pila de Rust imprime él mismo `thread 'main' has overflowed its
   stack` y `fatal runtime error: stack overflow`, que el parser reconoce.
5. **Teclado al depurar.** `Capabilities.DebugInput` = true, como C++: `runInTerminal` con el
   Supervisor compartido y `Job.Events` → `debug:output`.
6. **Rust-lldb y formateadores.** Los tipos de `std` (`Vec`, `String`, `Option`) se ven en crudo en LLDB
   sin los formateadores del proyecto Rust. Hay un spike (R2a) que los carga por `initCommands` con
   respaldo (§4.4).
7. **Estado del ayudante de código por lenguaje:** ya es de M2 (`lsp:status` lleva `codeLanguage`); R3
   sólo pasa `CodeLanguage: "rust"`. rust-analyzer tarda y consume (indexa el crate): se arranca
   perezoso y con `IdleTimeout` de 5 minutos como los demás.
8. **depguard:** regla `rust-adapter-stays-in-rust` (un `deny` por carpeta hermana: golang, python, cpp
   y los adaptadores neutrales) y exclusión de `adapters/rust/**` en `adapters-are-independent`.
   `adapters/rust/lldbdap` importa `protocol/dap/lldb`, no `adapters/cpp` (los adaptadores no se importan
   entre sí).
9. **Actualizaciones:** `updates.DetectInstallation` no cambia (no hay `toolchain/rust`); una copia
   `lite`, `full-cpp`, etc. conserva su variante y gana Rust sólo con rustup en el equipo.
10. **QA:** el arnés E2E (`wails/packaging/qa/e2e`) gana una fase Rust (`steps-rust.mjs`, como
    `steps-python.mjs`) en EN y ES; ver §8.
11. **Edición de Rust.** `rustc` suelto compila con `--edition 2024` si `rustc --version` es 1.85 o más
    y con `2021` si no. Un proyecto usa la edición de su `Cargo.toml`.
12. **Versión mínima.** `rustc` 1.70 o más: trae `cargo add` y `cargo remove` (1.62 y 1.66) y el protocolo
    `sparse` por defecto, que hace rápida la primera descarga de crates, y el formato actual de los
    mensajes de `panic` (1.73) se soporta junto al anterior.

## 4. Diseño: `wails/internal/adapters/rust/`

```
adapters/rust/
├── profile.go        LanguageProfile de Rust
├── locator.go        rustc, cargo, rust-analyzer, lldb-dap: configurado → ~/.cargo/bin → PATH
├── project.go        busca Cargo.toml hacia arriba; detecta workspace y bin
├── toolchain.go      versión (rustc --version), host (rustc -vV), edición, sysroot, enlazador
├── runner/           ProgramRunner: compilar (etapa 1) → ejecutar (etapa 2) con protocol/process
├── lldbdap/          flavor de Rust sobre protocol/dap/lldb (+ formateadores, parada en panic)
├── analyzer/         Flavor LSP de rust-analyzer
├── rustfmt/          Format
├── clippy/           Check (cargo clippy y clippy-driver)
├── packages/         PackageManager con cargo
└── errors/           OutputParser de JSON de rustc/cargo y de panics + data/catalog.rust.json
```

### 4.1 Perfil

```go
var Profile = domain.LanguageProfile{
    ID: domain.CodeLanguageRust, NameKey: "codeLanguage.rust", Extensions: []string{".rs"},
    Indent: domain.IndentStyle{UseTabs: false, Size: 4},
    Capabilities: domain.Capabilities{Build: true, Console: false, Format: true, Check: true, DebugInput: true,
        PackageActions: []domain.PackageAction{domain.PackageInit, domain.PackageAdd, domain.PackageRemove, domain.PackageList},
        ThreadsLabel: "debug.threads"},  // Tidy no existe en Cargo
    Tools: []domain.ToolSpec{
        {ID: "rustc", Role: domain.RoleCompiler, LabelKey: "settings.toolRustc", MissingKey: "errors.rustNotFound", InstallURL: "https://rustup.rs"},
        {ID: "cargo", Role: domain.RoleRuntime, LabelKey: "settings.toolCargo", MissingKey: "errors.cargoNotFound", InstallURL: "https://rustup.rs"},
        {ID: "rust-analyzer", Role: domain.RoleLanguageServer, LabelKey: "settings.toolRustAnalyzer", MissingKey: "errors.rustAnalyzerNotFound", InstallCommand: "rustup component add rust-analyzer"},
        {ID: "lldb-dap", Role: domain.RoleDebugAdapter, LabelKey: "settings.toolLldbDap", MissingKey: "errors.lldbDapNotFound", InstallURL: "https://github.com/llvm/llvm-project/releases"},
        {ID: "clippy", Role: domain.RoleCompiler, ProvidedBy: "cargo", LabelKey: "settings.toolClippy", MissingKey: "errors.clippyNotFound", InstallCommand: "rustup component add clippy"},
        {ID: "rustfmt", Role: domain.RoleFormatter, ProvidedBy: "cargo", LabelKey: "settings.toolRustfmt", MissingKey: "errors.rustfmtNotFound", InstallCommand: "rustup component add rustfmt"},
    },
}
```

`clippy` y `rustfmt` son componentes del mismo toolchain: llevan `ProvidedBy: "cargo"` (fila de estado y
aviso propios, sin "Elegir" en Ajustes), como `debugpy` con Python. `Tools()` devuelve un `ToolStatus` por
cada uno, con su versión (`cargo clippy --version`, `rustfmt --version`). `lldb-dap` reutiliza la clave
`errors.lldbDapNotFound` de M2, y su aviso dice que ejecutar sigue funcionando. Los `InstallURL` e
`InstallCommand` cambian por sistema en `support_rust.go` (§4.2: en Windows el texto del paso de rustup
lleva el host GNU).

### 4.2 Localizador y toolchain (`locator.go`, `toolchain.go`)

- Orden de búsqueda de `rustc`, `cargo` y `rust-analyzer`: `ToolPaths[id]` → `CARGO_HOME/bin` (si no está
  definida, `~/.cargo/bin`; en Windows `%USERPROFILE%\.cargo\bin`) → PATH. Rustup instala allí unos
  proxies (`rustc`, `cargo`, `rust-analyzer`, `rustfmt`, `cargo-clippy`, `clippy-driver`) que eligen el
  toolchain por `rustup default` o por un `rust-toolchain.toml` del proyecto; se ejecutan tal cual. La
  primera ejecución de un proxy en un proyecto con `rust-toolchain.toml` puede descargar el toolchain: es
  una etapa silenciosa larga, así que lleva el mismo aviso de silencio (`Notice`) que la compilación.
- `lldb-dap`: configurado → `toolchain/cpp/bin/` (cuando la variante lo trae, PLAN_CPP.md §7) → carpeta
  `bin` de `llvm-mingw` o LLVM en el PATH → PATH. Rustup no distribuye lldb-dap: sin él, Rust se compila y
  ejecuta pero no se depura, y el aviso lo dice (igual que GCC sin LLDB en M2).
- Validación de `rustc`: `rustc --version` con 5 s de tiempo límite; se exige 1.70 o más (§3.1 punto 12).
  Un proxy de rustup sin toolchain instalado responde con error pidiendo `rustup default`: el candidato
  cuenta como ausente y el aviso lo explica con `errors.rustupNoToolchain`.
- Triple del host (`toolchain.go`): `rustc -vV` da `host: x86_64-pc-windows-msvc`. En Windows con host
  `msvc` se muestra un aviso (no bloqueante) `errors.rustMsvcHost` con el comando `rustup default
  stable-x86_64-pc-windows-gnu`, porque si el alumno no tiene el enlazador de Visual Studio el enlace
  fallará con `linker \`link.exe\` not found` (entrada `RS-LINKER-MISSING` del catálogo, que da el mismo
  consejo). En macOS y Linux no hay aviso (el enlazador es `cc` del sistema; si falta, la misma entrada
  explica `xcode-select --install` o `build-essential`).
- `rustc --print sysroot` (una vez, en caché) da la carpeta para los formateadores de LLDB (§4.4) y para
  `rust-src` (la biblioteca estándar, que rust-analyzer usa para navegar a `Vec::push`; su ausencia sólo
  quita la navegación y no se avisa).
- Enlazador: en Windows GNU, si no hay `gcc` en el PATH y existe el `clang` de llvm-mingw, el
  localizador devuelve `-C linker=<ruta>` (rustc) y la variable de entorno de Cargo
  `CARGO_TARGET_X86_64_PC_WINDOWS_GNU_LINKER` (§3). El resultado se decide con el spike de R0.
- Entorno para todos los procesos de Rust: `CARGO_TERM_COLOR=never`, `NO_COLOR=1`, `RUST_BACKTRACE=0`
  (el principiante ve el mensaje del `panic` sin pila de núcleo; `RUST_BACKTRACE=1` queda como opción en
  una versión posterior) y `CARGO_INCREMENTAL=0` para ahorrar disco en `target/`.

### 4.3 Ejecutor (`runner/`)

- `project.go`: `FindProject(path)` sube desde la carpeta del archivo buscando `Cargo.toml`; el primero
  con tabla `[package]` es el proyecto (`Kind: domain.ProjectCargo`, `Name` = nombre del paquete,
  `Root` = su carpeta). Si el primero encontrado es un manifiesto virtual (`[workspace]` sin
  `[package]`), no hay proyecto y el aviso `errors.rustWorkspaceRoot` pide abrir un archivo dentro de un
  crate. Un crate dentro de un workspace se ejecuta con `--manifest-path`; el `target/` es el del
  workspace y no hay selector de miembros (fuera de alcance, §2).
- `Configure(path, args)`: si hay proyecto, `Mode: project`, `Target: Root`, `WorkingDir: Root`; si no,
  `Mode: file`, `Target: path`, `WorkingDir: carpeta del archivo`.
- `Run`: un `Job` de dos etapas con `Then`, en el `Supervisor` compartido:
  1. **Compilar**:
     - archivo suelto: `rustc --edition <e> -g -C opt-level=0 --error-format=json --color never -o
       <salida> <archivo>`, con `Dir: carpeta`. La salida va a
       `<UserCacheDir>/VizcachaIDE/build/<hash de la ruta>/<nombre>(.exe)` para no ensuciar la carpeta
       del alumno, como en C++.
     - proyecto: `cargo build --message-format=json --manifest-path <toml>` (más `--bin <nombre>` si el
       archivo abierto está en `src/bin/<nombre>.rs`). Cargo escribe en el `target/` del proyecto, que
       es lo habitual y lo que esperan los profesores; `cargo init` ya crea el `.gitignore`.
     - La salida pasa por el filtro de §3.1 punto 3: cada diagnóstico se emite como su texto `rendered`
       en `stderr`, la línea `compiler-artifact` con `executable` se guarda para la etapa 2 y se
       descarta, y `build-finished` decide el código. Si el código no es 0, `run:finished` con él y sin
       segunda etapa. `Notice` = `run.compiling` tras 2 s de silencio; la primera compilación de un
       proyecto con dependencias tarda de verdad (descarga y compila las crates), y el texto dice
       "Compilando… la primera vez tarda más".
  2. **Ejecutar**: el binario (la ruta guardada en proyectos; la de caché en archivos sueltos) con
     `Mode: Terminal` (`Echo: true`), `Dir` = la carpeta de trabajo y los argumentos del programa.
     `Job.Finished(exitCode)` (§3.1 punto 4): código 101 → sin línea (es un `panic`, ya impreso);
     `SIGSEGV`/`0xC0000005` → `Segmentation fault`; `SIGABRT`/`0xC0000409` sin texto previo de Rust →
     `Aborted`.
- Un crate de sólo biblioteca (sin `src/main.rs` ni `[[bin]]`) no se ejecuta: `ErrNoBinary` con el aviso
  `errors.rustNoBinary`.
- `Build`: sólo la etapa 1 con salida junto al fuente (`main.exe` / `main`) en archivos sueltos y en
  `target/debug` en proyectos.
- `RunUntitled`: escribe `main.rs` en `<temp>/vizcacha_rs_<pid>/` y ejecuta; `Cleanup` borra la carpeta.
  No tiene `Cargo.toml`, así que usa `rustc` y la edición de §3.1 punto 11.
- `Check` y `Format` no son del runner: los implementan `clippy/` y `rustfmt/` como `app.CodeChecker` y
  `app.CodeFormatter` (§4.5).

### 4.4 Depurador (`lldbdap/`)

El flavor genérico de lldb-dap (launch, `runInTerminal`, motivos de parada, filtro de frames de sistema,
descripciones de caída) es el de `protocol/dap/lldb` (PLAN_CPP.md §3.1 punto 9). Este paquete sólo añade:

- Antes del `launch`, compila con la etapa 1 del runner (misma salida). Un error de compilación termina
  la sesión con `debug:terminated` y su texto por `debug:output`, para que el Assistant lo explique.
- `Launch`: el del flavor de LLDB con `program` = el binario, `cwd`, `args`, `runInTerminal: true`
  (`Capabilities.DebugInput`) y estos `initCommands`:
  ```json
  ["command script import <sysroot>/lib/rustlib/etc/lldb_lookup.py",
   "command source -s 0 <sysroot>/lib/rustlib/etc/lldb_commands"]
  ```
  (`<sysroot>` sale de `rustc --print sysroot`.) Son los mismos comandos que carga `rust-lldb`: añaden
  los formateadores y "synthetic children" de `Vec`, `String`, `&str`, `Option`, `Result`, `HashMap` y
  otros. **Spike R2a (dos días como máximo):** comprobar con el `lldb-dap` que se usa (llvm-mingw 20260922
  u otro) que el LLDB incluye Python (los formateadores son scripts de Python: sin él, `command script
  import` falla) y que `Vec<i32>` y `String` se leen bien en la vista Variables. Respaldo en este orden:
  1. formateadores del sysroot (lo ideal);
  2. un archivo de comandos propio embebido que sólo usa resúmenes por plantilla de LLDB sin Python
     (`type summary add --summary-string` para la longitud de `Vec` y `String`; `Option` y `Result` ya
     muestran su variante por el DWARF);
  3. valores en crudo y una línea de ayuda en el panel Variables (`debug.rustRawValues`), sin error.
  Un fallo de `initCommands` nunca aborta la sesión: se ignora y se pasa al siguiente respaldo.
- **Parada en `panic!`.** LLDB no se detiene al entrar en pánico. Tras `initialize`, el adaptador envía
  `setFunctionBreakpoints` con `rust_panic` (el punto donde el runtime ya imprimió el mensaje y empieza a
  desenrollar), y la parada se mapea a `StopException` con la descripción `panicked`. Así el alumno ve la
  línea culpable en el depurador, igual que Python con una excepción no capturada, y el mensaje original
  (`thread 'main' panicked at src/main.rs:5:10`) ya está en `debug:output`. Al continuar, el programa
  termina con 101.
- `KeepFrame` (sobre el del flavor): oculta frames cuya ruta empieza por `/rustc/` (la biblioteca estándar
  que compiló rustup) o por `<sysroot>`, y los de `core::`, `std::`, `rust_panic` y
  `rust_begin_unwind`, salvo el primero si no queda ninguno. En proyectos también oculta lo que está en
  `~/.cargo/registry` (código de dependencias): el alumno depura su crate.
- `Output`: `stdout`/`stderr` del programa; el resto a `console`.
- Windows: sólo funciona con binarios del target GNU (DWARF); si el host es MSVC, el depurador responde
  con el aviso `errors.rustMsvcHost` en lugar de intentar leer PDB.

### 4.5 Inteligencia, comprobador y formato (`analyzer/`, `clippy/`, `rustfmt/`)

- **rust-analyzer**: `Command` = `rust-analyzer` (stdio, sin argumentos). `RootOf(path)` = la carpeta del
  proyecto (el `Cargo.toml` que encuentra `FindProject`) o la carpeta del archivo si es suelto
  (rust-analyzer lo trata como "detached file" con soporte reducido). `Configuration()` y
  `InitializationOptions()`:
  ```json
  {"cargo":{"buildScripts":{"enable":false}},"procMacro":{"enable":true},
   "cachePriming":{"enable":false},"lru":{"capacity":64},
   "check":{"command":"check"},"checkOnSave":true,
   "inlayHints":{"typeHints":{"enable":true},"parameterHints":{"enable":true},"chainingHints":{"enable":false}}}
  ```
  `check` en lugar de `clippy` para el ciclo de guardado: la experiencia en vivo (diagnósticos por
  guardado) debe ser rápida; los consejos de clippy van aparte, después de ejecutar (comprobador).
  Los diagnósticos propios de rust-analyzer (sintaxis, nombres sin resolver, tipos) llegan mientras se
  escribe; los de `cargo check` (préstamos, tiempos de vida) llegan al guardar. Ambos traen `code`
  (`E0382`), el mismo que el catálogo de §4.7.
- Primera apertura: rust-analyzer carga el crate y la biblioteca estándar (de segundos a más de un minuto
  según la máquina). `lsp:status` por lenguaje (M2) muestra "Iniciando…" y el resto de la IDE sigue
  utilizable. Si se cierra sin que termine, no hay que hacer nada.
- Fixtures de rust-analyzer real (`testdata/rust-analyzer/*.json`): diagnósticos con rango y código,
  completado con `detail`, hover (markdown con la firma y la documentación), firma, símbolos (anidados
  en `impl`), definición, y, si se hace R6, inlay hints.
- **Comprobador** (`clippy/`): `Check(ctx, config)` devuelve las líneas JSON (§3): `cargo clippy
  --message-format=json --quiet --manifest-path <toml>` en proyectos y `clippy-driver … --error-format=json`
  en archivos sueltos. Tiempo límite de 60 s; "" si no hay nada o clippy no está. Sin clippy, el mismo
  comando con `rustc --emit=metadata` (sólo avisos del compilador). Un archivo con errores devuelve los
  errores, sin duplicar los que ya mostró la ejecución (los ids ya existen en el parser).
- **Formato** (`rustfmt/`): `Format(path, text)` =
  `rustfmt --edition <e> --emit stdout` con el texto por stdin y `Dir` = carpeta del archivo (así
  rustfmt busca allí un `rustfmt.toml`/`.rustfmt.toml` del profesor). La edición de un proyecto se lee de
  su `Cargo.toml`; la de un archivo suelto sigue §3.1 punto 11. Errores de sintaxis → `ErrFormat` con la
  línea (el frontend ya muestra `errors.formatRejected`).

### 4.6 Paquetes (`packages/`)

| Verbo | Comando | Notas |
|---|---|---|
| `Init(dir, name)` | `cargo init --vcs none --name <name>` en `dir` | si `dir` ya tiene `Cargo.toml`, `ErrAlreadyProject`; `--vcs none` para no crear un repositorio de git que el alumno no pidió |
| `Add(dir, pkg)` | `cargo add <pkg>` | `pkg` acepta `serde` y `serde@1`; validado con `app.SingleWordArgument` |
| `Remove(dir, pkg)` | `cargo remove <pkg>` | |
| `List(dir)` | `cargo tree --depth 1` | |
| `Tidy` | `ErrUnsupported` | no existe en Cargo |

Corren por el `Supervisor` compartido y emiten `run:*` (la salida aparece en el diálogo de Paquetes como
la de `go mod` y `pip`). `Add`, `Remove` y `List` exigen un proyecto: sin `Cargo.toml`, el diálogo ofrece
`Init` primero (`errors.rustNeedsProject`). `cargo add` sin red falla con un mensaje de red que el
Assistant explica con `RS-NETWORK`. Las dependencias se descargan a `~/.cargo/registry`, fuera de la
carpeta del alumno.

### 4.7 Errores (`errors/`)

Parser (`parser.go`, con `protocol/errorcatalog.{SplitLines,LocationFrom,NamedGroups,ResolvePath}`),
cuatro fuentes:

- **JSON de rustc o cargo** (cada línea es un objeto). De `reason: "compiler-message"` (cargo; rustc suelto
  emite el mensaje directamente) se toma `message`: `code.code` (`E0382` o el nombre del lint, como
  `unused_variables`), `level` (`error` y `error: internal compiler error` → `SeverityError`; `warning` →
  `SeverityWarning`; `note`/`help` se anexan al diagnóstico anterior, porque rustc las trae en
  `children`), `message` y `spans`. `Location` = el span con `is_primary: true` (`file_name`,
  `line_start`, `column_start` y, gratis, `line_end`/`column_end` para subrayar el rango exacto, mejor
  que en C++). `RawText` = `rendered`. `Source` = `"compiler"` (`"clippy"` para lints cuyo
  `code.code` empieza por `clippy::`).
- **Errores de sintaxis sin código** (`expected one of …, found …`, `this file contains an unclosed
  delimiter`, `cannot find macro …`): llegan en el mismo JSON, con `code` nulo; el catálogo los
  reconoce por texto.
- **`panic`** (stderr del programa, formato actual desde 1.73):
  `^thread '(?P<thread>[^']+)' panicked at (?P<path>[^:\r\n]+):(?P<line>\d+):(?P<column>\d+):$` y la línea
  siguiente es el mensaje (`index out of bounds: the len is 3 but the index is 5`). Formato anterior:
  `thread 'main' panicked at '(?P<message>.*)', (?P<path>[^:]+):(?P<line>\d+):(?P<column>\d+)`. La
  ubicación de `path` (`src/main.rs`, relativa a la carpeta de trabajo) se resuelve con `ResolvePath` y
  apunta al código del alumno; si apunta a `/rustc/<hash>/library/…` (el `panic` ocurrió en la
  biblioteca estándar, por ejemplo un `unwrap` anidado), el parser busca la línea siguiente de pila si
  hay `RUST_BACKTRACE`, y si no, deja el diagnóstico sin ubicación (el depurador la da). Ruido: la línea
  `note: run with \`RUST_BACKTRACE=1\` environment variable…`. `Source` = `"runtime"`.
- **Otras caídas:** `thread '(?P<thread>[^']+)' has overflowed its stack`, `fatal runtime error:
  stack overflow`, y el `Error: …` que imprime `main` cuando devuelve `Err` (código de salida 1; se
  reconoce por la forma `^Error: (?P<debug>.+)$` de la última línea de stderr, con prudencia: sólo si el
  programa terminó con 1 y no hay otro diagnóstico).

Catálogo `data/catalog.rust.json`, mismos campos que el de Go (`id`, `patterns` con grupos nombrados, `en`
y `es` con `title`, `body`, `fix`). A diferencia de C++, la clave es el **código de error estable de
rustc**, no una regex del texto (que cambia entre versiones): `{"codes": ["E0382"]}` y, sólo para lo que no
tiene código, un patrón. Unas 30 entradas de errores más cinco avisos:

| Id | Código o patrón | Qué explica |
|---|---|---|
| `RS-MOVED` | E0382 | uso de un valor movido (`borrow of moved value: \`s\``) |
| `RS-BORROW-MUT-TWICE` | E0499 | dos préstamos mutables a la vez |
| `RS-BORROW-CONFLICT` | E0502 | préstamo mutable e inmutable a la vez (`v.push` dentro de `for x in &v`) |
| `RS-ASSIGN-BORROWED` | E0506 | asignar a un valor prestado |
| `RS-MOVE-BORROWED` | E0505 | mover un valor mientras está prestado |
| `RS-MOVE-OUT-OF-REF` | E0507 | mover fuera de una referencia o de un índice |
| `RS-DANGLING-REF` | E0597 | `borrowed value does not live long enough` |
| `RS-RETURN-LOCAL-REF` | E0515 | devolver una referencia a una variable local |
| `RS-MISSING-LIFETIME` | E0106 | falta el tiempo de vida (`missing lifetime specifier`) |
| `RS-MISMATCHED-TYPES` | E0308 | tipos distintos (`expected \`i32\`, found \`&str\``; también falta de `return`) |
| `RS-UNRESOLVED-NAME` | E0425 | nombre sin definir (`cannot find value \`x\` in this scope`) |
| `RS-FAILED-RESOLVE` | E0433 | ruta que no se resuelve (`use of undeclared type`, falta `use`) |
| `RS-UNRESOLVED-IMPORT` | E0432 | `unresolved import` (crate sin añadir a `Cargo.toml`) |
| `RS-NO-METHOD` | E0599 | método que no existe o trait no importado |
| `RS-NO-FIELD` | E0609 | campo que no existe en la estructura |
| `RS-ASSIGN-TWICE` | E0384 | asignar dos veces a una variable inmutable |
| `RS-NOT-MUTABLE` | E0596 | tomar prestado como mutable algo declarado sin `mut` |
| `RS-TRAIT-BOUND` | E0277 | no se cumple un trait (`doesn't implement Display`, `cannot add`, el `?` fuera de una función que devuelve `Result`) |
| `RS-WRONG-ARGS` | E0061 | número de argumentos incorrecto |
| `RS-BINARY-OP` | E0369 | operador binario no soportado (`cannot add &str to &str`) |
| `RS-DEREF` | E0614 | desreferenciar algo que no es referencia |
| `RS-NON-EXHAUSTIVE` | E0004 | `match` que no cubre todos los casos |
| `RS-PRIVATE` | E0603 | elemento privado |
| `RS-TYPE-ANNOTATIONS` | E0282 | el compilador no puede inferir el tipo (`collect`, `parse`) |
| `RS-MISSING-ITEMS` | E0046 | falta un método requerido al implementar un trait |
| `RS-NO-MAIN` | E0601 | falta `fn main` |
| `RS-EXPECTED-TOKEN` | sin código, `expected (one of )?.*, found` | sintaxis: falta un `;`, `)` u otro símbolo |
| `RS-UNCLOSED-DELIMITER` | sin código, `this file contains an unclosed delimiter` | llave o paréntesis sin cerrar |
| `RS-UNKNOWN-MACRO` | sin código, `cannot find macro \`(?P<name>\w+)\`` | macro mal escrita (`printn!`) o sin `!` |
| `RS-LINKER-MISSING` | `linker \`(?P<linker>link.exe\|cc\|gcc)\` not found` | falta el enlazador (§4.2: consejo por sistema) |
| `RS-NETWORK` | `failed to (download\|get)\|network failure\|Could not resolve host` | `cargo` sin red |
| `RS-PANIC-INDEX` | `index out of bounds: the len is (?P<len>\d+) but the index is (?P<index>\d+)` | índice fuera de rango |
| `RS-PANIC-UNWRAP-NONE` | `called \`Option::unwrap()\` on a \`None\` value` | `unwrap` sobre `None` |
| `RS-PANIC-UNWRAP-ERR` | `called \`Result::unwrap()\` on an \`Err\` value: (?P<error>.+)` | `unwrap` sobre `Err` |
| `RS-PANIC-EXPECT` | `(?P<message>.+): (?P<error>.+)` tras un `expect` (sólo cuando la ubicación apunta a un `expect`; si no, `RS-PANIC-EXPLICIT`) | `expect` falló |
| `RS-PANIC-OVERFLOW` | `attempt to (add\|subtract\|multiply\|negate) with overflow` | desbordamiento aritmético |
| `RS-PANIC-DIV-ZERO` | `attempt to (divide\|calculate the remainder) by zero` | división entre cero |
| `RS-PANIC-STR-BOUNDARY` | `byte index (?P<index>\d+) is not a char boundary` | cortar un `&str` por la mitad de un carácter |
| `RS-PANIC-REFCELL` | `already (borrowed\|mutably borrowed)` | `RefCell` prestado dos veces |
| `RS-PANIC-EXPLICIT` | `explicit panic`, `not implemented`, `internal error: entered unreachable code` y cualquier `panic!` | `panic!` escrito por el alumno |
| `RS-STACK-OVERFLOW` | `has overflowed its stack` | recursión infinita |
| `RS-MAIN-ERR` | `^Error: (?P<debug>.+)$` | `main` devolvió `Err` |
| `W-RS-UNUSED-VAR` | lint `unused_variables` | variable sin usar (con la ayuda `_x`) |
| `W-RS-UNUSED-MUT` | lint `unused_mut` | `mut` sobrante |
| `W-RS-UNUSED-IMPORT` | lint `unused_imports` | `use` sin usar |
| `W-RS-DEAD-CODE` | lint `dead_code` | función o campo nunca usado |
| `W-RS-CLIPPY` | cualquier `clippy::…` (genérica) | consejo de clippy con el nombre del lint (`needless_range_loop`, `redundant_clone`); el texto remite a su `help` |

Como en Go, el texto EN/ES dice qué pasó, por qué y cómo se arregla, con un ejemplo mínimo en las
entradas de propiedad y préstamos (las más difíciles: `RS-MOVED`, `RS-BORROW-*`, `RS-DANGLING-REF`,
`RS-MISSING-LIFETIME`); el mensaje original nunca se traduce. Pedir a los profesores errores reales de
alumnos antes de redactar; en Rust los más frecuentes son `RS-MOVED`, `RS-MISMATCHED-TYPES`,
`RS-BORROW-CONFLICT` y `RS-UNRESOLVED-NAME`.

### 4.8 `support_rust.go` (en `wails/`, paquete `main`)

```go
func newRustSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts, supervisor *process.Supervisor) (app.LanguageSupport, func(ctx context.Context), error)
```

Recibe también el `Supervisor` compartido. Crea el locator, el runner, el flavor de lldb-dap, rust-analyzer,
`rustfmt` (`Formatter`), `clippy` (`Checker`), `packages` y el explainer. `Console` es `nil`. `main.go` lo
añade al registro después de C++ en lugar del `app.UnavailableSupport` de 2.3 y borra el perfil provisional
de Rust de `support_unavailable.go`. `.golangci.yml` gana la regla `rust-adapter-stays-in-rust` y la
exclusión de §3.1 punto 8.

## 5. Frontend

| Pieza | Cambio |
|---|---|
| `package.json` | `@codemirror/lang-rust` |
| `lib/editor/languageSupport.ts` | caso `rust`: `rust()`, 4 espacios; `.rs` |
| `lib/editor/templates.ts` | plantilla `fn main() { println!("Hola, Rust"); }` (en varias líneas, con la sangría de rustfmt) y plantilla en blanco |
| `lib/bridge/mock*.ts` | escenarios de Rust: ejecución correcta, error de compilación (`RS-MOVED` explicado con la línea subrayada), `panic` (`RS-PANIC-INDEX` explicado), depuración con frames de `factorial(n: u64)` |
| `lib/panels/OutputPanel.svelte` | los enlaces `archivo:línea:col` ya existen; comprobar el formato `--> src/main.rs:5:10` de rustc (con flecha) y `src/main.rs:5:10:` del `panic` |
| `lib/panels/BottomPanel.svelte` | sin pestaña Consola para Rust (capacidad apagada); botón Build visible |
| `lib/shell/PackagesDialog.svelte` | verbos `init`, `add`, `remove`, `list` con los textos de Cargo (`packages.cargoHint`); sin Tidy |
| `lib/shell/FirstRunWizard.svelte` | paso de herramientas: si Rust está entre los lenguajes elegidos y falta `rustc`, enseña el enlace a https://rustup.rs y, en Windows, el comando con host GNU (copiable); si lldb-dap falta, lo explica |
| `lib/shell/AboutDialog.svelte` | créditos: Rust (rustc, cargo, clippy, rustfmt, rust-analyzer), LLVM (lldb) |
| `lib/stores/codeLanguages.ts` | `rust` en la lista de lenguajes (Nuevo archivo de…, Ajustes → Herramientas) |
| `lib/editor/` (opcional, R6) | decoraciones de pistas en línea (inlay hints), §9 |

## 6. Textos nuevos (`tools/po2json/ux_copy.json`)

| Clave | EN | ES |
|---|---|---|
| `codeLanguage.rust` | Rust | Rust |
| `settings.toolRustc` | Rust compiler (rustc) | Compilador de Rust (rustc) |
| `settings.toolCargo` | Cargo (packages and projects) | Cargo (paquetes y proyectos) |
| `settings.toolRustAnalyzer` | Code helper (rust-analyzer) | Ayudante de código (rust-analyzer) |
| `settings.toolClippy` | Advice (clippy) | Consejos (clippy) |
| `settings.toolRustfmt` | Formatter (rustfmt) | Formateador (rustfmt) |
| `errors.rustNotFound` | Rust isn't installed. Install it from rustup.rs (free) or choose it in Settings. | Rust no está instalado. Instálalo desde rustup.rs (gratis) o elígelo en Ajustes. |
| `errors.cargoNotFound` | Cargo isn't available, so projects and packages are off. It comes with rustup. | Cargo no está disponible, así que no hay proyectos ni paquetes. Viene con rustup. |
| `errors.rustupNoToolchain` | rustup is installed but has no Rust toolchain. Run the command below. | rustup está instalado pero sin un toolchain de Rust. Ejecuta el comando de abajo. |
| `errors.rustMsvcHost` | This Rust uses the Visual Studio linker, which isn't free. Switch to the GNU toolchain with the command below. | Este Rust usa el enlazador de Visual Studio, que no es libre. Cambia al toolchain GNU con el comando de abajo. |
| `errors.rustAnalyzerNotFound` | The code helper (rust-analyzer) isn't installed. Suggestions and live problems are off. | El ayudante de código (rust-analyzer) no está instalado. No habrá sugerencias ni problemas en vivo. |
| `errors.clippyNotFound` | clippy isn't installed, so only compiler warnings are shown. | clippy no está instalado, así que sólo se muestran avisos del compilador. |
| `errors.rustfmtNotFound` | rustfmt isn't installed, so the file was saved as it is. | rustfmt no está instalado, así que el archivo se guardó tal cual. |
| `errors.rustWorkspaceRoot` | This is a workspace folder. Open a file inside one of its crates to run it. | Esta es la carpeta de un workspace. Abre un archivo dentro de uno de sus crates para ejecutarlo. |
| `errors.rustNoBinary` | This crate is a library, it has no main program to run. | Este crate es una biblioteca, no tiene programa principal que ejecutar. |
| `errors.rustNeedsProject` | Packages need a Cargo project. Create one first. | Los paquetes necesitan un proyecto de Cargo. Crea uno primero. |
| `packages.cargoHint` | Packages are added with cargo to the project's Cargo.toml. | Los paquetes se añaden con cargo al Cargo.toml del proyecto. |
| `debug.rustRawValues` | Rust values are shown raw because the debugger has no Rust viewers. | Los valores de Rust se ven en crudo porque el depurador no tiene visores de Rust. |
| `firstRun.rustInstall` | Install Rust with rustup (free). On Windows choose the GNU toolchain. | Instala Rust con rustup (gratis). En Windows elige el toolchain GNU. |

Ya existen desde M0 y M2, y no se repiten: `packages.add`, `packages.remove`, `packages.list`,
`packages.init`, `run.compiling`, `run.compileFailed`, `run.crashed`, `errors.lldbDapNotFound`,
`errors.formatRejected`. **Dueño:** R0 añade todas las claves de esta tabla; las que pidan los tracks las
fusiona el orquestador al integrar.

## 7. Empaquetado

- **Rust no se empaqueta.** Un toolchain de Rust (rustc, biblioteca estándar, cargo, clippy, rustfmt,
  rust-analyzer) pesa más de 600 MB descomprimido; sumado al de `full-cpp` (objetivo de menos de 200 MB comprimido) el
  instalador ya no sería una descarga razonable para un colegio. Rust funciona en `lite` y en
  cualquier otra variante con rustup instalado, y la guía de instalación vive en el primer arranque y en
  el aviso de herramienta ausente (https://rustup.rs, gratis, §6).
- **No hay variante `full-rust` en M3.** Alternativas medidas y descartadas por ahora: `rustup` offline
  (la carpeta `RUSTUP_HOME` copiada: 600 MB, no se actualiza), o un toolchain mínimo (`--profile
  minimal`, unos 200 MB, sin clippy ni rustfmt ni rust-analyzer). Queda como opción futura, junto al
  empaquetado de LLVM para macOS y Linux, si los profesores lo piden y se acepta el tamaño. El costo de
  no empaquetar es que el alumno instala una herramienta más; el beneficio es una descarga pequeña,
  siempre el Rust actual y los componentes al día.
- **Depurador sin empaquetar Rust:** lldb-dap viene en `full-cpp` y `full`; con `lite` hay que instalar
  LLVM aparte, y la IDE lo dice (§4.2). El localizador de Rust busca `lldb-dap` en `toolchain/cpp/bin`.
- Sin cambios en `packaging/versions.toml`, `build_release.py` ni en el instalador. La prueba de humo
  `smoke_test.py --rust` (opcional) compila y ejecuta un `hola.rs` desde una carpeta con espacios y
  acentos si hay rustup, y se salta si no.
- Actualizaciones automáticas: sin cambios (§3.1 punto 9).

## 8. Tests

- Unitarios Go: `locator_test` (orden configurado → `CARGO_HOME` → PATH con `exec` falso, proxy de rustup
  sin toolchain, versión mínima), `toolchain_test` (`rustc -vV`, host msvc, edición según versión),
  `project_test` (`Cargo.toml` hacia arriba, manifiesto virtual, `src/bin/x.rs`), `runner_test`
  (Configure suelto y de proyecto, dos etapas con `Then`, caché de salida, filtro JSON → `rendered`, línea
  de caída por código de salida, 101 sin línea), `errors/parser_test` con
  `testdata/rust_output/*.json|*.txt` (E0382 con hijos y notas, varios diagnósticos, advertencia
  `unused_variables`, JSON de cargo con `compiler-message`, `panic` en formato nuevo y antiguo, `panic`
  con ruta de `/rustc/`, `has overflowed its stack`, `Error: …` de `main`, rutas de Windows),
  `catalog_test` (cada id tiene fixture grabado de un rustc real, placeholders válidos, códigos sin
  duplicados), `lldbdap/flavor_test` (launch con `initCommands`, filtro de frames `/rustc/`, punto de
  parada en `rust_panic`), `analyzer/flavor_test` (opciones, `RootOf` con y sin `Cargo.toml`),
  `rustfmt_test` y `clippy_test` (comando por modo).
- Integración (se saltan sin rustup o sin lldb-dap): compilar y ejecutar `hola.rs` con `read_line`; un
  proyecto `cargo init` con `cargo add` de un crate pequeño (sin red, se salta); un programa con
  `v[10]` imprime y muestra el `panic` con `RS-PANIC-INDEX`; depurar `factorial.rs` hasta un breakpoint,
  un paso y variables, y una parada por `panic!` con `StopException`; un error de préstamo llega por
  `debug:output` y termina la sesión; rust-analyzer da un diagnóstico de nombre no resuelto en un crate
  de prueba; `rustfmt` de un archivo mal sangrado; `clippy-driver` sobre un archivo suelto.
- Frontend: escenarios mock de Rust, plantilla, botón Build visible, pestaña Consola ausente, verbos del
  diálogo de Paquetes (sin Tidy).
- **QA E2E** (`wails/packaging/qa/e2e/steps-rust.mjs`, como `steps-python.mjs`): en EN y ES, ejecutar
  `hola.rs` con entrada por teclado, un error `E0382` explicado con la línea subrayada, un `panic` de
  índice explicado, depurar `factorial.rs` con variables y teclado, un aviso de variable sin usar en
  Problemas, y el aviso de rustup al apuntar `ToolPaths["rustc"]` a una carpeta vacía.
- QA manual (sección "Rust" en `docs/wails/QA_WAILS.md`): la lista de §10.

## 9. Tracks y orden

```
R0 (orquestador) ──► R1 · R2 · R3 · R4 · R5 en paralelo ──► integración (support_rust.go) ──► QA
                                                       └──► R6 (opcional, genérico)
```

| Track | Dueño de | Entrega | Hecho cuando |
|---|---|---|---|
| **R0 · Preparación y contrato** (orquestador) | `domain.CodeLanguageRust`, `ProjectCargo`, claves de §6, `process.Job.OutputFilter` (§3.1 punto 3), reglas depguard (§3.1 punto 8), `adapters/rust/profile.go`, `locator.go`, `project.go`, `toolchain.go`, `rusttest`, fixtures; spike del enlazador GNU en Windows | contrato compilando; perfil, localizador y toolchain con tests; `testdata/rust_output` grabado de un rustc estable real (con y sin `--error-format=json`, `cargo build`, `clippy`); decisión sobre el enlazador GNU y sobre `OutputFilter` | los demás tracks compilan contra el locator; un commit del que parten |
| **R1 · Compilar, ejecutar y paquetes** | `adapters/rust/runner`, `packages` | §4.3 y §4.6 | `hola.rs` con `read_line` por la PTY, proyecto con `cargo run` equivalente, Build, `panic` sin línea extra, caída por `unsafe`; `cargo init/add/remove/tree` por el diálogo |
| **R2 · Depurador** | `adapters/rust/lldbdap` (usa `protocol/dap/lldb`) | §4.4 (con el spike R2a de formateadores) | breakpoints, pasos, variables con "acaba de cambiar", pila sin frames de `/rustc/`, parada por `panic!` con la línea; `Vec`/`String` legibles o respaldo documentado |
| **R3 · Inteligencia, comprobador y formato** | `adapters/rust/analyzer`, `clippy`, `rustfmt` | §4.5 | diagnósticos en vivo, completado, hover, definición, símbolos; avisos de clippy tras ejecutar en proyecto y en archivo suelto; formato al guardar |
| **R4 · Assistant** | `adapters/rust/errors` | §4.7 | unas 30 entradas más 5 avisos con fixtures grabados de rustc real y textos EN/ES; panics y caídas explicados |
| **R5 · Frontend** | `frontend/src/lib/*` (§5) | editor, plantillas, mock, primer arranque, paquetes, créditos | `npm run dev` muestra los cuatro escenarios de Rust sin backend |
| **R6 · Pistas en línea** (opcional, genérico) | `protocol/lsp` (soporte de `textDocument/inlayHint`), puerto de `CodeIntelligence`, `lib/editor` | §9.1 | las pistas de tipo aparecen en Rust (y en Go, Python y C++ si el servidor las da); se apagan en Ajustes; no cambia el comportamiento sin ellas |

Integración: R0 → R1 → R4 → R2 → R3 → R5 (→ R6). Un commit por track; el orquestador escribe
`support_rust.go`, resuelve los CCR y fusiona los textos.

### 9.1 R6: pistas en línea (inlay hints), propuesta genérica

Rust infiere casi todos los tipos y un principiante no los ve: `let x = v.iter().map(...).collect();`
oculta el tipo de `x`. rust-analyzer da esas pistas por LSP (`textDocument/inlayHint`), y gopls, pylsp
(con pyright) y clangd también. Propuesta: `protocol/lsp` pide las pistas del rango visible con
debounce, el puerto `CodeIntelligence` gana `InlayHints(path, range)` con `domain.InlayHint{Line,
Column, Label, Kind}` (la forma exacta la decide R0 mirando `app/ports.go`), y el editor las pinta con
`Decoration.widget` de CodeMirror. Un interruptor en Ajustes (por defecto activado sólo para Rust; ver
preguntas abiertas). Es un track pequeño e independiente: puede ir en M3, después o nunca, sin bloquear el
resto. Si no se hace, el campo `inlayHints` de §4.5 se envía igual y rust-analyzer lo ignora sin que el
cliente lo declare.

## 10. Criterios de salida (2.4.0)

Con `lite` y rustup instalado (toolchain GNU en Windows, `lldb-dap` del `full-cpp` o de LLVM), en Windows,
macOS y Linux, y en EN y ES:

1. Nuevo archivo de Rust → plantilla → F5 compila e imprime `Hola, Rust`; la primera compilación muestra
   `Compilando…` si tarda.
2. Un programa que lee con `read_line` recibe lo escrito en Output; el aviso aparece antes de escribir.
3. Mover un `String` y usarlo después → el Assistant explica `RS-MOVED` con la línea subrayada y la
   propuesta (`.clone()` o prestar).
4. Un `v.push` dentro de `for x in &v` → `RS-BORROW-CONFLICT`; asignar dos veces a un `let` sin `mut` →
   `RS-ASSIGN-TWICE`; un tipo equivocado → `RS-MISMATCHED-TYPES`.
5. Un índice fuera de rango → lo impreso antes se ve y el `panic` se explica (`RS-PANIC-INDEX`) con su
   línea; F6 para en `rust_panic` con `StopException` en la línea culpable.
6. Breakpoint en `factorial` → F6 para ahí; F7/F8/F9; Variables marca cambios; Llamadas muestra
   `factorial(n=3)`; Hilos muestra uno; una `Vec<i32>` y un `String` se ven legibles (o el respaldo
   documentado de R2a, con la línea de ayuda).
7. Una carpeta con `Cargo.toml` y `src/main.rs` se compila y ejecuta como proyecto; Paquetes → `add` de un
   crate se refleja en `Cargo.toml`; `list` muestra el árbol de un nivel.
8. `Vec::ne` ofrece `new`; hover sobre `println` muestra su documentación; F12 salta a una función
   propia; un nombre mal escrito se subraya mientras se escribe y los errores de préstamo aparecen al
   guardar.
9. Guardar con formato activado reacomoda sangría y llaves con rustfmt.
10. Una variable sin usar muestra `W-RS-UNUSED-VAR` y un consejo de clippy aparece en Problemas tras
    ejecutar, en un proyecto y en un archivo suelto.
11. En Windows con host MSVC y sin Visual Studio, el aviso `errors.rustMsvcHost` enseña el comando
    `rustup default stable-x86_64-pc-windows-gnu`; tras ejecutarlo, F5 funciona.
12. Sin rustup instalado: abrir un `.rs` muestra el aviso con el enlace a https://rustup.rs; el
    primer arranque lo ofrece al elegir Rust.
13. Un `.rs`, un `.cpp`, un `.py` y un `.go` abiertos: cada uno ejecuta con su lenguaje; la pestaña
    Consola desaparece al pasar al `.rs`.
14. Tests Go y frontend en verde; `golangci-lint` limpio; CI en los tres sistemas (los tests que
    necesitan rustup se saltan si falta); `steps-rust.mjs` pasa.

## 11. Riesgos

| Riesgo | Mitigación |
|---|---|
| El toolchain `windows-gnu` no enlaza con sólo lo que trae rustup (necesita un `gcc`) | spike en R0 con una máquina limpia; respaldo: el `clang` de llvm-mingw con `-C linker` (§3) y, si no hay ni eso, el aviso `RS-LINKER-MISSING` con la guía |
| Alumnos con rustup en host MSVC y sin enlazador | aviso explícito `errors.rustMsvcHost` y comando de cambio; entrada de catálogo para el error de enlace |
| `lldb-dap` no está disponible con `lite` | ejecutar sigue funcionando y el aviso lo dice; `full-cpp` lo trae; documentar LLVM aparte |
| LLDB sin Python no carga los formateadores de Rust | spike R2a y respaldos en orden (§4.4): formateadores, resúmenes propios, valores en crudo con ayuda |
| LLDB con binarios de Rust se pierde con `async`, `Rc`, `Box<dyn Trait>` | fuera de alcance; el valor sale en crudo y se documenta como limitación |
| `rust_panic` no existe con ese nombre en todas las versiones, o el punto de parada no se alcanza | probar con la versión mínima y la última estable; respaldo: parar en `rust_begin_unwind`; si ninguna funciona, el depurador no se detiene en `panic` pero el mensaje se ve |
| rust-analyzer tarda y consume memoria (1 a 2 GB en crates grandes) | arranque perezoso, `IdleTimeout` de 5 min, `lsp:status` por lenguaje y opciones ligeras (`cachePriming`, `lru`); proyectos de alumnos son pequeños |
| rust-analyzer con archivos sueltos da menos (modo "detached file") | se documenta; los diagnósticos completos vienen de `cargo check` en proyectos y del comprobador (`clippy-driver`) en archivos sueltos |
| La primera compilación de un proyecto con dependencias tarda minutos | aviso de silencio `run.compiling` con texto de "la primera vez tarda más"; la salida de cargo se muestra |
| Cargo descarga un toolchain por `rust-toolchain.toml` sin avisar | etapa silenciosa con el mismo aviso; se documenta |
| El mensaje de `panic` cambió en 1.73 (`panicked at 'msg', ruta` → dos líneas) | el parser soporta los dos formatos y la versión mínima es 1.70 |
| `Error: …` de `main` es ambiguo (cualquier línea que empiece así) | sólo con código de salida 1, última línea de stderr y sin otro diagnóstico (§4.7) |
| `target/` ocupa mucho en las carpetas de los alumnos | `CARGO_INCREMENTAL=0`; el `.gitignore` de `cargo init`; una opción de limpieza en una versión posterior |
| Los códigos de error de rustc son estables, pero el texto de los lints cambia entre versiones | el catálogo se indexa por código y por nombre de lint, no por regex del mensaje; fixtures grabados con la versión mínima y la última estable |
| `cargo clippy` en proyectos compite con la ejecución por el bloqueo de `target/` | corre después de ejecutar, nunca en paralelo (§3) |
| `clippy-driver` por el proxy de rustup falla fuera de cargo en algún sistema | spike en R3; respaldo: `rustc --emit=metadata` (sólo avisos del compilador) |
| Dev-Cpp o Code::Blocks de los alumnos pone otro `gcc` en el PATH y rompe el enlace de windows-gnu | sin riesgo conocido: el GNU de rustup usa el `gcc` que encuentra; documentar la prueba en QA con ese entorno |

## 12. Prompt común para los coders

```text
You are a coder on VizcachaIDE (wails/: Go 1.25 + Wails v2 + Svelte 5 + CodeMirror 6), a beginner IDE,
bilingual EN/ES, that has a multi-language core (docs/PLAN_NUCLEO_MULTILENGUAJE.md) with Go, Python and
C++ already done. You work in an isolated git worktree, branched from m3-rust, on track <R?> of
docs/PLAN_RUST.md. First read that plan (sections 0, 3, 3.1, 4 and your track in 9), docs/PLAN_CPP.md
sections 3.1 and 4, docs/EXTENSION_MULTILENGUAJE.md section 4, wails/README.md, docs/wails/PLAN_WAILS.md
sections 2 and 4, and the adapters in wails/internal/adapters/{golang,python,cpp} as reference
implementations (python for packages and the checker, cpp for the two-stage runner and the LLDB flavor).
- Edit ONLY your track's folders. The v3 contract (internal/domain, internal/app/ports.go,
  internal/bridge/events.go, frontend/src/lib/{events,domain}.ts, bridge/types.ts) and internal/protocol
  (including protocol/dap/lldb) are read-only: request changes as a "Contract change request" in your
  report. Adapters never import each other: the shared LLDB flavor lives in protocol/dap/lldb.
- Use internal/protocol (process, dap, lsp, errorcatalog, toollocator): never copy its code.
- Only free tools: rustc, cargo, clippy, rustfmt and rust-analyzer from rustup, plus lldb-dap from LLVM.
  On Windows only the x86_64-pc-windows-gnu toolchain (the MSVC target needs a non-free linker). Pin
  nothing in code; there is nothing to bundle for Rust.
- Never build command lines through a shell: exec.Command with separate arguments (paths with spaces
  and accents are the norm in student folders).
- Catalog entries are keyed by the stable rustc error code (E0382...) or the lint name, with a text
  pattern only for errors without a code; record every fixture from a real rustc, never write it by hand.
- Library-first. Files under 200 lines, functions under 50, early return, domain names. Visible texts
  only through i18n (tools/po2json/ux_copy.json, then go run ./tools/po2json). Messages from rustc, cargo
  and the program are never translated.
- Tests that need rustup, lldb-dap or rust-analyzer must skip cleanly when they are missing.
- Verify: go vet ./... && go test ./... && golangci-lint run; cd frontend && npm run check && npm run test;
  wails build when your track affects it.
- When done: ONE local commit in English ending with "Co-Authored-By: Claude <noreply@anthropic.com>",
  no push. Report: commit hash, files, how to test, limitations, CCRs and new i18n keys.
```
