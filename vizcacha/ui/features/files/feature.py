"""File menu: New, Open, Save, Save As, Exit; window title; last-file restore."""

from pathlib import Path

from PyQt5.QtGui import QKeySequence
from PyQt5.QtWidgets import QAction, QFileDialog

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import _
from vizcacha.ui.workbench import Workbench


class FilesFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.editor = workbench.editor
        self.settings = workbench.services.settings

    def register(self) -> None:
        window = self.workbench.window
        self._action("file", _("&New"), QKeySequence.New, self.editor.new_tab)
        self._action("file", _("&Open..."), QKeySequence.Open, self.open_file)
        self._action("file", _("&Save"), QKeySequence.Save, self.editor.save_current_tab)
        self._action("file", _("Save &As..."), QKeySequence.SaveAs, self.editor.save_current_tab_as)
        exit_action = self._action(
            "file", _("E&xit"), QKeySequence.Quit, window.close, separator=True
        )
        exit_action.setMenuRole(QAction.QuitRole)
        self.editor.active_file_changed.connect(self._update_window_title)
        self.editor.file_opened.connect(self._remember_file)
        self.editor.file_saved.connect(self._remember_file)
        self.workbench.events.navigate_to.connect(self._navigate_to)
        self.workbench.add_close_guard(self._confirm_unsaved_changes)
        self._restore_last_file()

    def open_file(self) -> None:
        filename, _filter = QFileDialog.getOpenFileName(
            self.workbench.window, _("Open Go File"), "", _("Go Files (*.go);;All Files (*)")
        )
        if filename:
            self.editor.open_file(Path(filename))

    def _action(self, menu: str, text: str, shortcut, slot, separator: bool = False) -> QAction:
        action = QAction(text, self.workbench.window)
        action.setShortcut(shortcut)
        action.triggered.connect(lambda _checked=False: slot())
        return self.workbench.add_action(menu, action, separator=separator)

    def _navigate_to(self, location) -> None:
        if not self.editor.open_file(location.file):
            return
        editor = self.editor.current_editor()
        block = editor.document().findBlockByNumber(max(location.line - 1, 0))
        cursor = editor.textCursor()
        cursor.setPosition(block.position() + max(location.column - 1, 0))
        editor.setTextCursor(cursor)
        editor.setFocus()

    def _remember_file(self, path: str) -> None:
        self.settings.set(SettingsKeys.LAST_FILE, path)

    def _restore_last_file(self) -> None:
        last_file = self.settings.get(SettingsKeys.LAST_FILE, "")
        if last_file and Path(last_file).is_file():
            self.editor.open_file(Path(last_file))

    def _update_window_title(self, path: str) -> None:
        name = Path(path).name if path else _("Untitled")
        self.workbench.window.setWindowTitle(_("VizcachaIDE - {name}").format(name=name))

    def _confirm_unsaved_changes(self) -> bool:
        return all(self.editor.check_save_needed(editor) for editor in list(self.editor.editors()))


def register(workbench: Workbench) -> None:
    FilesFeature(workbench).register()
