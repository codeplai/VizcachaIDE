"""Builds a DebugState after a DAP ``stopped`` event.

Asynchronous chain: threads → stackTrace → scopes → variables (Locals of the top
user frame). Each step is a response callback, so the UI never blocks.

In parallel with the chain, a ``stackTrace`` with ``levels=1`` per goroutine (the first
``GOROUTINE_LOCATION_LIMIT`` of them) gives each goroutine its location. The state is
emitted once every answer is in.
"""

from collections.abc import Callable
from dataclasses import dataclass, field, replace

from vizcacha.domain.debugging import DebugState, Goroutine, StackFrame, StopReason, Variable
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.infrastructure.delve_dap import state_mapping
from vizcacha.infrastructure.delve_dap.connection import DapConnection

STACK_LEVELS = 50
GOROUTINE_LOCATION_LIMIT = 50
StateCallback = Callable[[DebugState], None]


@dataclass
class _Snapshot:
    reason: StopReason
    thread_id: int
    description: str
    goroutines: tuple[Goroutine, ...] = ()
    frames: tuple[StackFrame, ...] = ()
    variables: tuple[Variable, ...] = ()
    locations: dict[int, SourceLocation] = field(default_factory=dict)
    pending: int = 0  # answers still missing; the threads → variables chain counts as one

    def located(self, goroutine: Goroutine) -> Goroutine:
        location = self.locations.get(goroutine.goroutine_id)
        if goroutine.goroutine_id == self.thread_id and self.frames:
            location = self.frames[0].location  # the user frame, not the panic machinery
        return replace(goroutine, location=location)

    def to_state(self) -> DebugState:
        return DebugState(
            reason=self.reason,
            frames=self.frames,
            variables=self.variables,
            goroutines=tuple(self.located(goroutine) for goroutine in self.goroutines),
            current_goroutine=self.thread_id,
            description=self.description,
        )


class StopInspector:
    def __init__(self, connection: DapConnection, on_state: StateCallback) -> None:
        self._connection = connection
        self._on_state = on_state
        self._generation = 0
        self._snapshot: _Snapshot | None = None

    def cancel(self) -> None:
        """Ignore any answer still in flight (the program resumed or ended)."""
        self._generation += 1
        self._snapshot = None

    def inspect(self, stopped_body: dict) -> None:
        self.cancel()
        self._snapshot = _Snapshot(
            reason=state_mapping.stop_reason(stopped_body.get("reason", "")),
            thread_id=int(stopped_body.get("threadId", 0) or 0),
            description=state_mapping.stop_description(stopped_body),
        )
        self._request("threads", None, self._on_threads)

    def _request(self, command: str, arguments: dict | None, handler) -> None:
        generation = self._generation

        def on_response(response: dict) -> None:
            if generation != self._generation or self._snapshot is None:
                return
            handler((response.get("body") or {}) if response.get("success") else {})

        self._connection.request(command, arguments, on_response)

    def _on_threads(self, body: dict) -> None:
        self._snapshot.goroutines = state_mapping.map_goroutines(body)
        if self._snapshot.thread_id == 0 and self._snapshot.goroutines:
            self._snapshot.thread_id = self._snapshot.goroutines[0].goroutine_id
        others = [
            goroutine.goroutine_id
            for goroutine in self._snapshot.goroutines[:GOROUTINE_LOCATION_LIMIT]
            if goroutine.goroutine_id != self._snapshot.thread_id
        ]
        self._snapshot.pending = 1 + len(others)  # set before any (possibly instant) answer
        arguments = {"threadId": self._snapshot.thread_id, "startFrame": 0, "levels": STACK_LEVELS}
        self._request("stackTrace", arguments, self._on_stack_trace)
        for goroutine_id in others:
            self._request_location(goroutine_id)

    def _request_location(self, goroutine_id: int) -> None:
        def on_top_frame(body: dict) -> None:
            frames = state_mapping.map_frames(body)
            if frames and frames[0].location is not None:
                self._snapshot.locations[goroutine_id] = frames[0].location
            self._settle()

        arguments = {"threadId": goroutine_id, "startFrame": 0, "levels": 1}
        self._request("stackTrace", arguments, on_top_frame)

    def _on_stack_trace(self, body: dict) -> None:
        skip_runtime = self._snapshot.reason is StopReason.PANIC
        self._snapshot.frames = state_mapping.map_frames(body, skip_runtime=skip_runtime)
        if not self._snapshot.frames:
            self._settle()
            return
        arguments = {"frameId": self._snapshot.frames[0].frame_id}
        self._request("scopes", arguments, self._on_scopes)

    def _on_scopes(self, body: dict) -> None:
        reference = state_mapping.locals_reference(body)
        if reference == 0:
            self._settle()
            return
        self._request("variables", {"variablesReference": reference}, self._on_variables)

    def _on_variables(self, body: dict) -> None:
        self._snapshot.variables = tuple(state_mapping.map_variables(body))
        self._settle()

    def _settle(self) -> None:
        self._snapshot.pending -= 1
        if self._snapshot.pending == 0:
            self._finish()

    def _finish(self) -> None:
        state = self._snapshot.to_state()
        self._snapshot = None
        self._on_state(state)
