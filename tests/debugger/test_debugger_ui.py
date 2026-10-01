from pathlib import Path

from PyQt5.QtWidgets import QTabWidget

from vizcacha.application.ports import TERMINATED_BY_USER
from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.domain.debugging import DebugState, Goroutine, StackFrame, StopReason, Variable
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.ui.features.debugger.goroutines_view import GoroutinesView


def _debug_action(workbench, text: str):
    return next(a for a in workbench.menu("debug").actions() if a.text() == text)


def _open(workbench, tmp_path: Path, name: str = "main.go") -> Path:
    source = tmp_path / name
    source.write_text("package main\n\nfunc main() {\n\tx := 1\n\t_ = x\n}\n", encoding="utf-8")
    workbench.editor.open_file(source)
    return source


def test_goroutines_view_marks_the_current_one(qtbot):
    view = GoroutinesView()
    qtbot.addWidget(view)

    view.show_goroutines((Goroutine(1, "[Go 1] main.main"), Goroutine(2, "[Go 2] gopark")), 1)

    assert view.item(0).text().startswith("▶")
    assert not view.item(1).text().startswith("▶")


def test_debug_menu_and_panels(workbench):
    texts = [action.text() for action in workbench.menu("debug").actions()]
    tabs = workbench.window.findChild(object, "callstack").widget()

    assert {"⏩ Continue", "Run to &Cursor", "Stop Debugging", "Toggle &Breakpoint"} <= set(texts)
    assert isinstance(tabs, QTabWidget)
    assert [tabs.tabText(i) for i in range(tabs.count())] == ["Call Stack", "Goroutines"]
    assert not _debug_action(workbench, "Run to &Cursor").isEnabled()


def test_stopped_state_fills_goroutines_and_frame_click_navigates(workbench, tmp_path: Path):
    source = _open(workbench, tmp_path)
    frames = (StackFrame(1, "main.main", SourceLocation(source, 4)),)
    state = DebugState(StopReason.BREAKPOINT, frames, goroutines=(Goroutine(1, "[Go 1] main"),))
    targets = []
    workbench.events.navigate_to.connect(targets.append)

    workbench.services.debugger.stopped.emit(state)
    tabs = workbench.window.findChild(object, "callstack").widget()
    callstack, goroutines = tabs.widget(0), tabs.widget(1)
    callstack.itemClicked.emit(callstack.item(0))

    assert goroutines.count() == 1
    assert targets == [SourceLocation(source, 4), SourceLocation(source, 4)]


def test_goroutine_click_navigates_to_its_location(workbench, tmp_path: Path):
    source = _open(workbench, tmp_path)
    parked = Goroutine(2, "[Go 2] main.worker", SourceLocation(source, 5))
    unknown = Goroutine(3, "[Go 3] runtime.gopark")
    state = DebugState(StopReason.PAUSE, (), goroutines=(parked, unknown))
    targets = []
    workbench.events.navigate_to.connect(targets.append)

    workbench.services.debugger.stopped.emit(state)
    goroutines = workbench.window.findChild(object, "callstack").widget().widget(1)
    goroutines.itemClicked.emit(goroutines.item(0))
    goroutines.itemClicked.emit(goroutines.item(1))

    assert goroutines.item(0).text().endswith("main.go:5")
    assert targets == [SourceLocation(source, 5)]


def test_variables_panel_is_fed_by_variables_loaded(workbench, monkeypatch):
    debugger = workbench.services.debugger
    monkeypatch.setattr(debugger, "is_active", lambda: True)
    view = workbench.window.findChild(object, "variables").widget()
    view.request_children = lambda reference: debugger.variables_loaded.emit(
        reference, [Variable("x", "int", "1")]
    )
    view.show_variables([Variable("p", "main.Point", "{...}", reference=3)])

    view.topLevelItem(0).setExpanded(True)

    assert view.topLevelItem(0).child(0).text(0) == "x"


def test_user_stop_is_reported_as_stopped(workbench):
    workbench.services.debugger.terminated.emit(TERMINATED_BY_USER)

    assert "[Debugging stopped]" in workbench.console.toPlainText()


def test_missing_dlv_is_reported_in_the_console(workbench, settings, tmp_path: Path):
    settings.set(SettingsKeys.DELVE_PATH, str(tmp_path / "missing" / "dlv"))
    _open(workbench, tmp_path)

    workbench.menu("debug").actions()[0].trigger()

    console = workbench.console.toPlainText()
    assert "stdin" in console
    assert "go install github.com/go-delve/delve/cmd/dlv@latest" in console
    assert workbench.menu("debug").actions()[0].isEnabled()


def test_breakpoints_are_synchronised_during_a_session(workbench, tmp_path: Path, monkeypatch):
    debugger = workbench.services.debugger
    calls = []
    monkeypatch.setattr(debugger, "is_active", lambda: True)
    monkeypatch.setattr(debugger, "set_breakpoints", lambda f, lines: calls.append((f, lines)))
    source = _open(workbench, tmp_path)
    editor = workbench.editor.current_editor()

    editor.toggle_breakpoint_at_line(4)

    assert calls == [(source, [4])]


def test_run_to_cursor_uses_the_cursor_line(workbench, tmp_path: Path, monkeypatch):
    debugger = workbench.services.debugger
    targets = []
    monkeypatch.setattr(debugger, "run_to", targets.append)
    source = _open(workbench, tmp_path)
    editor = workbench.editor.current_editor()
    cursor = editor.textCursor()
    cursor.setPosition(editor.document().findBlockByNumber(4).position())
    editor.setTextCursor(cursor)

    _debug_action(workbench, "Run to &Cursor").setEnabled(True)
    _debug_action(workbench, "Run to &Cursor").trigger()

    assert targets == [SourceLocation(source, 5)]
