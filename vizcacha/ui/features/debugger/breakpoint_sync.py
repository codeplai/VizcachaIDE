"""Keeps Delve's breakpoints in step with the editors' gutters during a session."""

from vizcacha.application.ports import DebuggerPort
from vizcacha.domain.debugging import Breakpoint
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.ui.editor import CodeEditor, TabbedEditor


def collect_breakpoints(tabs: TabbedEditor) -> list[Breakpoint]:
    """Breakpoints of every saved file open in the editor."""
    return [
        Breakpoint(SourceLocation(editor.file_path, line))
        for editor in tabs.editors()
        if editor.file_path is not None
        for line in editor.get_breakpoints()
    ]


class BreakpointSync:
    def __init__(self, tabs: TabbedEditor, debugger: DebuggerPort) -> None:
        self._debugger = debugger
        tabs.editor_created.connect(self.attach)
        for editor in tabs.editors():
            self.attach(editor)

    def attach(self, editor: CodeEditor) -> None:
        editor.breakpoints_changed.connect(lambda: self._sync(editor))

    def _sync(self, editor: CodeEditor) -> None:
        if editor.file_path is None or not self._debugger.is_active():
            return
        self._debugger.set_breakpoints(editor.file_path, editor.get_breakpoints())
