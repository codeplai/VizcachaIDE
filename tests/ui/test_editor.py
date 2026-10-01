from pathlib import Path

from PyQt5.QtCore import Qt
from PyQt5.QtWidgets import QTextEdit

from vizcacha.domain.completion import CompletionItem, CompletionKind
from vizcacha.ui.editor import CodeEditor, TabbedEditor


def test_breakpoints_toggle_and_notify(qtbot):
    editor = CodeEditor()
    qtbot.addWidget(editor)
    editor.setPlainText("a\nb\nc\n")

    with qtbot.waitSignal(editor.breakpoints_changed):
        editor.toggle_breakpoint_at_line(2)
    editor.toggle_breakpoint_at_line(3)
    editor.toggle_breakpoint_at_line(3)

    assert editor.get_breakpoints() == [2]


def test_selection_layers_are_merged_and_replaced(qtbot):
    editor = CodeEditor()
    qtbot.addWidget(editor)
    editor.setPlainText("package main\nfunc main() {}\n")

    editor.set_selection_layer("search", [QTextEdit.ExtraSelection()])
    editor.highlight_current_line(2)
    assert len(editor.extraSelections()) == 2

    editor.clear_current_line_highlight()
    assert len(editor.extraSelections()) == 1
    assert editor.current_line is None


def test_completion_provider_can_be_replaced(qtbot):
    editor = CodeEditor()
    qtbot.addWidget(editor)
    editor.setPlainText("fmt.Pr")
    editor.moveCursor(editor.textCursor().End)

    assert {"Print", "Printf", "Println"} <= {
        item.label for item in editor.completion_provider(editor.toPlainText(), 6, None)
    }
    editor.completion_provider = lambda text, pos, path: [
        CompletionItem("Println", CompletionKind.FUNCTION)
    ]
    editor.insert_completion(editor.completion_provider("", 0, None)[0])
    assert editor.toPlainText() == "fmt.Println"


def test_enter_keeps_indentation_and_indents_after_brace(qtbot):
    editor = CodeEditor()
    qtbot.addWidget(editor)
    editor.setPlainText("func main() {")
    editor.moveCursor(editor.textCursor().End)

    qtbot.keyClick(editor, Qt.Key_Return)

    assert editor.toPlainText() == "func main() {\n\t"  # Go indents with real TABs


def test_tabbed_editor_save_and_reopen(qtbot, tmp_path: Path):
    tabs = TabbedEditor()
    qtbot.addWidget(tabs)
    editor = tabs.new_tab()
    editor.setPlainText("package main\n")
    target = tmp_path / "saved.go"

    with qtbot.waitSignal(tabs.file_saved):
        assert tabs.save_to_file(editor, target)

    assert target.read_text(encoding="utf-8") == "package main\n"
    assert tabs.tabText(0) == "saved.go"
    assert tabs.open_file(target)
    assert tabs.count() == 1  # already open: switched instead of duplicating


def test_modified_tab_shows_marker(qtbot):
    tabs = TabbedEditor()
    qtbot.addWidget(tabs)
    editor = tabs.new_tab()

    editor.insertPlainText("x")

    assert tabs.tabText(0) == "Untitled*"
    editor.document().setModified(False)
