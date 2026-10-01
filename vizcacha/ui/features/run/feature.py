"""Run menu: Run (F5), Stop (Shift+F5), Build (Ctrl+B); program arguments; Environment options.

The toolchain is shared with other features (e.g. ``go mod`` commands from the
project feature), so this feature only reacts to output of processes it started.
"""

from PyQt5.QtWidgets import QAction, QMessageBox

from vizcacha.application.run_program import ProgramArgumentsError, configuration_for_file
from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.ui.features.run.arguments_field import ProgramArgumentsField
from vizcacha.ui.features.run.environment_page import EnvironmentPage
from vizcacha.ui.workbench import Workbench


class RunFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.toolchain = workbench.services.toolchain
        self.console = workbench.console
        self._interactive = False
        self._owns_process = False

    def register(self) -> None:
        self.run_action = self._action(_("▶ Run"), "F5", self.run)
        self.stop_action = self._action(_("⬛ Stop"), "Shift+F5", self.toolchain.stop)
        self.build_action = self._action(_("🔨 Build"), "Ctrl+B", self.build, separator=True)
        self.arguments = ProgramArgumentsField(self.workbench.editor)
        self.workbench.toolbar.addWidget(self.arguments)
        self.workbench.add_toolbar_separator()
        self.stop_action.setEnabled(False)
        environment = self.workbench.services.environment
        self.workbench.add_settings_page(lambda: EnvironmentPage(environment))
        self.toolchain.output_received.connect(self._on_stdout)
        self.toolchain.error_received.connect(self._on_stderr)
        self.toolchain.execution_finished.connect(self._on_finished)
        self.console.input_submitted.connect(self.toolchain.write_input)

    def run(self) -> None:
        editor = self.workbench.editor.current_editor()
        if editor is None or self._busy():
            return
        program_args = self._program_arguments()
        if program_args is None:
            return
        self.console.clear()
        self._begin(interactive=True)
        config = self._start_run(editor, program_args)
        if config is None:
            self._end()
            return
        self.workbench.events.program_started.emit(config)

    def build(self) -> None:
        path = self.workbench.editor.current_file_path()
        if path is None:
            QMessageBox.warning(
                self.workbench.window, _("No File"), _("Please save your file before building.")
            )
            return
        if self._busy():
            return
        self.console.clear()
        self._begin(interactive=False)
        self.toolchain.build(configuration_for_file(path))

    def _start_run(self, editor, program_args: tuple[str, ...]) -> RunConfiguration | None:
        if editor.file_path is None:
            return self.toolchain.run_untitled(editor.toPlainText(), program_args)
        config = configuration_for_file(editor.file_path, program_args)
        self.toolchain.run(config)
        return config

    def _program_arguments(self) -> tuple[str, ...] | None:
        try:
            return self.arguments.arguments()
        except ProgramArgumentsError as error:
            QMessageBox.warning(self.workbench.window, _("Program arguments"), str(error))
            return None

    def _busy(self) -> bool:
        if not self.toolchain.is_running():
            return False
        self.workbench.show_status_message(_("A process is already running."))
        return True

    def _begin(self, interactive: bool) -> None:
        self._interactive = interactive
        self._owns_process = True
        self.console.set_waiting_for_input(interactive)
        self._set_running(True)

    def _end(self) -> None:
        self._owns_process = False
        self.console.set_waiting_for_input(False)
        self.console.disable_input()
        self._set_running(False)

    def _on_stdout(self, text: str) -> None:
        if not self._owns_process:
            return
        self.console.append_output(text)
        self.workbench.events.process_output.emit(text, "stdout")

    def _on_stderr(self, text: str) -> None:
        if not self._owns_process:
            return
        self.console.append_error(text)
        self.workbench.events.process_output.emit(text, "stderr")

    def _on_finished(self, exit_code: int) -> None:
        if not self._owns_process:
            return
        self._end()
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
