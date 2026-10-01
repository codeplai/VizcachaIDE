"""JSON-RPC connection to ``gopls serve`` over stdio, driven by a QProcess.

Notifications arrive asynchronously through Qt signals. Requests are answered
synchronously with :meth:`GoplsConnection.wait_until`, which blocks the caller for
a bounded time (see ``GoplsLanguageServer`` for why).
"""

import logging
import shutil
import time
from collections.abc import Callable
from pathlib import Path

from lsprotocol import types
from PyQt5.QtCore import QObject, QProcess, QProcessEnvironment, pyqtSignal

from vizcacha.application.errors import LanguageServerError
from vizcacha.infrastructure.gopls_lsp.jsonrpc import (
    LanguageServerTimeoutError,
    MessageReader,
    encode_message,
)
from vizcacha.infrastructure.gopls_lsp.lsp_codec import (
    notification_payload,
    request_payload,
    typed_result,
)

LOG = logging.getLogger(__name__)
STOP_TIMEOUT_MS = 1000


class GoplsConnection(QObject):
    notification_received = pyqtSignal(str, object)  # method, raw params (dict)
    response_received = pyqtSignal(int)  # request id; take it with take_result()
    failed = pyqtSignal(str)  # human-readable reason (English, for logs)

    def __init__(self, parent: QObject | None = None) -> None:
        super().__init__(parent)
        self._process: QProcess | None = None
        self._reader = MessageReader()
        self._next_id = 1
        self._expected: set[int] = set()
        self._responses: dict[int, dict] = {}
        self._closing = False

    # --- lifecycle --------------------------------------------------------
    def start(self, executable: str, environment: dict[str, str], folder: Path) -> None:
        """Launch ``<executable> serve``. Raises LanguageServerError if it cannot be found."""
        resolved = shutil.which(executable)
        if resolved is None:
            raise LanguageServerError(f"gopls not found: {executable}")
        process = QProcess(self)
        process_env = QProcessEnvironment()
        for name, value in environment.items():
            process_env.insert(name, value)
        process.setProcessEnvironment(process_env)
        process.setWorkingDirectory(str(folder))
        process.readyReadStandardOutput.connect(self._read_output)
        process.readyReadStandardError.connect(self._discard_log)
        process.errorOccurred.connect(self._on_error)
        process.finished.connect(self._on_finished)
        self._process = process
        process.start(resolved, ["serve"])

    def is_running(self) -> bool:
        return self._process is not None and self._process.state() != QProcess.NotRunning

    def close(self) -> None:
        """Polite ``shutdown`` + ``exit``; kills gopls if it does not stop in time."""
        if not self.is_running():
            return
        self._closing = True
        try:
            self.request(types.SHUTDOWN, None, STOP_TIMEOUT_MS)
        except LanguageServerError as error:
            LOG.info("gopls shutdown: %s", error)
        self.notify(types.EXIT, None)
        if not self._process.waitForFinished(STOP_TIMEOUT_MS):
            self._process.kill()
            self._process.waitForFinished(STOP_TIMEOUT_MS)

    # --- messages -----------------------------------------------------------
    def send_request(self, method: str, params: object) -> int:
        request_id = self._next_id
        self._next_id += 1
        self._expected.add(request_id)
        self._write(request_payload(request_id, method, params))
        return request_id

    def notify(self, method: str, params: object) -> None:
        self._write(notification_payload(method, params))

    def request(
        self,
        method: str,
        params: object,
        timeout_ms: int,
        ready: Callable[[], bool] = lambda: True,
    ) -> object:
        """Send ``method`` once ``ready()`` and wait for its typed result.

        The whole call takes at most ``timeout_ms``. Raises LanguageServerTimeoutError
        when it expires and LanguageServerError when gopls answers with an error.
        """
        deadline = time.monotonic() + timeout_ms / 1000
        if not self.wait_until(ready, timeout_ms):
            raise LanguageServerTimeoutError(f"gopls is not ready for {method}")
        remaining_ms = max(int((deadline - time.monotonic()) * 1000), 1)
        request_id = self.send_request(method, params)
        if not self.wait_until(lambda: self.has_response(request_id), remaining_ms):
            self.forget(request_id)
            raise LanguageServerTimeoutError(f"{method} timed out")
        return self.take_result(request_id, method)

    def has_response(self, request_id: int) -> bool:
        return request_id in self._responses

    def forget(self, request_id: int) -> None:
        """Stop waiting for ``request_id``: a late answer will be dropped."""
        self._expected.discard(request_id)
        self._responses.pop(request_id, None)

    def take_result(self, request_id: int, method: str) -> object:
        """Typed ``result`` of an answered request. Raises LanguageServerError on errors."""
        self._expected.discard(request_id)
        return typed_result(self._responses.pop(request_id), method)

    def wait_until(self, condition: Callable[[], bool], timeout_ms: int) -> bool:
        """Read from gopls until ``condition()`` holds or ``timeout_ms`` elapse."""
        deadline = time.monotonic() + timeout_ms / 1000
        while not condition():
            remaining_ms = int((deadline - time.monotonic()) * 1000)
            if remaining_ms <= 0 or not self.is_running():
                return condition()
            self._process.waitForReadyRead(remaining_ms)
            self._read_output()
        return True

    # --- internals ------------------------------------------------------------
    def _write(self, payload: dict) -> None:
        if self._process is None:
            raise LanguageServerError("gopls is not running")
        self._process.write(encode_message(payload))

    def _read_output(self) -> None:
        if self._process is None:
            return
        data = bytes(self._process.readAllStandardOutput())
        if not data:
            return
        try:
            messages = self._reader.feed(data)
        except LanguageServerError as error:
            self._fail(str(error))
            return
        # Answers first: a notification handler may wait for one of them.
        for message in sorted(messages, key=lambda message: "id" not in message):
            self._dispatch(message)

    def _dispatch(self, message: dict) -> None:
        if "method" in message and "id" in message:
            self._write({"jsonrpc": "2.0", "id": message["id"], "result": None})
            return
        if "method" in message:
            self.notification_received.emit(message["method"], message.get("params"))
            return
        request_id = message.get("id")
        if request_id in self._expected:
            self._responses[request_id] = message
            self.response_received.emit(request_id)

    def _discard_log(self) -> None:
        if self._process is not None:
            LOG.debug(
                "gopls: %s", bytes(self._process.readAllStandardError()).decode(errors="replace")
            )

    def _on_error(self, error: QProcess.ProcessError) -> None:
        if error == QProcess.FailedToStart:
            self._fail("gopls failed to start")

    def _on_finished(self, exit_code: int, _status: QProcess.ExitStatus) -> None:
        if not self._closing:
            self._fail(f"gopls exited with code {exit_code}")

    def _fail(self, reason: str) -> None:
        LOG.warning("Language server stopped: %s", reason)
        self.failed.emit(reason)
