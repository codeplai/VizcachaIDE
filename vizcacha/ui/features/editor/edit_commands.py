"""Edit / View commands: find & replace, go to line, comments, gofmt and zoom."""

from collections.abc import Callable

from PyQt5.QtGui import QKeySequence
from PyQt5.QtWidgets import QAction, QInputDialog

from vizcacha.i18n import N_, _
from vizcacha.ui.editor.code_editor import CodeEditor
from vizcacha.ui.editor.find_replace_bar import FindReplaceBar
from vizcacha.ui.editor.line_comments import toggle_line_comment
from vizcacha.ui.workbench import Workbench

MIN_ZOOM = -6
MAX_ZOOM = 24

# (menu, text, shortcuts, method name, separator before)
COMMANDS = (
    ("edit", N_("&Find..."), ("Ctrl+F",), "show_find", True),
    ("edit", N_("&Replace..."), ("Ctrl+H",), "show_replace", False),
    ("edit", N_("Find &Next"), ("F3",), "find_next", False),
    ("edit", N_("Find &Previous"), ("Shift+F3",), "find_previous", False),
    ("edit", N_("&Go to Line..."), ("Ctrl+G",), "go_to_line", False),
    ("edit", N_("Toggle &Comment"), ("Ctrl+/",), "toggle_comment", True),
    ("edit", N_("F&ormat Code"), ("Ctrl+Shift+F",), "format_code", False),
    ("view", N_("Zoom &In"), ("Ctrl+=", "Ctrl++"), "zoom_in", False),
    ("view", N_("Zoom &Out"), ("Ctrl+-",), "zoom_out", False),
    ("view", N_("&Reset Zoom"), ("Ctrl+0",), "reset_zoom", False),
)


class EditCommands:
    def __init__(self, workbench: Workbench, format_code: Callable[[], None]) -> None:
        self.workbench = workbench
        self.tabs = workbench.editor
        self.zoom_steps = 0
        self._format_code = format_code
        self.find_bar = FindReplaceBar(self.tabs.current_editor, self.tabs)
        self.tabs.currentChanged.connect(lambda _index: self.find_bar.refresh_highlights())

    def register(self) -> None:
        for menu, text, shortcuts, method, separator in COMMANDS:
            action = QAction(_(text), self.workbench.window)
            action.setShortcuts([QKeySequence(shortcut) for shortcut in shortcuts])
            action.triggered.connect(lambda _checked=False, m=method: getattr(self, m)())
            self.workbench.add_action(menu, action, separator=separator)

    # --- find / replace -------------------------------------------------------
    def show_find(self) -> None:
        self.find_bar.show_find()

    def show_replace(self) -> None:
        self.find_bar.show_replace()

    def find_next(self) -> None:
        self._find(self.find_bar.find_next)

    def find_previous(self) -> None:
        self._find(self.find_bar.find_previous)

    def _find(self, search: Callable[[], bool]) -> None:
        if not self.find_bar.query().text:
            self.find_bar.show_find()
            return
        search()

    # --- editing --------------------------------------------------------------
    def go_to_line(self) -> None:
        editor = self.tabs.current_editor()
        if editor is None:
            return
        count = editor.blockCount()
        line, accepted = QInputDialog.getInt(
            self.workbench.window,
            _("Go to Line"),
            _("Line number (1 - {count}):").format(count=count),
            editor.textCursor().blockNumber() + 1,
            1,
            count,
        )
        if accepted:
            editor.move_cursor_to(line)
            editor.setFocus()

    def toggle_comment(self) -> None:
        editor = self.tabs.current_editor()
        if editor is not None:
            toggle_line_comment(editor)

    def format_code(self) -> None:
        self._format_code()

    # --- zoom -------------------------------------------------------------------
    def zoom_in(self) -> None:
        self._set_zoom(self.zoom_steps + 1)

    def zoom_out(self) -> None:
        self._set_zoom(self.zoom_steps - 1)

    def reset_zoom(self) -> None:
        self._set_zoom(0)

    def apply_zoom(self, editor: CodeEditor) -> None:
        editor.set_zoom(self.zoom_steps)

    def _set_zoom(self, steps: int) -> None:
        self.zoom_steps = max(MIN_ZOOM, min(MAX_ZOOM, steps))
        for editor in self.tabs.editors():
            self.apply_zoom(editor)
