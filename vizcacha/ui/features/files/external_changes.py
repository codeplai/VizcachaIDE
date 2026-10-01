"""Reloads open files that were changed outside the IDE (QFileSystemWatcher).

* No local changes: the editor reloads silently, keeping cursor and scroll.
* Local changes: the user is asked whether to reload (losing them) or keep them.
* Our own saves are ignored: the disk content equals what the IDE last wrote.
* If the user keeps the local version, the question is not repeated until the
  file changes on disk again.
"""

from collections.abc import Callable
from pathlib import Path

from PyQt5.QtCore import QFileSystemWatcher, QObject, QTimer
from PyQt5.QtWidgets import QMessageBox

from vizcacha.i18n import _
from vizcacha.ui.editor.tabbed_editor import TabbedEditor
from vizcacha.ui.editor.text_replacement import replace_text_keeping_view

DEBOUNCE_MS = 150  # editors often write a file in several steps

ReloadQuestion = Callable[[Path], bool]


class ExternalChangeWatcher(QObject):
    def __init__(self, tabs: TabbedEditor, ask_reload: ReloadQuestion | None = None) -> None:
        super().__init__(tabs)
        self.tabs = tabs
        self.ask_reload: ReloadQuestion = ask_reload or self._ask_user
        self._watcher = QFileSystemWatcher(self)
        self._pending: set[str] = set()
        self._asking: set[str] = set()
        self._timer = QTimer(self)
        self._timer.setSingleShot(True)
        self._timer.setInterval(DEBOUNCE_MS)
        self._timer.timeout.connect(self._process_pending)
        self._watcher.fileChanged.connect(self._on_file_changed)
        self._known_content: dict[str, str] = {}  # what the IDE last read or wrote
        tabs.file_opened.connect(self._remember_and_watch)
        tabs.file_saved.connect(self._remember_and_watch)

    def watch(self, path: str) -> None:
        if Path(path).is_file() and path not in self._watcher.files():
            self._watcher.addPath(path)

    def _remember_and_watch(self, path: str) -> None:
        editor = self.tabs.editor_for_path(Path(path))
        if editor is not None:
            self._known_content[str(Path(path))] = editor.toPlainText()
        self.watch(path)

    def watched_files(self) -> list[str]:
        return self._watcher.files()

    def _on_file_changed(self, path: str) -> None:
        self._pending.add(path)
        self._timer.start()

    def _process_pending(self) -> None:
        pending, self._pending = self._pending, set()
        for path in pending:
            self.check(path)

    def check(self, path: str) -> None:
        """Compare the file on disk with its editor and reload or ask."""
        editor = self.tabs.editor_for_path(Path(path))
        if editor is None:
            self._watcher.removePath(path)
            return
        self.watch(path)  # atomic saves replace the file and drop the watch
        try:
            disk_text = Path(path).read_text(encoding="utf-8")
        except (OSError, UnicodeDecodeError):
            return  # deleted or unreadable: keep the editor as it is
        known = self._known_content.get(str(Path(path)))
        if disk_text in (editor.toPlainText(), known):
            return  # nothing new on disk (e.g. our own save)
        self._known_content[str(Path(path))] = disk_text
        if editor.document().isModified() and not self._confirm(path):
            return
        replace_text_keeping_view(editor, disk_text)
        editor.document().setModified(False)

    def _confirm(self, path: str) -> bool:
        if path in self._asking:  # a question for this file is already open
            return False
        self._asking.add(path)
        try:
            return self.ask_reload(Path(path))
        finally:
            self._asking.discard(path)

    def _ask_user(self, path: Path) -> bool:
        reply = QMessageBox.question(
            self.tabs,
            _("File Changed on Disk"),
            _(
                "'{name}' was changed outside VizcachaIDE and you have unsaved changes.\n"
                "Reload it and lose your changes?"
            ).format(name=path.name),
            QMessageBox.Yes | QMessageBox.No,
            QMessageBox.No,
        )
        return reply == QMessageBox.Yes
