# VizcachaIDE vs Thonny — Análisis de brechas

> Referencia: Thonny 5.0.0 (abril 2026). VizcachaIDE en el commit `ba14318`. Ver [ARQUITECTURA.md](ARQUITECTURA.md) para el detalle del código.
>
> **Alcance:** VizcachaIDE se publicará a nivel mundial con soporte **inglés y español**. Quedan **fuera de alcance** MicroPython, TinyGo y los microcontroladores. El plan ejecutable por fases y agentes está en [PLAN_DESARROLLO.md](PLAN_DESARROLLO.md).
>
> **Leyenda de estado:** ✅ tiene · 🟡 parcial / sólo UI · 🔴 simulado · ❌ falta
> **Prioridad:** P0 = imprescindible para ser "Thonny para Go" · P1 = importante · P2 = deseable · P3 = nicho

## 1. Resumen

Thonny no destaca por tener muchas funciones. Destaca por **hacer visible la ejecución** a un principiante: depurador paso a paso, variables en vivo, frames por llamada, un asistente que explica errores y un instalador con todo incluido.

Lo que VizcachaIDE ya tiene es la **carcasa**: el editor, la ejecución con `go run` y los paneles de variables y de pila. Le falta el **motor**, porque el depurador es una simulación. Hasta que no haya un depurador real con Delve, la propuesta de valor frente a Thonny no existe. Esa es la prioridad número uno.

## 2. Tabla comparativa

### 2.1 Editor

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| Resaltado de sintaxis | ✅ (regex propio) | Funciona; opcionalmente usar el lexer Go de Pygments, que ya es dependencia | — |
| Resaltado de errores de sintaxis en vivo (comillas y paréntesis sin cerrar) | ❌ | Diagnósticos de `gopls` o `gofmt -e` en segundo plano | P1 |
| Emparejado de paréntesis | ❌ (el README lo anuncia, pero no está implementado) | `QTextEdit.ExtraSelection` sobre el par | P2 |
| Resaltar apariciones del nombre y locales vs globales | ❌ | `gopls` `textDocument/documentHighlight` | P2 |
| Autocompletado inteligente (Jedi) | 🟡 (listas estáticas, ~10 paquetes) | **gopls** `textDocument/completion` | P1 |
| Call-tips (parámetros al escribir `(`) | ❌ | `gopls` `signatureHelp` | P2 |
| Ir a la definición (Ctrl+clic) | ❌ | `gopls` `definition` | P2 |
| Indentar o desindentar bloque, comentar o descomentar | ❌ | Acciones de editor simples | P2 |
| Auto-indent | ✅ (4 espacios fijos, también tras `:`) | Go usa **tabs**: integrar `gofmt` y respetar `tab_size` | P1 |
| Formatear código | ❌ | `gofmt` / `goimports` al guardar | **P0** (es idiomático en Go) |
| Buscar / Reemplazar, Ir a línea | ❌ | `QDialog` no modal + Ctrl+F / Ctrl+H / Ctrl+G | P1 |
| Pestañas múltiples | ✅ | — | — |
| Archivos recientes | 🟡 (sólo el último) | Lista MRU en `QSettings` | P2 |
| Recarga automática por cambios externos | ❌ | `QFileSystemWatcher` | P2 |
| Ejecutar sin guardar ("Untitled") | ❌ (obliga a guardar) | Escribir en un temporal y ejecutar `go run` sobre él | P1 |
| Guías de longitud de línea, zoom Ctrl +/- | ❌ | Simple | P3 |
| Imprimir | ❌ | `QPrintDialog` | P3 |

### 2.2 Ejecución y Shell

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| Run (F5) | ✅ `go run` | Ampliar a paquete o módulo (`go run .`) | — |
| Entrada stdin en la consola | ✅ | — | — |
| Stop / Interrupt | 🟡 (sólo `kill`) | Enviar Ctrl+C / SIGINT antes de matar el proceso | P2 |
| **Shell / REPL interactivo** | ❌ | Go no trae REPL oficial: embeber **yaegi** (intérprete Go) o **gomacro** | P1 (muy de Thonny, pero difícil) |
| Argumentos del programa | ❌ | Campo "Program arguments" pasado a `go run archivo.go -- args` | P1 |
| Ejecutar en terminal del sistema / abrir terminal con PATH | ❌ | Lanzar `cmd` o `wt` con el entorno Go configurado | P2 |
| Colores ANSI y `\r` (barras de progreso) | ❌ | Parser ANSI en `ConsoleWidget` | P2 |
| Enlaces en los errores hacia la línea | ❌ | Parsear `archivo.go:L:C` en el stderr del compilador o en el panic y hacerlo clicable | **P0** |
| Plotter de números impresos | ❌ | Vista con `QtCharts` / pyqtgraph | P3 |
| Compilar | ✅ (`go build`, pero bloquea la UI) | Pasarlo a `QProcess` | P1 |
| Tests | ❌ (Thonny tampoco lo tiene) | `go test` con panel de resultados: valor añadido propio de Go | P2 |

### 2.3 Depurador — el núcleo de Thonny

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| Depuración real por líneas | 🔴 **simulada** | **Delve**: `dlv dap` (Debug Adapter Protocol) o `dlv debug --headless --api-version=2` con JSON-RPC | **P0** |
| Step over / into / out | 🔴 (línea aleatoria) | Comandos `next` / `step` / `stepOut` de Delve | **P0** |
| Breakpoints | 🟡 (se dibujan; el depurador los ignora) | `setBreakpoints` (DAP) | **P0** |
| Resume y Run to cursor | ❌ | `continue` + breakpoint temporal | P1 |
| Resaltado de la línea actual | ✅ (UI) | Alimentarlo con el `stopped` real de Delve | **P0** |
| Panel de Variables | 🟡 (UI lista, datos falsos) | `scopes` / `variables` (DAP), con expansión perezosa de structs, slices y maps | **P0** |
| Panel de Stack (pila) | 🟡 (UI lista, datos falsos) | `stackTrace`; clic en un frame lleva al código | **P0** |
| **Frames como ventanas anidadas** (recursión visible) | ❌ | Factible con datos de `stackTrace` y `variables` por frame | P1 — **diferenciador** |
| **Evaluación de expresiones paso a paso** ("nicer debug") | ❌ | Muy difícil en un lenguaje compilado. Alternativa: instrumentar el AST con `go/ast` y emitir trazas, o usar **yaegi** como motor de "modo didáctico" | P2 — investigación |
| Step back | ❌ | `rr` + Delve (sólo Linux); o grabar los snapshots de cada paso y navegarlos en la UI | P3 |
| Heap / modelo de referencias (nombre → id → valor) | ❌ | Mostrar direcciones de punteros y punteros dentro de slices y maps: muy pedagógico en Go (punteros, slices que comparten un array) | P2 — diferenciador |
| Inspector de objetos | ❌ | Panel con detalle del valor seleccionado: `len` y `cap` de un slice, campos de un struct | P2 |
| Goroutines | ❌ (Python no las tiene) | Vista de goroutines de Delve: **valor añadido propio de Go** | P1 |

### 2.4 Ayuda para principiantes

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| **Assistant**: explica errores en lenguaje llano | ❌ | Catálogo de errores frecuentes del compilador Go (`declared and not used`, `imported and not used`, `missing return`, `cannot use x (type) as …`, panics `index out of range` y `nil map`) con explicación **bilingüe EN/ES** (se muestra también el mensaje original de Go) | **P0** — diferenciador |
| Avisos estáticos (Pylint, MyPy) | ❌ | `go vet` + `staticcheck` y mostrarlos en el Assistant | P1 |
| Aviso de "sombra de módulo" | ❌ | Equivalente en Go: avisar de un `package` distinto de `main`, de que falta `func main` o de que falta `go.mod` | P2 |
| Vista Outline | ❌ | `gopls documentSymbol` o `go/ast` | P2 |
| Vista TODO, Notes, Help integrada | ❌ | Simples; la ayuda puede enlazar a *A Tour of Go* | P3 |

### 2.5 Paquetes, intérpretes y entorno

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| **Python incluido en el instalador** | ❌ (exige instalar Go y Delve a mano) | Empaquetar un **toolchain Go + dlv + gopls** portable dentro del instalador | **P0** para principiantes |
| Gestor de paquetes gráfico (pip) | ❌ | Interfaz para `go mod init`, `go get`, `go mod tidy` y búsqueda en pkg.go.dev | P1 |
| Selector de intérprete y venvs | 🟡 (ruta de Go configurable) | Selector de versión de Go (toolchains múltiples, `GOTOOLCHAIN`) en la barra de estado | P2 |
| Proyecto con varios archivos | ❌ (sólo un archivo suelto) | Abrir carpeta, detectar `go.mod` y ejecutar `go run .` | P1 |
| Remoto por SSH | ❌ | `GOOS`/`GOARCH` + `scp` + ejecución remota | P3 |
| Vista Files (local y remoto) | ❌ | `QTreeView` + `QFileSystemModel` | P1 |

### 2.6 Apariencia, i18n y distribución

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| Temas claro y oscuro | 🟡 (se guardan, **no se aplican**) | Aplicar la paleta y los colores del highlighter desde `QSettings` | P1 (bajo esfuerzo) |
| Fuentes configurables | ✅ | — | — |
| Modos simple / regular / expert | ❌ | Ocultar menús y paneles según el nivel elegido | P2 |
| Escalado HiDPI | ❌ | `Qt.AA_EnableHighDpiScaling` | P2 |
| **Internacionalización inglés/español** | ❌ (UI sólo en inglés) | gettext + Babel; detección automática del idioma del sistema y selector; errores, ayuda, README e instalador en EN/ES | **P0** (publicación mundial) |
| Diálogo de primer arranque | ❌ | Idioma, tema y detección o instalación de Go y Delve | P1 |
| Instalador y versión portable | ❌ | PyInstaller/Nuitka + Inno Setup; zip portable | **P0** |
| Plugins | ❌ | Entry points `vizcacha.plugins` con `load_plugin(workbench)` | P3 |

### 2.7 Aula e investigación

| Función en Thonny | VizcachaIDE | Equivalente en Go / cómo hacerlo | Prioridad |
|---|---|---|---|
| Logs de acciones y exportación | ❌ | Registro JSONL opcional | P3 |
| Replayer | ❌ | Reproducir esos logs | P3 |
| `defaults.ini` para aulas | ❌ | `QSettings` con un archivo de defaults del sistema | P3 |

## 3. Hoja de ruta sugerida

> El detalle por fases, contratos y tracks paralelos está en [PLAN_DESARROLLO.md](PLAN_DESARROLLO.md).

**Fase 1: que lo anunciado funcione (P0)**
1. Depurador real con `dlv dap`: breakpoints, pasos, variables, pila y línea actual. Las señales Qt actuales ya encajan.
2. Errores del compilador y panics clicables, más un **Assistant** con explicaciones bilingües EN/ES de los ~25 errores más comunes.
3. `gofmt` al guardar e indentación con tabs.
4. Instalador para Windows con Go, Delve y gopls incluidos.

**Fase 2: paridad práctica con Thonny (P1)**
5. gopls por LSP: autocompletado real, diagnósticos en vivo y hover.
6. Aplicar los temas y opciones ya guardados, interfaz bilingüe EN/ES y asistente de primer arranque.
7. Proyectos con `go.mod`, vista Files e interfaz de `go get` / `go mod tidy`.
8. Frames en ventanas anidadas y vista de goroutines.
9. Argumentos del programa, Buscar/Reemplazar, Run to cursor y `go vet`/`staticcheck`.
10. REPL con yaegi.

**Fase 3: diferenciadores propios de Go (P2+)**
11. Visualizador de memoria: punteros, slices que comparten array, `len` y `cap`.
12. Panel de `go test`.
13. Modo didáctico de evaluación de expresiones (yaegi o instrumentación del AST).
14. Modos simple/expert, plugins y logs de aula.

## 4. Fuentes sobre Thonny

- https://thonny.org · https://thonny.org/blog/
- CHANGELOG oficial: https://github.com/thonny/thonny/blob/master/CHANGELOG.rst
- Notas de versión 3.1 a 5.0: https://newreleases.io/project/github/thonny/thonny
- https://pypi.org/project/thonny/ · https://en.wikipedia.org/wiki/Thonny
- Wiki: Plugins, User-action-logs y DeploymentOptions en https://github.com/thonny/thonny/wiki
- Annamaa, A. (2015). *Thonny, a Python IDE for learning programming*. ACM Koli Calling. https://dl.acm.org/doi/10.1145/2828959.2828969
