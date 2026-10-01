"""Go code editor: line numbers, breakpoints, highlighting and completion.

Extension points for parallel tracks (so they never need to edit this file):

* ``set_selection_layer(name, selections)``: each feature owns a named layer of
  ``QTextEdit.ExtraSelection`` (e.g. "debug_line", "diagnostics", "search").
* ``completion_provider``: callable ``(text, cursor_position, file_path) ->
  list[CompletionItem]``. Track C replaces the static default with gopls.
* Qt event filters (``installEventFilter`` on the editor or its viewport) for
  extra keys, Ctrl+click, hover tooltips, etc.
"""

from collections.abc import Callable
from pathlib import Path

from PyQt5.QtCore import QRect, Qt, pyqtSignal
from PyQt5.QtGui import QColor, QFont, QPalette, QTextCursor, QTextFormat
from PyQt5.QtWidgets import QPlainTextEdit, QTextEdit

from vizcacha.domain.completion import CompletionItem
from vizcacha.infrastructure.static_completion import GoAnalyzer
from vizcacha.ui.editor.autocomplete_popup import AutocompletePopup
from vizcacha.ui.editor.line_number_area import LineNumberArea
from vizcacha.ui.editor.syntax_highlighter import GoSyntaxHighlighter

CompletionProvider = Callable[[str, int, Path | None], list[CompletionItem]]
DEBUG_LINE_LAYER = "debug_line"
INDENT = "    "


def static_completion_provider() -> CompletionProvider:
    analyzer = GoAnalyzer()
    return lambda text, position, _path: analyzer.get_completions(text, position)


class CodeEditor(QPlainTextEdit):
    breakpoints_changed = pyqtSignal()

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.file_path: Path | None = None
        self.breakpoints: set[int] = set()
        self.current_line: int | None = None
        self.completion_provider: CompletionProvider = static_completion_provider()
        self._selection_layers: dict[str, list[QTextEdit.ExtraSelection]] = {}
        self._popup: AutocompletePopup | None = None
        self._configure_appearance()
        self.line_number_area = LineNumberArea(self)
        self.highlighter = GoSyntaxHighlighter(self.document())
        self.blockCountChanged.connect(self._update_margins)
        self.updateRequest.connect(self._update_line_number_area)
        self._update_margins()

    def _configure_appearance(self) -> None:
        font = QFont("Consolas", 11)
        if not font.exactMatch():
            font = QFont("Courier New", 11)
        self.setFont(font)
        self.setTabStopDistance(self.fontMetrics().horizontalAdvance(" ") * 4)
        palette = self.palette()
        palette.setColor(QPalette.Base, QColor("#FFFFFF"))
        palette.setColor(QPalette.Text, QColor("#000000"))
        self.setPalette(palette)

    # --- extension points -------------------------------------------------
    def set_selection_layer(self, name: str, selections: list[QTextEdit.ExtraSelection]) -> None:
        self._selection_layers[name] = list(selections)
        merged = [s for layer in self._selection_layers.values() for s in layer]
        self.setExtraSelections(merged)

    # --- breakpoints and current line ------------------------------------
    def toggle_breakpoint(self) -> None:
        self.toggle_breakpoint_at_line(self.textCursor().blockNumber() + 1)

    def toggle_breakpoint_at_line(self, line: int) -> None:
        self.breakpoints.symmetric_difference_update({line})
        self.line_number_area.update()
        self.breakpoints_changed.emit()

    def get_breakpoints(self) -> list[int]:
        return sorted(self.breakpoints)

    def highlight_current_line(self, line: int) -> None:
        self.current_line = line
        block = self.document().findBlockByNumber(line - 1)
        cursor = QTextCursor(block)
        self.setTextCursor(cursor)
        self.centerCursor()
        selection = QTextEdit.ExtraSelection()
        selection.format.setBackground(QColor("#FFFF00").lighter(160))
        selection.format.setProperty(QTextFormat.FullWidthSelection, True)
        selection.cursor = cursor
        self.set_selection_layer(DEBUG_LINE_LAYER, [selection])
        self.line_number_area.update()

    def clear_current_line_highlight(self) -> None:
        self.current_line = None
        self.set_selection_layer(DEBUG_LINE_LAYER, [])
        self.line_number_area.update()

    # --- completion ---------------------------------------------------------
    def show_autocomplete(self) -> None:
        position = self.textCursor().position()
        completions = self.completion_provider(self.toPlainText(), position, self.file_path)
        if not completions:
            return
        if self._popup is None:
            self._popup = AutocompletePopup(self)
            self._popup.completion_selected.connect(self.insert_completion)
        self._popup.show_completions(completions, self.mapToGlobal(self.cursorRect().bottomLeft()))

    def insert_completion(self, item: CompletionItem) -> None:
        cursor = self.textCursor()
        cursor.select(QTextCursor.WordUnderCursor)
        partial = cursor.selectedText()
        if partial and (partial[0].isalnum() or partial[0] == "_"):
            cursor.removeSelectedText()
        cursor.insertText(item.text_to_insert)
        self.setTextCursor(cursor)

    # --- Qt overrides -----------------------------------------------------
    def keyPressEvent(self, event) -> None:  # noqa: N802 - Qt override
        if event.key() == Qt.Key_Space and event.modifiers() == Qt.ControlModifier:
            self.show_autocomplete()
            return
        popup_keys = (Qt.Key_Up, Qt.Key_Down, Qt.Key_Return, Qt.Key_Enter, Qt.Key_Escape)
        if self._popup is not None and self._popup.isVisible() and event.key() in popup_keys:
            self._popup.keyPressEvent(event)
            return
        if event.key() in (Qt.Key_Return, Qt.Key_Enter):
            self._insert_newline_with_indent(event)
            return
        super().keyPressEvent(event)

    def _insert_newline_with_indent(self, event) -> None:
        line = self.textCursor().block().text()
        indent = line[: len(line) - len(line.lstrip())]
        super().keyPressEvent(event)
        extra = INDENT if line.rstrip().endswith(("{", ":")) else ""
        self.textCursor().insertText(indent + extra)

    def resizeEvent(self, event) -> None:  # noqa: N802 - Qt override
        super().resizeEvent(event)
        rect = self.contentsRect()
        self.line_number_area.setGeometry(
            QRect(rect.left(), rect.top(), self.line_number_area.preferred_width(), rect.height())
        )

    def _update_margins(self, _block_count: int = 0) -> None:
        self.setViewportMargins(self.line_number_area.preferred_width(), 0, 0, 0)

    def _update_line_number_area(self, rect, dy: int) -> None:
        if dy:
            self.line_number_area.scroll(0, dy)
        else:
            self.line_number_area.update(0, rect.y(), self.line_number_area.width(), rect.height())
        if rect.contains(self.viewport().rect()):
            self._update_margins()
