"""Debug menu and the Variables, Call Stack and Goroutines panels, wired to DebuggerPort."""

from PyQt5.QtWidgets import QAction, QMessageBox, QTabWidget

from vizcacha.domain.debugging import DebugState, StackFrame
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.ui.features.debugger.breakpoint_sync import BreakpointSync, collect_breakpoints
from vizcacha.ui.features.debugger.callstack_view import CallStackView
from vizcacha.ui.features.debugger.goroutines_view import GoroutinesView
from vizcacha.ui.features.debugger.variables_view import VariablesView
from vizcacha.ui.workbench import Workbench


class DebuggerFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.debugger = workbench.services.debugger
        self.variables_view = VariablesView(children_provider=self.debugger.variables)
        self.callstack_view = CallStackView()
        self.goroutines_view = GoroutinesView()
        self.breakpoint_sync: BreakpointSync | None = None

    def register(self) -> None:
        self.start_action = self._action(_("🐛 Debug"), "F6", self.start, toolbar=True)
        self.session_actions = [
            self._action(_("⏩ Continue"), "Shift+F6", self.debugger.resume, toolbar=True),
            self._action(_("↷ Step Over"), "F7", self.debugger.step_over, toolbar=True),
            self._action(_("↓ Step Into"), "F8", self.debugger.step_into, toolbar=True),
            self._action(_("↑ Step Out"), "F9", self.debugger.step_out, toolbar=True),
            self._action(_("Run to &Cursor"), "Ctrl+F10", self.run_to_cursor),
            self._action(_("Stop Debugging"), "Ctrl+F6", self.debugger.stop),
        ]
        self._action(
            _("Toggle &Breakpoint"), "F10", self.workbench.editor.toggle_breakpoint, separator=True
        )
        self._add_panels()
        self.breakpoint_sync = BreakpointSync(self.workbench.editor, self.debugger)
        self.callstack_view.frame_activated.connect(self._navigate_to_frame)
        self.debugger.stopped.connect(self._on_stopped)
        self.debugger.output.connect(self._on_output)
        self.debugger.terminated.connect(self._on_terminated)
        self._set_debugging(False)

    def _add_panels(self) -> None:
        self.workbench.add_panel("variables", _("Variables"), self.variables_view)
        tabs = QTabWidget()
        tabs.addTab(self.callstack_view, _("Call Stack"))
        tabs.addTab(self.goroutines_view, _("Goroutines"))
        self.workbench.add_panel("callstack", _("Call Stack"), tabs)

    def start(self) -> None:
        path = self.workbench.editor.current_file_path()
        if path is None:
            QMessageBox.warning(
                self.workbench.window, _("No File"), _("Please save your file before debugging.")
            )
            return
        self.workbench.console.clear()
        self._clear_panels()
        self.workbench.console.append_output(
            _("Note: the program cannot read keyboard input (stdin) while debugging.") + "\n"
        )
        self._set_debugging(True)
        breakpoints = collect_breakpoints(self.workbench.editor)
        self.debugger.start(RunConfiguration.for_file(path), breakpoints)

    def run_to_cursor(self) -> None:
        editor = self.workbench.editor.current_editor()
        if editor is None or editor.file_path is None:
            return
        line = editor.textCursor().blockNumber() + 1
        self.debugger.run_to(SourceLocation(editor.file_path, line))

    def _on_stopped(self, state: DebugState) -> None:
        self.variables_view.show_variables(state.variables)
        self.callstack_view.show_frames(state.frames)
        self.goroutines_view.show_goroutines(state.goroutines, state.current_goroutine)
        self.workbench.editor.clear_current_line_highlight()
        if state.description:
            self.workbench.show_status_message(state.description)
        location = state.current_location
        if location is None:
            return
        self.workbench.events.navigate_to.emit(location)
        self.workbench.editor.highlight_current_line(location.line)

    def _on_output(self, text: str, category: str) -> None:
        if category == "stderr":
            self.workbench.console.append_error(text)
        else:
            self.workbench.console.append_output(text)
        if category in ("stdout", "stderr"):
            self.workbench.events.process_output.emit(text, category)

    def _on_terminated(self, exit_code: int) -> None:
        self._set_debugging(False)
        self._clear_panels()
        self.workbench.editor.clear_current_line_highlight()
        if exit_code < 0:
            message = _("[Debugging stopped]")
        else:
            message = _("[Debugging finished with code {code}]").format(code=exit_code)
        self.workbench.console.append_success("\n" + message + "\n")

    def _clear_panels(self) -> None:
        self.variables_view.clear()
        self.callstack_view.clear()
        self.goroutines_view.clear()

    def _navigate_to_frame(self, frame: StackFrame) -> None:
        if frame.location is not None:
            self.workbench.events.navigate_to.emit(frame.location)

    def _set_debugging(self, active: bool) -> None:
        self.start_action.setEnabled(not active)
        for action in self.session_actions:
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
