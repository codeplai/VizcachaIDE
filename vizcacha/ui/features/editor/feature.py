"""Edit menu and editor settings (font, word wrap; themes are pending for track D)."""

from PyQt5.QtGui import QFont, QKeySequence
from PyQt5.QtWidgets import QAction, QPlainTextEdit

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import N_, _
from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.features.editor.appearance_page import AppearancePage
from vizcacha.ui.features.editor.editor_page import EditorPage
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

    def register(self) -> None:
        for text, shortcut, method, separator in EDIT_ACTIONS:
            action = QAction(_(text), self.workbench.window)
            action.setShortcut(shortcut)
            action.triggered.connect(lambda _checked=False, m=method: self._on_current(m))
            self.workbench.add_action("edit", action, separator=separator)
        self.workbench.add_settings_page(EditorPage)
        self.workbench.add_settings_page(AppearancePage)
        self.workbench.editor.editor_created.connect(self.apply_to_editor)
        self.workbench.events.settings_changed.connect(self.apply_to_all_editors)

    def apply_to_all_editors(self) -> None:
        for editor in self.workbench.editor.editors():
            self.apply_to_editor(editor)

    def apply_to_editor(self, editor: CodeEditor) -> None:
        font = QFont(
            self.settings.get(SettingsKeys.FONT_FAMILY, "Consolas"),
            self.settings.get(SettingsKeys.FONT_SIZE, 11),
        )
        editor.setFont(font)
        wrap = self.settings.get(SettingsKeys.WORD_WRAP, False)
        editor.setLineWrapMode(QPlainTextEdit.WidgetWidth if wrap else QPlainTextEdit.NoWrap)

    def _on_current(self, method: str) -> None:
        editor = self.workbench.editor.current_editor()
        if editor is not None:
            getattr(editor, method)()


def register(workbench: Workbench) -> None:
    EditorFeature(workbench).register()
