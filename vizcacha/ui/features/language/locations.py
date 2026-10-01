"""Conversions between QTextCursor positions and domain SourceLocations.

Qt counts positions in UTF-16 units; the domain counts columns in code points.
"""

import os
from pathlib import Path

from PyQt5.QtGui import QTextCursor

from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.infrastructure.gopls_lsp.positions import code_point_index, utf16_offset
from vizcacha.ui.editor import CodeEditor

GO_SUFFIX = ".go"


def is_go_file(path: Path | None) -> bool:
    return path is not None and Path(path).suffix == GO_SUFFIX


def same_file(first: Path | None, second: Path | None) -> bool:
    if first is None or second is None:
        return False
    return os.path.normcase(os.path.abspath(first)) == os.path.normcase(os.path.abspath(second))


def cursor_location(editor: CodeEditor, cursor: QTextCursor) -> SourceLocation | None:
    if not is_go_file(editor.file_path):
        return None
    block = cursor.block()
    column = code_point_index(block.text(), cursor.positionInBlock())
    return SourceLocation(editor.file_path, block.blockNumber() + 1, column + 1)


def offset_location(editor: CodeEditor, position: int) -> SourceLocation | None:
    cursor = QTextCursor(editor.document())
    cursor.setPosition(min(max(position, 0), editor.document().characterCount() - 1))
    return cursor_location(editor, cursor)


def cursor_at(editor: CodeEditor, location: SourceLocation) -> QTextCursor:
    """A cursor placed at ``location`` (clamped to the document)."""
    document = editor.document()
    block = document.findBlockByNumber(max(location.line - 1, 0))
    if not block.isValid():
        block = document.lastBlock()
    offset = utf16_offset(block.text(), location.column - 1)
    cursor = QTextCursor(block)
    cursor.setPosition(block.position() + min(offset, max(block.length() - 1, 0)))
    return cursor


def span_cursor(editor: CodeEditor, start: SourceLocation, end: SourceLocation) -> QTextCursor:
    cursor = cursor_at(editor, start)
    cursor.setPosition(cursor_at(editor, end).position(), QTextCursor.KeepAnchor)
    return cursor


def word_cursor(editor: CodeEditor, location: SourceLocation) -> QTextCursor:
    """The word at ``location``, or one character if there is no word there."""
    cursor = cursor_at(editor, location)
    cursor.select(QTextCursor.WordUnderCursor)
    if cursor.hasSelection():
        return cursor
    cursor = cursor_at(editor, location)
    if not cursor.atBlockEnd():
        cursor.movePosition(QTextCursor.Right, QTextCursor.KeepAnchor)
    else:
        cursor.movePosition(QTextCursor.Left, QTextCursor.KeepAnchor)
    return cursor
