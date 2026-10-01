"""The language feature wired into the real Workbench, with a fake language server."""

from PyQt5.QtCore import QEvent, QObject, Qt
from PyQt5.QtGui import QHelpEvent, QTextCharFormat
from PyQt5.QtWidgets import QApplication, QToolTip

from vizcacha.domain.code_structure import DocumentSymbol, SymbolKind
from vizcacha.domain.completion import CompletionItem, CompletionKind
from vizcacha.domain.diagnostics import Diagnostic, Severity, SourceLocation

NOT_FOUND_TEXT = "gopls not found: using basic completion"


def publish(qtbot, workbench, fake, source, diagnostics):
    with qtbot.waitSignal(workbench.events.diagnostics_changed) as blocker:
        fake.diagnostics_published.emit(source, diagnostics)
    return blocker.args


def test_without_language_server_static_completion_and_notice(make_workbench):
    workbench = make_workbench(None)
    editor = workbench.editor.current_editor()

    labels = {item.label for item in editor.completion_provider("fmt.Pr", 6, None)}

    assert {"Print", "Printf", "Println"} <= labels
    assert workbench.window.statusBar().currentMessage() == NOT_FOUND_TEXT


def test_outline_panel_is_registered_at_startup(make_workbench, fake):
    workbench = make_workbench(fake)

    outline = workbench.window.findChild(QObject, "outline")

    assert outline is not None
    assert outline.widget().topLevelItemCount() == 0
    assert "Outline" in [action.text() for action in workbench.menu("view").actions()]


def test_documents_are_opened_and_changes_debounced(qtbot, opened, fake, source):
    _workbench, editor = opened

    editor.moveCursor(editor.textCursor().End)
    editor.insertPlainText("// a")
    editor.insertPlainText("b")

    qtbot.waitUntil(lambda: "change" in fake.names(), timeout=2000)
    text = source.read_text(encoding="utf-8")
    assert fake.calls[0] == ("open", source, text)
    changes = [call for call in fake.calls if call[0] == "change"]
    assert changes == [("change", source, text + "// ab", 2)]


def test_completion_uses_server_and_falls_back_to_static(opened, fake, source):
    _workbench, editor = opened
    text = editor.toPlainText()
    position = text.index("fmt.Pr") + len("fmt.Pr")

    fallback = {item.label for item in editor.completion_provider(text, position, source)}
    fake.completions = [CompletionItem("Println", CompletionKind.FUNCTION, "from gopls")]
    served = editor.completion_provider(text, position, source)

    assert "Printf" in fallback
    assert [item.detail for item in served] == ["from gopls"]
    assert ("completion", SourceLocation(source, 6, 11)) in fake.calls


def test_diagnostics_are_underlined_and_republished(qtbot, opened, fake, source):
    workbench, editor = opened
    error = Diagnostic(SourceLocation(source, 6, 5), Severity.ERROR, "boom", "boom", "gopls")
    fake.symbols = [DocumentSymbol("main", SymbolKind.FUNCTION, SourceLocation(source, 5, 6))]

    args = publish(qtbot, workbench, fake, source, [error])

    assert args == [source, [error]]
    underlined = [
        s
        for s in editor.extraSelections()
        if s.format.underlineStyle() == QTextCharFormat.WaveUnderline
    ]
    assert [s.cursor.selectedText() for s in underlined] == ["fmt"]
    outline = workbench.window.findChild(QObject, "outline")
    assert outline is not None
    assert outline.widget().topLevelItem(0).text(0) == "main"


def test_outline_double_click_navigates(qtbot, opened, fake, source):
    workbench, _editor = opened
    fake.symbols = [DocumentSymbol("main", SymbolKind.FUNCTION, SourceLocation(source, 5, 6))]
    publish(qtbot, workbench, fake, source, [])
    tree = workbench.window.findChild(QObject, "outline").widget()

    with qtbot.waitSignal(workbench.events.navigate_to) as blocker:
        tree.itemDoubleClicked.emit(tree.topLevelItem(0), 0)

    assert blocker.args == [SourceLocation(source, 5, 6)]


def test_ctrl_click_goes_to_definition(qtbot, opened, fake, source):
    workbench, editor = opened
    fake.target = SourceLocation(source, 5, 6)
    cursor = editor.textCursor()
    cursor.setPosition(editor.toPlainText().index("main()"))
    point = editor.cursorRect(cursor).center()

    with qtbot.waitSignal(workbench.events.navigate_to) as blocker:
        qtbot.mouseClick(editor.viewport(), Qt.LeftButton, Qt.ControlModifier, point)

    assert blocker.args == [fake.target]
    assert fake.calls[-1][0] == "definition"


def test_hover_shows_diagnostics_and_server_text(qtbot, opened, fake, source):
    workbench, editor = opened
    workbench.window.show()
    error = Diagnostic(SourceLocation(source, 6, 5), Severity.ERROR, "boom", "boom", "gopls")
    publish(qtbot, workbench, fake, source, [error])
    cursor = editor.textCursor()
    cursor.setPosition(editor.toPlainText().index("fmt.Pr") + 1)
    point = editor.cursorRect(cursor).center()

    event = QHelpEvent(QEvent.ToolTip, point, editor.viewport().mapToGlobal(point))
    QApplication.sendEvent(editor.viewport(), event)

    assert "boom" in QToolTip.text()
    assert "func fake()" in QToolTip.text()


def test_open_paren_requests_call_tip(qtbot, opened, fake):
    _workbench, editor = opened

    qtbot.keyClicks(editor, "(")

    qtbot.waitUntil(lambda: "signature_help" in fake.names(), timeout=2000)


def test_unavailable_notice_is_shown_once(opened, fake):
    workbench, _editor = opened
    bar = workbench.window.statusBar()

    fake.server_unavailable.emit("not_found")
    first = bar.currentMessage()
    bar.clearMessage()
    fake.server_unavailable.emit("crashed")

    assert first == NOT_FOUND_TEXT
    assert bar.currentMessage() == ""


def test_closing_tab_and_window_close_document_and_shutdown(qtbot, opened, fake, source):
    workbench, _editor = opened
    workbench.editor.new_tab()

    workbench.editor.tabCloseRequested.emit(0)
    workbench.window.close()

    assert ("close", source) in fake.calls
    assert fake.calls[-1] == ("shutdown",)
