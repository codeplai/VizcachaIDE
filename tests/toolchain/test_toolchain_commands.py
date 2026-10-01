"""Commands GoToolchain builds, with a simulated QProcess (no Go needed)."""

from pathlib import Path

import pytest
from PyQt5.QtCore import QProcess

from vizcacha.domain.project import RunConfiguration
from vizcacha.infrastructure.go_toolchain import (
    GoCommandArgumentError,
    GoEnvironment,
    GoToolchain,
    mod_get_arguments,
    mod_init_arguments,
    mod_tidy_arguments,
)
from vizcacha.infrastructure.go_toolchain import toolchain as toolchain_module
from vizcacha.infrastructure.settings import InMemorySettingsRepository


class FakeProcess(QProcess):
    started: list["FakeProcess"] = []

    def start(self, program, arguments):
        self.program = program
        self.arguments = list(arguments)
        self.calls: list[str] = []
        self.fake_state = QProcess.Running
        FakeProcess.started.append(self)

    def state(self):
        return getattr(self, "fake_state", QProcess.NotRunning)

    def processId(self):  # noqa: N802 - Qt override
        return 4242

    def kill(self):
        self.calls.append("kill")

    def terminate(self):
        self.calls.append("terminate")

    def finish(self, code: int) -> None:
        self.fake_state = QProcess.NotRunning
        self.finished.emit(code, QProcess.NormalExit)


@pytest.fixture
def fake_process(monkeypatch):
    FakeProcess.started = []
    monkeypatch.setattr(toolchain_module, "QProcess", FakeProcess)
    return FakeProcess


@pytest.fixture
def system_calls(monkeypatch):
    calls: list[tuple] = []
    monkeypatch.setattr(
        toolchain_module, "kill_tree_windows", lambda pid: calls.append(("tree", pid))
    )
    monkeypatch.setattr(
        toolchain_module,
        "signal_children_posix",
        lambda pid, name: calls.append((name, pid)),
    )
    return calls


@pytest.fixture
def toolchain(fake_process, tmp_path):
    settings = InMemorySettingsRepository({"env/go_path": "/fake/go"})
    environment = GoEnvironment(settings, {"PATH": ""}, app_directory=tmp_path / "app")
    return GoToolchain(environment)


def _last() -> FakeProcess:
    return FakeProcess.started[-1]


def test_run_file_passes_program_arguments(toolchain, tmp_path: Path):
    toolchain.run(RunConfiguration.for_file(tmp_path / "hello.go", ("a", "b c")))

    process = _last()
    assert process.program == "/fake/go"
    assert process.arguments == ["run", "hello.go", "a", "b c"]
    assert Path(process.workingDirectory()) == tmp_path


def test_build_writes_named_binary_in_working_dir(toolchain, tmp_path: Path, monkeypatch):
    monkeypatch.setattr(toolchain_module, "IS_WINDOWS", True)

    toolchain.build(RunConfiguration.for_file(tmp_path / "hello.go"))

    assert _last().arguments == ["build", "-o", "hello.exe", "hello.go"]


def test_second_process_is_refused(toolchain, tmp_path: Path):
    errors: list[str] = []
    toolchain.error_received.connect(errors.append)
    toolchain.run(RunConfiguration.for_file(tmp_path / "a.go"))

    assert toolchain.run_go_command(tmp_path, ["mod", "tidy"]) is False
    assert len(FakeProcess.started) == 1
    assert "already running" in "".join(errors)


def test_untitled_source_runs_from_a_temporary_directory(toolchain):
    config = toolchain.run_untitled("package main\n", ("x",))

    scratch = config.working_dir
    assert config.target.read_text(encoding="utf-8") == "package main\n"
    assert _last().arguments == ["run", "main.go", "x"]
    assert Path(_last().workingDirectory()) == scratch

    _last().finish(0)

    assert not scratch.exists()


def test_module_command_streams_header_and_uses_folder(toolchain, tmp_path: Path):
    output: list[str] = []
    toolchain.output_received.connect(output.append)

    assert toolchain.run_go_command(tmp_path, mod_tidy_arguments()) is True

    assert _last().arguments == ["mod", "tidy"]
    assert Path(_last().workingDirectory()) == tmp_path
    assert "go mod tidy" in output[0]


def test_stop_on_windows_kills_process_tree(toolchain, tmp_path, system_calls, monkeypatch):
    monkeypatch.setattr(toolchain_module, "IS_WINDOWS", True)
    toolchain.run(RunConfiguration.for_file(tmp_path / "a.go"))

    toolchain.stop()

    assert system_calls == [("tree", 4242)]
    assert _last().calls == ["kill"]


def test_stop_on_posix_terminates_then_kills(qtbot, toolchain, tmp_path, system_calls, monkeypatch):
    monkeypatch.setattr(toolchain_module, "IS_WINDOWS", False)
    monkeypatch.setattr(toolchain_module, "STOP_GRACE_MS", 10)
    toolchain.run(RunConfiguration.for_file(tmp_path / "a.go"))

    toolchain.stop()

    assert _last().calls == ["terminate"]
    qtbot.waitUntil(lambda: "kill" in _last().calls, timeout=2000)
    assert system_calls == [("TERM", 4242), ("KILL", 4242)]


def test_module_arguments_are_validated():
    assert mod_init_arguments(" example.com/hi ") == ["mod", "init", "example.com/hi"]
    assert mod_get_arguments("github.com/google/uuid@v1.6.0") == [
        "get",
        "github.com/google/uuid@v1.6.0",
    ]
    for bad in ("", "  ", "-toolexec=evil", "two words"):
        with pytest.raises(GoCommandArgumentError):
            mod_get_arguments(bad)
