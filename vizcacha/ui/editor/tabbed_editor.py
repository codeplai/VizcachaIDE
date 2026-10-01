"""Tabs of CodeEditors: open, save, close, unsaved-changes guard."""

from collections.abc import Iterator
from pathlib import Path

from PyQt5.QtCore import pyqtSignal
from PyQt5.QtWidgets import QFileDialog, QMessageBox, QTabWidget

from vizcacha.i18n import _
from vizcacha.ui.editor.code_editor import CodeEditor

MODIFIED_MARK = "*"


class TabbedEditor(QTabWidget):
    active_file_changed = pyqtSignal(str)  # path or "" for untitled
    editor_created = pyqtSignal(object)  # CodeEditor, so features can attach to it
    file_opened = pyqtSignal(str)
    file_saved = pyqtSignal(str)

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setTabsClosable(True)
        self.setMovable(True)
        self.tabCloseRequested.connect(self.close_tab)
        self.currentChanged.connect(self._on_tab_changed)
        self._untitled_counter = 0

    # --- queries ----------------------------------------------------------
    def editors(self) -> Iterator[CodeEditor]:
        for index in range(self.count()):
            yield self.widget(index)

    def current_editor(self) -> CodeEditor | None:
        return self.currentWidget()

    def current_file_path(self) -> Path | None:
        editor = self.current_editor()
        return editor.file_path if editor else None

    def get_all_breakpoints(self) -> list[int]:
        editor = self.current_editor()
        return editor.get_breakpoints() if editor else []

    # --- tabs -------------------------------------------------------------
    def new_tab(self, file_path: Path | None = None, content: str = "") -> CodeEditor:
        editor = CodeEditor()
        editor.setPlainText(content)
        editor.file_path = file_path
        editor.document().setModified(False)
        editor.untitled_name = self._next_untitled_name() if file_path is None else ""
        index = self.addTab(editor, self._base_title(editor))
        editor.document().modificationChanged.connect(
            lambda modified, e=editor: self._refresh_title(e, modified)
        )
        self.editor_created.emit(editor)
        self.setCurrentIndex(index)
        return editor

    def open_file(self, file_path: Path) -> bool:
        file_path = Path(file_path)
        for index, editor in enumerate(self.editors()):
            if editor.file_path == file_path:
                self.setCurrentIndex(index)
                return True
        try:
            content = file_path.read_text(encoding="utf-8")
        except OSError as error:
            QMessageBox.critical(
                self, _("Error"), _("Could not open file: {error}").format(error=error)
            )
            return False
        self._reuse_or_create_tab(file_path, content)
        self.file_opened.emit(str(file_path))
        return True

    def close_tab(self, index: int) -> bool:
        editor = self.widget(index)
        if not self.check_save_needed(editor):
            return False
        self.removeTab(index)
        if self.count() == 0:
            self.new_tab()
        return True

    # --- saving -----------------------------------------------------------
    def check_save_needed(self, editor: CodeEditor) -> bool:
        """Ask to save unsaved changes. Returns False if the user cancelled."""
        if not editor.document().isModified():
            return True
        reply = QMessageBox.question(
            self,
            _("Unsaved Changes"),
            _("Do you want to save changes to '{name}'?").format(name=self._base_title(editor)),
            QMessageBox.Save | QMessageBox.Discard | QMessageBox.Cancel,
        )
        if reply == QMessageBox.Save:
            self.setCurrentWidget(editor)
            return self.save_current_tab()
        return reply == QMessageBox.Discard

    def save_current_tab(self) -> bool:
        editor = self.current_editor()
        if editor is None:
            return False
        if editor.file_path is None:
            return self.save_current_tab_as()
        return self.save_to_file(editor, editor.file_path)

    def save_current_tab_as(self) -> bool:
        editor = self.current_editor()
        if editor is None:
            return False
        filename, _filter = QFileDialog.getSaveFileName(
            self, _("Save Go File"), "", _("Go Files (*.go);;All Files (*)")
        )
        if not filename:
            return False
        path = Path(filename)
        if path.suffix != ".go":
            path = path.with_name(path.name + ".go")
        return self.save_to_file(editor, path)

    def save_to_file(self, editor: CodeEditor, path: Path) -> bool:
        try:
            Path(path).write_text(editor.toPlainText(), encoding="utf-8")
        except OSError as error:
            QMessageBox.critical(
                self, _("Error"), _("Could not save file: {error}").format(error=error)
            )
            return False
        editor.file_path = Path(path)
        editor.document().setModified(False)
        self._refresh_title(editor, False)
        self.file_saved.emit(str(path))
        if editor is self.current_editor():
            self.active_file_changed.emit(str(path))
        return True

    # --- current-line helpers used by the debugger -------------------------
    def highlight_current_line(self, line: int) -> None:
        editor = self.current_editor()
        if editor:
            editor.highlight_current_line(line)

    def clear_current_line_highlight(self) -> None:
        for editor in self.editors():
            editor.clear_current_line_highlight()

    def toggle_breakpoint(self) -> None:
        editor = self.current_editor()
        if editor:
            editor.toggle_breakpoint()

    # --- internals ----------------------------------------------------------
    def _reuse_or_create_tab(self, file_path: Path, content: str) -> None:
        editor = self.current_editor()
        pristine = editor is not None and editor.file_path is None and not editor.toPlainText()
        if not pristine:
            self.new_tab(file_path, content)
            return
        editor.setPlainText(content)
        editor.file_path = file_path
        editor.document().setModified(False)
        self._refresh_title(editor, False)
        self.active_file_changed.emit(str(file_path))

    def _next_untitled_name(self) -> str:
        self._untitled_counter += 1
        if self._untitled_counter == 1:
            return _("Untitled")
        return _("Untitled {number}").format(number=self._untitled_counter)

    def _base_title(self, editor: CodeEditor) -> str:
        return editor.file_path.name if editor.file_path else editor.untitled_name

    def _refresh_title(self, editor: CodeEditor, modified: bool) -> None:
        index = self.indexOf(editor)
        if index < 0:
            return
        suffix = MODIFIED_MARK if modified else ""
        self.setTabText(index, self._base_title(editor) + suffix)

    def _on_tab_changed(self, index: int) -> None:
        if index < 0:
            return
        path = self.widget(index).file_path
        self.active_file_changed.emit(str(path) if path else "")
