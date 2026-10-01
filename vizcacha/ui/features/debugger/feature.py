"""Debug menu, Variables and Call Stack panels, wired to the DebuggerPort.

Phase 0 uses NullDebugger, so "Start Debugging" only explains that the debugger
is not available yet. Track A replaces it with the Delve DAP adapter.
"""

from PyQt5.QtWidgets import QAction, QMessageBox

from vizcacha.domain.debugging import Breakpoint, DebugState, StackFrame
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.ui.features.debugger.callstack_view import CallStackView
from vizcacha.ui.features.debugger.variables_view import VariablesView
from vizcacha.ui.workbench import Workbench


class DebuggerFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.debugger = workbench.services.debugger
        self.variables_view = VariablesView()
        self.callstack_view = CallStackView()

    def register(self) -> None:
        self.start_action = self._action(_("🐛 Debug"), "F6", self.start, toolbar=True)
        self.step_actions = [
            self._action(_("↷ Step Over"), "F7", self.debugger.step_over, toolbar=True),
            self._action(_("↓ Step Into"), "F8", self.debugger.step_into, toolbar=True),
            self._action(_("↑ Step Out"), "F9", self.debugger.step_out, toolbar=True),
        ]
        self.stop_action = self._action(_("Stop Debugging"), "Ctrl+F6", self.debugger.stop)
        self._action(
            _("Toggle &Breakpoint"), "F10", self.workbench.editor.toggle_breakpoint, separator=True
        )
        self.workbench.add_panel("variables", _("Variables"), self.variables_view)
        self.workbench.add_panel("callstack", _("Call Stack"), self.callstack_view)
        self.callstack_view.frame_activated.connect(self._navigate_to_frame)
        self.debugger.stopped.connect(self._on_stopped)
        self.debugger.output.connect(self._on_output)
        self.debugger.terminated.connect(self._on_terminated)
        self._set_debugging(False)

    def start(self) -> None:
        path = self.workbench.editor.current_file_path()
        if path is None:
            QMessageBox.warning(
                self.workbench.window, _("No File"), _("Please save your file before debugging.")
            )
            return
        self.workbench.console.clear()
        self.variables_view.clear()
        self.callstack_view.clear()
        breakpoints = [
            Breakpoint(SourceLocation(path, line))
            for line in self.workbench.editor.get_all_breakpoints()
        ]
        self._set_debugging(True)
        self.debugger.start(RunConfiguration.for_file(path), breakpoints)

    def _on_stopped(self, state: DebugState) -> None:
        self.variables_view.show_variables(state.variables)
        self.callstack_view.show_frames(state.frames)
        location = state.current_location
        if location is not None:
            self.workbench.events.navigate_to.emit(location)
            self.workbench.editor.highlight_current_line(location.line)

    def _on_output(self, text: str, category: str) -> None:
        if category == "stderr":
            self.workbench.console.append_error(text)
        else:
            self.workbench.console.append_output(text)

    def _on_terminated(self, _exit_code: int) -> None:
        self._set_debugging(False)
        self.workbench.editor.clear_current_line_highlight()

    def _navigate_to_frame(self, frame: StackFrame) -> None:
        if frame.location is not None:
            self.workbench.events.navigate_to.emit(frame.location)

    def _set_debugging(self, active: bool) -> None:
        self.start_action.setEnabled(not active)
        self.stop_action.setEnabled(active)
        for action in self.step_actions:
            action.setEnabled(active)

    def _action(
        self, text: str, shortcut: str, slot, toolbar: bool = False, separator: bool = False
    ) -> QAction:
        action = QAction(text, self.workbench.window)
        action.setShortcut(shortcut)
        action.triggered.connect(lambda _checked=False: slot())
        return self.workbench.add_action("debug", action, toolbar=toolbar, separator=separator)


def register(workbench: Workbench) -> None:
    DebuggerFeature(workbench).register()
