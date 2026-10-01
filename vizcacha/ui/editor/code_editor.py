"""Go code editor: line numbers, breakpoints, highlighting and completion.

Extension points for parallel tracks (so they never need to edit this file):

* ``set_selection_layer(name, selections)``: each feature owns a named layer of
  ``QTextEdit.ExtraSelection`` ("debug_line", "diagnostics", "search", "brackets").
* ``completion_provider``: callable ``(text, cursor_position, file_path) ->
  list[CompletionItem]``. Track C replaces the static default with gopls.
* Qt event filters (``installEventFilter`` on the editor or its viewport) for
  extra keys, Ctrl+click, hover tooltips, etc.

Theme, font, zoom, tab width and gutter geometry live in ``EditorAppearance``.
"""

from collections.abc import Callable
from pathlib import Path

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtGui import QColor, QTextCursor, QTextFormat
from PyQt5.QtWidgets import QPlainTextEdit, QTextEdit

from vizcacha.domain.completion import CompletionItem
from vizcacha.infrastructure.static_completion import GoAnalyzer
from vizcacha.ui.editor.autocomplete_popup import AutocompletePopup
from vizcacha.ui.editor.block_selection import spans_several_lines
from vizcacha.ui.editor.bracket_matcher import BracketMatcher
from vizcacha.ui.editor.editor_appearance import EditorAppearance, default_editor_font
from vizcacha.ui.editor.indentation import (
    indent_lines,
    insert_newline,
    outdent_before_closer,
    outdent_lines,
)
from vizcacha.ui.editor.line_number_area import LineNumberArea
from vizcacha.ui.editor.syntax_highlighter import GoSyntaxHighlighter
from vizcacha.ui.editor.themes import LIGHT, EditorTheme

CompletionProvider = Callable[[str, int, Path | None], list[CompletionItem]]
DEBUG_LINE_LAYER = "debug_line"
DEFAULT_TAB_SIZE = 4
NEWLINE_KEYS = (Qt.Key_Return, Qt.Key_Enter)


def static_completion_provider() -> CompletionProvider:
    analyzer = GoAnalyzer()
    return lambda text, position, _path: analyzer.get_completions(text, position)


class CodeEditor(EditorAppearance, QPlainTextEdit):
    breakpoints_changed = pyqtSignal()

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.file_path: Path | None = None
        self.breakpoints: set[int] = set()
        self.current_line: int | None = None
        self.completion_provider: CompletionProvider = static_completion_provider()
        self.theme: EditorTheme = LIGHT
        self.tab_size = DEFAULT_TAB_SIZE
        self.auto_indent = True
        self.line_numbers_visible = True
        self.zoom_steps = 0
        self._base_font = default_editor_font()
        self._selection_layers: dict[str, list[QTextEdit.ExtraSelection]] = {}
        self._popup: AutocompletePopup | None = None
        self.line_number_area = LineNumberArea(self)
        self.highlighter = GoSyntaxHighlighter(self.document(), self.theme)
        self.bracket_matcher = BracketMatcher(self)
        self.blockCountChanged.connect(self._update_margins)
        self.updateRequest.connect(self._update_line_number_area)
        self.set_base_font(self._base_font)
        self.apply_theme(LIGHT)

    def move_cursor_to(self, line: int, column: int = 1) -> None:
        """Place the cursor at 1-based ``line``/``column`` (clamped) and centre it."""
        block = self.document().findBlockByNumber(max(line - 1, 0))
        if not block.isValid():
            block = self.document().lastBlock()
        cursor = self.textCursor()
        cursor.setPosition(block.position() + min(max(column - 1, 0), len(block.text())))
        self.setTextCursor(cursor)
        self.centerCursor()

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
        self.setTextCursor(QTextCursor(self.document().findBlockByNumber(line - 1)))
        self.centerCursor()
        self._show_debug_line(line)

    def clear_current_line_highlight(self) -> None:
        self.current_line = None
        self.set_selection_layer(DEBUG_LINE_LAYER, [])
        self.line_number_area.update()

    def _show_debug_line(self, line: int) -> None:
        selection = QTextEdit.ExtraSelection()
        selection.format.setBackground(QColor(self.theme.current_line))
        selection.format.setProperty(QTextFormat.FullWidthSelection, True)
        selection.cursor = QTextCursor(self.document().findBlockByNumber(line - 1))
        self.set_selection_layer(DEBUG_LINE_LAYER, [selection])
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

    # --- keys -------------------------------------------------------------
    def keyPressEvent(self, event) -> None:  # noqa: N802 - Qt override
        if event.key() == Qt.Key_Space and event.modifiers() == Qt.ControlModifier:
            self.show_autocomplete()
            return
        popup_keys = (Qt.Key_Up, Qt.Key_Down, *NEWLINE_KEYS, Qt.Key_Escape)
        if self._popup is not None and self._popup.isVisible() and event.key() in popup_keys:
            self._popup.keyPressEvent(event)
            return
        if self._handle_editing_key(event.key()):
            return
        if event.text() == "}" and self.auto_indent:
            outdent_before_closer(self)
        super().keyPressEvent(event)

    def _handle_editing_key(self, key: int) -> bool:
        if key in NEWLINE_KEYS:
            insert_newline(self, self.auto_indent)
        elif key == Qt.Key_Tab and spans_several_lines(self):
            indent_lines(self)
        elif key == Qt.Key_Backtab:
            outdent_lines(self, self.tab_size)
        else:
            return False
        return True
