"""One debugging session: a ``dlv dap`` process, its DAP connection and stop inspector."""

from PyQt5.QtCore import QObject, QTimer

from vizcacha.infrastructure.delve_dap.adapter_process import DelveProcess
from vizcacha.infrastructure.delve_dap.connection import DapConnection
from vizcacha.infrastructure.delve_dap.stop_inspector import StateCallback, StopInspector

KILL_DELAY_MS = 1500


class DapSession:
    def __init__(self, owner: QObject, on_state: StateCallback) -> None:
        self.process = DelveProcess(owner)
        self.connection = DapConnection(owner)
        self.inspector = StopInspector(self.connection, on_state)
        self.process.listening.connect(self.connection.connect_to)

    def close(self) -> None:
        """Ask Delve to end the program, then kill whatever is left after a grace period."""
        self.inspector.cancel()
        self.connection.request("disconnect", {"terminateDebuggee": True})
        QTimer.singleShot(KILL_DELAY_MS, self._release)

    def _release(self) -> None:
        self.connection.close()
        self.process.kill()
        self.connection.deleteLater()
        self.process.deleteLater()
