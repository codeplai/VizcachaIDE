"""Applies the stored editor / appearance settings to editors and to the console."""

from PyQt5.QtGui import QColor, QFont, QPalette
from PyQt5.QtWidgets import QPlainTextEdit, QWidget

from vizcacha.application.ports import SettingsRepository
from vizcacha.application.settings_keys import SettingsKeys as Keys
from vizcacha.ui.editor.code_editor import CodeEditor
from vizcacha.ui.editor.themes import console_theme, editor_theme

# New keys owned by track D (Contract change request: move to SettingsKeys).
FORMAT_ON_SAVE_KEY = "editor/format_on_save"
USE_CUSTOM_COLORS_KEY = "appearance/use_custom_colors"

DEFAULTS = {
    Keys.FONT_FAMILY: "Consolas",
    Keys.FONT_SIZE: 11,
    Keys.TAB_SIZE: 4,
    Keys.AUTO_INDENT: True,
    Keys.SHOW_LINE_NUMBERS: True,
    Keys.WORD_WRAP: False,
    Keys.EDITOR_THEME: "Light",
    Keys.CONSOLE_THEME: "Dark",
    Keys.BACKGROUND_COLOR: "",
    Keys.TEXT_COLOR: "",
    FORMAT_ON_SAVE_KEY: True,
    USE_CUSTOM_COLORS_KEY: False,
}


def setting(settings: SettingsRepository, key: str):
    return settings.get(key, DEFAULTS[key])


def custom_colors(settings: SettingsRepository) -> tuple[str, str]:
    """(background, text) overrides, or empty strings to use the theme colours."""
    if not setting(settings, USE_CUSTOM_COLORS_KEY):
        return "", ""
    return _valid(setting(settings, Keys.BACKGROUND_COLOR)), _valid(
        setting(settings, Keys.TEXT_COLOR)
    )


def apply_editor_preferences(editor: CodeEditor, settings: SettingsRepository) -> None:
    editor.set_base_font(
        QFont(setting(settings, Keys.FONT_FAMILY), setting(settings, Keys.FONT_SIZE))
    )
    wrap = setting(settings, Keys.WORD_WRAP)
    editor.setLineWrapMode(QPlainTextEdit.WidgetWidth if wrap else QPlainTextEdit.NoWrap)
    editor.set_tab_size(setting(settings, Keys.TAB_SIZE))
    editor.auto_indent = setting(settings, Keys.AUTO_INDENT)
    editor.set_line_numbers_visible(setting(settings, Keys.SHOW_LINE_NUMBERS))
    background, text = custom_colors(settings)
    editor.apply_theme(editor_theme(setting(settings, Keys.EDITOR_THEME)), background, text)


def apply_console_theme(console: QWidget, settings: SettingsRepository) -> None:
    """Colours the console from outside (the widget belongs to another track)."""
    theme = console_theme(setting(settings, Keys.CONSOLE_THEME))
    palette = console.palette()
    palette.setColor(QPalette.Base, QColor(theme.background))
    palette.setColor(QPalette.Text, QColor(theme.text))
    console.setPalette(palette)


def _valid(color: str) -> str:
    return color if color and QColor(color).isValid() else ""
