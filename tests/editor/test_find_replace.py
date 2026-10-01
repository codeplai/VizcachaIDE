import pytest
from PyQt5.QtGui import QTextCursor

from vizcacha.ui.editor.find_replace_bar import FindReplaceBar
from vizcacha.ui.editor.text_search import (
    SearchQuery,
    find_all,
    find_next,
    replace_all,
)

SOURCE = "Count := 1\ncount++\naccount := count\n"


@pytest.fixture
def editor(workbench):
    current = workbench.editor.current_editor()
    current.setPlainText(SOURCE)
    current.moveCursor(QTextCursor.Start)
    return current


def _edit_actions(workbench) -> dict:
    return {action.text(): action for action in workbench.menu("edit").actions()}


@pytest.mark.parametrize(
    ("query", "count"),
    [
        (SearchQuery("count"), 4),
        (SearchQuery("count", case_sensitive=True), 3),
        (SearchQuery("count", whole_word=True), 3),
        (SearchQuery("count", case_sensitive=True, whole_word=True), 2),
        (SearchQuery(""), 0),
    ],
)
def test_find_all_respects_options(editor, query, count):
    assert len(find_all(editor.document(), query)) == count


def test_find_next_wraps_around(editor):
    query = SearchQuery("account")
    assert find_next(editor, query)
    first = editor.textCursor().selectionStart()

    assert find_next(editor, query)

    assert editor.textCursor().selectionStart() == first
    assert editor.textCursor().selectedText() == "account"


def test_replace_all_is_one_undo_step(editor):
    assert replace_all(editor, SearchQuery("count", whole_word=True), "total") == 3
    assert editor.toPlainText() == "total := 1\ntotal++\naccount := total\n"

    editor.undo()

    assert editor.toPlainText() == SOURCE


def test_find_bar_highlights_matches_in_search_layer(workbench, editor):
    _edit_actions(workbench)["&Find..."].trigger()
    bar = _find_bar(workbench)

    bar.find_field.setText("count")

    assert len(editor.extraSelections()) == 4
    assert bar.status.text() == "4 matches"
    assert editor.textCursor().selectedText() == "Count"

    bar.close_bar()
    assert editor.extraSelections() == []


def test_find_bar_replace_and_options(workbench, editor):
    _edit_actions(workbench)["&Replace..."].trigger()
    bar = _find_bar(workbench)
    bar.match_case.setChecked(True)
    bar.whole_word.setChecked(True)
    bar.find_field.setText("count")
    bar.replace_field.setText("n")

    bar.replace()

    assert editor.toPlainText() == "Count := 1\nn++\naccount := count\n"
    assert bar.replace_all() == 1
    assert editor.toPlainText() == "Count := 1\nn++\naccount := n\n"


def test_find_next_and_previous_shortcuts(workbench, editor):
    actions = _edit_actions(workbench)
    assert actions["&Find..."].shortcut().toString() == "Ctrl+F"
    assert actions["&Replace..."].shortcut().toString() == "Ctrl+H"
    assert actions["Find &Next"].shortcut().toString() == "F3"
    assert actions["Find &Previous"].shortcut().toString() == "Shift+F3"
    assert actions["&Go to Line..."].shortcut().toString() == "Ctrl+G"
    _find_bar(workbench).find_field.setText("count")
    start = editor.textCursor().selectionStart()

    actions["Find &Next"].trigger()
    assert editor.textCursor().selectionStart() > start
    actions["Find &Previous"].trigger()
    assert editor.textCursor().selectionStart() == start


def _find_bar(workbench) -> FindReplaceBar:
    return workbench.editor.findChild(FindReplaceBar)
