# Plan M1 · Python en VizcachaIDE

> Requiere [PLAN_NUCLEO_MULTILENGUAJE.md](PLAN_NUCLEO_MULTILENGUAJE.md) (M0) terminado: contrato v3,
> `internal/protocol`, `LanguageRegistry` y el frontend por perfiles. Si no está hecho, hazlo
> primero. Diseño general en [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §4.5 y en el
> diagrama [arquitectura-multilenguaje.svg](arquitectura-multilenguaje.svg). Escrito el 2026-10-03.

## 0. Cómo usar este plan en un chat nuevo

1. Comprueba que M0 está integrado: existe `wails/internal/protocol/`, `wails/support_go.go` y
   `app.LanguageRegistry`. Si no, ejecuta primero el plan M0.
2. Lee: este archivo; [EXTENSION_MULTILENGUAJE.md](EXTENSION_MULTILENGUAJE.md) §4 y §8;
   [`wails/README.md`](../wails/README.md); [`docs/wails/PLAN_WAILS.md`](wails/PLAN_WAILS.md) §2 y §4;
   [`docs/wails/UX_COPY.md`](wails/UX_COPY.md); y el adaptador de Go en `wails/internal/adapters/golang/`
   como ejemplo de cada pieza.
3. Para desarrollar y probar hace falta un Python 3.10 o más en la máquina con `debugpy`,
   `python-lsp-server` y `ruff` instalados (`pip install debugpy python-lsp-server ruff`). Los tests de
   integración se saltan solos si faltan.
4. Verificación antes de cada commit:
   ```sh
   cd wails && go vet ./... && go test ./... && golangci-lint run
   cd frontend && npm run check && npm run test
   wails build
   ```
5. Reglas de siempre: archivos de menos de 200 líneas, funciones de menos de 50, retorno temprano,
   nombres de dominio, library-first, textos sólo por i18n. Commits locales en inglés, sin push salvo
   indicación.
6. Resultado esperado: **versión 2.2.0** con Python como segundo lenguaje y la variante
   `full-python` empaquetada.

## 1. Objetivo

Que un principiante escriba, ejecute, entienda sus errores y depure un programa Python con la misma
experiencia que tiene con Go: F5 ejecuta, el Assistant explica en EN/ES, F6 depura paso a paso con
variables que marcan "acaba de cambiar", hay sugerencias y problemas en vivo, y una consola
interactiva como la de Thonny.

## 2. Alcance

Dentro:

- Ejecutar un archivo `.py` (con entrada por teclado), detener, argumentos del programa, archivo sin
  título.
- Depurar con debugpy: breakpoints, siguiente línea, entrar, salir, continuar, ejecutar hasta aquí,
  variables (con hijos perezosos), pila, vista de llamadas, parada en excepción no capturada.
- Inteligencia de código con python-lsp-server: diagnósticos en vivo (pyflakes), completado, hover,
  firma, definición, ocurrencias, símbolos.
- Formato al guardar con `ruff format`; `ruff check` después de ejecutar (como `go vet`).
- Consola interactiva (REPL) con la sesión persistente.
- Gestor de paquetes con pip: instalar, desinstalar, listar.
- Catálogo de unas 25 errores de Python explicados en EN/ES.
- Empaquetado `full-python` para Windows, macOS y Linux, y detección de un Python ya instalado en la
  variante `lite`.

Fuera (primera versión): entornos virtuales por proyecto (sólo se detecta `.venv`), Jupyter o
notebooks, tipado estático (pyright), `pyproject.toml` como proyecto, gráficos.

## 3. Decisiones (library-first) y licencias

| Necesidad | Decisión | Licencia | Por qué |
|---|---|---|---|
| Intérprete | CPython 3.12.x empaquetado desde **python-build-standalone** (archivo `install_only`) | PSF | builds portables para los tres sistemas, es lo que usa `uv`; sin instalador |
| Depurador | **debugpy** (`python -m debugpy.adapter` por stdio) | MIT | DAP nativo, mantenido con VS Code; `runInTerminal` para que `input()` funcione |
| Inteligencia | **python-lsp-server** (`python -m pylsp`) con jedi y pyflakes | MIT | puro Python, se instala con pip en el intérprete empaquetado; sin Node |
| Formato y check | **ruff** (`python -m ruff format` / `check`) | MIT | un binario, rápido; una sola herramienta para las dos cosas |
| Consola | módulo `code` de la stdlib en un script auxiliar embebido | PSF | la stdlib ya trae el intérprete interactivo; el script sólo añade JSON por stdio |
| Paquetes | `python -m pip` | PSF | viene con el intérprete |
| Editor | `@codemirror/lang-python` | MIT | igual que `lang-go` |

Alternativas descartadas por ahora: pyright o basedpyright (mejor análisis, pero requieren Node o un
paquete pesado), IPython o ipykernel para la consola (demasiado grandes para el paquete), cling no
aplica.

## 4. Diseño: `wails/internal/adapters/python/`

```
adapters/python/
├── profile.go        LanguageProfile de Python
├── locator.go        encuentra el intérprete: configurado → empaquetado → .venv del proyecto → PATH
├── environment.go    variables de entorno para los procesos de Python
├── runner/           ProgramRunner sobre protocol/process (modo PTY)
├── debugpy/          Flavor DAP + Transport stdio + ReverseHandler (runInTerminal)
├── pylsp/            Flavor LSP
├── ruff/             Format y Check
├── repl/             Console: proceso auxiliar + repl.py embebido
├── packages/         PackageManager con pip
└── errors/           OutputParser de tracebacks + data/catalog.python.json
```

### 4.1 Perfil

```go
var Profile = domain.LanguageProfile{
    ID: domain.CodeLanguagePython, NameKey: "codeLanguage.python", Extensions: []string{".py", ".pyw"},
    Indent: domain.IndentStyle{UseTabs: false, Size: 4},
    Capabilities: domain.Capabilities{Build: false, Console: true, Format: true, Check: true,
        PackageActions: []domain.PackageAction{domain.PackageAdd, domain.PackageRemove, domain.PackageList},
        ThreadsLabel: "debug.threads"},
    Tools: []domain.ToolSpec{
        {ID: "python", Role: domain.RoleRuntime, LabelKey: "settings.toolPython",
            MissingKey: "errors.pythonNotFound", InstallURL: "https://www.python.org/downloads/"},
        {ID: "debugpy", Role: domain.RoleDebugAdapter, ProvidedBy: "python", LabelKey: "settings.modulePython",
            MissingKey: "errors.debugpyMissing", InstallCommand: "python -m pip install debugpy"},
        {ID: "pylsp", Role: domain.RoleLanguageServer, ProvidedBy: "python", LabelKey: "settings.modulePython",
            MissingKey: "errors.pylspMissing", InstallCommand: "python -m pip install python-lsp-server"},
        {ID: "ruff", Role: domain.RoleFormatter, ProvidedBy: "python", LabelKey: "settings.modulePython",
            MissingKey: "errors.ruffMissing", InstallCommand: "python -m pip install ruff"},
    },
}
```

Un solo ejecutable: debugpy, pylsp y ruff son módulos del mismo intérprete (`python -m …`), así que
sólo hay que localizar `python`; por eso sus `ToolSpec` llevan `ProvidedBy: "python"` (M0 §3.1):
tienen fila y aviso propios, pero Ajustes no ofrece "Elegir" para ellos. `Tools()` devuelve un
`ToolStatus` por módulo con su versión, obtenido con un solo `python -c` que imprime las tres versiones en
JSON; así Ajustes muestra qué falta.

### 4.2 Localizador (`locator.go`)

Orden: ruta configurada (`ToolPaths["python"]`) → empaquetado (`toolchain/python/python.exe` en
Windows, `toolchain/python/bin/python3` en Unix) → `.venv` de la carpeta del archivo (`.venv/Scripts/python.exe`
o `.venv/bin/python`) → PATH (`python3`, luego `python`). Cada candidato se valida con
`<python> -c "import sys; print(sys.version_info[:2])"` con 5 s de tiempo límite y se exige 3.10 o más.

Windows: el alias `python.exe` de `WindowsApps` abre la Microsoft Store cuando no hay Python; un
candidato en una ruta que contiene `WindowsApps` sólo vale si la validación anterior responde. Si
existe el lanzador `py`, `py -3 -c "import sys; print(sys.executable)"` da la ruta real y se prefiere.

### 4.3 Entorno (`environment.go`)

Para todos los procesos de Python: `PYTHONUTF8=1`, `PYTHONIOENCODING=utf-8`, `PYTHONUNBUFFERED=1`,
`PYTHONDONTWRITEBYTECODE=1` (nada de `__pycache__` en las carpetas de los alumnos). Con el intérprete
empaquetado, además `PYTHONNOUSERSITE=1` para no mezclar paquetes de otro Python del sistema.

### 4.4 Ejecutor (`runner/`)

- `Configure(path, args)`: `CodeLanguage: python`, `Mode: file`, `Target: path`, `WorkingDir: carpeta`,
  `Project: {Root: carpeta, Kind: "folder"}` (si hay `.venv`, `Name: ".venv"`). Nunca hay modo
  proyecto: Python ejecuta un archivo.
- `Run`: en el `Supervisor` compartido (M0 §4.1), `Job{Command: python, Args: ["-X", "utf8", path, args...], Mode: Terminal}` con
  `Echo: true` en la configuración (la PTY devuelve el eco; ver M0 §7). Si la PTY no está disponible
  (M0 la dejó en `Pipes`), `Mode: Pipes` y `-u` ya cubre la salida.
- `RunUntitled`: escribe el texto en `<temp>/vizcacha_py_<pid>/main.py` y ejecuta; `Cleanup` borra la
  carpeta.
- `Build`: `ErrUnsupported`.
- `Check` y `Format` no son del runner: los implementa `ruff/` como `app.CodeChecker` y
  `app.CodeFormatter` (M0 §3.5):
  - `Check`: `python -m ruff check --output-format json <target>`; devuelve el JSON (el parser de
    §4.9 lo entiende); `""` si no hay nada o ruff no está.
  - `Format(path, text)`: `python -m ruff format --stdin-filename <nombre de path> -` con el texto
    por stdin; errores de sintaxis → `ErrFormat` con la línea (el frontend ya muestra
    `errors.formatRejected`).
- `Stop`: lo hace `protocol/process` (Ctrl+C por la PTY, luego árbol de procesos).

### 4.5 Depurador (`debugpy/`)

- Transport: stdio sobre `python -m debugpy.adapter` (sin `--port` habla DAP por stdin/stdout).
- `Flavor.AdapterID()` = `"python"`. `Launch` produce:
  ```json
  {"request":"launch","type":"python","name":"VizcachaIDE","program":"<path>","cwd":"<dir>",
   "args":[...],"python":["<interprete>"],"console":"integratedTerminal","justMyCode":true,
   "stopOnEntry":false,"redirectOutput":false,"env":{...}}
  ```
- `console: integratedTerminal` hace que el adaptador envíe la petición inversa **`runInTerminal`**
  con el comando del lanzador de debugpy. El `ReverseHandler` del adaptador la contesta arrancando
  ese comando con `protocol/process` en modo PTY (misma salida y entrada que Ejecutar) y devolviendo
  el pid. Así `input()` funciona depurando. Si la PTY no existe, `console: internalConsole`
  y el aviso `run.debugStdin` de Go (sin teclado al depurar).
- `ExceptionFilters()` = `["uncaught"]`: una excepción no capturada para el programa en la línea
  culpable con `StopException` y la descripción `NameError: name 'x' is not defined`. Al continuar,
  el programa termina y el traceback llega por la PTY, así que el Assistant lo explica igual que en
  Ejecutar.
- `KeepVariable` oculta los grupos `special variables`, `function variables`, `class variables` y
  los nombres `__dunder__` (como hace Thonny). `IsLocalsScope` acepta `Locals`; `Globals` se expone
  como un segundo grupo plegado sólo en el frame de nivel módulo (opcional en esta versión).
- `KeepFrame` deja todos (justMyCode ya filtra la stdlib). `StopReason` mapea `breakpoint`, `step`,
  `exception`, `pause`, `entry`.
- `Output`: categorías `stdout`/`stderr`; lo demás a `console` (mensajes del adaptador).

### 4.6 Inteligencia (`pylsp/`)

- `Command`: `python -m pylsp` (sin `--tcp`). `RootOf(path)` = carpeta del archivo (o la carpeta
  abierta en Files si el archivo está dentro).
- `Configuration()` enviado como `workspace/didChangeConfiguration`:
  ```json
  {"pylsp":{"plugins":{"pyflakes":{"enabled":true},"pycodestyle":{"enabled":false},
   "mccabe":{"enabled":false},"pylint":{"enabled":false},"jedi_completion":{"include_params":false},
   "jedi_signature_help":{"enabled":true},"jedi_symbols":{"enabled":true}}}}
  ```
  pycodestyle apagado a propósito: los avisos de estilo abruman a un principiante.
- El mapeo genérico de `protocol/lsp` ya convierte `publishDiagnostics` (pyflakes: `undefined name`,
  `imported but unused`), completion (kinds de LSP), hover (markdown), signatureHelp,
  documentHighlight y documentSymbol. Verificar con fixtures grabadas de pylsp real
  (`testdata/pylsp/*.json`), igual que hoy con gopls.

### 4.7 Consola (`repl/`)

- `repl.py` embebido con `go:embed`, copiado a `<UserCacheDir>/VizcachaIDE/repl/repl.py` al primer
  uso. Se ejecuta `python -X utf8 -u repl.py`. Protocolo de líneas JSON por stdio: petición
  `{"code":"..."}`, respuesta `{"result":"...","output":"...","error":""}`.
- Dentro: `code.InteractiveInterpreter`; primero `compile(src, "<console>", "eval")`: si compila, se
  evalúa y `result = repr(valor)` salvo `None`; si no, `compile(..., "exec")` y se ejecuta.
  `stdout`/`stderr` se capturan con `contextlib.redirect_stdout`/`redirect_stderr`; el error es el
  traceback sin los frames del propio `repl.py`. `input()` dentro de la consola devuelve un error
  claro (`errors.consoleNoInput`).
- Go: `Console.Eval` escribe la petición y espera la respuesta con el tiempo límite de 5 s de la
  consola de Go; si vence, mata el proceso y lo olvida (`Reset` implícito), igual que yaegi.
  `Reset` mata el proceso. El proceso se crea perezosamente en el primer `Eval`.

### 4.8 Paquetes (`packages/`)

`Add`: `python -m pip install <pkg>`; `Remove`: `python -m pip uninstall -y <pkg>`; `List`:
`python -m pip list`. `Init` y `Tidy`: `ErrUnsupported`. Argumentos validados con
`app.SingleWordArgument` (sin espacios, sin `-` inicial). Corren por el `Supervisor` compartido y emiten
`run:*`, así que la salida aparece en el diálogo de Paquetes como hoy la de `go mod`. Con el
intérprete empaquetado, pip instala en su `site-packages` (instalación por usuario, carpeta
escribible); con un `.venv` detectado, en el `.venv`.

### 4.9 Errores (`errors/`)

Parser de la salida de Python (`parser.go`, con `protocol/errorcatalog.{SplitLines,LocationFrom,NamedGroups}`):

- Traceback: desde `Traceback (most recent call last):` hasta la última línea `Tipo: mensaje`.
  `Location` = el último frame `File "<ruta>", line N, in <f>` cuya ruta está dentro de `workingDir`
  o del archivo ejecutado (nunca la stdlib). `Message` = la última línea. `RawText` = todo el bloque.
  `Source` = `"runtime"`.
- Errores de sintaxis sin traceback: `File "<ruta>", line N` + línea de código + `^` + `SyntaxError: …`
  (también `IndentationError`, `TabError`). `Source` = `"syntax"`.
- `ruff check` (JSON de `Check`): un diagnóstico por entrada con `Code` = el código de ruff (`F821`,
  `F841`, `F401`, `E999`), `Severity` warning, `Source` = `"ruff"`.
- Ruido ignorado: `python: can't open file` (se convierte en diagnóstico sin ubicación), avisos
  `DeprecationWarning` (fuera de alcance).

Catálogo `data/catalog.python.json`, mismos campos que el de Go (`id`, `patterns` con grupos
nombrados, `en`, `es` con `title`, `body`, `fix`). Ids propuestos, unas 25 entradas:

| Id | Patrón (resumen) |
|---|---|
| `PY-NAME-ERROR` | `NameError: name '(?P<name>\w+)' is not defined` |
| `PY-UNBOUND-LOCAL` | `UnboundLocalError: … '(?P<name>\w+)' referenced before assignment` / `cannot access local variable` |
| `PY-INDENTATION` | `IndentationError: (expected an indented block\|unexpected indent\|unindent does not match)` |
| `PY-TAB-ERROR` | `TabError: inconsistent use of tabs and spaces` |
| `PY-SYNTAX-INVALID` | `SyntaxError: invalid syntax` |
| `PY-SYNTAX-MISSING-COLON` | `SyntaxError: expected ':'` |
| `PY-SYNTAX-UNCLOSED` | `SyntaxError: '(?P<open>[(\[{])' was never closed` |
| `PY-SYNTAX-STRING` | `SyntaxError: unterminated string literal` |
| `PY-SYNTAX-ASSIGN` | `SyntaxError: cannot assign to …` / `invalid syntax. Maybe you meant '=='` |
| `PY-TYPE-OPERANDS` | `TypeError: unsupported operand type\(s\) for (?P<op>\S+): '(?P<left>\w+)' and '(?P<right>\w+)'` |
| `PY-TYPE-CONCAT` | `TypeError: can only concatenate str \(not "(?P<type>\w+)"\) to str` |
| `PY-TYPE-NOT-SUBSCRIPTABLE` | `TypeError: '(?P<type>\w+)' object is not subscriptable` |
| `PY-TYPE-NOT-CALLABLE` | `TypeError: '(?P<type>\w+)' object is not callable` |
| `PY-TYPE-NOT-ITERABLE` | `TypeError: '(?P<type>\w+)' object is not iterable` |
| `PY-TYPE-MISSING-ARGS` | `TypeError: (?P<function>\w+)\(\) missing (?P<count>\d+) required positional argument` |
| `PY-TYPE-TOO-MANY-ARGS` | `TypeError: (?P<function>\w+)\(\) takes (?P<expected>\d+) positional argument` |
| `PY-TYPE-INT-STR` | `TypeError: '(?P<op>[<>]=?)' not supported between instances of 'str' and 'int'` (y al revés) |
| `PY-VALUE-LITERAL` | `ValueError: invalid literal for int\(\) with base 10: '(?P<text>[^']*)'` |
| `PY-ZERO-DIVISION` | `ZeroDivisionError: (division\|integer division or modulo) by zero` |
| `PY-INDEX-RANGE` | `IndexError: (?P<kind>\w+) index out of range` |
| `PY-KEY-ERROR` | `KeyError: (?P<key>.+)` |
| `PY-ATTRIBUTE` | `AttributeError: '(?P<type>\w+)' object has no attribute '(?P<name>\w+)'` |
| `PY-MODULE-NOT-FOUND` | `ModuleNotFoundError: No module named '(?P<module>[\w.]+)'` |
| `PY-IMPORT-NAME` | `ImportError: cannot import name '(?P<name>\w+)'` |
| `PY-RECURSION` | `RecursionError: maximum recursion depth exceeded` |
| `PY-FILE-NOT-FOUND` | `FileNotFoundError: \[Errno 2\] No such file or directory: '(?P<file>[^']*)'` |
| `PY-EOF-INPUT` | `EOFError: EOF when reading a line` |
| `PY-ASSERTION` | `AssertionError` |
| `W-PY-UNUSED-VAR` | ruff `F841`: `Local variable \`(?P<name>\w+)\` is assigned to but never used` |
| `W-PY-UNUSED-IMPORT` | ruff `F401`: `\`(?P<module>[\w.]+)\` imported but unused` |
| `W-PY-UNDEFINED` | ruff `F821`: `Undefined name \`(?P<name>\w+)\`` |

Los textos EN/ES siguen el estilo de los de Go (qué pasó, por qué, cómo se arregla; el mensaje
original nunca se traduce). Antes de redactarlos conviene pedir a los profesores diez o quince
errores reales de alumnos para ajustar los patrones.

### 4.10 `support_python.go` (en `wails/`, paquete `main`)

```go
func newPythonSupport(sink *bridge.WailsEventSink, store app.SettingsStore, texts *backendTexts) (app.LanguageSupport, func(ctx context.Context))
```

Recibe también el `Supervisor` compartido. Crea locator, runner, debugpy, pylsp, ruff (asignado a
`Formatter` y `Checker`), repl, packages y explainer, y devuelve el soporte más la función de
apagado. `main.go` lo añade al registro después de Go en lugar del `app.UnavailableSupport` de 2.1,
y borra el perfil provisional de Python de `support_unavailable.go`. `.golangci.yml` gana la regla
`python-adapter-stays-in-python` (copia de la de Go de M0 §4.6) y la exclusión correspondiente en
`adapters-are-independent`.

## 5. Frontend

| Pieza | Cambio |
|---|---|
| `package.json` | `@codemirror/lang-python` |
| `lib/editor/languageSupport.ts` | caso `python`: `python()`, 4 espacios |
| `lib/editor/templates.ts` | plantilla `print("Hola, Python")` y plantilla en blanco |
| `lib/bridge/mock*.ts` | escenarios de Python: ejecución correcta (`Hola, Python`), error (`NameError` explicado en EN/ES), depuración (frames y variables de ejemplo de una función `factorial(n)` en Python); `?lang=…&language=python` en la barra de desarrollo |
| `lib/stores/commands.ts` | nada específico: `missingToolIn` ya es genérico desde M0 |
| `lib/panels/ConsolePanel.svelte` | prompt `>>>` cuando el lenguaje activo es Python (hoy `>` para Go); `errors.consoleNoInput` |
| `lib/shell/PackagesDialog.svelte` | verbos `add`, `remove`, `list` (pip) con los textos de §6 |
| `lib/shell/FirstRunWizard.svelte` | paso nuevo "¿Qué lenguajes vas a usar?" que escribe `EnabledLanguages`; el paso de herramientas comprueba las de los lenguajes elegidos |
| `lib/shell/AboutDialog.svelte` | créditos: CPython, debugpy, python-lsp-server, ruff |

## 6. Textos nuevos (`tools/po2json/ux_copy.json`)

| Clave | EN | ES |
|---|---|---|
| `settings.toolPython` | Python | Python |
| `settings.modulePython` | {module} module | Módulo {module} |
| `errors.pythonNotFound` | Python isn't available. Install it or choose it in Settings. | Python no está disponible. Instálalo o elígelo en Ajustes. |
| `errors.debugpyMissing` | The debugger for Python (debugpy) isn't installed in this Python. Install it with the command below. | El depurador de Python (debugpy) no está instalado en este Python. Instálalo con el comando de abajo. |
| `errors.pylspMissing` | The code helper for Python (python-lsp-server) isn't installed. Suggestions and live problems are off. | El ayudante de código de Python (python-lsp-server) no está instalado. No habrá sugerencias ni problemas en vivo. |
| `errors.ruffMissing` | The formatter for Python (ruff) isn't installed, so the file was saved as it is. | El formateador de Python (ruff) no está instalado, así que el archivo se guardó tal cual. |
| `errors.consoleNoInput` | The console can't read the keyboard. Try it in a file with F5. | La consola no puede leer el teclado. Pruébalo en un archivo con F5. |
| `packages.add` | Install a package | Instalar un paquete |
| `packages.remove` | Uninstall | Desinstalar |
| `packages.list` | Show installed packages | Ver paquetes instalados |
| `packages.pipHint` | Packages are installed with pip into the Python VizcachaIDE uses. | Los paquetes se instalan con pip en el Python que usa VizcachaIDE. |
| `firstRun.languages` | Which languages will you use? | ¿Qué lenguajes vas a usar? |
| `run.pythonVenv` | Using the project's .venv | Usando el .venv del proyecto |

## 7. Empaquetado

- `packaging/versions.toml` gana:
  ```toml
  [python]
  version = "3.12.x"            # fijada
  release = "YYYYMMDD"          # tag de python-build-standalone
  url_template = "https://github.com/astral-sh/python-build-standalone/releases/download/{release}/cpython-{version}+{release}-{triple}-install_only.tar.gz"
  [python.sha256]
  windows-amd64 = "…"           # triple x86_64-pc-windows-msvc
  darwin-arm64 = "…"            # aarch64-apple-darwin
  darwin-amd64 = "…"            # x86_64-apple-darwin
  linux-amd64 = "…"             # x86_64-unknown-linux-gnu
  linux-arm64 = "…"             # aarch64-unknown-linux-gnu
  [python.wheels]
  debugpy = "1.8.x"
  python-lsp-server = "1.12.x"
  ruff = "0.x.y"
  ```
- `packaging/fetch_python.py`: descarga y verifica el archivo, lo extrae en
  `<stage>/toolchain/python/`, y descarga los wheels del target con
  `pip download --only-binary=:all: --platform <tag> --python-version 3.12 --dest cache/wheels`
  (python-lsp-server y sus dependencias son puros; debugpy y ruff tienen wheel por plataforma),
  los instala con `pip install --no-index --find-links cache/wheels --target
  <stage>/toolchain/python/<site-packages del target>` y copia las licencias a
  `toolchain/licenses/`. Escribe `VERSIONS.txt`.
- `wails/packaging/build_release.py`: variantes `full-python` y `full` (Go + Python). El NSIS y el
  `.dmg` no cambian: todo está bajo `toolchain/`.
- Prueba de humo: `smoke_test.py --python <bundled>` ejecuta
  `python -c "import debugpy, pylsp, ruff; print('ok')"`.
- Tamaño objetivo de `full-python`: 50 a 70 MB comprimido. Poder `toolchain/python/lib/python3.12/test`
  e `idlelib`/`tkinter` si no se usan.
- CI (`wails-release.yml`): matriz × variante; subir `SHA256SUMS`.

## 8. Tests

- Unitarios Go: `locator_test` (WindowsApps, `.venv`, versión mínima, con `exec` falso),
  `runner_test` (Configure, argumentos, `RunUntitled` limpia), `errors/parser_test` con
  `testdata/python_output/*.txt` (traceback simple, anidado en función, SyntaxError con caret,
  IndentationError, archivo inexistente, ruff JSON), `catalog_test` (cada id tiene fixture y
  placeholders válidos), `repl` con el proceso real si hay Python (salta si no), `debugpy/flavor_test`
  (launch JSON, filtros, variables ocultas), `pylsp/flavor_test` (configuración, root).
- Integración (se saltan sin Python o sin los módulos): ejecutar `hola.py` con `input()`; depurar
  `factorial.py` hasta un breakpoint, un paso y leer variables; abrir un archivo en pylsp y recibir
  un diagnóstico de nombre indefinido; `ruff format` de un archivo mal sangrado.
- Frontend: tests de los escenarios mock de Python, plantilla, prompt de consola, verbos de pip.
- QA manual (añadir a `docs/wails/QA_WAILS.md` una sección "Python"): la lista de §9.

## 9. Tracks y orden

```
P0 (orquestador) ──► P1 · P2 · P3 · P4 · P5 · P6 en paralelo ──► integración (support_python.go) ──► QA
```

| Track | Dueño de | Entrega | Hecho cuando |
|---|---|---|---|
| **P0 · Preparación** | `adapters/python/profile.go`, `locator.go`, `environment.go`, fixtures | perfil, localizador con tests, entorno, `testdata/python_output` grabado de un Python real | los demás tracks compilan contra el locator |
| **P1 · Ejecutor y paquetes** | `adapters/python/runner`, `packages` | §4.4 y §4.8 | ejecutar `hola.py` con `input()` por la PTY, stop, untitled, pip install por el diálogo |
| **P2 · Depurador** | `adapters/python/debugpy` | §4.5 | breakpoints, pasos, variables con "acaba de cambiar", pila, excepción no capturada; `input()` depurando si hay PTY |
| **P3 · Inteligencia y formato** | `adapters/python/pylsp`, `ruff` | §4.6, `Format`, `Check` | diagnósticos de pyflakes en vivo, completado, hover, definición, símbolos; formato al guardar |
| **P4 · Assistant y consola** | `adapters/python/errors`, `repl` | §4.9 y §4.7 | 25 ids con fixtures y textos EN/ES; REPL con sesión persistente y tiempo límite |
| **P5 · Frontend** | `frontend/src/lib/*` (§5) | editor, plantillas, mock, consola, paquetes, primer arranque | `npm run dev` muestra los tres escenarios de Python sin backend |
| **P6 · Empaquetado** | `packaging/fetch_python.py`, `versions.toml`, `wails/packaging/*` | §7 | `full-python` arranca sin Python instalado en una máquina limpia de Windows; macOS y Linux por CI |

Integración: P0 → P1 → P4 → P2 → P3 → P5 → P6. Un commit por track; el orquestador escribe
`support_python.go`, resuelve los CCR y fusiona los textos.

## 10. Criterios de salida (2.2.0)

Con la variante `full-python` en una máquina Windows sin Python, y con `lite` en una con Python
instalado, en EN y ES:

1. Nuevo archivo de Python → plantilla → F5 imprime `Hola, Python`.
2. Un programa con `input()` lee lo que se escribe en Output.
3. `print(x)` con `x` sin definir → el Assistant explica `PY-NAME-ERROR` con la línea subrayada.
4. Un error de sangría muestra `PY-INDENTATION` con la ubicación correcta.
5. Breakpoint en `factorial` → F6 para ahí; F7/F8/F9 funcionan; Variables marca los cambios; la vista
   de llamadas muestra `factorial(n=3)`; Hilos muestra uno.
6. Una división por cero sin capturar para el depurador en la línea con la descripción del error.
7. Escribir `pri` ofrece `print`; hover sobre `len` muestra su documentación; F12 en una función
   propia salta a su `def`.
8. Guardar con formato activado corrige la sangría; `ruff check` muestra una variable sin usar en
   Problemas después de ejecutar.
9. La consola evalúa `2 + 2`, recuerda `x = 5` entre entradas y explica un error.
10. Paquetes → instalar `cowsay` → un programa que lo importa funciona.
11. Un `.py` y un `.go` abiertos a la vez: cada uno ejecuta con su lenguaje; cambiar de pestaña
    cambia los paneles (Consola `>>>` frente a `>`).
12. Los tests Go y frontend pasan; `golangci-lint` limpio; el CI construye los tres sistemas.

## 11. Riesgos

| Riesgo | Mitigación |
|---|---|
| El alias de la Microsoft Store se detecta como Python | validación por versión y preferencia por `py -3` (§4.2) |
| debugpy y la versión de Python no casan | versiones fijadas juntas en `versions.toml`; test de humo importa los tres módulos |
| `runInTerminal` sin PTY | `console: internalConsole` y el aviso de "sin teclado al depurar" (comportamiento de Go hoy) |
| pylsp tarda en arrancar la primera vez (1 a 2 s) | el arranque perezoso y el `lsp:status: starting` ya existen; mensaje de estado en la barra |
| Eco duplicado de la entrada con PTY | `RunConfiguration.Echo` (M0) desactiva el eco manual del frontend |
| Programas que usan `tkinter`, `turtle` o `pygame` | fuera de alcance; `turtle` funciona si el Python empaquetado trae tk (decidir en P6 según tamaño) |
| Alumnos que instalan paquetes incompatibles en el intérprete empaquetado | el instalador deja `toolchain/python` íntegro en una reinstalación; documentar "Reparar" |
| Unicode en Windows (acentos en `print`) | `PYTHONUTF8=1` y `-X utf8`; la PTY ya entrega UTF-8 |

## 12. Prompt común para los coders

```text
You are a coder on VizcachaIDE (wails/: Go 1.25 + Wails v2 + Svelte 5 + CodeMirror 6), a beginner IDE,
bilingual EN/ES, that now has a multi-language core (docs/PLAN_NUCLEO_MULTILENGUAJE.md). You work in an
isolated git worktree on track <P?> of docs/PLAN_PYTHON.md. First read that plan (sections 0, 3, 4 and
your track in 9), docs/EXTENSION_MULTILENGUAJE.md section 4, wails/README.md, docs/wails/PLAN_WAILS.md
sections 2 and 4, and the Go adapter in wails/internal/adapters/golang as the reference implementation.
- Edit ONLY your track's folders. The v3 contract (internal/domain, internal/app/ports.go,
  internal/bridge/events.go, frontend/src/lib/{events,domain}.ts, bridge/types.ts) and internal/protocol
  are read-only: request changes as a "Contract change request" in your report.
- Use internal/protocol (process, dap, lsp, errorcatalog, toollocator): never copy its code.
- Only free tools: CPython, debugpy, python-lsp-server, ruff, pip. Pin nothing in code; versions live
  in packaging/versions.toml.
- Library-first. Files under 200 lines, functions under 50, early return, domain names. Visible texts
  only through i18n (tools/po2json/ux_copy.json, then go run ./tools/po2json). Error messages from Python
  are never translated.
- Tests that need Python or its modules must skip cleanly when they are missing.
- Verify: go vet ./... && go test ./... && golangci-lint run; cd frontend && npm run check && npm run test;
  wails build when your track affects it.
- When done: ONE local commit in English ending with "Co-Authored-By: Claude <noreply@anthropic.com>",
  no push. Report: commit hash, files, how to test, limitations, CCRs and new i18n keys.
```
