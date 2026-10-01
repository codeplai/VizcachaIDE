"""Lines (QTextBlocks) touched by the editor's selection."""

from PyQt5.QtGui import QTextBlock, QTextCursor
from PyQt5.QtWidgets import QPlainTextEdit


def selected_blocks(editor: QPlainTextEdit) -> list[QTextBlock]:
    """Blocks covered by the selection (the current block when nothing is selected).

    A selection that ends at column 0 does not include that last line, which is
    what users expect after selecting whole lines with Shift+Down.
    """
    cursor = editor.textCursor()
    document = editor.document()
    first = document.findBlock(cursor.selectionStart())
    last = document.findBlock(cursor.selectionEnd())
    ends_at_line_start = cursor.selectionEnd() == last.position()
    if cursor.hasSelection() and ends_at_line_start and last != first:
        last = last.previous()
    blocks = [first]
    while blocks[-1] != last and blocks[-1].next().isValid():
        blocks.append(blocks[-1].next())
    return blocks


def spans_several_lines(editor: QPlainTextEdit) -> bool:
    cursor = editor.textCursor()
    if not cursor.hasSelection():
        return False
    document = editor.document()
    return document.findBlock(cursor.selectionStart()) != document.findBlock(cursor.selectionEnd())


def select_whole_blocks(editor: QPlainTextEdit, blocks: list[QTextBlock]) -> None:
    """Select from the start of the first block to the end of the last one."""
    cursor = editor.textCursor()
    cursor.setPosition(blocks[0].position())
    last = blocks[-1]
    cursor.setPosition(last.position() + len(last.text()), QTextCursor.KeepAnchor)
    editor.setTextCursor(cursor)
