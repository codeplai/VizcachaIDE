"""Regex-based Go syntax highlighter; colours come from an EditorTheme."""

import re

from PyQt5.QtGui import QColor, QFont, QSyntaxHighlighter, QTextCharFormat

from vizcacha.ui.editor.themes import LIGHT, SYNTAX_CATEGORIES, EditorTheme, SyntaxStyle

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


def _format(style: SyntaxStyle) -> QTextCharFormat:
    text_format = QTextCharFormat()
    text_format.setForeground(QColor(style.color))
    if style.bold:
        text_format.setFontWeight(QFont.Bold)
    text_format.setFontItalic(style.italic)
    return text_format


def _words(words: tuple[str, ...]) -> re.Pattern:
    return re.compile(r"\b(?:" + "|".join(words) + r")\b")


# (category, pattern) in application order: later rules win over earlier ones.
RULES = (
    ("keyword", _words(KEYWORDS)),
    ("type", _words(TYPES)),
    ("builtin", _words(BUILTINS)),
    ("string", re.compile(r'"[^"\\]*(\\.[^"\\]*)*"')),
    ("string", re.compile(r"`[^`]*`")),
    ("number", re.compile(r"\b\d+\.?\d*\b")),
    ("comment", re.compile(r"//[^\n]*")),
)


class GoSyntaxHighlighter(QSyntaxHighlighter):
    def __init__(self, document=None, theme: EditorTheme = LIGHT) -> None:
        super().__init__(document)
        self.comment_start = re.compile(r"/\*")
        self.comment_end = re.compile(r"\*/")
        self.formats: dict[str, QTextCharFormat] = {}
        self.theme = theme
        self.set_theme(theme)

    def set_theme(self, theme: EditorTheme) -> None:
        self.theme = theme
        fallback = SyntaxStyle(theme.text)
        self.formats = {
            category: _format(theme.syntax.get(category, fallback))
            for category in SYNTAX_CATEGORIES
        }
        self.rehighlight()

    def format_for(self, category: str) -> QTextCharFormat:
        return self.formats[category]

    def highlightBlock(self, text: str) -> None:  # noqa: N802 - Qt override
        for category, pattern in RULES:
            text_format = self.formats[category]
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
                self.setFormat(start, len(text) - start, self.formats["comment"])
                return
            self.setFormat(start, end_match.end() - start, self.formats["comment"])
            next_match = self.comment_start.search(text, end_match.end())
            start = next_match.start() if next_match else -1
