"""Edit menu, editor commands and every editor / appearance setting.

Settings are applied at start-up (to each new editor) and again on
``events.settings_changed``: theme, custom colours, font, zoom, tab size,
auto-indent, line numbers, word wrap and console theme. gofmt runs on save.
"""

from PyQt5.QtGui import QKeySequence
from PyQt5.QtWidgets import QAction

from vizcacha.i18n import N_, _
from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.features.editor.appearance_page import AppearancePage
from vizcacha.ui.features.editor.edit_commands import EditCommands
from vizcacha.ui.features.editor.editor_page import EditorPage
from vizcacha.ui.features.editor.editor_preferences import (
    apply_console_theme,
    apply_editor_preferences,
)
from vizcacha.ui.features.editor.source_formatting import SourceFormatter
from vizcacha.ui.workbench import Workbench

EDIT_ACTIONS = (
    (N_("&Undo"), QKeySequence.Undo, "undo", False),
    (N_("&Redo"), QKeySequence.Redo, "redo", False),
    (N_("Cu&t"), QKeySequence.Cut, "cut", True),
    (N_("&Copy"), QKeySequence.Copy, "copy", False),
    (N_("&Paste"), QKeySequence.Paste, "paste", False),
)


class EditorFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.settings = workbench.services.settings
        self.formatter = SourceFormatter(workbench)
        self.commands = EditCommands(workbench, self.formatter.format_current)

    def register(self) -> None:
        for text, shortcut, method, separator in EDIT_ACTIONS:
            action = QAction(_(text), self.workbench.window)
            action.setShortcut(shortcut)
            action.triggered.connect(lambda _checked=False, m=method: self._on_current(m))
            self.workbench.add_action("edit", action, separator=separator)
        self.commands.register()
        self.workbench.add_settings_page(EditorPage)
        self.workbench.add_settings_page(AppearancePage)
        self.workbench.editor.add_save_hook(self.formatter.format_before_save)
        self.workbench.editor.editor_created.connect(self.apply_to_editor)
        self.workbench.events.settings_changed.connect(self.apply_settings)
        self.apply_settings()

    def apply_settings(self) -> None:
        for editor in self.workbench.editor.editors():
            self.apply_to_editor(editor)
        apply_console_theme(self.workbench.console, self.settings)
        self.commands.find_bar.refresh_highlights()

    def apply_to_editor(self, editor: CodeEditor) -> None:
        apply_editor_preferences(editor, self.settings)
        self.commands.apply_zoom(editor)

    def _on_current(self, method: str) -> None:
        editor = self.workbench.editor.current_editor()
        if editor is not None:
            getattr(editor, method)()


def register(workbench: Workbench) -> None:
    EditorFeature(workbench).register()
