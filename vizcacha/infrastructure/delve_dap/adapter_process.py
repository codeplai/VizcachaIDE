"""Runs ``dlv dap --listen=127.0.0.1:0`` and reports the port it listens on."""

import re
import shutil
from collections.abc import Mapping
from pathlib import Path

from PyQt5.QtCore import QObject, QProcess, QProcessEnvironment, pyqtSignal

from vizcacha.application.errors import DebugAdapterNotFoundError
from vizcacha.i18n import _

LISTEN_ADDRESS = "127.0.0.1:0"
LISTENING_PATTERN = re.compile(r"DAP server listening at:\s*(?P<host>[^\s:]+):(?P<port>\d+)")
KILL_WAIT_MS = 500
INSTALL_COMMAND = "go install github.com/go-delve/delve/cmd/dlv@latest"


def parse_listening_port(line: str) -> tuple[str, int] | None:
    match = LISTENING_PATTERN.search(line)
    if match is None:
        return None
    return match.group("host"), int(match.group("port"))


def delve_not_found_error(executable: str) -> DebugAdapterNotFoundError:
    return DebugAdapterNotFoundError(
        _(
            "Delve (dlv) was not found: {executable}\n"
            "Install it with:\n    {command}\n"
            "or set its path in Tools > Options > Environment."
        ).format(executable=executable, command=INSTALL_COMMAND)
    )


def resolve_executable(executable: str) -> str:
    """Absolute path of ``executable``. Raises DebugAdapterNotFoundError."""
    if Path(executable).is_file():
        return executable
    found = shutil.which(executable)
    if found is None:
        raise delve_not_found_error(executable)
    return found


class DelveProcess(QObject):
    listening = pyqtSignal(str, int)  # host, port
    log_line = pyqtSignal(str)  # anything else Delve prints on its own console
    finished = pyqtSignal(int)
    failed_to_start = pyqtSignal(object)  # DebugAdapterNotFoundError

    def __init__(self, parent: QObject | None = None) -> None:
        super().__init__(parent)
        self._process: QProcess | None = None
        self._pending = ""
        self._executable = ""

    def start(self, executable: str, working_dir: Path, environment: Mapping[str, str]) -> None:
        self._executable = resolve_executable(executable)
        self._pending = ""
        process = QProcess(self)
        process.setProcessChannelMode(QProcess.MergedChannels)
        process.setWorkingDirectory(str(working_dir))
        qt_environment = QProcessEnvironment()
        for name, value in environment.items():
            qt_environment.insert(name, value)
        process.setProcessEnvironment(qt_environment)
        process.readyReadStandardOutput.connect(self._read_output)
        process.finished.connect(lambda code, _status: self.finished.emit(code))
        process.errorOccurred.connect(self._on_error)
        self._process = process
        process.start(self._executable, ["dap", f"--listen={LISTEN_ADDRESS}"])

    def is_running(self) -> bool:
        return self._process is not None and self._process.state() != QProcess.NotRunning

    def kill(self) -> None:
        """Kill Delve (and with it the debuggee) without emitting ``finished`` any more."""
        if not self.is_running():
            return
        self._process.finished.disconnect()
        self._process.errorOccurred.disconnect()
        self._process.kill()
        self._process.waitForFinished(KILL_WAIT_MS)

    def _read_output(self) -> None:
        data = bytes(self._process.readAllStandardOutput()).decode("utf-8", errors="replace")
        self._pending += data
        *lines, self._pending = self._pending.split("\n")
        for line in lines:
            self._handle_line(line.rstrip("\r"))

    def _handle_line(self, line: str) -> None:
        address = parse_listening_port(line)
        if address is not None:
            self.listening.emit(*address)
            return
        if line.strip():
            self.log_line.emit(line)

    def _on_error(self, error: QProcess.ProcessError) -> None:
        if error == QProcess.FailedToStart:
            self.failed_to_start.emit(delve_not_found_error(self._executable))
