# Plan M4 · C++ con CMake y vcpkg en VizcachaIDE

> Rama: `feature/cpp-cmake-vcpkg` (desde main 70a4576). Versión objetivo: **2.5.0**.
> Orquestador Opus; coders Sonnet en worktrees (`Agent`, `model: "sonnet"`, `isolation: "worktree"`).
> Lee antes `docs/PLAN_CPP.md` (M2: el C++ actual) y `docs/PLAN_RUST.md` §4.6 (paquetes).

## 1. Objetivo

Que un estudiante use librerías C++ como usa crates en Rust: busca `fmt`, la elige de una lista,
escribe `#include <fmt/core.h>` y F5 funciona. Para eso todo proyecto C++ es un proyecto **CMake**
y las librerías vienen de **vcpkg** (modo manifiesto). El estudiante no escribe CMake si no quiere.

## 2. Decisiones del usuario (2026-10-05)

1. **Gestor:** vcpkg (MIT, ~2.500 librerías).
2. **Todo proyecto C++ es CMake.** Proyecto nuevo = CMakeLists.txt + CMakePresets.json + vcpkg.json.
3. **Carpeta sin CMakeLists.txt:** al ejecutarla (F5, Construir, Depurar) la IDE **crea el
   CMakeLists.txt automáticamente** y desde ahí todo es CMake. (Solo un archivo sin guardar, sin
   carpeta, sigue compilando directo como hoy.)
4. **Entrega:** CMake, Ninja y vcpkg **incluidos en el instalador** full-cpp.
5. **Agregar librería:** la IDE edita `vcpkg.json` **y** un bloque marcado del CMakeLists
   (`find_package` + `target_link_libraries`). El estudiante solo hace `#include`.
6. **Estructura:** `main.cpp` en la raíz; todo `.cpp` de la carpeta entra al programa
   (`file(GLOB ... CONFIGURE_DEPENDS)`), así que crear un archivo nuevo basta.

## 3. Decisiones técnicas (library-first, todo libre)

| Pieza | Elección | Licencia |
|---|---|---|
| Sistema de construcción | CMake ≥ 3.25 (binario oficial, sin docs) | BSD-3 |
| Generador | Ninja (un solo exe) | Apache-2.0 |
| Librerías | vcpkg, **instantánea de ports sin `.git`**, modo manifiesto **sin `builtin-baseline`**: las versiones son las de la instantánea, sin git | MIT |
| Triplet Windows | `x64-mingw-static` (target **y host**, para no necesitar MSVC), compilado con llvm-mingw | — |
| Triplets otros | `x64-linux` / `arm64-osx` / `x64-osx` | — |
| Ejecutable a correr | CMake File API (`codemodel-v2`): los targets ejecutables y su ruta | — |
| Ayuda UTF-8 de consola | `CMAKE_PROJECT_TOP_LEVEL_INCLUDES=<ide>/vizcacha.cmake`, que con `cmake_language(DEFER)` agrega `vizcacha_console_utf8.cpp` a cada target ejecutable (solo Windows) | — |
| Inteligencia | clangd lee `build/compile_commands.json` (lo busca solo en `build/`) | — |

**Carpeta de construcción:** `build/` dentro del proyecto (convención CMake, como `target/` en
Rust; clangd la encuentra sola). El scaffold escribe un `.gitignore` con `build/`.
**Configuración:** una sola, `Debug` (F5 y depurar comparten construcción). "Construir ejecutable"
usa `build/release` (`Release`) y copia el `.exe` junto al código, como hoy.
**Caché de vcpkg:** descargas y binarios en `<UserCacheDir>/VizcachaIDE/vcpkg/{downloads,archives}`
(`VCPKG_DOWNLOADS`, `VCPKG_DEFAULT_BINARY_CACHE`): una librería se compila una vez por usuario.
**Configure al abrir:** al abrir un proyecto CMake (y cuando cambian CMakeLists.txt o vcpkg.json) se
configura en segundo plano, para que clangd vea los includes de las librerías antes del primer F5.

### 3.1 Plantillas

`CMakeLists.txt` (el target es el nombre de la carpeta sin acentos ni espacios: `Ñandú` → `nandu`):

```cmake
cmake_minimum_required(VERSION 3.25)
project(nandu LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# Every .cpp in this folder is part of the program: a new file is enough.
file(GLOB SOURCES CONFIGURE_DEPENDS *.cpp *.cc *.cxx)
add_executable(nandu ${SOURCES})
target_compile_options(nandu PRIVATE -Wall -Wextra)

# VizcachaIDE libraries (vcpkg) · begin
# VizcachaIDE libraries (vcpkg) · end
```

`vcpkg.json`: `{ "name": "nandu", "version": "0.1.0", "dependencies": [] }`.
`CMakePresets.json`: preset `debug` (Ninja, `build/`, toolchain `$env{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake`),
para que el terminal integrado (que exporta `VCPKG_ROOT`, cmake y ninja) pueda `cmake --preset debug`.
Los comentarios del CMakeLists van en el idioma de la IDE (ES/EN).

## 4. Diseño backend

### 4.0 Contrato C1 ↔ C2 (fijo; ninguno cambia el del otro)

```go
// package cmake (C1) — wails/internal/adapters/cpp/cmake
type Project struct { Root, Target string } // Target: nombre ASCII del ejecutable principal
func FindProject(path, stopAt string) (Project, bool)
func TargetName(folderName string) string     // "Ñandú" → "nandu", "Mi Programa" → "mi_programa"
// Dependencies lo implementa vcpkg.Setup (C2); cmake no importa vcpkg.
type Dependencies interface {
    ConfigureArgs(root string) []string           // -DCMAKE_TOOLCHAIN_FILE=…, triplets, VCPKG_INSTALL_OPTIONS
    Environment(base []string) []string           // VCPKG_ROOT, VCPKG_DOWNLOADS, … y PATH
    Prepare(ctx context.Context) error            // la semilla; idempotente y barata la 2.ª vez
}
func (b *Builder) Configure(ctx context.Context, root string) (output string, err error) // ErrConfigureFailed

// package vcpkg (C2) — wails/internal/adapters/cpp/vcpkg
type Configurer interface { Configure(ctx context.Context, root string) (string, error) } // = *cmake.Builder
const LibrariesBegin = "# VizcachaIDE libraries (vcpkg) · begin"
const LibrariesEnd   = "# VizcachaIDE libraries (vcpkg) · end"
```
`profile.go` (IDs de herramienta `cmake`, `ninja`, `vcpkg`) lo toca **solo C1**; C2 usa la constante
`"vcpkg"` y su propio localizador de raíz. Catálogo de errores: C1 agrega `CPP-CMAKE-*`, C2
`CPP-VCPKG-*` (el orquestador resuelve el conflicto de JSON al mezclar).

### 4.1 `adapters/cpp/cmake/` (track C1)
- `project.go`: `FindProject(path)` sube hasta el CMakeLists.txt raíz (sin pasar la carpeta abierta);
  `Generate(folder, language)` escribe el CMakeLists de §3.1 (+ vcpkg.json, .gitignore si faltan).
- `configure.go`: `cmake -S <root> -B <root>/build -G Ninja -DCMAKE_BUILD_TYPE=Debug
  -DCMAKE_CXX_COMPILER=<clang++> -DCMAKE_MAKE_PROGRAM=<ninja> -DCMAKE_TOOLCHAIN_FILE=<vcpkg.cmake>
  -DVCPKG_TARGET_TRIPLET=… -DVCPKG_HOST_TRIPLET=… -DCMAKE_PROJECT_TOP_LEVEL_INCLUDES=<vizcacha.cmake>`;
  reconfigura solo si cambió CMakeLists/vcpkg.json/CMakePresets o falta `build/CMakeCache.txt`.
- `fileapi.go`: escribe la query `codemodel-v2`, lee la respuesta → targets ejecutables + artefacto.
  Varios ejecutables: el que contiene el archivo activo, si no el primero.
- `filter.go`: `OutputFilter` que oculta el ruido de Ninja (`[3/7] Building…`) y deja diagnósticos.
- Runner: `Configure` de C++ con CMakeLists → `RunProject`; sin CMakeLists en carpeta guardada →
  `Generate` y luego igual. Etapas: configure → `cmake --build build` → ejecutar. `CompileForDebug`
  construye y devuelve el artefacto Debug. `Check` = build.
- Locator: `Tool(ctx, "cmake")`, `"ninja"` (configurado → incluido → junto al compilador → PATH).

### 4.2 `adapters/cpp/vcpkg/` (track C2)
- `locator.go`: raíz de vcpkg (configurada → `toolchain/cpp/vcpkg` → `VCPKG_ROOT`), triplets,
  entorno (`VCPKG_ROOT`, `VCPKG_DOWNLOADS`, `VCPKG_DEFAULT_BINARY_CACHE`, `VCPKG_DISABLE_METRICS=1`,
  PATH con llvm-mingw primero).
- `search.go`: `app.PackageSearch` **sin internet**: índice de `ports/*/vcpkg.json` (nombre,
  versión, descripción), ordenado como PyPI (empieza-con primero).
- `manifest.go`: agrega/quita en `vcpkg.json` (conserva el formato) y en el bloque marcado.
- `usage.go`: las líneas CMake de una librería: `share/<port>/usage` tras instalar (vcpkg las
  imprime: `find_package(fmt CONFIG REQUIRED)` + `target_link_libraries(main PRIVATE fmt::fmt)`),
  con el nombre del target del proyecto. Librerías header-only sin usage: `find_path` + include.
- `packages.go`: `PackageManager` (Listar/Agregar/Quitar). Agregar = editar manifiesto + configure
  (vcpkg instala en `build/vcpkg_installed`) con aviso "Compilando la librería, la primera vez tarda".
- Errores: catálogo `CPP-VCPKG-*` (sin internet, port inexistente, falló la compilación de un port,
  `find_package` no encontrado) y `CPP-CMAKE-*` (error en CMakeLists con línea, target sin fuentes).

### 4.3 Cableado (`support_cpp.go`)
`LanguageSupport.Packages/Search` para C++; `Shell` agrega cmake/ninja al PATH y `VCPKG_ROOT`;
clangd: `Manifests` + `build/compile_commands.json`; `Scaffold` nuevo.

## 5. Frontend
La pestaña Paquetes ya es genérica: aparece para C++ (buscar → lista → instalar). Texto del aviso
largo de la primera compilación. Explorador: `build/` atenuado como `target/`.

## 6. Empaquetado
full-cpp: `toolchain/cpp/cmake/` (bin + share/cmake-x.y, sin doc/man), `toolchain/cpp/bin/ninja.exe`,
`toolchain/cpp/vcpkg/` (vcpkg(.exe), ports, scripts, triplets, LICENSE; sin .git). Licencias en
THIRD_PARTY. Dev: `VIZCACHA_TEST_CMAKE_BIN`, `VIZCACHA_TEST_VCPKG_ROOT` (en `.toolchain-dev`).

## 7. Tracks
| Track | Quién | Qué |
|---|---|---|
| R0 | Orquestador | Prueba real (§8). Fija triplet, flags y tiempos. |
| C1 | Sonnet | §4.1 + tests (unitarios + integración con cmake real bajo env var). |
| C2 | Sonnet | §4.2 + tests. En paralelo con C1 (contrato: `cmake.Project{Root,Target}`, `cmake.Configure`). |
| C3 | Sonnet | Scaffold §3.1, §4.3, §5, ux_copy EN/ES, empaquetado, E2E `cpp-cmake` (nuevo proyecto → fmt → F5 → depurar → carpeta vieja autogenera CMake). |
| QA | Orquestador | QA completo de todas las fases; CHANGELOG/QA_WAILS 2.5.0. |

## 8. Resultado de R0 (2026-10-05) ✅

Entorno dev (`wails/.toolchain-dev`, ignorado por git):
- `cmake/` = CMake 4.4.4 oficial + `ninja.exe` 1.13.2 en `cmake/bin` (153 MB; sin `doc/`, `man/`,
  `share/cmake-4.4/Help`, `cmake-gui.exe` quedan ~50 MB → eso se empaqueta).
- `vcpkg/` = instantánea de microsoft/vcpkg (2026-09-26) **sin `.git` ni `versions/`**: `vcpkg.exe`,
  `ports/` (2.871 ports), `scripts/`, `triplets/`, `LICENSE.txt`, `NOTICE.txt`,
  `vcpkg.disable-metrics` y **`.vcpkg-root` (obligatorio; sin él vcpkg.cmake falla)**. 39 MB.
- `vcpkg-seed/` = descargas propias de vcpkg ya bajadas: `PowerShell-7.6.6-win-x64.zip` (101 MB),
  `7z2603-x64.7z.exe`, `tools/7zr-26.03-windows/7zr.exe`.
- Variables de test: `VIZCACHA_TEST_CMAKE_BIN=.toolchain-dev/cmake/bin`,
  `VIZCACHA_TEST_VCPKG_ROOT=.toolchain-dev/vcpkg`, `VIZCACHA_TEST_VCPKG_SEED=.toolchain-dev/vcpkg-seed`
  (y las de siempre: `VIZCACHA_TEST_LLVM_BIN`).

Hallazgos:
1. ✅ vcpkg sin git, manifiesto sin baseline, `x64-mingw-static` como target **y** host, instala
   `fmt` 12.2.0 con llvm-mingw (el toolchain `mingw.cmake` de vcpkg busca
   `x86_64-w64-mingw32-g++`, que llvm-mingw trae: **llvm-mingw debe ir primero en el PATH**; el
   triplet pasa `PATH` al build de los ports).
2. ⚠️ vcpkg descarga **PowerShell 7.6.6** (lo usa para el hash del compilador; obligatorio),
   **7-Zip** y, para ports con pkg-config, **msys2 pkgconf**. Si el archivo ya está en
   `VCPKG_DOWNLOADS`, no lo descarga (solo lo extrae, una vez por usuario: pwsh extraído = 247 MB).
   → **Semilla:** el instalador lleva `toolchain/cpp/vcpkg-seed/` y la IDE copia lo que falte a
   `VCPKG_DOWNLOADS` (archivos sueltos y `tools/`) antes del primer configure.
3. ⚠️ vcpkg escribe en su raíz (`buildtrees/`, `packages/`): instalado en Program Files falla.
   → `-DVCPKG_INSTALL_OPTIONS=--x-buildtrees-root=<cache>/vcpkg/buildtrees;--x-packages-root=<cache>/vcpkg/packages`
   (verificado: con eso la raíz queda intacta). Lo instalado va a `build/vcpkg_installed/`.
4. ✅ clangd usa `build/compile_commands.json` **aunque exista** un `compile_flags.txt` viejo, e
   incluye `-isystem build/vcpkg_installed/x64-mingw-static/include`: `#include <fmt/core.h>` sin errores.
5. ✅ `CMAKE_PROJECT_TOP_LEVEL_INCLUDES=vizcacha.cmake` con `cmake_language(DEFER DIRECTORY
   ${CMAKE_SOURCE_DIR} CALL …)` + `BUILDSYSTEM_TARGETS` agrega `vizcacha_console_utf8.cpp` solo a
   los ejecutables (visto en el log de Ninja). Archivo probado en el scratchpad de R0; C1 lo embebe.
6. ✅ File API: con `build/.cmake/api/v1/query/codemodel-v2` vacío antes de configurar, la respuesta
   trae `codemodel-v2-*.json` y `target-*.json` con el artefacto.
7. ⏱️ Tiempos: configure **frío** con fmt 178 s (incluye bajar pwsh/7zip ~100 s y compilar fmt
   56 s); **con caché binaria** 11 s; build incremental 3 s.
8. ⚠️ `CMAKE_CXX_STANDARD 17` da `-std=gnu++17`: la plantilla agrega `set(CMAKE_CXX_EXTENSIONS OFF)`.
9. ⚠️ Rutas largas: CMake avisa si `build/CMakeFiles/...` pasa 250 caracteres (carpetas muy profundas).
   Solo aviso; el catálogo lo explica si llega a fallar.
10. Acentos: carpeta `Ñandú` funciona; target ASCII `nandu`.
11. lldb-dap sobre el ejecutable Debug: lo cubre el test de integración de C1 (el build tiene `-g`).

## 9. Criterios de salida (2.5.0)
Proyecto nuevo C++ con CMake; carpeta vieja se convierte sola; buscar/agregar/quitar librería;
F5, depurar, Construir y Problemas funcionan con librerías; terminal con cmake/vcpkg; acentos;
QA completo verde; textos EN/ES; licencias.

## 10. Riesgos
- Primera compilación de una librería grande (boost) tarda minutos: aviso claro y caché binaria.
- Ports que piden herramientas de host extra (python, perl, meson): error del catálogo que lo explica.
- Proyectos con varios `main()` en la misma carpeta (ejercicios sueltos juntos) ya fallaban; el
  error de "main definido dos veces" debe sugerir separar en carpetas.
