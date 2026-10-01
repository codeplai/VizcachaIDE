import pytest
from PyQt5.QtGui import QTextCursor
from PyQt5.QtWidgets import QInputDialog

from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.editor.bracket_matcher import bracket_pair_at, find_matching_bracket
from vizcacha.ui.editor.line_comments import toggle_line_comment


@pytest.fixture
def editor(qtbot):
    widget = CodeEditor()
    qtbot.addWidget(widget)
    return widget


def _select(editor: CodeEditor, start: int, end: int) -> None:
    cursor = editor.textCursor()
    cursor.setPosition(start)
    cursor.setPosition(end, QTextCursor.KeepAnchor)
    editor.setTextCursor(cursor)


def test_toggle_comment_on_selection_round_trips(editor):
    source = "func main() {\n\tx := 1\n\n\t\ty := 2\n}"
    editor.setPlainText(source)
    _select(editor, source.index("\tx"), source.index("y := 2"))

    toggle_line_comment(editor)
    assert editor.toPlainText() == "func main() {\n\t// x := 1\n\n\t// \ty := 2\n}"

    toggle_line_comment(editor)
    assert editor.toPlainText() == source


def test_toggle_comment_on_current_line_without_selection(editor):
    editor.setPlainText("// fmt.Println(1)")

    toggle_line_comment(editor)

    assert editor.toPlainText() == "fmt.Println(1)"


def test_mixed_lines_get_commented(editor):
    editor.setPlainText("// a\nb")
    _select(editor, 0, len("// a\nb"))

    toggle_line_comment(editor)

    assert editor.toPlainText() == "// // a\n// b"


def test_toggle_comment_action_uses_ctrl_slash(workbench):
    editor = workbench.editor.current_editor()
    editor.setPlainText("x := 1")
    action = {a.text(): a for a in workbench.menu("edit").actions()}["Toggle &Comment"]

    action.trigger()

    assert action.shortcut().toString() == "Ctrl+/"
    assert editor.toPlainText() == "// x := 1"


@pytest.mark.parametrize(
    ("text", "index", "expected"),
    [
        ("f(a[1], {b})", 1, 11),
        ("f(a[1], {b})", 11, 1),
        ("f(a[1], {b})", 3, 5),
        ("{ { } }", 0, 6),
        ("(()", 0, None),
        ("abc", 1, None),
    ],
)
def test_find_matching_bracket(text, index, expected):
    assert find_matching_bracket(text, index) == expected


def test_bracket_pair_prefers_the_bracket_before_the_cursor():
    assert bracket_pair_at("f()", 2) == (1, 2)
    assert bracket_pair_at("f()", 1) == (1, 2)
    assert bracket_pair_at("abc", 1) is None


def test_bracket_layer_highlights_both_brackets(editor):
    editor.setPlainText("func main() {\n}")
    cursor = editor.textCursor()
    cursor.setPosition(len("func main() {"))
    editor.setTextCursor(cursor)

    positions = sorted(s.cursor.selectionStart() for s in editor.extraSelections())

    assert positions == [len("func main() "), len("func main() {\n")]
    editor.moveCursor(QTextCursor.Start)
    assert editor.extraSelections() == []


def test_go_to_line(workbench, monkeypatch):
    editor = workbench.editor.current_editor()
    editor.setPlainText("a\nb\nc\nd")
    monkeypatch.setattr(QInputDialog, "getInt", lambda *args: (3, True))
    action = {a.text(): a for a in workbench.menu("edit").actions()}["&Go to Line..."]

    action.trigger()

    assert editor.textCursor().blockNumber() == 2
