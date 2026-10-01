"""DAP client over ``QTcpSocket``: requests with callbacks, events as a Qt signal.

Everything is asynchronous: responses are dispatched from the event loop as they
arrive, so the UI never waits for Delve.
"""

from collections.abc import Callable

from PyQt5.QtCore import QObject, pyqtSignal
from PyQt5.QtNetwork import QAbstractSocket, QTcpSocket

from vizcacha.application.errors import DebugAdapterError
from vizcacha.infrastructure.delve_dap.framing import DapFrameDecoder, encode_message

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
