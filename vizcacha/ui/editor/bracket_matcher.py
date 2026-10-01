"""Highlights the bracket next to the cursor and its partner: () [] {}.

Uses the "brackets" selection layer of the CodeEditor. Brackets inside strings
or comments are not skipped (a known, beginner-acceptable simplification).
"""

from PyQt5.QtCore import QObject
from PyQt5.QtGui import QColor, QTextCursor
from PyQt5.QtWidgets import QTextEdit

BRACKETS_LAYER = "brackets"
PARTNERS = {"(": ")", "[": "]", "{": "}", ")": "(", "]": "[", "}": "{"}
OPENERS = "([{"
MAX_SCAN_CHARACTERS = 200_000


def find_matching_bracket(text: str, index: int) -> int | None:
    """Index of the bracket that matches ``text[index]``, or None."""
    if not 0 <= index < len(text) or text[index] not in PARTNERS:
        return None
    same = text[index]
    partner = PARTNERS[same]
    step = 1 if same in OPENERS else -1
    depth = 0
    position = index
    limit = MAX_SCAN_CHARACTERS
    while 0 <= position < len(text) and limit > 0:
        character = text[position]
        if character == same:
            depth += 1
        elif character == partner:
            depth -= 1
        if depth == 0:
            return position
        position += step
        limit -= 1
    return None


def bracket_pair_at(text: str, cursor_position: int) -> tuple[int, int] | None:
    """The pair touching the cursor: the bracket before it first, then the one after it."""
    for index in (cursor_position - 1, cursor_position):
        match = find_matching_bracket(text, index)
        if match is not None:
            return index, match
    return None


class BracketMatcher(QObject):
    def __init__(self, editor) -> None:
        super().__init__(editor)
        self.editor = editor
        editor.cursorPositionChanged.connect(self.refresh)

    def refresh(self) -> None:
        position = self.editor.textCursor().position()
        if not self._touches_bracket(position):
            self.editor.set_selection_layer(BRACKETS_LAYER, [])
            return
        pair = bracket_pair_at(self.editor.toPlainText(), position)
        indexes = pair if pair is not None else ()
        self.editor.set_selection_layer(BRACKETS_LAYER, [self._selection(i) for i in indexes])

    def _touches_bracket(self, position: int) -> bool:
        document = self.editor.document()
        before = document.characterAt(position - 1) if position > 0 else ""
        return before in PARTNERS or document.characterAt(position) in PARTNERS

    def _selection(self, index: int) -> QTextEdit.ExtraSelection:
        selection = QTextEdit.ExtraSelection()
        selection.format.setBackground(QColor(self.editor.theme.bracket_match))
        cursor = QTextCursor(self.editor.document())
        cursor.setPosition(index)
        cursor.setPosition(index + 1, QTextCursor.KeepAnchor)
        selection.cursor = cursor
        return selection
