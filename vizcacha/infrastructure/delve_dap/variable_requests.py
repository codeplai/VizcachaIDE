"""Asynchronous ``variables`` requests for the lazy expansion of the Variables panel.

Delve hands out new ``variablesReference`` numbers at every stop and reuses the old
ones, so an answer that arrives after the program resumed is dropped: the next
``stopped`` state replaces every variable anyway.
"""

from collections.abc import Callable

from vizcacha.domain.debugging import Variable
from vizcacha.infrastructure.delve_dap import state_mapping
from vizcacha.infrastructure.delve_dap.connection import DapConnection

LoadedCallback = Callable[[int, list[Variable]], None]


class VariableRequests:
    def __init__(self, on_loaded: LoadedCallback) -> None:
        self._on_loaded = on_loaded
        self._generation = 0

    def invalidate(self) -> None:
        """The program resumed or ended: answers still in flight are meaningless."""
        self._generation += 1

    def request(self, connection: DapConnection, reference: int) -> None:
        if reference == 0:
            self._on_loaded(reference, [])
            return
        generation = self._generation

        def on_response(response: dict) -> None:
            if generation != self._generation:
                return
            body = (response.get("body") or {}) if response.get("success") else {}
            self._on_loaded(reference, state_mapping.map_variables(body))

        arguments = {"variablesReference": reference}
        if connection.request("variables", arguments, on_response) == 0:
            self._on_loaded(reference, [])  # the connection is already closed
