import pytest
from PyQt5.QtCore import Qt
from PyQt5.QtGui import QTextCursor

from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.editor.indentation import new_line_indent


@pytest.fixture
def editor(qtbot):
    widget = CodeEditor()
    qtbot.addWidget(widget)
    return widget


def _type_at_end(qtbot, editor: CodeEditor, text: str, key=Qt.Key_Return) -> str:
    editor.setPlainText(text)
    editor.moveCursor(QTextCursor.End)
    qtbot.keyClick(editor, key)
    return editor.toPlainText()


@pytest.mark.parametrize(
    ("line", "expected"),
    [
        ("func main() {", "\t"),
        ("\tif x > 0 {", "\t\t"),
        ("import (", "\t"),
        ("\tx := 1", "\t"),
        ("\tcase 1:", "\t"),  # no extra level after ":" (that was Python)
        ("", ""),
    ],
)
def test_new_line_indent(line, expected):
    assert new_line_indent(line) == expected


def test_enter_after_brace_inserts_a_real_tab(qtbot, editor):
    assert _type_at_end(qtbot, editor, "func main() {") == "func main() {\n\t"


def test_enter_keeps_nested_tab_indentation(qtbot, editor):
    text = "func main() {\n\tif ok {"
    assert _type_at_end(qtbot, editor, text) == text + "\n\t\t"


def test_enter_between_braces_puts_closing_brace_on_its_own_line(qtbot, editor):
    editor.setPlainText("func main() {}")
    cursor = editor.textCursor()
    cursor.setPosition(len("func main() {"))
    editor.setTextCursor(cursor)

    qtbot.keyClick(editor, Qt.Key_Return)

    assert editor.toPlainText() == "func main() {\n\t\n}"
    assert editor.textCursor().blockNumber() == 1


def test_auto_indent_disabled_inserts_a_plain_newline(qtbot, editor):
    editor.auto_indent = False
    assert _type_at_end(qtbot, editor, "\tfunc main() {") == "\tfunc main() {\n"


def test_closing_brace_removes_one_level(qtbot, editor):
    editor.setPlainText("func main() {\n\t\t")
    editor.moveCursor(QTextCursor.End)

    qtbot.keyClicks(editor, "}")

    assert editor.toPlainText() == "func main() {\n\t}"


def test_tab_and_shift_tab_indent_and_outdent_the_selection(qtbot, editor):
    editor.setPlainText("a := 1\n\nb := 2\nc := 3")
    cursor = editor.textCursor()
    cursor.setPosition(0)
    cursor.setPosition(len("a := 1\n\nb"), QTextCursor.KeepAnchor)
    editor.setTextCursor(cursor)

    qtbot.keyClick(editor, Qt.Key_Tab)
    assert editor.toPlainText() == "\ta := 1\n\n\tb := 2\nc := 3"  # blank lines stay empty

    qtbot.keyClick(editor, Qt.Key_Backtab, Qt.ShiftModifier)
    assert editor.toPlainText() == "a := 1\n\nb := 2\nc := 3"


def test_shift_tab_removes_spaces_up_to_tab_size(qtbot, editor):
    editor.set_tab_size(4)
    editor.setPlainText("      x := 1")

    qtbot.keyClick(editor, Qt.Key_Backtab, Qt.ShiftModifier)

    assert editor.toPlainText() == "  x := 1"


def test_tab_without_selection_inserts_a_tab(qtbot, editor):
    assert _type_at_end(qtbot, editor, "", key=Qt.Key_Tab) == "\t"
