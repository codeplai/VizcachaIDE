"""Run menu: Run (F5), Stop (Shift+F5), Build (Ctrl+B); console stdin; Environment options."""

from PyQt5.QtWidgets import QAction, QMessageBox

from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.ui.features.run.environment_page import EnvironmentPage
from vizcacha.ui.workbench import Workbench


class RunFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.toolchain = workbench.services.toolchain
        self.console = workbench.console
        self._interactive = False

    def register(self) -> None:
        self.run_action = self._action(_("▶ Run"), "F5", self.run)
        self.stop_action = self._action(_("⬛ Stop"), "Shift+F5", self.toolchain.stop)
        self.build_action = self._action(_("🔨 Build"), "Ctrl+B", self.build, separator=True)
        self.workbench.add_toolbar_separator()
        self.stop_action.setEnabled(False)
        self.workbench.add_settings_page(EnvironmentPage)
        self.toolchain.output_received.connect(self._on_stdout)
        self.toolchain.error_received.connect(self._on_stderr)
        self.toolchain.execution_finished.connect(self._on_finished)
        self.console.input_submitted.connect(self.toolchain.write_input)

    def run(self) -> None:
        config = self._configuration(_("Please save your file before running."))
        if config is None:
            return
        self._interactive = True
        self.console.set_waiting_for_input(True)
        self._set_running(True)
        self.workbench.events.program_started.emit(config)
        self.toolchain.run(config)

    def build(self) -> None:
        config = self._configuration(_("Please save your file before building."))
        if config is None:
            return
        self._interactive = False
        self._set_running(True)
        self.toolchain.build(config)

    def _configuration(self, unsaved_message: str) -> RunConfiguration | None:
        path = self.workbench.editor.current_file_path()
        if path is None:
            QMessageBox.warning(self.workbench.window, _("No File"), unsaved_message)
            return None
        self.console.clear()
        return RunConfiguration.for_file(path)

    def _on_stdout(self, text: str) -> None:
        self.console.append_output(text)
        self.workbench.events.process_output.emit(text, "stdout")

    def _on_stderr(self, text: str) -> None:
        self.console.append_error(text)
        self.workbench.events.process_output.emit(text, "stderr")

    def _on_finished(self, exit_code: int) -> None:
        self.console.set_waiting_for_input(False)
        self.console.disable_input()
        self._set_running(False)
        if exit_code == 0:
            message = (
                _("[Process finished successfully]")
                if self._interactive
                else _("[Build successful]")
            )
            self.console.append_success("\n" + message)
        else:
            self.console.append_error(
                "\n" + _("[Process exited with code {code}]").format(code=exit_code)
            )
        if self._interactive:
            self.workbench.events.program_finished.emit(exit_code)

    def _set_running(self, running: bool) -> None:
        self.run_action.setEnabled(not running)
        self.build_action.setEnabled(not running)
        self.stop_action.setEnabled(running)

    def _action(self, text: str, shortcut: str, slot, separator: bool = False) -> QAction:
        action = QAction(text, self.workbench.window)
        action.setShortcut(shortcut)
        action.triggered.connect(lambda _checked=False: slot())
        return self.workbench.add_action("run", action, toolbar=True, separator=separator)


def register(workbench: Workbench) -> None:
    RunFeature(workbench).register()
