# Changelog

All notable changes to VizcachaIDE are documented in this file. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project aims to follow
[Semantic Versioning](https://semver.org/).

*La versión en español está más abajo: [Historial de cambios (español)](#historial-de-cambios-español).*

## [Unreleased]

### Added
- **C++ projects use CMake**: a new C++ project is a CMake project (`CMakeLists.txt`, `CMakePresets.json`,
  `vcpkg.json`, `.gitignore`, `main.cpp` and `.clang-format`; no more `compile_flags.txt`). Every `.cpp` in the
  folder is part of the program, so a new file is enough. A folder without a `CMakeLists.txt` gets one
  automatically the first time you run, build or debug it. F5, Build, Debug and Problems build with CMake and
  Ninja; the code helper reads `build/compile_commands.json`.
- **C++ libraries from vcpkg** in the Packages dialog: search a library by name (offline), choose it from
  the list and install it; VizcachaIDE edits `vcpkg.json` and the marked block of `CMakeLists.txt`, so you
  only write the `#include`. The first install of a library compiles it and takes a few minutes (a notice
  says so); later ones come from a cache. Uninstall removes it again.
- The full-cpp variant bundles CMake 4.4.4, Ninja 1.13.2 and a vcpkg snapshot (with the downloads vcpkg
  needs the first time, so no waiting for them). The integrated terminal has `cmake`, `ninja` and
  `VCPKG_ROOT`, so `cmake --preset debug` works there.
- The Files panel shows the folders the build tools generate (`build/`, `target/`) dimmed.
- **Integrated terminal**: a Terminal tab in the bottom panel (Ctrl+ the key left of 1, whatever it prints; also Ctrl+Ñ on keyboards with a Ñ) with a real shell (PowerShell on Windows, your shell elsewhere)
  in the open folder. The IDE's own Go, Python, C++ and Rust come first in PATH, so `go`, `python`, `pip`, `clang++` and `cargo`
  work there exactly as with F5. Several terminals, New terminal and Kill terminal, copy/paste (Ctrl+Shift+C / Ctrl+Shift+V,
  right click) and light/dark colours.
- **File → New project…** (Ctrl+Shift+N): a name, the programming language and the folder where it
  goes (always chosen, no default); VizcachaIDE writes the project and opens it ready for F5. Each
  one starts with a program that asks your name and greets you. Go: `go.mod` + `main.go`; Python:
  `main.py`; C++: a CMake project with `main.cpp`; Rust: `Cargo.toml`
  + `src/main.rs` + `.gitignore`. Names with spaces and accents keep their folder name; the Go module
  and the Rust crate get a valid version ("Mi Tienda Ñandú" → `mi-tienda-nandu`). A C++ project also
  gets a `.clang-format` with the IDE's style (LLVM, 4 spaces), so another editor formats it the same.
- **Search packages by name** in the Packages dialog: while you type, a list shows the matching
  packages with their version and description, and you choose the one to install (Python: PyPI,
  among its ~15 000 most downloaded projects plus any exact name; Go: pkg.go.dev; Rust: crates.io).
  Without network the dialog says so and still installs an exact name.
- **Close folder** in the Files panel (the ✕ next to Refresh, or a right click on the folder) and in
  the File menu: it closes the
  tabs of that folder (asking about unsaved changes), empties the panel and is not reopened at the
  next start. Files from elsewhere stay open.

### Changed
- **New file** (Ctrl+N and the toolbar button) is written in the language you are working in: the
  open file's, else the open folder's (`Cargo.toml`, `go.mod`, `compile_flags.txt`… or most of its
  sources), else the default of Settings. It used to be always a `.go`.

### Fixed
- pip's "A new release of pip is available" notice no longer shows after installing (it looked
  like an error), and installing a package no longer leaves a false problem ("os error 2"): the
  check that follows a successful run skipped Go's commands but not pip's or cargo's.
- C++ programs on Windows showed accents wrong ("¿Cómo" as "┐C├│mo") and lost the accented letters
  typed in Output ("Ñandú" read as "and"): every C++ program VizcachaIDE builds on Windows sets the
  console to UTF-8 and reads `std::cin` through the console's Unicode input.

## [2.4.0] - 2026-10-04

Wails edition. **Rust** is the fourth language: write, compile and run, understand compiler errors
and panics, debug with the keyboard working, format and manage Cargo packages, with the same
experience as Go, Python and C++. **Inlay hints** arrive for every language whose code helper offers
them. Go, Python and C++ work as in 2.3.0 (80/80 QA steps in English and Spanish, see
`docs/wails/QA_WAILS.md`).

### Added
- **Run Rust** (F5): a loose `.rs` file compiles with `rustc`, a Cargo project with `cargo`; the
  program runs in a terminal, so `read_line` reads what you type in Output. *More → Build* compiles
  without running. Cargo **workspaces** work: the member of the open file runs, and from the
  workspace root a dialog asks which program to run.
- **The Assistant explains about 45 Rust errors** in English and Spanish, recognised by rustc's
  stable codes: ownership and borrowing (moved value, two mutable borrows, a reference that does not
  live long enough…) with a tiny example, mismatched types, unknown names and imports, missing `;` or
  `}`, panics (index out of bounds, `unwrap` on `None` or `Err`, overflow, division by zero…) at the
  line of your code, a stack overflow, and clippy advice.
- **Debugger for Rust** (lldb-dap, shared with C++): breakpoints, steps, readable `String`, `Vec`,
  `Option`, `Result` and `HashMap`, a stop at the line of a `panic!`, and **keyboard input while
  debugging**.
- **Code intelligence for Rust** (rust-analyzer): live problems, completion, hover, signature help,
  go to definition and the Outline, for Cargo projects and loose files.
- **Format on save** with rustfmt, **clippy** advice after a run, and the **Packages** dialog for
  Cargo (init, add, remove, list).
- **Inlay hints** (inferred types and parameter names) for Rust, Go and C++, drawn inside the code
  without changing it; Settings → Editor turns them off. Python's code helper has none.
- Settings → Tools tells what to fix in a Rust install (Visual Studio toolchain, not the stable one,
  too old, rustup without toolchain) with the command to copy; the first run shows how to install
  Rust with rustup (free).

### Changed
- One program at a time still: Rust shares the run supervisor with the other languages.
- The language servers learn about a file opened again after the window reloaded (problems and
  status come back).

### Fixed
- The debugger showed the payload of Rust enums wrong with LLDB 23 on Windows (`Some(7)` read as
  `Some(440)`); the IDE ships a fix for LLDB's Rust formatters.
- A Rust install could fail to link when llvm-mingw was on PATH: the toolchain's own MinGW linker is
  now always used.
- Debugging did not stop at breakpoints set in another file of the project (only the open file's
  were sent), in every language.
- After `go get` or `cargo add` the code helper kept saying the package could not be found: the
  language servers now hear when `go.mod`, `go.sum`, `Cargo.toml` or `Cargo.lock` change.
- Crates such as `rand` did not build with rustup's Windows GNU toolchain (dlltool): VizcachaIDE
  passes llvm-mingw's `llvm-dlltool`, and the Assistant explains what to install when there is none.
- The call stack of Rust and C++ showed the standard library frames below `main` again.

### Notes
- Rust is not bundled: install it with rustup (on Windows the `x86_64-pc-windows-gnu` toolchain,
  since the MSVC one needs Visual Studio's linker, which is not free). Debugging needs lldb-dap, which
  the `full-cpp` and `full` downloads include.
- For developers: `process.Job.OutputFilter`, catalog entries by compiler code, `LanguageServer.InlayHints`,
  `lsp.PulledConfiguration`, `app.MemberRunner`, `ToolStatus.Advice` and the Rust adapter in
  `wails/internal/adapters/rust` (see `docs/PLAN_RUST.md`).

## [2.3.0] - 2026-10-04

Wails edition. **C++** is the third language: write, compile and run, understand errors and crashes,
debug with the keyboard working, and format, with the same experience as Go and Python. Go and
Python work as in 2.2.0 (62/62 QA steps in English and Spanish, see `docs/wails/QA_WAILS.md`).

### Added
- **Compile and run C++** (F5) with g++ or clang++ (C++17, warnings on): the program runs in a
  terminal, so `std::cin` reads what you type in Output. A folder with several `.cpp` files is
  compiled as one program; *More → Build* leaves the executable next to the source without running it.
- **Crashes are named**: a program that dies prints "Segmentation fault", "Stack overflow",
  "Floating point exception" or "Aborted", and the run ends with "The program crashed. Debug it with
  F6 to see the line."
- **The Assistant explains 27 common C++ errors** in English and Spanish, for both GCC and Clang:
  undeclared names (and `cout` without `#include <iostream>` or `std::`), missing `;` or `}`, wrong
  arguments, conversions, linker errors (`undefined reference`), crashes, uncaught exceptions,
  `.at()` out of range and the most useful warnings.
- **Debugger for C++** (lldb-dap): breakpoints, steps, variables (`std::string` and `std::vector`
  readable), call stack without the C runtime, stop on a crash at the right line, and **keyboard
  input while debugging**.
- **Code intelligence for C++** (clangd): live problems, completion, hover, signature help, go to
  definition and the Outline, also with g++'s headers.
- **Format on save** with clang-format (4 spaces; a teacher's `.clang-format` wins).
- **First run** shows how to get a compiler on each system (bundled on Windows, `xcode-select
  --install` on macOS, `apt install` on Linux).
- New download **`full-cpp`** for Windows (IDE + llvm-mingw: clang, lldb, clangd, clang-format; about
  120 MB). `full` now bundles Go, Python and C++.
- **Automatic updates.** Once a day, when it opens, VizcachaIDE looks for a new version on its
  GitHub releases (only `wails-v*`), downloads the file of this installation (same variant, system,
  installed or portable) in the background and checks its SHA-256 against the release's checksum
  file. A notice offers *Install and restart* (installed Windows copy: the new installer runs and the
  IDE closes) or *Show the file* (portable, macOS, Linux). Settings → *Updates* shows the version,
  the progress and the last check, has *Check now* and turns the automatic check off; the *More*
  menu has *Check for updates…*.

### Changed
- The Output panel keeps what a debugged program printed after the session ends ("Debugging
  finished."), until the next run.
- "Your program didn't run" no longer says "Go found N problems" for every language.
- The call stack hides the start-up frames below `main` (C runtime, Go's `runtime.main`).
- A program killed by a signal on macOS and Linux reports which one (exact crash line).

### Fixed
- Crash lines and other problems without a file were explained with the Go catalog; they now use
  the language of the last run.
- An older Assistant answer could replace a newer one (cards disappearing after a run).
- After the window reloaded, an open file lost its live problems and the status bar stayed on
  "Code helper starting…".

### Notes
- The C++ compiler is found in this order: the one chosen in Settings, the bundled one
  (`toolchain/cpp`), and `g++`, `clang++`, `c++` on PATH; lldb-dap, clangd and clang-format are
  looked for next to it. On Windows programs are linked with `-static`.
- For developers: `process.Job.Finished`, `lsp:status` per language, the shared LLDB flavor in
  `protocol/dap/lldb` (Rust will use it) and the C++ adapter in `wails/internal/adapters/cpp` (see
  `docs/PLAN_CPP.md`).

## [2.2.0] - 2026-10-04

Wails edition. **Python** is the second language: write, run, understand errors, debug and use a
console, with the same experience as Go. Go works as in 2.1.0 (46/46 QA steps in English and Spanish,
see `docs/wails/QA_WAILS.md`).

### Added
- **Run Python** (F5) in a terminal, so `input()` reads what you type in Output and what a program
  prints before crashing is never lost; program arguments and untitled files work too.
- **The Assistant explains about 30 common Python errors** in English and Spanish (NameError,
  IndentationError, TypeError, missing colon, unclosed bracket, ZeroDivisionError and more), with
  the line underlined and the original message kept.
- **Debugger for Python** (debugpy): breakpoints, steps, variables marked "just changed" (without
  Python's internal variables), call stack, stop on an uncaught exception, and **keyboard input while
  debugging**.
- **Code intelligence for Python** (python-lsp-server with pyflakes): live problems, completion,
  hover, signature help, go to definition and the Outline.
- **Format on save** and checks after a run with ruff (only syntax errors and pyflakes warnings,
  ignoring any personal ruff configuration of the machine).
- **Python console** (>>>) that remembers variables between entries.
- **Packages** dialog for Python: install, uninstall and list packages with pip.
- The first-run wizard asks **which programming languages** you will use; *New file of…* and
  Settings → Tools show only those (Settings → General changes it).
- Packages `full-python` (IDE + Python) and `full` (Go + Python). The Go-only package is now
  `full-go`.
- **Hide the side panel and the Assistant** for more editor room: click the active rail icon, use
  the two buttons in the title bar, or press Ctrl+B / Ctrl+Alt+B. The Assistant opens again when you
  start debugging, and its button shows the number of problems while it is hidden.

### Changed
- The status bar shows the open file's language and version ("Python 3.12.14") and says
  "Code helper …" instead of "gopls …".
- The Go format error now names the line ("errors.formatRejected" said "line ?").

### Notes
- Python is found in this order: the one chosen in Settings, the bundled one, the folder's
  `.venv`, the `py` launcher (Windows) and PATH; it must be 3.10 or newer.
- For developers: `process.Job.Events`, `Capabilities.DebugInput`, `dap.StdioTransport` and the
  Python adapter in `wails/internal/adapters/python` (see `docs/PLAN_PYTHON.md`).

## [2.1.0] - 2026-10-03

Wails edition. The IDE is now built around **language profiles**, so Python and C++ can be added
next without touching the rest of the IDE. Go works exactly as in 2.0.0-rc1 (34/34 parity steps in
English and Spanish, see `docs/wails/QA_WAILS.md`).

### Added
- *File → New file of…*: new files of Go, Python or C++, each with its own template and
  extension. Settings has a "Language for new files" option.
- Settings → Tools shows one group per language with the tools it found and their versions.
- A missing tool now shows which one is missing, with *Install* or *Copy the command* and
  *Choose in Settings*.
- The language server stops after 5 minutes without open files and starts again when one opens.
- Programs of future languages can run in a pseudoterminal (ConPTY on Windows), so what they
  print before crashing is not lost.

### Changed
- `settings.json` keeps tool paths in `toolPaths`. A 2.0 file (`goPath`, `delvePath`,
  `goplsPath`) is converted automatically the first time 2.1 starts.
- The Console tab and the package actions follow the language of the open file. The debug panel
  says "Goroutines" for Go and "Threads" for other languages.
- The "Go modules" dialog is now the "Packages" dialog. For Go it shows the same texts and
  actions (`go mod init`, `go get`, `go mod tidy`).

### Notes
- Python and C++ files can be created and edited (plain text), but running them says "This isn't
  available for Python/C++" until their support arrives (plans in `docs/PLAN_PYTHON.md` and
  `docs/PLAN_CPP.md`).
- For developers: the debugger (DAP), language server (LSP), process supervisor and error catalog
  engine moved to `wails/internal/protocol`; the Go adapter lives in `wails/internal/adapters/golang`.
  See `docs/PLAN_NUCLEO_MULTILENGUAJE.md`.

## [1.0.0-rc1] - 2026-10-01

First release candidate of the public, bilingual 1.0. The code base was rewritten
around a layered architecture; most of what the 0.1 README promised is now real.

### Added
- **Real debugger based on Delve** (`dlv dap`): breakpoints (also added or removed while the
  program runs), Continue, Step Over, Step Into, Step Out, Run to Cursor and Stop; Variables
  panel with lazily expanded structs, slices and maps; Call Stack (click to jump to a frame)
  and Goroutines panels; current-line highlight.
- **Assistant** panel that explains Go problems for beginners in English and Spanish:
  25 kinds of messages (16 compiler errors, 7 runtime panics, 2 `go vet` checks), each with
  a title, an explanation, a suggested fix, the untranslated original message, *Go to line* and
  *Search this error*. One example program per error in `examples/errors/<ID>/`.
- Clickable `file.go:LINE:COL` links in the console.
- **gopls integration:** completion (with a static fallback when gopls is missing), live
  diagnostics underlined over the exact range of the error, hover documentation, Ctrl+click to
  go to a definition, call-tips when typing `(`, highlighting of other occurrences and an
  **Outline** panel.
- Editor: 4 editor themes and 2 console themes that are actually applied, real tabs with a
  configurable width, gofmt on save and *Format Code* (Ctrl+Shift+F), Find/Replace bar,
  Go to Line, toggle comment, indent/unindent selection, zoom, bracket matching, up to 10
  recent files, and reload/prompt when a file changes outside the IDE.
- Run: **program arguments** field, running unsaved tabs, **Go modules** (`go run .` inside a
  folder with `go.mod`), a real Stop (interrupt, then kill after 2 s), non-blocking Build
  (Ctrl+B), keyboard input in the console.
- *File → Open Folder…* with a **Files** panel, and *Tools → Go Modules…* for `go mod init`,
  `go get` and `go mod tidy`.
- *Tools → Options…* with Environment (paths to go/dlv/gopls, GOPATH, GOROOT, extra variables,
  detected tools), Editor and Appearance pages.
- **English and Spanish user interface** with gettext/Babel catalogs; the language follows the
  operating system.
- **Packaging:** PyInstaller builds in two variants, *lite* (IDE only) and *full* (with Go
  1.25.14, Delve 1.27.2 and gopls 0.21.1 pinned and checksum-verified); Windows installer
  (Inno Setup, per-user, English/Spanish, optional `.go` association) and portable zip; macOS
  `.dmg` (arm64 and x86_64) and Linux AppImage scripts; GitHub Actions workflows for CI and
  for draft releases.
- `NOTICE.md` explaining the license of the distributed binaries, also shown in *Help → About*.
- Automated tests (pytest + pytest-qt, 334 tests including integration tests with real Go,
  Delve and gopls), ruff, and import-linter architecture contracts.

### Changed
- New layered architecture (`domain`, `application`, `infrastructure`, `ui`) where every
  feature plugs into the window through `register(workbench)`. `python -m vizcacha` is the new
  entry point (`python main.py` still works).
- Indentation uses tabs, as `gofmt` does, instead of 4 spaces; the automatic indent after `:`
  was removed.
- Settings dialog moved to *Tools → Options…*; the settings stored by 0.1 (`env/*`, `editor/*`,
  `appearance/*`, last file) are kept.
- Running and building no longer block the window.
- Python 3.10 or newer is now required.

### Removed
- The **simulated debugger** of 0.1, which faked stepping and variables without running Delve.
  It was replaced by the real Delve debugger.
- Pygments is no longer a dependency.

### Fixed
- Theme, tab size, auto-indent, line numbers and status bar options were saved but ignored;
  they are now applied.
- Bracket matching, which 0.1 announced but did not implement.
- Stop now ends the whole process tree of `go run`, not only the `go` command.

### Known limitations
- macOS and Linux packages are a **preview**: not tested on real machines yet. The CI and
  release workflows have not been run yet, and the Windows Inno Setup installer has not been
  compiled in the release environment (the PyInstaller builds and the portable zip were
  verified on Windows).
- No code signing (Windows SmartScreen warning; macOS app not notarized).
- The Spanish translation is incomplete, and there is no language selector in the app yet.
- A program being debugged cannot read from stdin.
- The default panel layout is still being tuned; the Assistant panel can start small.

## [0.1.0] - 2025-10-01

### Added
- First prototype: PyQt5 editor with Go syntax highlighting, line numbers, tabs and a static
  autocomplete list (Ctrl+Space).
- Run (F5) and Stop (Shift+F5) through `go run`, with an output console.
- Settings dialog (Environment, Editor and Appearance).
- A **simulated** debugger user interface (Variables and Call Stack panels) that did not use
  Delve.

---

# Historial de cambios (español)

Aquí se documentan los cambios importantes de VizcachaIDE. El formato sigue
[Keep a Changelog](https://keepachangelog.com/es-ES/1.1.0/) y el proyecto intenta seguir el
[versionado semántico](https://semver.org/lang/es/).

## [Sin publicar]

### Añadido
- **Los proyectos C++ usan CMake**: un proyecto C++ nuevo es un proyecto CMake (`CMakeLists.txt`,
  `CMakePresets.json`, `vcpkg.json`, `.gitignore`, `main.cpp` y `.clang-format`; ya no hay `compile_flags.txt`).
  Todo `.cpp` de la carpeta forma parte del programa, así que basta crear un archivo nuevo. A una carpeta sin
  `CMakeLists.txt` se le crea uno solo la primera vez que la ejecutas, construyes o depuras. F5, Construir,
  Depurar y Problemas compilan con CMake y Ninja; el ayudante de código lee `build/compile_commands.json`.
- **Librerías de C++ con vcpkg** en el diálogo de Paquetes: busca una librería por nombre (sin internet),
  elígela de la lista e instálala; VizcachaIDE edita `vcpkg.json` y el bloque marcado del `CMakeLists.txt`, así
  que solo escribes el `#include`. La primera instalación de una librería la compila y tarda unos minutos
  (un aviso lo dice); las siguientes salen de una caché. Desinstalar la quita de nuevo.
- La variante full-cpp incluye CMake 4.4.4, Ninja 1.13.2 y una instantánea de vcpkg (con las descargas que
  vcpkg necesita la primera vez, para no esperarlas). El terminal integrado tiene `cmake`, `ninja` y
  `VCPKG_ROOT`, así que `cmake --preset debug` funciona ahí.
- El panel de Archivos muestra atenuadas las carpetas que generan las herramientas (`build/`, `target/`).
- **Terminal integrado**: una pestaña Terminal en el panel inferior (Ctrl+ la tecla a la izquierda del 1, sea cual sea su símbolo; también Ctrl+Ñ en teclados con Ñ)
  con un shell real (PowerShell en Windows, tu shell en macOS y Linux) en la carpeta abierta. El Go,
  Python, C++ y Rust de la IDE van primero en el PATH, así que `go`, `python`, `pip`, `clang++` y `cargo`
  funcionan ahí igual que con F5. Varios terminales, Nuevo terminal y Cerrar terminal, copiar y pegar
  (Ctrl+Shift+C / Ctrl+Shift+V, clic derecho) y colores claros u oscuros.
- **Archivo → Nuevo proyecto…** (Ctrl+Shift+N): un nombre, el lenguaje de programación y la carpeta
  donde va (siempre se elige, sin valor por defecto); VizcachaIDE escribe el proyecto y lo abre listo
  para F5. Cada uno empieza con un programa que pregunta tu nombre y te saluda. Go: `go.mod` +
  `main.go`; Python: `main.py`; C++: un proyecto CMake con `main.cpp` (C++17 para el ayudante de
  código); Rust: `Cargo.toml` + `src/main.rs` + `.gitignore`. Los nombres con espacios y acentos
  conservan su carpeta; el módulo de Go y el crate de Rust reciben una versión válida ("Mi Tienda
  Ñandú" → `mi-tienda-nandu`). Un proyecto C++ también lleva un `.clang-format` con el estilo de la IDE
  (LLVM, 4 espacios), para que otro editor lo formatee igual.
- **Buscar paquetes por nombre** en el diálogo de Paquetes: mientras escribes, una lista muestra los
  paquetes que coinciden con su versión y descripción, y eliges cuál instalar (Python: PyPI, entre
  sus ~15 000 proyectos más descargados más cualquier nombre exacto; Go: pkg.go.dev; Rust:
  crates.io). Sin red el diálogo lo dice y deja instalar un nombre exacto.
- **Cerrar carpeta** en el panel de Archivos (la ✕ junto a Actualizar, o clic derecho sobre la
  carpeta) y en el menú Archivo: cierra
  las pestañas de esa carpeta (preguntando por los cambios sin guardar), vacía el panel y no vuelve a
  abrirse al siguiente inicio. Los archivos de otros lugares siguen abiertos.

### Cambiado
- **Archivo nuevo** (Ctrl+N y el botón de la barra) se escribe en el lenguaje con el que estás
  trabajando: el del archivo abierto, si no el de la carpeta abierta (`Cargo.toml`, `go.mod`,
  `compile_flags.txt`… o la mayoría de sus fuentes), si no el de Ajustes. Antes era siempre un `.go`.

### Corregido
- Ya no aparece el aviso de pip "A new release of pip is available" al instalar (parecía un error),
  e instalar un paquete ya no deja un problema falso ("os error 2"): la comprobación que sigue a una
  ejecución correcta omitía los comandos de Go pero no los de pip ni cargo.
- Los programas de C++ en Windows mostraban mal los acentos ("¿Cómo" como "┐C├│mo") y perdían las
  letras acentuadas escritas en Salida ("Ñandú" se leía "and"): todo programa de C++ que VizcachaIDE
  compila en Windows pone la consola en UTF-8 y lee `std::cin` con la entrada Unicode de la consola.

## [2.4.0] - 2026-10-04

Edición Wails. **Rust** es el cuarto lenguaje: escribir, compilar y ejecutar, entender los errores
del compilador y los `panic`, depurar con el teclado funcionando, formatear y manejar paquetes de
Cargo, con la misma experiencia que Go, Python y C++. Llegan las **pistas en línea** para todos los
lenguajes cuyo ayudante de código las ofrece. Go, Python y C++ funcionan como en 2.3.0 (80/80
pasos de QA en inglés y español, ver `docs/wails/QA_WAILS.md`).

### Añadido
- **Ejecutar Rust** (F5): un `.rs` suelto se compila con `rustc` y un proyecto Cargo con `cargo`; el
  programa corre en una terminal, así que `read_line` lee lo que escribes en Salida. *Más → Compilar*
  compila sin ejecutar. Funcionan los **workspaces** de Cargo: se ejecuta el miembro del archivo
  abierto y, desde la raíz del workspace, un diálogo pregunta qué programa ejecutar.
- **El Asistente explica unos 45 errores de Rust** en inglés y español, reconocidos por los códigos
  estables de rustc: propiedad y préstamos (valor movido, dos préstamos mutables, una referencia que
  no vive lo suficiente…) con un ejemplo mínimo, tipos que no coinciden, nombres e importaciones
  desconocidos, `;` o `}` que faltan, `panic` (índice fuera de rango, `unwrap` sobre `None` o `Err`,
  desbordamiento, división entre cero…) en la línea de tu código, desbordamiento de pila y consejos
  de clippy.
- **Depurador de Rust** (lldb-dap, compartido con C++): puntos de interrupción, pasos, `String`,
  `Vec`, `Option`, `Result` y `HashMap` legibles, parada en la línea de un `panic!` y **entrada por
  teclado mientras depuras**.
- **Inteligencia de código para Rust** (rust-analyzer): problemas en vivo, completado, información al
  pasar el ratón, ayuda de firmas, ir a la definición y el Esquema, en proyectos Cargo y en archivos
  sueltos.
- **Formato al guardar** con rustfmt, **consejos de clippy** después de ejecutar y el diálogo de
  **Paquetes** para Cargo (crear, añadir, quitar, listar).
- **Pistas en línea** (tipos deducidos y nombres de parámetros) para Rust, Go y C++, dibujadas dentro
  del código sin cambiarlo; se apagan en Ajustes → Editor. El ayudante de código de Python no las da.
- Ajustes → Herramientas dice qué arreglar en una instalación de Rust (toolchain de Visual Studio, no
  estable, demasiado antigua, rustup sin toolchain) con el comando a copiar; el primer arranque
  explica cómo instalar Rust con rustup (gratis).

### Cambiado
- Sigue habiendo un solo programa a la vez: Rust comparte el supervisor de ejecución con los demás.
- Los ayudantes de código se enteran de un archivo abierto otra vez tras recargar la ventana (vuelven
  los problemas y el estado).

### Corregido
- El depurador mostraba mal el contenido de los enums de Rust con LLDB 23 en Windows (`Some(7)` se
  leía `Some(440)`); la IDE incluye un arreglo para los formateadores de Rust de LLDB.
- Una instalación de Rust podía fallar al enlazar si llvm-mingw estaba en el PATH: ahora se usa
  siempre el enlazador MinGW propio del toolchain.
- Depurar no se detenía en los puntos de interrupción puestos en otro archivo del proyecto (sólo se
  enviaban los del archivo abierto), en todos los lenguajes.
- Después de `go get` o `cargo add` el ayudante de código seguía diciendo que no encontraba el
  paquete: ahora los servidores de lenguaje se enteran cuando cambian `go.mod`, `go.sum`,
  `Cargo.toml` o `Cargo.lock`.
- Crates como `rand` no compilaban con el toolchain GNU de rustup en Windows (dlltool): VizcachaIDE
  pasa el `llvm-dlltool` de llvm-mingw, y el Asistente explica qué instalar cuando no lo hay.
- La pila de llamadas de Rust y C++ volvía a mostrar los marcos de la biblioteca estándar bajo `main`.

### Notas
- Rust no viene incluido: se instala con rustup (en Windows el toolchain `x86_64-pc-windows-gnu`,
  porque el de MSVC necesita el enlazador de Visual Studio, que no es libre). Depurar necesita
  lldb-dap, que traen las descargas `full-cpp` y `full`.
- Para desarrolladores: `process.Job.OutputFilter`, entradas del catálogo por código del compilador,
  `LanguageServer.InlayHints`, `lsp.PulledConfiguration`, `app.MemberRunner`, `ToolStatus.Advice` y el
  adaptador de Rust en `wails/internal/adapters/rust` (ver `docs/PLAN_RUST.md`).

## [2.3.0] - 2026-10-04

Edición Wails. **C++** es el tercer lenguaje: escribir, compilar y ejecutar, entender los errores y
las caídas, depurar con el teclado funcionando y formatear, con la misma experiencia que Go y Python.
Go y Python funcionan como en 2.2.0 (62/62 pasos de QA en inglés y español, ver
`docs/wails/QA_WAILS.md`).

### Añadido
- **Compilar y ejecutar C++** (F5) con g++ o clang++ (C++17, con avisos): el programa corre en una
  terminal, así que `std::cin` lee lo que escribes en Salida. Una carpeta con varios `.cpp` se
  compila como un solo programa; *Más → Compilar* deja el ejecutable junto al código sin ejecutarlo.
- **Las caídas tienen nombre**: un programa que se cae muestra "Segmentation fault", "Stack
  overflow", "Floating point exception" o "Aborted", y la ejecución termina con "El programa se cayó.
  Depúralo con F6 para ver la línea."
- **El Asistente explica 27 errores comunes de C++** en inglés y español, para GCC y Clang: nombres
  no declarados (y `cout` sin `#include <iostream>` o sin `std::`), `;` o `}` que faltan, argumentos
  equivocados, conversiones, errores del enlazador (`undefined reference`), caídas, excepciones no
  capturadas, `.at()` fuera de rango y los avisos más útiles.
- **Depurador de C++** (lldb-dap): puntos de interrupción, pasos, variables (`std::string` y
  `std::vector` legibles), pila de llamadas sin el arranque de C, parada en la línea de una caída y
  **entrada por teclado mientras depuras**.
- **Inteligencia de código para C++** (clangd): problemas en vivo, completado, información al pasar
  el ratón, ayuda de firmas, ir a la definición y el Esquema, también con los encabezados de g++.
- **Formato al guardar** con clang-format (4 espacios; manda el `.clang-format` del profesor).
- **Primer arranque** indica cómo conseguir un compilador en cada sistema (incluido en Windows,
  `xcode-select --install` en macOS, `apt install` en Linux).
- Nueva descarga **`full-cpp`** para Windows (IDE + llvm-mingw: clang, lldb, clangd, clang-format;
  unos 120 MB). `full` ahora incluye Go, Python y C++.
- **Actualizaciones automáticas.** Una vez al día, al abrirse, VizcachaIDE busca una versión nueva
  en sus releases de GitHub (sólo `wails-v*`), descarga en segundo plano el archivo de esta
  instalación (misma variante, sistema, instalada o portable) y verifica su SHA-256 con el archivo
  de sumas de la release. Un aviso ofrece *Instalar y reiniciar* (copia instalada en Windows: se
  ejecuta el instalador nuevo y el IDE se cierra) o *Mostrar el archivo* (portable, macOS, Linux).
  Ajustes → *Actualizaciones* muestra la versión, el progreso y la última comprobación, tiene
  *Buscar ahora* y permite apagar la búsqueda automática; el menú *Más* tiene *Buscar
  actualizaciones…*.

### Cambiado
- El panel Salida conserva lo que imprimió un programa depurado cuando termina la sesión
  ("Depuración terminada."), hasta la siguiente ejecución.
- "No se pudo ejecutar" ya no dice "Go encontró N problemas" en todos los lenguajes.
- La pila de llamadas oculta los marcos de arranque bajo `main` (arranque de C, `runtime.main` de Go).
- Un programa terminado por una señal en macOS y Linux dice cuál (línea de caída exacta).

### Corregido
- Las líneas de caída y otros problemas sin archivo se explicaban con el catálogo de Go; ahora usan
  el lenguaje de la última ejecución.
- Una respuesta vieja del Asistente podía reemplazar a una nueva (tarjetas que desaparecían tras
  ejecutar).
- Tras recargar la ventana, un archivo abierto perdía sus problemas en vivo y la barra de estado se
  quedaba en "Ayudante de código iniciando…".

### Notas
- El compilador de C++ se busca en este orden: el elegido en Ajustes, el empaquetado
  (`toolchain/cpp`) y `g++`, `clang++`, `c++` en el PATH; lldb-dap, clangd y clang-format se buscan
  junto a él. En Windows los programas se enlazan con `-static`.
- Para desarrolladores: `process.Job.Finished`, `lsp:status` por lenguaje, el flavor de LLDB
  compartido en `protocol/dap/lldb` (lo usará Rust) y el adaptador de C++ en
  `wails/internal/adapters/cpp` (ver `docs/PLAN_CPP.md`).

## [2.2.0] - 2026-10-04

Edición Wails. **Python** es el segundo lenguaje: escribir, ejecutar, entender los errores, depurar
y usar una consola, con la misma experiencia que Go. Go funciona como en 2.1.0 (46/46 pasos de QA en
inglés y español, ver `docs/wails/QA_WAILS.md`).

### Añadido
- **Ejecutar Python** (F5) en una terminal: `input()` lee lo que escribes en Salida y nunca se
  pierde lo que el programa imprime antes de caerse; también argumentos y archivos sin título.
- **El Asistente explica unos 30 errores comunes de Python** en inglés y español (NameError,
  IndentationError, TypeError, falta de dos puntos, paréntesis sin cerrar, ZeroDivisionError y más),
  con la línea subrayada y el mensaje original intacto.
- **Depurador de Python** (debugpy): puntos de interrupción, pasos, variables marcadas "acaba de
  cambiar" (sin las variables internas de Python), pila de llamadas, parada en una excepción no
  capturada y **entrada por teclado mientras depuras**.
- **Inteligencia de código para Python** (python-lsp-server con pyflakes): problemas en vivo,
  completado, ayuda al pasar el ratón, firma, ir a la definición y la Estructura.
- **Formato al guardar** y revisión tras ejecutar con ruff (sólo errores de sintaxis y avisos de
  pyflakes, sin la configuración personal de ruff que tenga la máquina).
- **Consola de Python** (>>>) que recuerda las variables entre entradas.
- Diálogo **Paquetes** para Python: instalar, desinstalar y listar paquetes con pip.
- El asistente de primer arranque pregunta **qué lenguajes de programación** usarás; *Nuevo archivo
  de…* y Ajustes → Herramientas muestran sólo esos (se cambia en Ajustes → General).
- Paquetes `full-python` (IDE + Python) y `full` (Go + Python). El paquete sólo con Go pasa a
  llamarse `full-go`.
- **Oculta el panel lateral y el Asistente** para tener más espacio en el editor: pulsa el icono
  activo de la barra lateral, usa los dos botones de la barra de título o pulsa Ctrl+B / Ctrl+Alt+B.
  El Asistente vuelve a abrirse al empezar a depurar, y su botón muestra el número de problemas
  mientras está oculto.

### Cambiado
- La barra de estado muestra el lenguaje del archivo abierto y su versión ("Python 3.12.14") y dice
  "Ayudante de código …" en vez de "gopls …".
- El error de formato de Go ahora dice la línea ("errors.formatRejected" decía "línea ?").

### Notas
- Python se busca en este orden: el elegido en Ajustes, el empaquetado, el `.venv` de la carpeta,
  el lanzador `py` (Windows) y el PATH; debe ser 3.10 o más nuevo.
- Para desarrolladores: `process.Job.Events`, `Capabilities.DebugInput`, `dap.StdioTransport` y el
  adaptador de Python en `wails/internal/adapters/python` (ver `docs/PLAN_PYTHON.md`).

## [2.1.0] - 2026-10-03

Edición Wails. El IDE se organiza ahora en torno a **perfiles de lenguaje**, así que Python y C++
se podrán añadir sin tocar el resto del IDE. Go funciona exactamente igual que en 2.0.0-rc1
(34/34 pasos de paridad en inglés y español, ver `docs/wails/QA_WAILS.md`).

### Añadido
- *Archivo → Nuevo archivo de…*: archivos nuevos de Go, Python o C++, cada uno con su plantilla y
  su extensión. Ajustes tiene la opción "Lenguaje de los archivos nuevos".
- Ajustes → Herramientas muestra un grupo por lenguaje con las herramientas encontradas y sus
  versiones.
- Cuando falta una herramienta, el aviso dice cuál es y ofrece *Instalar* o *Copiar el comando* y
  *Elegir en Ajustes*.
- El servidor de lenguaje se detiene tras 5 minutos sin archivos abiertos y vuelve a arrancar al
  abrir uno.
- Los programas de los próximos lenguajes pueden ejecutarse en una pseudoterminal (ConPTY en
  Windows), para que no se pierda lo que imprimen antes de caerse.

### Cambiado
- `settings.json` guarda las rutas de las herramientas en `toolPaths`. Un archivo de la 2.0
  (`goPath`, `delvePath`, `goplsPath`) se convierte solo la primera vez que arranca la 2.1.
- La pestaña Consola y las acciones de paquetes siguen al lenguaje del archivo abierto. El panel de
  depuración dice "Goroutines" en Go e "Hilos" en los demás lenguajes.
- El diálogo "Módulos de Go" es ahora el diálogo "Paquetes". Para Go muestra los mismos textos y
  acciones (`go mod init`, `go get`, `go mod tidy`).

### Notas
- Los archivos de Python y C++ se pueden crear y editar (como texto plano), pero al ejecutarlos el
  IDE dice "Esto no está disponible para Python/C++" hasta que llegue su soporte (planes en
  `docs/PLAN_PYTHON.md` y `docs/PLAN_CPP.md`).
- Para desarrolladores: el depurador (DAP), el servidor de lenguaje (LSP), el supervisor de
  procesos y el motor del catálogo de errores pasaron a `wails/internal/protocol`; el adaptador de
  Go vive en `wails/internal/adapters/golang`. Ver `docs/PLAN_NUCLEO_MULTILENGUAJE.md`.

## [1.0.0-rc1] - 2026-10-01

Primera versión candidata de la 1.0 pública y bilingüe. El código se reescribió sobre una
arquitectura en capas, y casi todo lo que prometía el README de la 0.1 ahora existe de verdad.

### Añadido
- **Depurador real basado en Delve** (`dlv dap`): puntos de interrupción (también se ponen y se
  quitan con el programa en marcha), Continuar, Paso sobre, Paso adentro, Paso afuera, Ejecutar
  hasta el cursor y Detener; panel de Variables con structs, slices y maps que se despliegan
  bajo demanda; paneles de Pila de llamadas (un clic lleva al frame) y Goroutines; resaltado de
  la línea actual.
- Panel **Asistente**, que explica los problemas de Go a principiantes en español y en inglés:
  25 tipos de mensajes (16 errores de compilación, 7 panics y 2 avisos de `go vet`), cada uno
  con título, explicación, sugerencia, el mensaje original sin traducir y los botones *Ir a la
  línea* y *Buscar este error*. Hay un programa de ejemplo por error en `examples/errors/<ID>/`.
- Enlaces clicables `archivo.go:LÍNEA:COL` en la consola.
- **Integración con gopls:** autocompletado (con una lista estática de respaldo si falta gopls),
  diagnósticos en vivo subrayados en el rango exacto del error, documentación al pasar el ratón,
  Ctrl+clic para ir a la definición, ayuda de parámetros al escribir `(`, resaltado de las demás
  apariciones y panel **Outline**.
- Editor: 4 temas de editor y 2 de consola que ahora sí se aplican, tabulaciones reales con
  ancho configurable, gofmt al guardar y *Format Code* (Ctrl+Shift+F), barra de buscar y
  reemplazar, ir a línea, comentar, indentar/desindentar la selección, zoom, emparejado de
  llaves, hasta 10 archivos recientes y recarga (o pregunta) cuando un archivo cambia fuera del
  IDE.
- Ejecución: campo de **argumentos del programa**, ejecutar pestañas sin guardar, **módulos de
  Go** (`go run .` en una carpeta con `go.mod`), un Detener de verdad (interrumpe y, a los 2 s,
  mata el proceso), Compilar sin bloquear (Ctrl+B) y entrada por teclado en la consola.
- *Archivo → Open Folder…* con un panel **Files**, y *Herramientas → Go Modules…* para
  `go mod init`, `go get` y `go mod tidy`.
- *Herramientas → Opciones…* con las páginas Entorno (rutas de go/dlv/gopls, GOPATH, GOROOT,
  variables adicionales y herramientas detectadas), Editor y Apariencia.
- **Interfaz en español y en inglés** con catálogos gettext/Babel; el idioma sigue al del
  sistema operativo.
- **Empaquetado:** builds de PyInstaller en dos variantes, *lite* (sólo el IDE) y *full* (con
  Go 1.25.14, Delve 1.27.2 y gopls 0.21.1 fijados y verificados con sha256); instalador de
  Windows (Inno Setup, por usuario, español/inglés, asociación `.go` opcional) y zip portable;
  scripts para `.dmg` de macOS (arm64 y x86_64) y AppImage de Linux; workflows de GitHub Actions
  para CI y para releases en borrador.
- `NOTICE.md`, que explica la licencia de los binarios distribuidos; también aparece en
  *Ayuda → Acerca de*.
- Tests automáticos (pytest + pytest-qt, 334 tests, incluidos los de integración con Go, Delve y
  gopls reales), ruff y contratos de arquitectura con import-linter.

### Cambiado
- Nueva arquitectura en capas (`domain`, `application`, `infrastructure`, `ui`), en la que cada
  función se conecta a la ventana con `register(workbench)`. El nuevo punto de entrada es
  `python -m vizcacha` (`python main.py` sigue funcionando).
- La indentación usa tabulaciones, como `gofmt`, en lugar de 4 espacios; se quitó la
  indentación automática después de `:`.
- La configuración pasó a *Herramientas → Opciones…*; se conservan los ajustes guardados por la
  0.1 (`env/*`, `editor/*`, `appearance/*` y el último archivo).
- Ejecutar y compilar ya no bloquean la ventana.
- Ahora se necesita Python 3.10 o posterior.

### Eliminado
- El **depurador simulado** de la 0.1, que fingía los pasos y las variables sin ejecutar Delve.
  Lo sustituye el depurador real con Delve.
- Pygments ya no es una dependencia.

### Corregido
- Las opciones de tema, ancho del tabulador, autoindentado, números de línea y barra de estado
  se guardaban pero no se aplicaban; ahora sí.
- El emparejado de llaves, que la 0.1 anunciaba pero no tenía.
- Detener ahora termina todo el árbol de procesos de `go run`, no sólo el comando `go`.

### Limitaciones conocidas
- Los paquetes de macOS y Linux son una **vista previa**: aún no se han probado en máquinas
  reales. Los workflows de CI y de release todavía no se han ejecutado, y el instalador de
  Inno Setup para Windows aún no se ha compilado en el entorno de release (sí se verificaron en
  Windows los builds de PyInstaller y el zip portable).
- Sin firma de código (aviso de SmartScreen en Windows; app de macOS sin notarizar).
- La traducción al español está incompleta y todavía no hay selector de idioma en la app.
- Un programa en depuración no puede leer de stdin.
- La disposición inicial de los paneles aún se está ajustando; el Asistente puede empezar
  pequeño.

## [0.1.0] - 2025-10-01

### Añadido
- Primer prototipo: editor en PyQt5 con resaltado de sintaxis de Go, números de línea, pestañas
  y una lista estática de autocompletado (Ctrl+Space).
- Ejecutar (F5) y Detener (Shift+F5) con `go run`, con una consola de salida.
- Diálogo de configuración (Entorno, Editor y Apariencia).
- Una interfaz de depurador **simulada** (paneles de Variables y Pila de llamadas) que no usaba
  Delve.

[Unreleased]: https://github.com/codeplai/VizcachaIDE/compare/wails-v2.4.0...HEAD
[Sin publicar]: https://github.com/codeplai/VizcachaIDE/compare/wails-v2.4.0...HEAD
[2.4.0]: https://github.com/codeplai/VizcachaIDE/compare/wails-v2.3.0...wails-v2.4.0
[2.3.0]: https://github.com/codeplai/VizcachaIDE/compare/wails-v2.2.0...wails-v2.3.0
[2.2.0]: https://github.com/codeplai/VizcachaIDE/compare/wails-v2.1.0...wails-v2.2.0
[2.1.0]: https://github.com/codeplai/VizcachaIDE/releases/tag/wails-v2.1.0
[1.0.0-rc1]: https://github.com/codeplai/VizcachaIDE/compare/v0.1.0...v1.0.0-rc1
[0.1.0]: https://github.com/codeplai/VizcachaIDE/releases/tag/v0.1.0
