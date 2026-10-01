"""DebuggerPort on ``dlv dap``: initialize → launch → setBreakpoints → configurationDone.

Each session gets a fresh DapSession, so late signals of an old one are harmless.
"""

import os
from collections.abc import Sequence
from pathlib import Path

from PyQt5.QtCore import QObject, pyqtSignal

from vizcacha.application.errors import VizcachaError
from vizcacha.application.ports import TERMINATED_BY_USER
from vizcacha.domain.debugging import Breakpoint
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.infrastructure.delve_dap import launch_arguments as requests
from vizcacha.infrastructure.delve_dap import state_mapping
from vizcacha.infrastructure.delve_dap.breakpoint_registry import BreakpointRegistry
from vizcacha.infrastructure.delve_dap.connection import response_error
from vizcacha.infrastructure.delve_dap.session import DapSession
from vizcacha.infrastructure.delve_dap.variable_requests import VariableRequests
from vizcacha.infrastructure.go_toolchain import GoEnvironment

FAILED = 1


class DelveDapDebugger(QObject):
    stopped = pyqtSignal(object)  # DebugState
    output = pyqtSignal(str, str)  # text, "stdout" | "stderr" | "console"
    terminated = pyqtSignal(int)  # exit code or TERMINATED_BY_USER
    variables_loaded = pyqtSignal(int, object)  # reference, list[Variable]

    def __init__(self, environment: GoEnvironment, parent: QObject | None = None) -> None:
        super().__init__(parent)
        self._environment = environment
        self._registry = BreakpointRegistry()
        self._session: DapSession | None = None
        self._config: RunConfiguration | None = None
        self._active = False
        self._configured = False
        self._thread_id = 0
        self._exit_code = 0
        self._variables = VariableRequests(self.variables_loaded.emit)

    # --- DebuggerPort -------------------------------------------------------
    def is_active(self) -> bool:
        return self._active

    def start(self, config: RunConfiguration, breakpoints: Sequence[Breakpoint]) -> None:
        if self._active:
            return
        self._active, self._configured = True, False
        self._thread_id, self._exit_code, self._config = 0, 0, config
        self._registry.reset(breakpoints)
        self._session = self._create_session()
        try:
            self._session.process.start(
                self._environment.delve_executable(),
                config.working_dir,
                self._environment.variables(),
            )
        except VizcachaError as error:  # DebugAdapterNotFoundError
            self._fail(error)

    def set_breakpoints(self, file: Path, lines: Sequence[int]) -> None:
        key = self._registry.replace(file, lines)
        if self._configured:
            self._send_breakpoints(key)

    def step_over(self) -> None:
        self._execute("next")

    def step_into(self) -> None:
        self._execute("stepIn")

    def step_out(self) -> None:
        self._execute("stepOut")

    def resume(self) -> None:
        self._execute("continue")

    def run_to(self, location: SourceLocation) -> None:
        if not self._configured:
            return
        self._send_breakpoints(self._registry.set_temporary(location))
        self._execute("continue")

    def request_variables(self, reference: int) -> None:
        if not self._configured:
            self.variables_loaded.emit(reference, [])
            return
        self._variables.request(self._session.connection, reference)

    def stop(self) -> None:
        self._finish(TERMINATED_BY_USER)

    # --- session wiring -------------------------------------------------------
    def _create_session(self) -> DapSession:
        session = DapSession(self, self.stopped.emit)  # closing a session cancels its inspector

        def current(slot):  # ignore signals of a previous session that is still shutting down
            return lambda *args: slot(*args) if session is self._session else None

        session.process.failed_to_start.connect(current(self._fail))
        session.process.finished.connect(current(self._on_process_finished))
        session.connection.connected.connect(current(self._initialize))
        session.connection.event_received.connect(current(self._on_event))
        session.connection.failed.connect(current(self._fail))
        return session

    def _request(self, command: str, arguments: dict | None = None, on_response=None) -> None:
        self._session.connection.request(command, arguments, on_response)

    def _initialize(self) -> None:
        self._request("initialize", requests.INITIALIZE_ARGUMENTS, self._launch)

    def _launch(self, response: dict) -> None:
        if not self._succeeded(response):
            return
        output = requests.debug_binary_path(os.getpid())
        environment = self._environment.variables()
        arguments = requests.launch_arguments(self._config, environment, output)
        self._request("launch", arguments, self._succeeded)

    def _succeeded(self, response: dict) -> bool:
        """A failed initialize/launch (e.g. a build error) ends the session."""
        if not response.get("success"):
            self._fail(response_error(response))
        return bool(response.get("success"))

    def _configure(self) -> None:
        self._configured = True
        for file in self._registry.files():
            self._send_breakpoints(file)
        self._request("configurationDone")

    def _send_breakpoints(self, file: Path) -> None:
        lines = self._registry.lines_for(file)
        self._request("setBreakpoints", requests.set_breakpoints_arguments(file, lines))

    def _execute(self, command: str) -> None:
        if not self._configured or self._thread_id == 0:
            return
        self._session.inspector.cancel()
        self._variables.invalidate()
        self._request(command, {"threadId": self._thread_id})

    # --- events -----------------------------------------------------------------
    def _on_event(self, message: dict) -> None:
        body = message.get("body") or {}
        handlers = {
            "initialized": self._configure,
            "output": lambda: self._on_output(body),
            "stopped": lambda: self._on_stopped(body),
            "exited": lambda: self._remember_exit_code(body),
            "terminated": lambda: self._finish(self._exit_code),
        }
        handler = handlers.get(message.get("event", ""))
        if handler is not None:
            handler()

    def _on_output(self, body: dict) -> None:
        event = state_mapping.output_event(body)
        if event is not None and self._active:  # skip Delve's "Detaching..." after the end
            self.output.emit(*event)

    def _on_stopped(self, body: dict) -> None:
        self._thread_id = int(body.get("threadId", 0) or 0) or self._thread_id
        temporary_file = self._registry.clear_temporary()
        if temporary_file is not None:
            self._send_breakpoints(temporary_file)
        self._session.inspector.inspect(body)

    def _remember_exit_code(self, body: dict) -> None:
        self._exit_code = int(body.get("exitCode", 0) or 0)

    # --- shutdown -----------------------------------------------------------
    def _fail(self, error: VizcachaError) -> None:
        self.output.emit(str(error).rstrip("\n") + "\n", "stderr")
        self._finish(FAILED)

    def _on_process_finished(self, exit_code: int) -> None:
        if self._active:
            self.output.emit(_("Delve stopped unexpectedly.") + "\n", "console")
            self._finish(exit_code or FAILED)

    def _finish(self, exit_code: int) -> None:
        if not self._active:
            return
        self._active = self._configured = False
        self._variables.invalidate()
        self._session.close()
        self.terminated.emit(exit_code)
