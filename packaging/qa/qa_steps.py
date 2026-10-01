"""Steps of the headless functional QA (see functional_qa.py). Each step records a QaResult."""

from __future__ import annotations

import os
import shutil
import subprocess
import tempfile
import time
from collections.abc import Callable
from dataclasses import dataclass
from pathlib import Path

from PyQt5.QtWidgets import QApplication

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import _, active_language
from vizcacha.infrastructure.delve_dap import DelveDapDebugger
from vizcacha.infrastructure.error_catalog import GoErrorExplainer
from vizcacha.infrastructure.go_toolchain import GoEnvironment, GoToolchain
from vizcacha.infrastructure.gopls_lsp import GoplsLanguageServer
from vizcacha.infrastructure.settings import InMemorySettingsRepository
from vizcacha.ui.app import build_workbench
from vizcacha.ui.features.assistant.panel import AssistantPanel
from vizcacha.ui.features.debugger.variables_view import VariablesView
from vizcacha.ui.services import Services

REPO_ROOT = Path(__file__).resolve().parents[2]
EXAMPLES = REPO_ROOT / "examples"
RUN_TIMEOUT_S = 180
DEBUG_TIMEOUT_S = 420  # first debug build compiles the std library with -N -l
MESSY = 'package main\nimport "fmt"\nfunc main(){\nx:=1\nfmt.Println( x )\n}\n'
TIDY = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tx := 1\n\tfmt.Println(x)\n}\n'


@dataclass
class QaResult:
    name: str
    passed: bool
    seconds: float
    detail: str


class QaSession:
    def __init__(self, app: QApplication, work: Path, env: dict, bundle: Path | None) -> None:
        self.app, self.work, self.bundle = app, work, bundle
        self.results: list[QaResult] = []
        settings = InMemorySettingsRepository({SettingsKeys.LANGUAGE: active_language()})
        environment = GoEnvironment(settings, base_environment=env, app_directory=bundle)
        self.services = Services(
            settings=settings,
            environment=environment,
            toolchain=GoToolchain(environment),
            debugger=DelveDapDebugger(environment),
            language_server=GoplsLanguageServer(environment),
            explainer=GoErrorExplainer(),
        )
        started = time.monotonic()
        self.workbench = build_workbench(self.services)
        self.workbench.window.show()
        self.startup_seconds = time.monotonic() - started
        self.finished: list[int] = []
        self.workbench.events.program_finished.connect(self.finished.append)

    # --- helpers ------------------------------------------------------------------
    def wait_until(self, condition: Callable[[], bool], timeout_s: float) -> bool:
        deadline = time.monotonic() + timeout_s
        while time.monotonic() < deadline:
            self.app.processEvents()
            if condition():
                return True
            time.sleep(0.02)
        return False

    def record(self, name: str, started: float, passed: bool, detail: str) -> None:
        result = QaResult(name, passed, time.monotonic() - started, detail)
        self.results.append(result)
        print(f"[{'PASS' if passed else 'FAIL'}] {name} ({result.seconds:.1f} s) {detail}")

    def copy_example(self, source: Path) -> Path:
        folder = self.work / source.stem
        folder.mkdir(parents=True, exist_ok=True)
        return Path(shutil.copy(source, folder / source.name))

    def action(self, menu_id: str, index: int):
        return [a for a in self.workbench.menu(menu_id).actions() if not a.isSeparator()][index]

    def run_file(self, path: Path) -> tuple[bool, str]:
        self.workbench.editor.open_file(path)
        self.finished.clear()
        self.action("run", 0).trigger()  # ▶ Run
        done = self.wait_until(lambda: bool(self.finished), RUN_TIMEOUT_S)
        self.wait_until(lambda: False, 0.3)  # let the Assistant update its panel
        return done, self.workbench.console.toPlainText()

    def tool_origins(self) -> dict[str, str]:
        origins = self.services.environment.tool_origins()
        return {tool: f"{loc.origin.value}:{loc.path}" for tool, loc in origins.items()}

    # --- steps --------------------------------------------------------------------
    def check_bundled_toolchain(self) -> None:
        started = time.monotonic()
        environment = self.services.environment
        go = environment.go_executable()
        version = subprocess.run(
            [go, "version"], capture_output=True, text=True, env=environment.variables()
        ).stdout.strip()
        origins = environment.tool_origins()
        if self.bundle is None:
            self.record("toolchain", started, bool(version), f"{version} ({go})")
            return
        bundled = all(loc.origin.value == "bundled" for loc in origins.values())
        inside = all(Path(loc.path).is_relative_to(self.bundle) for loc in origins.values())
        detail = f"{version}; all tools bundled={bundled}, inside bundle={inside}"
        self.record("bundled toolchain", started, bundled and inside and bool(version), detail)

    def check_ui_language(self) -> None:
        """Menu and Assistant texts are in the active language (es must differ from en)."""
        started = time.monotonic()
        run_text = self.action("run", 0).text()
        empty_panel = self.workbench.window.findChild(AssistantPanel).summary.text()
        english = {"▶ Run", "Nothing to explain. When Go reports an error, I will explain it here."}
        texts = {run_text, empty_panel}
        ok = texts == english if active_language() == "en" else not (texts & english)
        self.record(f"ui language {active_language()}", started, ok, f"{sorted(texts)}")

    def run_example(self, name: str, expected: str) -> None:
        started = time.monotonic()
        done, console = self.run_file(self.copy_example(EXAMPLES / Path(name).stem / name))
        ok = done and self.finished == [0] and expected in console
        self.record(f"run {name}", started, ok, f"exit={self.finished} expected={expected!r}")

    def explain_error(self, error_id: str) -> None:
        started = time.monotonic()
        source = EXAMPLES / "errors" / error_id / f"{error_id}.go"
        done, _console = self.run_file(self.copy_example(source))
        panel = self.workbench.window.findChild(AssistantPanel)
        cards = panel.cards if panel else []
        codes = [card.diagnostic.code for card in cards]
        title = cards[0].findChild(object, "assistant_title").text() if cards else ""
        known = bool(cards) and cards[0].explanation is not None
        ok = done and codes[:1] == [error_id] and known
        self.record(f"assistant {error_id}", started, ok, f"ids={codes} title={title!r}")
        if known and active_language() != "en":
            message_id = cards[0].explanation.title
            translated = _(message_id) != message_id
            self.record(f"assistant {error_id} translated", started, translated, repr(message_id))

    def debug_functions(self, breakpoint_line: int, variable: str, value: str) -> None:
        started = time.monotonic()
        path = self.copy_example(EXAMPLES / "functions" / "functions.go")
        self.workbench.editor.open_file(path)
        self.workbench.editor.current_editor().toggle_breakpoint_at_line(breakpoint_line)
        terminated: list[int] = []
        self.services.debugger.terminated.connect(terminated.append)
        view = self.workbench.window.findChild(VariablesView)
        self.action("debug", 0).trigger()  # 🐛 Debug
        stopped = self.wait_until(
            lambda: view.topLevelItemCount() > 0 or terminated, DEBUG_TIMEOUT_S
        )
        items = [view.topLevelItem(i) for i in range(view.topLevelItemCount())]
        variables = {item.text(0): item.text(2) for item in items}
        line = self.workbench.editor.current_editor().current_line
        self.action("debug", 1).trigger()  # ⏩ Continue
        ended = self.wait_until(lambda: bool(terminated), RUN_TIMEOUT_S)
        output_ok = "5 * 7 = 35" in self.workbench.console.toPlainText()
        ok = stopped and variables.get(variable) == value and line == breakpoint_line
        ok = ok and ended and terminated[:1] == [0] and output_ok
        found = variables.get(variable)
        detail = f"line={line} {variable}={found} exit={terminated} output={output_ok}"
        self.wait_until(lambda: False, 2.0)  # dlv removes its binary when it exits
        leftover = list(Path(tempfile.gettempdir()).glob(f"vizcacha_debug_bin_{os.getpid()}*"))
        detail += f" leftover_debug_binary={[p.name for p in leftover]}"
        if not ok:
            detail += "\n--- console ---\n" + self.workbench.console.toPlainText()
        self.record("debug functions.go", started, ok, detail)

    def format_on_save(self) -> None:
        started = time.monotonic()
        target = self.work / "format" / "main.go"
        target.parent.mkdir(parents=True, exist_ok=True)
        editor = self.workbench.editor.current_editor()
        editor.setPlainText(MESSY)
        saved = self.workbench.editor.save_to_file(editor, target)
        on_disk = target.read_text(encoding="utf-8") if target.exists() else ""
        ok = saved and on_disk == TIDY and editor.toPlainText() == TIDY
        self.record("gofmt on save", started, ok, f"saved={saved} formatted={on_disk == TIDY}")

    # --- end ----------------------------------------------------------------------
    def close(self) -> None:
        self.services.debugger.stop()
        if self.services.language_server is not None:
            self.services.language_server.shutdown()
        for editor in self.workbench.editor.editors():
            editor.document().setModified(False)
        self.wait_until(lambda: False, 1.5)  # let dlv/gopls exit before removing the folder

    def all_passed(self) -> bool:
        return bool(self.results) and all(result.passed for result in self.results)

    def report(self, seconds: float) -> str:
        passed = sum(result.passed for result in self.results)
        return f"[qa] {passed}/{len(self.results)} checks passed in {seconds:.1f} s"
