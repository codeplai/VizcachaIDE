from pathlib import Path

from PyQt5.QtWidgets import QLabel

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.domain.debugging import DebugState, StackFrame, StopReason, Variable
from vizcacha.domain.diagnostics import SourceLocation


def _menu_texts(workbench, menu_id):
    return [action.text() for action in workbench.menu(menu_id).actions() if action.text()]


def test_features_register_menus_toolbar_and_panels(workbench):
    assert _menu_texts(workbench, "file")[0] == "&New"
    assert "&Undo" in _menu_texts(workbench, "edit")
    assert "▶ Run" in _menu_texts(workbench, "run")
    assert "Toggle &Breakpoint" in _menu_texts(workbench, "debug")
    assert {"Variables", "Call Stack"} <= set(_menu_texts(workbench, "view"))
    assert "&Options..." in _menu_texts(workbench, "tools")
    assert workbench.editor.count() == 1


def test_window_title_follows_active_file(workbench, tmp_path: Path):
    source = tmp_path / "hola.go"
    source.write_text("package main\n", encoding="utf-8")

    workbench.editor.open_file(source)

    assert workbench.window.windowTitle() == "VizcachaIDE - hola.go"
    assert workbench.editor.count() == 1  # the pristine Untitled tab was reused
    assert workbench.services.settings.get(SettingsKeys.LAST_FILE, "") == str(source)


def test_navigate_to_moves_cursor(workbench, tmp_path: Path):
    source = tmp_path / "lines.go"
    source.write_text("package main\n\nfunc main() {\n}\n", encoding="utf-8")

    workbench.events.navigate_to.emit(SourceLocation(source, 3, 6))

    cursor = workbench.editor.current_editor().textCursor()
    assert (cursor.blockNumber(), cursor.positionInBlock()) == (2, 5)


def test_debugger_panels_render_debug_state(workbench, tmp_path: Path):
    source = tmp_path / "main.go"
    source.write_text("package main\n\nfunc main() {\n\tx := 1\n}\n", encoding="utf-8")
    workbench.editor.open_file(source)
    state = DebugState(
        StopReason.STEP,
        frames=(StackFrame(1, "main.main", SourceLocation(source, 4)),),
        variables=(Variable("x", "int", "1"),),
    )

    workbench.services.debugger.stopped.emit(state)

    editor = workbench.editor.current_editor()
    assert editor.current_line == 4
    variables = workbench.window.findChild(object, "variables").widget()
    assert variables.topLevelItem(0).text(2) == "1"


def test_null_debugger_explains_it_is_not_available(workbench, tmp_path: Path):
    source = tmp_path / "main.go"
    source.write_text("package main\n", encoding="utf-8")
    workbench.editor.open_file(source)

    workbench.menu("debug").actions()[0].trigger()

    assert "not available yet" in workbench.console.toPlainText()
    assert workbench.menu("debug").actions()[0].isEnabled()


def test_status_bar_and_toolbar_follow_settings(workbench, settings):
    settings.set(SettingsKeys.SHOW_TOOLBAR, False)
    settings.set(SettingsKeys.SHOW_STATUS_BAR, True)

    workbench.events.settings_changed.emit()

    assert workbench.toolbar.isHidden()
    assert not workbench.window.statusBar().isHidden()
    workbench.add_status_widget(QLabel("ok"))


def test_settings_pages_round_trip(workbench, settings):
    from vizcacha.ui.settings_dialog import SettingsDialog

    settings.set(SettingsKeys.FONT_SIZE, 14)
    settings.set(SettingsKeys.EDITOR_THEME, "Dark")
    dialog = SettingsDialog(workbench)

    dialog.apply()

    assert [page.title for page in dialog.pages] == ["Editor", "Appearance", "Environment"]
    assert settings.get(SettingsKeys.FONT_SIZE, 0) == 14
    assert settings.get(SettingsKeys.EDITOR_THEME, "") == "Dark"
    assert workbench.editor.current_editor().font().pointSize() == 14
