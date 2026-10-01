"""Diagnostics underline their exact range; occurrences are painted from the server."""

from PyQt5.QtGui import QTextCharFormat

from vizcacha.domain.code_structure import SourceRange
from vizcacha.domain.diagnostics import Diagnostic, Severity, SourceLocation
from vizcacha.ui.features.language.editor_layers import OCCURRENCE_COLOR

PARAGRAPH_SEPARATOR = " "  # how QTextCursor.selectedText() returns line breaks


def underlined_texts(editor) -> list[str]:
    return [
        selection.cursor.selectedText()
        for selection in editor.extraSelections()
        if selection.format.underlineStyle() == QTextCharFormat.WaveUnderline
    ]


def highlighted_texts(editor) -> list[str]:
    return [
        selection.cursor.selectedText()
        for selection in editor.extraSelections()
        if selection.format.background().color() == OCCURRENCE_COLOR
    ]


def publish(qtbot, workbench, fake, source, diagnostics) -> None:
    with qtbot.waitSignal(workbench.events.diagnostics_changed):
        fake.diagnostics_published.emit(source, diagnostics)


def error_at(source, start: tuple[int, int], end: tuple[int, int] | None) -> Diagnostic:
    end_location = None if end is None else SourceLocation(source, *end)
    start_location = SourceLocation(source, *start)
    return Diagnostic(start_location, Severity.ERROR, "m", "m", "gopls", end=end_location)


def test_diagnostic_with_end_underlines_the_exact_range(qtbot, opened, fake, source):
    workbench, editor = opened

    publish(qtbot, workbench, fake, source, [error_at(source, (6, 5), (6, 11))])

    assert underlined_texts(editor) == ["fmt.Pr"]


def test_multiline_range_is_underlined_across_lines(qtbot, opened, fake, source):
    workbench, editor = opened

    publish(qtbot, workbench, fake, source, [error_at(source, (5, 13), (6, 8))])

    assert underlined_texts(editor) == ["{" + PARAGRAPH_SEPARATOR + "    fmt"]


def test_without_end_or_with_empty_range_the_word_is_underlined(qtbot, opened, fake, source):
    workbench, editor = opened
    diagnostics = [error_at(source, (6, 5), None), error_at(source, (6, 9), (6, 9))]

    publish(qtbot, workbench, fake, source, diagnostics)

    assert underlined_texts(editor) == ["fmt", "Pr"]


def test_occurrences_come_from_document_highlights(qtbot, opened, fake, source):
    workbench, editor = opened
    publish(qtbot, workbench, fake, source, [])
    fake.highlights = [SourceRange(SourceLocation(source, 5, 6), SourceLocation(source, 5, 10))]
    cursor = editor.textCursor()
    cursor.setPosition(editor.toPlainText().index("main()") + 1)

    editor.setTextCursor(cursor)

    qtbot.waitUntil(lambda: highlighted_texts(editor) == ["main"], timeout=2000)
    assert ("document_highlights", SourceLocation(source, 5, 7)) in fake.calls
