"""Regex-based Go syntax highlighter."""

import re

from PyQt5.QtGui import QColor, QFont, QSyntaxHighlighter, QTextCharFormat

KEYWORDS = (
    "break",
    "case",
    "chan",
    "const",
    "continue",
    "default",
    "defer",
    "else",
    "fallthrough",
    "for",
    "func",
    "go",
    "goto",
    "if",
    "import",
    "interface",
    "map",
    "package",
    "range",
    "return",
    "select",
    "struct",
    "switch",
    "type",
    "var",
)
TYPES = (
    "bool",
    "byte",
    "complex64",
    "complex128",
    "error",
    "float32",
    "float64",
    "int",
    "int8",
    "int16",
    "int32",
    "int64",
    "rune",
    "string",
    "uint",
    "uint8",
    "uint16",
    "uint32",
    "uint64",
    "uintptr",
)
BUILTINS = (
    "append",
    "cap",
    "close",
    "complex",
    "copy",
    "delete",
    "imag",
    "len",
    "make",
    "new",
    "panic",
    "print",
    "println",
    "real",
    "recover",
)

IN_BLOCK_COMMENT = 1


def _format(color: str, bold: bool = False, italic: bool = False) -> QTextCharFormat:
    text_format = QTextCharFormat()
    text_format.setForeground(QColor(color))
    if bold:
        text_format.setFontWeight(QFont.Bold)
    text_format.setFontItalic(italic)
    return text_format


def _words(words: tuple[str, ...]) -> re.Pattern:
    return re.compile(r"\b(?:" + "|".join(words) + r")\b")


class GoSyntaxHighlighter(QSyntaxHighlighter):
    def __init__(self, document=None) -> None:
        super().__init__(document)
        self.comment_format = _format("#808080", italic=True)
        self.rules = [
            (_words(KEYWORDS), _format("#0000FF", bold=True)),
            (_words(TYPES), _format("#008080", bold=True)),
            (_words(BUILTINS), _format("#800080")),
            (re.compile(r'"[^"\\]*(\\.[^"\\]*)*"'), _format("#008000")),
            (re.compile(r"`[^`]*`"), _format("#008000")),
            (re.compile(r"\b\d+\.?\d*\b"), _format("#FF6600")),
            (re.compile(r"//[^\n]*"), self.comment_format),
        ]
        self.comment_start = re.compile(r"/\*")
        self.comment_end = re.compile(r"\*/")

    def highlightBlock(self, text: str) -> None:  # noqa: N802 - Qt override
        for pattern, text_format in self.rules:
            for match in pattern.finditer(text):
                self.setFormat(match.start(), match.end() - match.start(), text_format)
        self._highlight_block_comments(text)

    def _highlight_block_comments(self, text: str) -> None:
        self.setCurrentBlockState(0)
        start = 0
        if self.previousBlockState() != IN_BLOCK_COMMENT:
            match = self.comment_start.search(text)
            start = match.start() if match else -1
        while start >= 0:
            end_match = self.comment_end.search(text, start)
            if not end_match:
                self.setCurrentBlockState(IN_BLOCK_COMMENT)
                self.setFormat(start, len(text) - start, self.comment_format)
                return
            self.setFormat(start, end_match.end() - start, self.comment_format)
            next_match = self.comment_start.search(text, end_match.end())
            start = next_match.start() if next_match else -1
