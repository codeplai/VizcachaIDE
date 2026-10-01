"""Selection layers painted by the language feature: diagnostics and occurrences."""

from PyQt5.QtGui import QColor, QTextCharFormat
from PyQt5.QtWidgets import QTextEdit

from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.infrastructure.gopls_lsp import TextRange
from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.features.language.locations import span_cursor, word_cursor

DIAGNOSTICS_LAYER = "diagnostics"
HIGHLIGHTS_LAYER = "highlights"
ERROR_COLOR = QColor("#E51400")
WARNING_COLOR = QColor("#D7A100")
OCCURRENCE_COLOR = QColor("#DCE6F7")


def _diagnostic_selection(editor: CodeEditor, diagnostic: Diagnostic) -> QTextEdit.ExtraSelection:
    selection = QTextEdit.ExtraSelection()
    selection.cursor = word_cursor(editor, diagnostic.location)
    text_format = QTextCharFormat()
    text_format.setUnderlineStyle(QTextCharFormat.WaveUnderline)
    color = ERROR_COLOR if diagnostic.severity == Severity.ERROR else WARNING_COLOR
    text_format.setUnderlineColor(color)
    selection.format = text_format
    return selection


def show_diagnostics(editor: CodeEditor, diagnostics: list[Diagnostic]) -> None:
    selections = [
        _diagnostic_selection(editor, diagnostic)
        for diagnostic in diagnostics
        if diagnostic.location is not None
    ]
    editor.set_selection_layer(DIAGNOSTICS_LAYER, selections)


def show_occurrences(editor: CodeEditor, ranges: list[TextRange]) -> None:
    selections = []
    for text_range in ranges:
        selection = QTextEdit.ExtraSelection()
        selection.cursor = span_cursor(editor, text_range.start, text_range.end)
        selection.format.setBackground(OCCURRENCE_COLOR)
        selections.append(selection)
    editor.set_selection_layer(HIGHLIGHTS_LAYER, selections)


def diagnostics_on_line(diagnostics: list[Diagnostic], line: int) -> list[Diagnostic]:
    return [d for d in diagnostics if d.location is not None and d.location.line == line]
