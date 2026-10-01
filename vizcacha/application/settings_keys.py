"""Names of every persisted setting.

The ``env/``, ``editor/``, ``appearance/`` and ``last_file`` keys come from version 0.1.
Keep them unchanged so that existing users keep their configuration.
"""


class SettingsKeys:
    LANGUAGE = "general/language"

    GO_PATH = "env/go_path"
    GOPATH = "env/gopath"
    GOROOT = "env/goroot"
    DELVE_PATH = "env/delve_path"
    GOPLS_PATH = "env/gopls_path"
    EXTRA_VARS = "env/extra_vars"

    FONT_FAMILY = "editor/font_family"
    FONT_SIZE = "editor/font_size"
    TAB_SIZE = "editor/tab_size"
    AUTO_INDENT = "editor/auto_indent"
    SHOW_LINE_NUMBERS = "editor/show_line_numbers"
    WORD_WRAP = "editor/word_wrap"
    FORMAT_ON_SAVE = "editor/format_on_save"
    BACKGROUND_COLOR = "editor/background_color"
    TEXT_COLOR = "editor/text_color"

    EDITOR_THEME = "appearance/editor_theme"
    CONSOLE_THEME = "appearance/console_theme"
    SHOW_TOOLBAR = "appearance/show_toolbar"
    SHOW_STATUS_BAR = "appearance/show_status_bar"
    USE_CUSTOM_COLORS = "appearance/use_custom_colors"

    LAST_FILE = "last_file"
    LAST_FOLDER = "project/last_folder"
    RECENT_FILES = "files/recent"  # newline-separated paths

    WINDOW_GEOMETRY = "window/geometry"  # base64 of QMainWindow.saveGeometry()
    WINDOW_STATE = "window/state"  # base64 of QMainWindow.saveState(): docks and toolbars
