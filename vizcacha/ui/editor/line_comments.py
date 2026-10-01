"""Comment / uncomment the selected lines with Go line comments (``//``)."""

from PyQt5.QtGui import QTextBlock, QTextCursor
from PyQt5.QtWidgets import QPlainTextEdit

from vizcacha.ui.editor.block_selection import select_whole_blocks, selected_blocks
from vizcacha.ui.editor.indentation import leading_whitespace

COMMENT_MARK = "//"


def toggle_line_comment(editor: QPlainTextEdit) -> None:
    """Uncomment if every non-blank selected line is commented, otherwise comment them."""
    blocks = selected_blocks(editor)
    code_blocks = [block for block in blocks if block.text().strip()]
    if not code_blocks:
        return
    had_selection = editor.textCursor().hasSelection()
    uncomment = all(_is_commented(block.text()) for block in code_blocks)
    column = min(len(leading_whitespace(block.text())) for block in code_blocks)
    cursor = QTextCursor(editor.document())
    cursor.beginEditBlock()
    for block in code_blocks:
        if uncomment:
            _remove_comment(cursor, block)
        else:
            cursor.setPosition(block.position() + column)
            cursor.insertText(COMMENT_MARK + " ")
    cursor.endEditBlock()
    if had_selection:
        select_whole_blocks(editor, blocks)


def _is_commented(line: str) -> bool:
    return line.lstrip().startswith(COMMENT_MARK)


def _remove_comment(cursor: QTextCursor, block: QTextBlock) -> None:
    text = block.text()
    start = text.index(COMMENT_MARK)
    end = start + len(COMMENT_MARK)
    if text[end : end + 1] == " ":
        end += 1
    cursor.setPosition(block.position() + start)
    cursor.setPosition(block.position() + end, QTextCursor.KeepAnchor)
    cursor.removeSelectedText()
