"""Project: File > Open Folder..., Files panel and Tools > Go Modules...

The Files panel is created the first time a folder is opened (or restored), so
the View menu only lists it when there is something to show.
"""

from pathlib import Path

from PyQt5.QtWidgets import QAction, QDockWidget, QFileDialog, QMenu, QMessageBox

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.i18n import _
from vizcacha.ui.features.project.files_panel import FilesPanel
from vizcacha.ui.features.project.modules_dialog import GoModulesDialog
from vizcacha.ui.workbench import Workbench

LAST_FOLDER_KEY = SettingsKeys.LAST_FOLDER
FILES_PANEL_ID = "project_files"


def _move_before_last_separator(menu: QMenu, action: QAction) -> None:
    separators = [item for item in menu.actions() if item.isSeparator()]
    if not separators:
        return
    menu.removeAction(action)
    menu.insertAction(separators[-1], action)


class ProjectFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.settings = workbench.services.settings
        self.toolchain = workbench.services.toolchain
        self.console = workbench.console
        self.folder: Path | None = None
        self.panel: FilesPanel | None = None
        self.dock: QDockWidget | None = None
        self.dialog: GoModulesDialog | None = None
        self._owns_process = False

    def register(self) -> None:
        open_folder = self._action(_("Open &Folder..."), self.open_folder)
        self.workbench.add_action("file", open_folder)
        _move_before_last_separator(self.workbench.menu("file"), open_folder)
        self.workbench.add_action("tools", self._action(_("Go &Modules..."), self.show_modules))
        self.toolchain.output_received.connect(self._on_stdout)
        self.toolchain.error_received.connect(self._on_stderr)
        self.toolchain.execution_finished.connect(self._on_finished)
        self._restore_last_folder()

    # --- folder and Files panel -------------------------------------------------
    def open_folder(self) -> None:
        start = str(self.folder) if self.folder else ""
        chosen = QFileDialog.getExistingDirectory(self.workbench.window, _("Open Folder"), start)
        if chosen:
            self.set_folder(Path(chosen))

    def set_folder(self, folder: Path) -> None:
        self.folder = Path(folder)
        self.settings.set(LAST_FOLDER_KEY, str(self.folder))
        self._ensure_panel().set_folder(self.folder)
        self.dock.show()
        if self.dialog is not None:
            self.dialog.set_folder(self.folder)

    def _ensure_panel(self) -> FilesPanel:
        if self.panel is None:
            self.panel = FilesPanel()
            self.panel.file_activated.connect(self._open_file)
            self.dock = self.workbench.add_panel(FILES_PANEL_ID, _("Files"), self.panel, "left")
        return self.panel

    def _open_file(self, path: Path) -> None:
        self.workbench.events.navigate_to.emit(SourceLocation(path, 1))

    def _restore_last_folder(self) -> None:
        last = self.settings.get(LAST_FOLDER_KEY, "")
        if last and Path(last).is_dir():
            self.set_folder(Path(last))

    # --- Go modules ---------------------------------------------------------------
    def show_modules(self) -> None:
        folder = self._modules_folder()
        if folder is None:
            QMessageBox.information(
                self.workbench.window,
                _("Go Modules"),
                _("Open a folder or save your file first."),
            )
            return
        if self.dialog is None:
            self.dialog = GoModulesDialog(self.workbench.window)
            self.dialog.command_requested.connect(self.run_module_command)
        self.dialog.set_folder(folder)
        self.dialog.show()
        self.dialog.raise_()

    def run_module_command(self, working_dir: Path, arguments: list[str]) -> None:
        if self.toolchain.is_running():
            self.workbench.show_status_message(_("A process is already running."))
            return
        self._owns_process = True
        self._refresh_dialog(busy=True)
        if not self.toolchain.run_go_command(working_dir, arguments):
            self._owns_process = False
            self._refresh_dialog(busy=False)

    def _modules_folder(self) -> Path | None:
        if self.folder is not None:
            return self.folder
        path = self.workbench.editor.current_file_path()
        return path.parent if path else None

    def _refresh_dialog(self, busy: bool) -> None:
        if self.dialog is not None:
            self.dialog.refresh(busy=busy)

    def _on_stdout(self, text: str) -> None:
        if self._owns_process:
            self.console.append_output(text)

    def _on_stderr(self, text: str) -> None:
        if self._owns_process:
            self.console.append_error(text)

    def _on_finished(self, exit_code: int) -> None:
        if not self._owns_process:
            return
        self._owns_process = False
        if exit_code == 0:
            self.console.append_success("\n" + _("[Command finished successfully]") + "\n")
        else:
            message = _("[Process exited with code {code}]").format(code=exit_code)
            self.console.append_error("\n" + message + "\n")
        self._refresh_dialog(busy=False)

    def _action(self, text: str, slot) -> QAction:
        action = QAction(text, self.workbench.window)
        action.triggered.connect(lambda _checked=False: slot())
        return action


def register(workbench: Workbench) -> None:
    ProjectFeature(workbench).register()
