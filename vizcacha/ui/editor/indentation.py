"""Idiomatic Go indentation: real TAB characters (gofmt style).

The visual width of a TAB is the ``tab_size`` setting; the file always stores
``\\t``. A new line keeps the indentation of the previous one and adds one level
after ``{`` or ``(`` (blocks, ``import (``, ``var (``...).
"""

from PyQt5.QtGui import QTextCursor
from PyQt5.QtWidgets import QPlainTextEdit

from vizcacha.ui.editor.block_selection import select_whole_blocks, selected_blocks

INDENT_UNIT = "\t"
BLOCK_OPENERS = ("{", "(")
BLOCK_CLOSERS = ("}", ")")


def leading_whitespace(line: str) -> str:
    return line[: len(line) - len(line.lstrip(" \t"))]


def new_line_indent(text_before_cursor: str) -> str:
    """Indentation for the line created by pressing Enter after ``text_before_cursor``."""
    indent = leading_whitespace(text_before_cursor)
    if text_before_cursor.rstrip().endswith(BLOCK_OPENERS):
        return indent + INDENT_UNIT
    return indent


def insert_newline(editor: QPlainTextEdit, auto_indent: bool) -> None:
    cursor = editor.textCursor()
    if not auto_indent:
        cursor.insertText("\n")
        editor.setTextCursor(cursor)
        return
    text = cursor.block().text()
    before = text[: cursor.positionInBlock()]
    after = text[cursor.positionInBlock() :]
    indent = new_line_indent(before)
    cursor.beginEditBlock()
    cursor.insertText("\n" + indent)
    opened = indent != leading_whitespace(before)
    if opened and after.lstrip().startswith(BLOCK_CLOSERS):
        # Enter between "{}" leaves the closing brace on its own line.
        inner_position = cursor.position()
        cursor.insertText("\n" + leading_whitespace(before))
        cursor.setPosition(inner_position)
    cursor.endEditBlock()
    editor.setTextCursor(cursor)


def outdent_before_closer(editor: QPlainTextEdit) -> None:
    """Typing ``}`` on a blank, indented line removes one indentation level first."""
    cursor = editor.textCursor()
    before = cursor.block().text()[: cursor.positionInBlock()]
    if cursor.hasSelection() or before.strip() or not before.endswith(INDENT_UNIT):
        return
    cursor.deletePreviousChar()
    editor.setTextCursor(cursor)


def indent_lines(editor: QPlainTextEdit) -> None:
    blocks = selected_blocks(editor)
    cursor = QTextCursor(editor.document())
    cursor.beginEditBlock()
    for block in blocks:
        if not block.text().strip():
            continue
        cursor.setPosition(block.position())
        cursor.insertText(INDENT_UNIT)
    cursor.endEditBlock()
    select_whole_blocks(editor, blocks)


def outdent_lines(editor: QPlainTextEdit, tab_size: int) -> None:
    blocks = selected_blocks(editor)
    had_selection = editor.textCursor().hasSelection()
    cursor = QTextCursor(editor.document())
    cursor.beginEditBlock()
    for block in blocks:
        removable = _outdent_width(block.text(), tab_size)
        if not removable:
            continue
        cursor.setPosition(block.position())
        cursor.setPosition(block.position() + removable, QTextCursor.KeepAnchor)
        cursor.removeSelectedText()
    cursor.endEditBlock()
    if had_selection:
        select_whole_blocks(editor, blocks)


def _outdent_width(line: str, tab_size: int) -> int:
    """Characters to remove for one level: a TAB, or up to ``tab_size`` spaces."""
    if line.startswith(INDENT_UNIT):
        return 1
    spaces = len(line) - len(line.lstrip(" "))
    return min(spaces, tab_size)
