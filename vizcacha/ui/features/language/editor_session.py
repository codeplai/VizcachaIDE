"""One CodeEditor connected to the language server: document sync and its state."""

from pathlib import Path

from PyQt5.QtCore import QObject, QTimer, pyqtSignal
from PyQt5.QtGui import QTextCursor

from vizcacha.application.ports import LanguageServerPort
from vizcacha.domain.diagnostics import Diagnostic, SourceLocation
from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.features.language.editor_layers import (
    DIAGNOSTICS_LAYER,
    HIGHLIGHTS_LAYER,
    show_diagnostics,
    show_occurrences,
)
from vizcacha.ui.features.language.locations import cursor_location, is_go_file, same_file

CHANGE_DEBOUNCE_MS = 300
HIGHLIGHT_DEBOUNCE_MS = 300


class EditorSession(QObject):
    """Keeps gopls' copy of ``editor``'s text up to date (didOpen/didChange/didClose)."""

    analyzed = pyqtSignal(object)  # CodeEditor: new diagnostics arrived for it

    def __init__(self, editor: CodeEditor, server: LanguageServerPort) -> None:
        super().__init__(editor)
        self.editor = editor
        self.server = server
        self.path: Path | None = None
        self.version = 0
        self.diagnostics: list[Diagnostic] = []
        self.is_analyzed = False  # gopls already answered for this document
        self._dirty = False
        self._change_timer = self._single_shot(CHANGE_DEBOUNCE_MS, self.flush)
        self._highlight_timer = self._single_shot(HIGHLIGHT_DEBOUNCE_MS, self.refresh_highlights)
        editor.textChanged.connect(self._on_text_changed)
        editor.cursorPositionChanged.connect(self._highlight_timer.start)

    def _single_shot(self, interval_ms: int, slot) -> QTimer:
        timer = QTimer(self)
        timer.setSingleShot(True)
        timer.setInterval(interval_ms)
        timer.timeout.connect(slot)
        return timer

    # --- synchronisation ------------------------------------------------------
    def sync_path(self) -> None:
        """Open/close the document when the editor's file changes (open, Save As…)."""
        path = self.editor.file_path if is_go_file(self.editor.file_path) else None
        if (path is None and self.path is None) or same_file(path, self.path):
            return
        self.release()
        if path is None:
            return
        self.path = Path(path)
        self.version = 1
        self.server.open_document(self.path, self.editor.toPlainText())

    def flush(self) -> None:
        """Send the pending change now (before a query, or when the debounce expires)."""
        self._change_timer.stop()
        if not self._dirty or self.path is None:
            return
        self._dirty = False
        self.version += 1
        self.server.change_document(self.path, self.editor.toPlainText(), self.version)

    def release(self) -> None:
        """didClose and clear everything painted for the old document."""
        self._change_timer.stop()
        if self.path is not None:
            self.server.close_document(self.path)
        self.forget()

    def forget(self) -> None:
        """Drop the document without telling the server (it was shut down)."""
        self.path = None
        self._dirty = False
        self.is_analyzed = False
        self.diagnostics = []
        self.editor.set_selection_layer(DIAGNOSTICS_LAYER, [])
        self.editor.set_selection_layer(HIGHLIGHTS_LAYER, [])

    def _on_text_changed(self) -> None:
        self.sync_path()
        if self.path is None:
            return
        self._dirty = True
        self._change_timer.start()

    # --- results ----------------------------------------------------------------
    def show_diagnostics(self, diagnostics: list[Diagnostic]) -> None:
        self.diagnostics = list(diagnostics)
        self.is_analyzed = True
        show_diagnostics(self.editor, self.diagnostics)
        self.analyzed.emit(self.editor)

    def location(self, cursor: QTextCursor | None = None) -> SourceLocation | None:
        if self.path is None:
            return None
        return cursor_location(self.editor, cursor or self.editor.textCursor())

    def refresh_highlights(self) -> None:
        query = getattr(self.server, "document_highlights", None)
        location = self.location()
        if query is None or location is None or not self.is_analyzed:
            return
        self.flush()
        show_occurrences(self.editor, query(location))
