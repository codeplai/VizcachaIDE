"""Recorded ``dlv dap`` transcripts and a fake DAP session to replay them without dlv."""

import json
from pathlib import Path

import pytest
from PyQt5.QtCore import QObject, pyqtSignal

from vizcacha.infrastructure.delve_dap import debugger as debugger_module
from vizcacha.infrastructure.delve_dap.stop_inspector import StopInspector

TRANSCRIPTS = Path(__file__).parent / "transcripts"
GOROOT = "C:/Program Files/Go"


def load_transcript(name: str, source_dir: Path) -> dict:
    text = (TRANSCRIPTS / name).read_text(encoding="utf-8")
    text = text.replace("{SRC}", source_dir.as_posix()).replace("{GOROOT}", GOROOT)
    return json.loads(text)


@pytest.fixture
def source_dir(tmp_path: Path) -> Path:
    return tmp_path


@pytest.fixture
def functions_transcript(source_dir: Path) -> list[dict]:
    return load_transcript("functions_session.json", source_dir)["messages"]


@pytest.fixture
def panic_transcript(source_dir: Path) -> dict:
    return load_transcript("panic_and_build_error.json", source_dir)


class FakeProcess(QObject):
    listening = pyqtSignal(str, int)
    log_line = pyqtSignal(str)
    finished = pyqtSignal(int)
    failed_to_start = pyqtSignal(object)

    def __init__(self) -> None:
        super().__init__()
        self.started_with: tuple | None = None
        self.killed = False

    def start(self, executable, working_dir, environment) -> None:
        self.started_with = (executable, working_dir, environment)

    def kill(self) -> None:
        self.killed = True


class FakeConnection(QObject):
    """Answers each request with the recorded response for its command."""

    connected = pyqtSignal()
    event_received = pyqtSignal(dict)
    failed = pyqtSignal(object)
    closed = pyqtSignal()

    def __init__(self) -> None:
        super().__init__()
        self.responses: dict[str, dict] = {}
        self.sent: list[tuple[str, dict | None]] = []
        self.events_before_response: dict[str, list[dict]] = {}

    def load(self, messages: list[dict]) -> None:
        """Events recorded before the launch response are replayed while launching."""
        pending: list[dict] = []
        for message in messages:
            if message.get("type") == "event":
                pending.append(message)
                continue
            self.responses.setdefault(message["command"], message)
            if message["command"] == "launch":
                self.events_before_response["launch"], pending = pending, []

    def request(self, command, arguments=None, on_response=None) -> int:
        self.sent.append((command, arguments))
        for event in self.events_before_response.get(command, []):
            self.event_received.emit(event)
        response = self.responses.get(command, {"success": True, "command": command})
        if on_response is not None:
            on_response(response)
        return len(self.sent)

    def request_blocking(self, command, arguments=None) -> dict:
        self.sent.append((command, arguments))
        return self.responses.get(command, {"success": True, "command": command})

    def commands(self) -> list[str]:
        return [command for command, _arguments in self.sent]

    def close(self) -> None:
        return


class FakeSession:
    instances: list["FakeSession"] = []

    def __init__(self, owner, on_state) -> None:
        self.process = FakeProcess()
        self.connection = FakeConnection()
        self.inspector = StopInspector(self.connection, on_state)
        self.closed = False
        FakeSession.instances.append(self)

    def close(self) -> None:
        self.inspector.cancel()
        self.connection.request("disconnect", {"terminateDebuggee": True})
        self.closed = True


@pytest.fixture
def fake_sessions(monkeypatch) -> list[FakeSession]:
    FakeSession.instances = []
    monkeypatch.setattr(debugger_module, "DapSession", FakeSession)
    return FakeSession.instances
