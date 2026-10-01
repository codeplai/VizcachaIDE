# VizcachaIDE — Arquitectura actual

> Estado del código a la fecha del commit `ba14318` (2025-10-01). Documento generado el 2026-09-30 a partir de la lectura del código fuente, no del README.

![Diagrama de arquitectura](arquitectura.svg)

## 1. Resumen

VizcachaIDE es una aplicación de escritorio **Python 3 + PyQt5** que pretende ser un "Thonny para Go". Tiene ~2 800 líneas en 12 módulos, organizadas en dos paquetes:

| Paquete | Responsabilidad | LOC |
|---|---|---|
| `gui/` | Widgets Qt: ventana principal, editor, consola, paneles de depuración, diálogo de configuración | ~2 140 |
| `core/` | Ejecución de `go run`, depurador, analizador para autocompletado | ~620 |

No hay tests, empaquetado (PyInstaller está previsto en el código pero no hay `.spec`), CI ni internacionalización. Todo el texto de la UI está en inglés.

## 2. Estructura de archivos

```
VizcachaIDE/
├── main.py                 # Crea QApplication y MainWindow
├── requirements.txt        # PyQt5, Pygments (Pygments no se usa)
├── logo.png / logo.ico
├── hola.go                 # archivo suelto de prueba
├── examples/               # 6 programas Go de ejemplo
├── core/
│   ├── runner.py           # GoRunner  – go run vía QProcess
│   ├── debugger.py         # GoDebugger – SIMULADO; DelveAPIDebugger – stub
│   └── go_analyzer.py      # GoAnalyzer – autocompletado con diccionarios estáticos
└── gui/
    ├── main_window.py      # MainWindow – orquestador
    ├── tabbed_editor.py    # TabbedEditor – pestañas y archivos
    ├── editor.py           # CodeEditor, LineNumberArea, GoSyntaxHighlighter
    ├── autocomplete.py     # AutocompleteWidget + delegate de pintado
    ├── console.py          # ConsoleWidget – salida y stdin
    ├── variables.py        # VariablesWidget – árbol de variables
    ├── callstack.py        # CallStackWidget – lista de frames
    └── settings_dialog.py  # SettingsDialog – 3 pestañas de configuración
```

## 3. Componentes

### 3.1 Presentación (`gui/`)

| Componente | Base Qt | Qué hace | Estado |
|---|---|---|---|
| `MainWindow` | `QMainWindow` | Arma el layout (editor + panel derecho arriba, consola abajo, con `QSplitter`), menús *File / Edit / Run / Debug / Help*, toolbar, habilita/deshabilita acciones, conecta señales. Contiene `build_code()` que llama `go build` con `subprocess.run` **en el hilo de UI**. | Funcional |
| `TabbedEditor` | `QTabWidget` | Abrir/guardar/cerrar pestañas, marca `*` de modificado, recopila breakpoints, reenvía el resaltado de línea. | Funcional |
| `CodeEditor` | `QPlainTextEdit` | Números de línea con clic para breakpoint, resaltado de sintaxis por regex propio (`GoSyntaxHighlighter`), auto‑indent (4 espacios fijos, también tras `:`), Ctrl+Space para autocompletar, resaltado amarillo de línea actual. | Funcional |
| `AutocompleteWidget` | `QListWidget` | Popup con nombre, firma y documentación; navegación con teclado. | Funcional |
| `ConsoleWidget` | `QTextEdit` | Salida coloreada (stdout/stderr/éxito) y entrada de usuario para `stdin`, emitiendo `input_submitted`. | Funcional |
| `VariablesWidget` | `QTreeWidget` | Árbol nombre/tipo/valor con hijos expandibles. | UI lista, datos falsos |
| `CallStackWidget` | `QListWidget` | Lista de frames. Sin navegación al código. | UI lista, datos falsos |
| `SettingsDialog` | `QDialog` | Pestañas Entorno / Editor / Apariencia, persiste en `QSettings`. | Parcial (ver §5) |

### 3.2 Lógica (`core/`)

| Componente | Qué hace | Estado |
|---|---|---|
| `GoRunner` | Lanza `go run <archivo>` con `QProcess` en la carpeta del archivo; inyecta `GOPATH`, `GOROOT` y variables extra; lee stdout/stderr de forma asíncrona; `write_input()` alimenta stdin; `stop()` hace `kill()`. | Funcional |
| `GoDebugger` | Verifica que `dlv version` responda y **nada más**: no arranca Delve. `step_over/into/out` emiten una **línea aleatoria entre 1 y 50**, variables y pila **hardcodeadas**. Ignora los breakpoints. | **Simulado** |
| `DelveAPIDebugger` | Esqueleto para Delve headless + JSON‑RPC (`--listen=127.0.0.1:2345`). Métodos vacíos; hace *fallback* a `GoDebugger`. | **Stub** |
| `GoAnalyzer` | Autocompletado con listas estáticas: keywords, tipos, builtins y ~10 paquetes de la stdlib (`fmt`, `strings`, `os`, `math`…) con documentación corta. Detecta contexto `paquete.`. No conoce los símbolos del código del usuario. `get_completions_with_gopls()` es un stub. | Parcial |

### 3.3 Persistencia y dependencias externas

- **`QSettings`** (registro de Windows / `.ini`): claves `env/*`, `editor/*`, `appearance/*`, `last_file`.
- **Toolchain Go**: `go run` (GoRunner) y `go build` (MainWindow).
- **Delve**: sólo se comprueba su existencia.
- **gopls**: no se usa.

## 4. Flujos principales

### Ejecutar (F5)
1. `MainWindow.run_code()` exige que el archivo esté guardado, limpia la consola y deshabilita *Run*.
2. `GoRunner.run()` crea un `QProcess` → `go run archivo.go`.
3. `readyReadStandardOutput/Error` → señales `output_received` / `error_received` → `ConsoleWidget`.
4. El usuario escribe en la consola → `input_submitted` → `GoRunner.write_input()` → stdin.
5. `finished` → `execution_finished(code)` → `MainWindow.on_execution_finished()` reactiva botones.

### Depurar (F6)
1. `MainWindow.start_debug()` toma breakpoints de todas las pestañas y llama a `GoDebugger.start()`.
2. Se ejecuta `dlv version` (bloqueante, timeout 5 s).
3. Con `QTimer` se emite línea 1 y una variable placeholder.
4. Cada *Step* emite una línea aleatoria + datos de ejemplo. **El programa del usuario nunca se ejecuta.**

### Autocompletar (Ctrl+Space)
`CodeEditor` importa perezosamente `GoAnalyzer` y `AutocompleteWidget` → `get_completions(texto, cursor)` → top 20 → popup → `insert_completion()` reemplaza el prefijo.

### Comunicación entre capas
Todo el acoplamiento core → UI es por **señales Qt** conectadas en `MainWindow.connect_signals()`. Es un patrón limpio y conviene mantenerlo al reemplazar el depurador.

## 5. Brechas entre README / configuración y el código

| Lo que se anuncia | Realidad en el código |
|---|---|
| Depuración paso a paso con Delve | Simulada (líneas aleatorias, datos fijos) |
| Breakpoints | Se dibujan y recolectan, pero el depurador los ignora |
| Temas Light/Dark/Solarized, tema de consola, colores personalizados | Se guardan en `QSettings`, nunca se aplican |
| Tamaño de tab, auto‑indent on/off, mostrar números de línea, barra de estado | Se guardan, nunca se leen fuera del diálogo |
| Ruta de Delve configurable | Se guarda; el depurador usa `dlv` hardcodeado |
| Resaltado con Pygments | Pygments está en `requirements.txt` pero no se importa |
| Lista de archivos recientes | Sólo se recuerda el último archivo |
| Emparejado de llaves y plegado de código (*code folding*) | No implementados en `editor.py` |
| `go build` respeta variables de entorno configuradas | No: `build_code()` no inyecta GOPATH/GOROOT/extras y bloquea la UI |

## 6. Observaciones técnicas

- **Monoproceso, sin hilos**: correcto para `go run` (QProcess es asíncrono), pero `build_code()` y la verificación de `dlv` congelan la UI.
- **Imports dentro de funciones** (`import os` en varios métodos) y lógica de entorno duplicada entre `GoRunner` y `build_code()`; conviene un `GoEnvironment` común.
- **Archivo único**: `go run archivo.go` no soporta paquetes con varios archivos ni `go.mod` del proyecto.
- **PyQt5** está en mantenimiento; migrar a **PySide6/PyQt6** es razonable antes de crecer.
- **Sin tests** ni separación modelo/vista para el depurador, lo que dificulta reemplazar el simulador.
- Carpeta `.kilo/worktrees/` (no versionada) contiene una copia del código; no forma parte de la arquitectura.

## 7. Siguientes pasos sugeridos (por impacto)

1. **Depurador real**: `dlv debug --headless --api-version=2 --listen=127.0.0.1:<puerto>` + cliente JSON‑RPC (o DAP: `dlv dap`) en un `QThread`/`QTcpSocket`, reutilizando las señales existentes.
2. **gopls vía LSP** para autocompletado, diagnósticos en vivo, ir a definición y hover.
3. **Aplicar la configuración ya guardada** (temas, tab, número de líneas, ruta de Delve) — bajo esfuerzo, alto impacto visible.
4. Unificar entorno Go y mover `go build` a `QProcess`.
5. `gofmt` al guardar y `go vet` como verificador.
6. **Internacionalización inglés/español** (interfaz, errores, ayuda, instalador), porque el proyecto se publicará a nivel mundial.

El plan detallado, con la arquitectura objetivo y los tracks para agentes en paralelo, está en [PLAN_DESARROLLO.md](PLAN_DESARROLLO.md) ([diagrama](arquitectura-objetivo.svg)).

Ver la comparación de funcionalidades contra Thonny en [COMPARATIVA_THONNY.md](COMPARATIVA_THONNY.md).
