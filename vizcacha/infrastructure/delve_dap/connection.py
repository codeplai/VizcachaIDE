"""DAP client over ``QTcpSocket``: requests with callbacks, events as a Qt signal.

Everything is asynchronous except ``request_blocking``, which exists only for the
synchronous ``DebuggerPort.variables``. It waits at most ``BLOCKING_TIMEOUT_MS``
with short ``waitForReadyRead`` slices (no extra threads), and every other message
that arrives meanwhile is dispatched later from the event loop, never re-entrantly.
"""

import time
from collections.abc import Callable

from PyQt5.QtCore import QObject, QTimer, pyqtSignal
from PyQt5.QtNetwork import QAbstractSocket, QTcpSocket

from vizcacha.application.errors import DebugAdapterError
from vizcacha.infrastructure.delve_dap.framing import DapFrameDecoder, encode_message

BLOCKING_TIMEOUT_MS = 2000
WAIT_SLICE_MS = 100

ResponseCallback = Callable[[dict], None]


def response_error(response: dict) -> DebugAdapterError:
    details = ((response.get("body") or {}).get("error") or {}).get("format", "")
    message = details or response.get("message") or response.get("command", "")
    return DebugAdapterError(message)


class DapConnection(QObject):
    connected = pyqtSignal()
    event_received = pyqtSignal(dict)
    failed = pyqtSignal(object)  # DebugAdapterError
    closed = pyqtSignal()

    def __init__(self, parent: QObject | None = None) -> None:
        super().__init__(parent)
        self._socket = QTcpSocket(self)
        self._decoder = DapFrameDecoder()
        self._seq = 0
        self._callbacks: dict[int, ResponseCallback] = {}
        self._awaiting: int | None = None
        self._awaited: dict | None = None
        self._deferred: list[dict] = []
        self._socket.connected.connect(self.connected)
        self._socket.readyRead.connect(self._read_available)
        self._socket.disconnected.connect(self.closed)
        self._socket.errorOccurred.connect(self._on_socket_error)

    def connect_to(self, host: str, port: int) -> None:
        self._socket.connectToHost(host, port)

    def is_open(self) -> bool:
        return self._socket.state() == QAbstractSocket.ConnectedState

    def close(self) -> None:
        self._callbacks.clear()
        self._socket.abort()

    def request(
        self,
        command: str,
        arguments: dict | None = None,
        on_response: ResponseCallback | None = None,
    ) -> int:
        if not self.is_open():
            return 0
        self._seq += 1
        message: dict = {"seq": self._seq, "type": "request", "command": command}
        if arguments is not None:
            message["arguments"] = arguments
        if on_response is not None:
            self._callbacks[self._seq] = on_response
        self._socket.write(encode_message(message))
        return self._seq

    def request_blocking(self, command: str, arguments: dict | None = None) -> dict:
        """Send a request and wait (bounded) for its response. Raises DebugAdapterError."""
        seq = self.request(command, arguments)
        if seq == 0:
            raise DebugAdapterError("The debug adapter is not connected.")
        self._awaiting, self._awaited = seq, None
        deadline = time.monotonic() + BLOCKING_TIMEOUT_MS / 1000
        try:
            while self._awaited is None and time.monotonic() < deadline and self.is_open():
                self._socket.waitForReadyRead(WAIT_SLICE_MS)
                self._read_available()
        finally:
            self._awaiting = None
            if self._deferred:
                QTimer.singleShot(0, self._flush_deferred)
        if self._awaited is None:
            raise DebugAdapterError(f"Timed out waiting for '{command}'.")
        return self._awaited

    # --- incoming -----------------------------------------------------------
    def _read_available(self) -> None:
        data = bytes(self._socket.readAll())
        if not data:
            return
        try:
            messages = self._decoder.feed(data)
        except DebugAdapterError as error:
            self.failed.emit(error)
            return
        for message in messages:
            self._route(message)

    def _route(self, message: dict) -> None:
        if self._awaiting is None:
            self._dispatch(message)
            return
        if message.get("type") == "response" and message.get("request_seq") == self._awaiting:
            self._awaited = message
            return
        self._deferred.append(message)

    def _flush_deferred(self) -> None:
        pending, self._deferred = self._deferred, []
        for message in pending:
            self._dispatch(message)

    def _dispatch(self, message: dict) -> None:
        kind = message.get("type")
        if kind == "event":
            self.event_received.emit(message)
            return
        if kind != "response":
            return
        callback = self._callbacks.pop(message.get("request_seq", -1), None)
        if callback is not None:
            callback(message)

    def _on_socket_error(self, socket_error: QAbstractSocket.SocketError) -> None:
        if socket_error == QAbstractSocket.RemoteHostClosedError:
            return
        self.failed.emit(DebugAdapterError(self._socket.errorString()))
