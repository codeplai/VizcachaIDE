"""Regenerate the README / release screenshots in docs/images/.

Usage (from the repository root, with the development environment active)::

    python docs/release/make_screenshots.py            # English and Spanish
    python docs/release/make_screenshots.py --lang es  # only one language

Requirements: ``go`` and ``dlv`` on PATH (the debugger shot uses a real Delve
session); ``gopls`` is optional (without it there is no Outline panel or live
underlining). Qt runs with ``QT_QPA_PLATFORM=offscreen``, so no window appears.

Each language runs in its own Python process, because menu titles are
translated when the workbench is built. The app is built with
``build_workbench`` on an in-memory settings repository: your real settings
(last file, last folder, theme...) are neither read nor modified. The examples
are copied to a temporary folder, which is opened as the project so that the
Files panel is visible.

Output (``<lang>`` = ``en`` | ``es``), as palette PNGs (Qt quantization, no Pillow)::

    docs/images/editor-run.<lang>.png   simple_loop.go and its output
    docs/images/assistant.<lang>.png    the Assistant explaining E-UNUSED-VAR
    docs/images/debugger.<lang>.png     stopped at a breakpoint in functions.go

``--assistant-draft`` (default: on) merges ``docs/i18n/assistant.es.po`` into
the Spanish catalog for the run, so that the Spanish Assistant shot shows the
reviewed draft translations even before they are merged into
``vizcacha/i18n/locale``. Pass ``--no-assistant-draft`` to capture exactly what
the shipped catalogs contain.
"""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
import time
from pathlib import Path

os.environ.setdefault("QT_QPA_PLATFORM", "offscreen")
# The offscreen platform only looks for fonts in QT_QPA_FONTDIR (default: Qt's own
# lib/fonts, which PyQt5 wheels do not ship), so without this every text is blank.
SYSTEM_FONT_DIRS = ("C:/Windows/Fonts", "/System/Library/Fonts", "/usr/share/fonts")
for _font_dir in SYSTEM_FONT_DIRS:
    if "QT_QPA_FONTDIR" not in os.environ and Path(_font_dir).is_dir():
        os.environ["QT_QPA_FONTDIR"] = _font_dir

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT))

OUTPUT_DIR = ROOT / "docs" / "images"
WINDOW_SIZE = (1280, 800)
BREAKPOINT_LINE = 13  # "return result" inside multiply(5, 7)
BUILD_TIMEOUT_S = 300  # first debug build compiles the standard library
STEP_TIMEOUT_S = 60
LANGUAGES = ("en", "es")
DEBUGGER_PANELS = ("variables", "callstack")
ASSISTANT_WIDTH = 430
EXAMPLES = {  # folder in the temporary project -> file in the repository
    "loop": "examples/simple_loop.go",
    "functions": "examples/functions.go",
    "errors": "examples/errors/E-UNUSED-VAR/E-UNUSED-VAR.go",
}


# --------------------------------------------------------------------- helpers
def wait_until(app, predicate, timeout_s: float, what: str) -> None:
    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        app.processEvents()
        if predicate():
            return
        time.sleep(0.05)
    raise TimeoutError(f"timed out waiting for {what}")


def settle(app, seconds: float) -> None:
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        app.processEvents()
        time.sleep(0.05)


def trigger_shortcut(window, shortcut: str) -> None:
    from PyQt5.QtGui import QKeySequence
    from PyQt5.QtWidgets import QAction

    wanted = QKeySequence(shortcut)
    for action in window.findChildren(QAction):
        if action.shortcut() == wanted and action.isEnabled():
            action.trigger()
            return
    raise LookupError(f"no enabled action with shortcut {shortcut}")


def save_png(window, name: str, language: str) -> Path:
    """Grab the window and store it as an 8-bit palette PNG (much smaller, still sharp)."""
    from PyQt5.QtCore import Qt
    from PyQt5.QtGui import QImage

    image = window.grab().toImage().convertToFormat(QImage.Format_RGB32)
    quantized = image.convertToFormat(QImage.Format_Indexed8, Qt.AvoidDither)
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    path = OUTPUT_DIR / f"{name}.{language}.png"
    if not quantized.save(str(path), "PNG", 9):
        raise OSError(f"could not write {path}")
    print(f"  {path.relative_to(ROOT)}  {path.stat().st_size // 1024} KB")
    return path


def prepare_language(language: str, use_draft: bool, scratch: Path) -> None:
    from vizcacha.i18n import install_language
    from vizcacha.i18n.translator import LOCALE_DIR

    if language != "es" or not use_draft:
        install_language(language)
        return
    from babel.messages.mofile import write_mo
    from babel.messages.pofile import read_po

    shipped_po = LOCALE_DIR / "es" / "LC_MESSAGES" / "vizcacha.po"
    draft_po = ROOT / "docs" / "i18n" / "assistant.es.po"
    with shipped_po.open("rb") as handle:
        catalog = read_po(handle)
    with draft_po.open("rb") as handle:
        for message in read_po(handle):
            shipped = catalog.get(message.id) if message.id else None
            if message.string and (shipped is None or not shipped.string):
                catalog[message.id] = message
    target = scratch / "locale" / "es" / "LC_MESSAGES"
    target.mkdir(parents=True)
    with (target / "vizcacha.mo").open("wb") as handle:
        write_mo(handle, catalog)
    install_language("es", scratch / "locale")


def copy_examples(scratch: Path) -> Path:
    """One folder per program: Go (and gopls) treat a folder as one package with one main."""
    project = scratch / "examples"
    for folder, source in EXAMPLES.items():
        (project / folder).mkdir(parents=True)
        shutil.copy(ROOT / source, project / folder / Path(source).name)
    return project


def use_readable_fonts(app) -> None:
    """Offscreen picks the first font file it finds; use the platform's UI font instead."""
    from PyQt5.QtGui import QFont

    for family, emoji in (("Segoe UI", "Segoe UI Emoji"), ("Helvetica Neue", "Apple Color Emoji")):
        if family in _font_families():
            QFont.insertSubstitution(family, emoji)
            app.setFont(QFont(family, 9))
            return


def _font_families() -> list[str]:
    from PyQt5.QtGui import QFontDatabase

    return QFontDatabase().families()


# ----------------------------------------------------------------------- shots
def shot_editor_run(app, workbench, project: Path, language: str) -> None:
    workbench.editor.open_file(project / "loop" / "simple_loop.go")
    settle(app, 2)
    trigger_shortcut(workbench.window, "F5")
    wait_until(app, lambda: not workbench.services.toolchain.is_running(), 120, "go run")
    settle(app, 2)
    save_png(workbench.window, "editor-run", language)


def shot_assistant(app, workbench, project: Path, language: str) -> None:
    from PyQt5.QtCore import Qt
    from PyQt5.QtWidgets import QDockWidget

    workbench.editor.open_file(project / "errors" / "E-UNUSED-VAR.go")
    settle(app, 1)
    trigger_shortcut(workbench.window, "F5")
    wait_until(app, lambda: not workbench.services.toolchain.is_running(), 120, "go run")
    window = workbench.window
    assistant = window.findChild(QDockWidget, "assistant")
    wait_until(app, lambda: assistant.isVisible(), 30, "the Assistant panel")
    # Staging: close the (empty) debugger panels, as a user can from the View menu,
    # so the Assistant gets the whole right column.
    debugger_docks = [window.findChild(QDockWidget, name) for name in DEBUGGER_PANELS]
    for dock in debugger_docks:
        dock.hide()
    window.resizeDocks([assistant], [ASSISTANT_WIDTH], Qt.Horizontal)
    settle(app, 4)  # let gopls underline the variable as well
    save_png(window, "assistant", language)
    assistant.hide()
    for dock in debugger_docks:
        dock.show()


def shot_debugger(app, workbench, project: Path, language: str) -> None:
    source = project / "functions" / "functions.go"
    workbench.editor.open_file(source)
    editor = workbench.editor.current_editor()
    editor.toggle_breakpoint_at_line(BREAKPOINT_LINE)
    debugger = workbench.services.debugger
    states, exits = [], []
    debugger.stopped.connect(states.append)
    debugger.terminated.connect(exits.append)
    trigger_shortcut(workbench.window, "F6")
    wait_until(app, lambda: bool(states or exits), BUILD_TIMEOUT_S, "the breakpoint")
    if exits:
        raise RuntimeError("the program ended before reaching the breakpoint")
    settle(app, 2)
    save_png(workbench.window, "debugger", language)
    debugger.stop()
    wait_until(app, lambda: bool(exits), STEP_TIMEOUT_S, "dlv to stop")
    settle(app, 1)


# ------------------------------------------------------------------------ main
def capture(language: str, use_draft: bool) -> None:
    from PyQt5.QtWidgets import QApplication

    from vizcacha.application.settings_keys import SettingsKeys
    from vizcacha.infrastructure.settings import InMemorySettingsRepository

    app = QApplication([sys.argv[0]])
    use_readable_fonts(app)
    with tempfile.TemporaryDirectory(prefix="vizcacha-shots-") as scratch_name:
        scratch = Path(scratch_name)
        prepare_language(language, use_draft, scratch)
        from vizcacha.ui.app import build_services, build_workbench

        project = copy_examples(scratch)
        settings = InMemorySettingsRepository(
            {SettingsKeys.LAST_FOLDER: str(project), SettingsKeys.FORMAT_ON_SAVE: False}
        )
        workbench = build_workbench(build_services(settings))
        window = workbench.window
        window.resize(*WINDOW_SIZE)
        window.show()
        settle(app, 2)
        print(f"[{language}]")
        try:
            shot_editor_run(app, workbench, project, language)
            shot_assistant(app, workbench, project, language)
            shot_debugger(app, workbench, project, language)
        finally:
            for editor in workbench.editor.editors():
                editor.document().setModified(False)
            workbench.services.language_server.shutdown()
            settle(app, 1)
            window.close()


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--lang", choices=LANGUAGES, help="capture a single language")
    parser.add_argument(
        "--assistant-draft",
        action=argparse.BooleanOptionalAction,
        default=True,
        help="merge docs/i18n/assistant.es.po into the Spanish catalog (default: on)",
    )
    args = parser.parse_args()
    if args.lang:
        capture(args.lang, args.assistant_draft)
        return 0
    draft_flag = "--assistant-draft" if args.assistant_draft else "--no-assistant-draft"
    for language in LANGUAGES:
        command = [sys.executable, __file__, "--lang", language, draft_flag]
        if subprocess.run(command, check=False).returncode != 0:
            return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
