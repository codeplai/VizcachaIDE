# Contributing to VizcachaIDE

Thank you for helping! VizcachaIDE is a Go IDE for beginners, so the most valuable
contributions are often small: a clearer error explanation, a better Spanish translation, a bug
report with the steps to reproduce it. *Resumen en español al final:
[Resumen en español](#resumen-en-español).*

## Getting started

```bash
git clone https://github.com/codeplai/VizcachaIDE.git
cd VizcachaIDE
python -m venv .venv
.venv/Scripts/activate              # Windows  (Linux/macOS: source .venv/bin/activate)
pip install -r requirements-dev.txt
python -m vizcacha                  # run the IDE from source
```

You need Python 3.10+, and Go on `PATH`. Install `dlv` and `gopls` too if you work on the
debugger or on code intelligence.

Before opening a pull request, run:

```bash
python -m pytest                    # all tests
ruff check .                        # lint
ruff format vizcacha tests          # formatting (CI runs `ruff format --check vizcacha tests`)
lint-imports                        # architecture contracts
```

## Architecture

The code follows Clean Architecture. Each layer may only import the layers below it:

```
vizcacha.ui               PyQt5 widgets and features            (may import everything below)
vizcacha.infrastructure   adapters: go toolchain, Delve DAP, gopls LSP, error catalog, settings
vizcacha.application      ports (typing.Protocol) and use cases
vizcacha.domain           pure Python model: debugging, diagnostics, explanations, project
vizcacha.i18n             translation functions (a leaf: imports nothing from vizcacha nor Qt)
```

These rules are enforced by [import-linter](https://import-linter.readthedocs.io/) through the
contracts in `pyproject.toml`:

1. **Layers:** `ui` → `infrastructure` → `application` → `domain`, never upwards.
2. **No Qt in the core:** `vizcacha.domain` and `vizcacha.application` must not import PyQt5.
3. **i18n is a leaf:** `vizcacha.i18n` imports no other vizcacha package and no Qt.
4. **Features are independent:** the packages in `vizcacha.ui.features.*` never import each
   other. They talk through `workbench.events` (Qt signals carrying plain or domain objects).

Other conventions:

- The ports in `vizcacha/application/ports.py` and the domain types are **contracts**. Change
  them in a dedicated pull request, together with every implementation and `tests/test_contracts.py`.
- `vizcacha/ui/app.py` is the composition root: the only place that creates concrete adapters
  (`build_services`) and lists the features (`FEATURES`).
- Small files and functions, early returns, domain names (no `utils`, `helpers` or `common`
  modules), and typed exceptions. Prefer a maintained library to custom code.
- Out of scope: MicroPython, TinyGo and microcontrollers.

More background, in Spanish: [docs/ARQUITECTURA.md](docs/ARQUITECTURA.md) and
[docs/PLAN_DESARROLLO.md](docs/PLAN_DESARROLLO.md).

## Adding a feature

A feature is a package in `vizcacha/ui/features/<name>/` whose `__init__.py` exposes
`register(workbench)`. The `Workbench` (`vizcacha/ui/workbench.py`) is the API it uses to plug
itself into the window:

| Workbench API | Use |
|---|---|
| `add_action(menu_id, action, toolbar=False, separator=False)` | Add a `QAction` to `file`, `edit`, `view`, `run`, `debug`, `tools` or `help` (and optionally the toolbar) |
| `add_panel(panel_id, title, widget, area="right")` | Add a dock panel (`right`, `bottom`, `left`); it is listed in the View menu |
| `add_settings_page(factory)` | A page in *Tools → Options…*: a widget with `title`, `load(settings)` and `save(settings)` |
| `add_status_widget(widget)`, `show_status_message(text)` | Status bar |
| `add_close_guard(guard)` | `guard()` returns `False` to cancel closing the window |
| `events` | `navigate_to`, `process_output`, `program_started`, `program_finished`, `diagnostics_changed`, `settings_changed` |
| `services` | `settings`, `environment`, `toolchain`, `debugger`, `language_server`, `explainer` |
| `editor`, `console`, `toolbar`, `window` | The shared widgets |

Minimal example:

```python
# vizcacha/ui/features/greeting/__init__.py
from PyQt5.QtWidgets import QAction

from vizcacha.i18n import _
from vizcacha.ui.workbench import Workbench


def register(workbench: Workbench) -> None:
    action = QAction(_("Say &Hello"), workbench.window)
    action.triggered.connect(
        lambda _checked=False: workbench.console.append_output(_("Hello, Go!") + "\n")
    )
    workbench.add_action("tools", action)
```

Then add `register_greeting` to `FEATURES` in `vizcacha/ui/app.py` (the order of that tuple is
the order of the menu entries and toolbar buttons) and write tests in `tests/<area>/`, using the
`workbench` fixture from `tests/conftest.py`.

If a feature needs a new external tool or service, define a port (`typing.Protocol`) in
`application/ports.py`, implement it in `infrastructure/`, and add it to `Services` and
`build_services`.

## User-visible text and translations

**Rule:** every text the user can see is written in **English** and wrapped in `_()`
(or `ngettext()` for plurals), imported from `vizcacha.i18n`:

```python
from vizcacha.i18n import N_, _, ngettext

label = _("Program arguments")
message = _("[Process exited with code {code}]").format(code=exit_code)   # named placeholders
summary = ngettext("Go reported {count} problem:", "Go reported {count} problems:", count)
TITLE = N_("&File")   # module-level constant: mark it with N_() and call _(TITLE) when used
```

- Never build sentences by concatenating translated pieces; use one string with `{named}`
  placeholders so translators can reorder words.
- Use `N_()` for strings defined at import time (the language is chosen after imports) and
  translate them with `_()` where they are displayed.
- Go's own error messages are **never** translated: the Assistant shows them as they are.

Catalogs live in `vizcacha/i18n/locale/<lang>/LC_MESSAGES/vizcacha.po` and are maintained with
[Babel](https://babel.pocoo.org/):

```bash
# 1. extract the strings from the code into the template
pybabel extract -F babel.cfg -o vizcacha/i18n/locale/vizcacha.pot --project=VizcachaIDE .
# 2. merge the new strings into each language
pybabel update -i vizcacha/i18n/locale/vizcacha.pot -d vizcacha/i18n/locale -D vizcacha
# 3. translate the empty msgstr entries in vizcacha/i18n/locale/es/LC_MESSAGES/vizcacha.po
# 4. compile the .mo files that the app loads
pybabel compile -d vizcacha/i18n/locale -D vizcacha
```

Spanish style: address the user as *tú*, keep it short and friendly, and keep the usual Go
terms (*slice*, *map*, *goroutine*, *panic*) in English. A breakpoint is a
*punto de interrupción*.

## Adding an error to the Assistant's catalog

The catalog lives in `vizcacha/infrastructure/error_catalog/`. Each entry maps one or more
regular expressions over Go's message to a **stable id** and three English texts:

1. **Pick an id** with a prefix: `E-` compiler error, `P-` runtime panic, `V-` `go vet`
   warning (for example `E-INVALID-OP`). Ids are permanent: never rename one.
2. **Add the entry** with `catalog_entry(...)` in the right module: `compile_errors.py`,
   `syntax_type_errors.py`, `runtime_panics.py` or `vet_warnings.py`.
   ```python
   catalog_entry(
       "E-UNUSED-VAR",
       (r"declared and not used: (?P<name>\w+)", r"^(?P<name>\w+) declared (?:and|but) not used"),
       N_('Variable "{name}" is never used'),          # title
       N_('You created the variable "{name}" but never read it. ...'),   # body
       N_('Use "{name}" somewhere (for example, print it), delete it, ...'),   # fix_hint
   ),
   ```
   - Every `{placeholder}` must be a **named group of every pattern** (a test checks it).
   - The texts are passed to `str.format`, so literal braces are doubled: `{{ }}`.
   - **Order matters:** the first matching entry wins, so put specific entries before
     general ones (as `E-UNEXPORTED` before `E-UNDEFINED`).
   - Write for a beginner: what happened, why Go complains, and what to try.
3. **Add an example program** `examples/errors/<ID>/<ID>.go` that produces exactly that error,
   with a first comment line `// <ID>: short description`.
4. **Record Go's output** in `tests/assistant/fixtures/go_output/<ID>.txt` (the output of
   `go run` or `go vet` on the example, run from the example's folder; write any absolute path
   of the working directory as `$WORKDIR`).
   `tests/assistant/test_error_catalog.py` checks that every id has an example and a fixture
   that maps to it.
5. **Translate** the three texts into Spanish (see the Babel workflow above).
6. Run `python -m pytest tests/assistant`.

## Tests

- `pytest` + `pytest-qt`; Qt runs with `QT_QPA_PLATFORM=offscreen` (set in `tests/conftest.py`).
- Every test starts in English; switch with `install_language("es")` when you test Spanish.
- Tests that need real tools are marked `@pytest.mark.requires_go`, `requires_dlv` or
  `requires_gopls` and are skipped when the tool is not on `PATH`.
- Debugger and gopls protocol code is tested with recorded transcripts
  (`tests/debugger/transcripts/`), so most tests run without the tools.
- Use the `settings` fixture (in-memory) instead of real `QSettings`.

## Screenshots

`python docs/release/make_screenshots.py` rebuilds `docs/images/*.png` in English and Spanish
with a real Go/Delve session (see the script's docstring). Re-run it when the UI changes
visibly, and check the images before committing them.

## Pull requests

- One topic per pull request, with tests.
- Describe how to try the change by hand in two or three steps.
- Update `CHANGELOG.md` (both the English and the Spanish sections) under *Unreleased*.
- By contributing you agree that your contribution is licensed under the MIT License of the
  project.

---

## Resumen en español

- **Preparar el entorno:** Python 3.10+, `pip install -r requirements-dev.txt` y Go en el
  `PATH` (más `dlv` y `gopls` si tocas el depurador o la inteligencia de código). Antes de un
  pull request: `python -m pytest`, `ruff check .`, `ruff format vizcacha tests` y
  `lint-imports`.
- **Arquitectura en capas:** `ui` → `infrastructure` → `application` → `domain`, nunca hacia
  arriba. `domain` y `application` no importan Qt; `i18n` es una hoja; las features de
  `ui/features/*` no se importan entre sí y se comunican con `workbench.events`. import-linter
  lo comprueba.
- **Añadir una función:** crea `vizcacha/ui/features/<nombre>/` con `register(workbench)`, usa
  `add_action`, `add_panel`, `add_settings_page` y los eventos del Workbench, y añádela a
  `FEATURES` en `vizcacha/ui/app.py`.
- **Textos:** todo texto visible se escribe en inglés dentro de `_()` (o `ngettext()`), con
  marcadores con nombre (`{code}`); `N_()` para constantes de módulo. Los mensajes originales de
  Go nunca se traducen. Flujo con Babel: `pybabel extract` → `pybabel update` → traducir el
  `.po` → `pybabel compile`. Estilo: tuteo, frases cortas, y *slice*, *map*, *goroutine* y
  *panic* en inglés.
- **Nuevo error en el Asistente:** id estable (`E-`, `P-` o `V-`), entrada `catalog_entry` en el
  módulo de su categoría (los marcadores deben ser grupos con nombre de todos los patrones; el
  orden importa), ejemplo en `examples/errors/<ID>/<ID>.go`, salida grabada en
  `tests/assistant/fixtures/go_output/<ID>.txt`, traducción al español y
  `python -m pytest tests/assistant`.
- **Tests:** pytest + pytest-qt en modo offscreen; marcadores `requires_go`, `requires_dlv` y
  `requires_gopls`; fixture `settings` en memoria.
- **Pull requests:** un tema por PR, con tests, pasos para probarlo a mano y una entrada en
  `CHANGELOG.md` (inglés y español). Las contribuciones se publican bajo la licencia MIT.
