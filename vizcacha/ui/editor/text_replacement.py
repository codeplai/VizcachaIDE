"""Replace the whole text of an editor (gofmt, external reload) keeping the view.

The cursor is mapped by counting non-whitespace characters, which survives the
whitespace-only changes that gofmt makes. The scroll position is restored.
The replacement is one undo step.
"""

from PyQt5.QtGui import QTextCursor
from PyQt5.QtWidgets import QPlainTextEdit


def map_cursor_offset(old_text: str, new_text: str, offset: int) -> int:
    """Position in ``new_text`` equivalent to ``offset`` in ``old_text``."""
    offset = max(0, min(offset, len(old_text)))
    before = old_text[:offset]
    tokens_before = sum(1 for character in before if not character.isspace())
    rest = old_text[offset:]
    gap = rest[: len(rest) - len(rest.lstrip())]
    # Cursor in the indentation of a line with code: stay in front of that code.
    snap_forward = (not before or before[-1].isspace()) and rest.strip() and "\n" not in gap
    position = _position_after_tokens(new_text, tokens_before)
    if snap_forward:
        while position < len(new_text) and new_text[position].isspace():
            position += 1
    return position


def _position_after_tokens(text: str, tokens: int) -> int:
    if tokens == 0:
        return 0
    seen = 0
    for index, character in enumerate(text):
        if character.isspace():
            continue
        seen += 1
        if seen == tokens:
            return index + 1
    return len(text)


def replace_text_keeping_view(editor: QPlainTextEdit, new_text: str) -> bool:
    """Returns False when the text was already identical (nothing changed)."""
    old_text = editor.toPlainText()
    if old_text == new_text:
        return False
    vertical = editor.verticalScrollBar().value()
    horizontal = editor.horizontalScrollBar().value()
    old_position = editor.textCursor().position()
    cursor = QTextCursor(editor.document())
    cursor.beginEditBlock()
    cursor.select(QTextCursor.Document)
    cursor.insertText(new_text)
    cursor.endEditBlock()
    restored = editor.textCursor()
    restored.setPosition(map_cursor_offset(old_text, new_text, old_position))
    editor.setTextCursor(restored)
    editor.verticalScrollBar().setValue(vertical)
    editor.horizontalScrollBar().setValue(horizontal)
    return True
