import pytest
from PyQt5.QtCore import QPoint, Qt
from PyQt5.QtGui import QPalette

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.editor.line_number_area import BREAKPOINT_MARGIN
from vizcacha.ui.editor.themes import CONSOLE_THEMES, DARK, EDITOR_THEMES, SYNTAX_CATEGORIES
from vizcacha.ui.features.editor.editor_preferences import USE_CUSTOM_COLORS_KEY


def _color(widget, role) -> str:
    return widget.palette().color(role).name().upper()


def _apply(workbench, settings, **values) -> CodeEditor:
    for key, value in values.items():
        settings.set(key, value)
    workbench.events.settings_changed.emit()
    return workbench.editor.current_editor()


@pytest.mark.parametrize("name", list(EDITOR_THEMES))
def test_editor_theme_sets_palette_gutter_and_syntax(workbench, settings, name):
    editor = _apply(workbench, settings, **{SettingsKeys.EDITOR_THEME: name})
    theme = EDITOR_THEMES[name]

    assert _color(editor, QPalette.Base) == theme.background
    assert _color(editor, QPalette.Text) == theme.text
    assert editor.theme is theme
    for category in SYNTAX_CATEGORIES:
        text_format = editor.highlighter.format_for(category)
        assert text_format.foreground().color().name().upper() == theme.syntax[category].color


def test_highlighter_paints_document_with_theme_colours(qtbot):
    editor = CodeEditor()
    qtbot.addWidget(editor)
    editor.setPlainText('func main() { s := "hi" }')

    editor.apply_theme(DARK)

    ranges = editor.document().firstBlock().layout().formats()
    colours = {(r.start, r.length): r.format.foreground().color().name().upper() for r in ranges}
    assert colours[(0, 4)] == DARK.syntax["keyword"].color
    assert colours[(19, 4)] == DARK.syntax["string"].color


def test_new_editors_follow_the_current_theme(workbench, settings):
    _apply(workbench, settings, **{SettingsKeys.EDITOR_THEME: "Solarized Dark"})

    editor = workbench.editor.new_tab()

    assert _color(editor, QPalette.Base) == EDITOR_THEMES["Solarized Dark"].background


def test_custom_colours_override_the_theme_only_when_enabled(workbench, settings):
    values = {
        SettingsKeys.EDITOR_THEME: "Dark",
        SettingsKeys.BACKGROUND_COLOR: "#112233",
        SettingsKeys.TEXT_COLOR: "#AABBCC",
    }
    editor = _apply(workbench, settings, **values)
    assert _color(editor, QPalette.Base) == DARK.background

    editor = _apply(workbench, settings, **{USE_CUSTOM_COLORS_KEY: True})

    assert _color(editor, QPalette.Base) == "#112233"
    assert _color(editor, QPalette.Text) == "#AABBCC"
    assert editor.highlighter.format_for("keyword").foreground().color().name().upper() == (
        DARK.syntax["keyword"].color
    )


@pytest.mark.parametrize("name", list(CONSOLE_THEMES))
def test_console_theme_is_applied(workbench, settings, name):
    _apply(workbench, settings, **{SettingsKeys.CONSOLE_THEME: name})

    assert _color(workbench.console, QPalette.Base) == CONSOLE_THEMES[name].background
    assert _color(workbench.console, QPalette.Text) == CONSOLE_THEMES[name].text


def test_tab_size_and_auto_indent_settings_reach_the_editor(workbench, settings):
    values = {SettingsKeys.TAB_SIZE: 8, SettingsKeys.AUTO_INDENT: False}
    editor = _apply(workbench, settings, **values)

    assert editor.tab_size == 8
    assert editor.tabStopDistance() == editor.fontMetrics().horizontalAdvance(" ") * 8
    assert editor.auto_indent is False


def test_hidden_line_numbers_keep_breakpoints_clickable(qtbot, workbench, settings):
    editor = _apply(workbench, settings, **{SettingsKeys.SHOW_LINE_NUMBERS: False})
    editor.setPlainText("package main\n\nfunc main() {\n}\n")
    workbench.window.show()
    qtbot.waitExposed(workbench.window)

    assert editor.line_number_area.width() == BREAKPOINT_MARGIN
    y = editor.blockBoundingGeometry(editor.document().firstBlock()).height() / 2
    qtbot.mouseClick(editor.line_number_area, Qt.LeftButton, pos=QPoint(5, int(y)))

    assert editor.get_breakpoints() == [1]


def test_zoom_actions_change_the_font_size_of_every_editor(workbench, settings):
    editor = _apply(workbench, settings, **{SettingsKeys.FONT_SIZE: 12})
    view = {action.text(): action for action in workbench.menu("view").actions()}

    view["Zoom &In"].trigger()
    view["Zoom &In"].trigger()
    other = workbench.editor.new_tab()

    assert editor.font().pointSize() == 14
    assert other.font().pointSize() == 14
    view["&Reset Zoom"].trigger()
    assert editor.font().pointSize() == 12
