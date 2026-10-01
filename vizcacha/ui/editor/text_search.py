"""Find / replace inside a CodeEditor, built on QTextDocument.find."""

from dataclasses import dataclass

from PyQt5.QtGui import QColor, QTextCursor, QTextDocument
from PyQt5.QtWidgets import QPlainTextEdit, QTextEdit

SEARCH_LAYER = "search"
MAX_HIGHLIGHTED_MATCHES = 2000


@dataclass(frozen=True)
class SearchQuery:
    text: str
    case_sensitive: bool = False
    whole_word: bool = False

    def flags(self, backward: bool = False) -> QTextDocument.FindFlags:
        flags = QTextDocument.FindFlags()
        if self.case_sensitive:
            flags |= QTextDocument.FindCaseSensitively
        if self.whole_word:
            flags |= QTextDocument.FindWholeWords
        if backward:
            flags |= QTextDocument.FindBackward
        return flags


def find_all(document: QTextDocument, query: SearchQuery) -> list[QTextCursor]:
    if not query.text:
        return []
    matches = []
    cursor = document.find(query.text, 0, query.flags())
    while not cursor.isNull() and len(matches) < MAX_HIGHLIGHTED_MATCHES:
        matches.append(cursor)
        cursor = document.find(query.text, cursor, query.flags())
    return matches


def find_next(editor: QPlainTextEdit, query: SearchQuery, backward: bool = False) -> bool:
    """Select the next (or previous) match, wrapping around the document."""
    if not query.text:
        return False
    document = editor.document()
    flags = query.flags(backward)
    found = document.find(query.text, editor.textCursor(), flags)
    if found.isNull():
        wrap_from = QTextCursor(document)
        wrap_from.movePosition(QTextCursor.End if backward else QTextCursor.Start)
        found = document.find(query.text, wrap_from, flags)
    if found.isNull():
        return False
    editor.setTextCursor(found)
    return True


def replace_current(editor: QPlainTextEdit, query: SearchQuery, replacement: str) -> bool:
    """Replace the selected match (if the selection is one) and move to the next match."""
    cursor = editor.textCursor()
    if _is_match(editor.document(), query, cursor):
        cursor.insertText(replacement)
        editor.setTextCursor(cursor)
    return find_next(editor, query)


def replace_all(editor: QPlainTextEdit, query: SearchQuery, replacement: str) -> int:
    matches = find_all(editor.document(), query)
    if not matches:
        return 0
    edit = QTextCursor(editor.document())
    edit.beginEditBlock()
    for match in matches:  # QTextCursors follow the edits made before them
        match.insertText(replacement)
    edit.endEditBlock()
    return len(matches)


def match_selections(matches: list[QTextCursor], color: str) -> list[QTextEdit.ExtraSelection]:
    selections = []
    for match in matches:
        selection = QTextEdit.ExtraSelection()
        selection.format.setBackground(QColor(color))
        selection.cursor = match
        selections.append(selection)
    return selections


def _is_match(document: QTextDocument, query: SearchQuery, cursor: QTextCursor) -> bool:
    if not cursor.hasSelection() or not query.text:
        return False
    start = QTextCursor(document)
    start.setPosition(cursor.selectionStart())
    found = document.find(query.text, start, query.flags())
    return (
        not found.isNull()
        and found.selectionStart() == cursor.selectionStart()
        and found.selectionEnd() == cursor.selectionEnd()
    )
