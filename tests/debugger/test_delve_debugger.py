"""DelveDapDebugger driven by recorded transcripts through a fake DAP session (no dlv)."""

from pathlib import Path

from PyQt5.QtCore import QMetaMethod, pyqtBoundSignal

from vizcacha.application.ports import DEBUGGER_SIGNALS, TERMINATED_BY_USER, DebuggerPort
from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.domain.debugging import Breakpoint, StopReason
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.infrastructure.delve_dap.adapter_process import INSTALL_COMMAND

QT_TYPES = {str: "QString", int: "int", object: "PyQt_PyObject"}


def _start(adapter, sessions, transcript, source: Path, lines=(13,)):
    breakpoints = [Breakpoint(SourceLocation(source, line)) for line in lines]
    adapter.start(RunConfiguration.for_file(source), breakpoints)
    connection = sessions[-1].connection
    connection.load(transcript)
    connection.connected.emit()
    return connection


def _event(transcript, name: str, index: int = 0) -> dict:
    return [message for message in transcript if message.get("event") == name][index]


def test_follows_the_debugger_port_contract(adapter):
    assert isinstance(adapter, DebuggerPort)
    meta = adapter.metaObject()
    signatures = {
        bytes(meta.method(i).methodSignature()).decode()
        for i in range(meta.methodCount())
        if meta.method(i).methodType() == QMetaMethod.Signal
    }
    for name, types in DEBUGGER_SIGNALS.items():
        assert isinstance(getattr(adapter, name), pyqtBoundSignal)
        assert f"{name}({','.join(QT_TYPES[t] for t in types)})" in signatures


def test_handshake_sends_breakpoints_then_configuration_done(
    adapter, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "functions.go")

    assert connection.commands() == ["initialize", "launch", "setBreakpoints", "configurationDone"]
    breakpoints = connection.sent[2][1]
    assert breakpoints["breakpoints"] == [{"line": 13}]
    assert adapter.is_active()


def test_stopped_event_becomes_a_debug_state(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "functions.go")

    connection.event_received.emit(_event(functions_transcript, "stopped"))

    state = recorder["stopped"][0]
    assert state.reason is StopReason.BREAKPOINT
    assert state.current_location == SourceLocation(source_dir / "functions.go", 13)
    assert {v.name: v.value for v in state.variables}["result"] == "35"
    assert len(state.goroutines) == 6 and state.current_goroutine == 1
    inspection = connection.commands()[4:]
    assert inspection[:4] == ["threads", "stackTrace", "scopes", "variables"]
    assert inspection[4:] == ["stackTrace"] * 5  # top frame of each other goroutine


def test_steps_and_resume_use_the_stopped_goroutine(
    adapter, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "functions.go")
    connection.event_received.emit(_event(functions_transcript, "stopped"))

    for operation in (adapter.step_over, adapter.step_into, adapter.step_out, adapter.resume):
        operation()

    assert connection.sent[-4:] == [
        ("next", {"threadId": 1}),
        ("stepIn", {"threadId": 1}),
        ("stepOut", {"threadId": 1}),
        ("continue", {"threadId": 1}),
    ]


def test_run_to_sets_a_temporary_breakpoint(
    adapter, fake_sessions, functions_transcript, source_dir
):
    source = source_dir / "functions.go"
    connection = _start(adapter, fake_sessions, functions_transcript, source)
    connection.event_received.emit(_event(functions_transcript, "stopped"))

    adapter.run_to(SourceLocation(source, 44))
    temporary = connection.sent[-2][1]["breakpoints"]
    connection.event_received.emit(_event(functions_transcript, "stopped", 1))
    restored = [args for command, args in connection.sent if command == "setBreakpoints"][-1]

    assert temporary == [{"line": 13}, {"line": 44}]
    assert restored["breakpoints"] == [{"line": 13}]
    assert ("continue", {"threadId": 1}) in connection.sent


def test_breakpoints_change_during_the_session(
    adapter, fake_sessions, functions_transcript, source_dir
):
    source = source_dir / "functions.go"
    connection = _start(adapter, fake_sessions, functions_transcript, source)

    adapter.set_breakpoints(source, [20, 13])

    assert connection.sent[-1] == ("setBreakpoints", connection.sent[-1][1])
    assert connection.sent[-1][1]["breakpoints"] == [{"line": 13}, {"line": 20}]


def test_program_output_and_exit(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "functions.go")

    for message in functions_transcript:
        if message.get("type") == "event" and message["event"] in (
            "output",
            "exited",
            "terminated",
        ):
            connection.event_received.emit(message)

    assert ("5 * 7 = 35\n", "stdout") in recorder["output"]
    assert not any("dlv help" in text for text, _category in recorder["output"])
    assert recorder["terminated"] == [0]
    assert not adapter.is_active()
    assert fake_sessions[-1].closed


def test_build_error_is_shown_as_stderr(
    adapter, recorder, fake_sessions, panic_transcript, source_dir
):
    _start(adapter, fake_sessions, panic_transcript["build_error"], source_dir / "bad.go")

    errors = [text for text, category in recorder["output"] if category == "stderr"]
    assert "declared and not used: x" in errors[0]
    assert "Failed to launch" in errors[-1]
    assert recorder["terminated"] == [1]


def test_stop_disconnects_and_terminates(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    _start(adapter, fake_sessions, functions_transcript, source_dir / "functions.go")

    adapter.stop()
    adapter.stop()

    assert fake_sessions[-1].connection.sent[-1] == ("disconnect", {"terminateDebuggee": True})
    assert recorder["terminated"] == [TERMINATED_BY_USER]


def test_missing_dlv_reports_install_instructions(adapter, recorder, settings, tmp_path):
    settings.set(SettingsKeys.DELVE_PATH, str(tmp_path / "missing" / "dlv"))

    adapter.start(RunConfiguration.for_file(tmp_path / "main.go"), [])

    assert INSTALL_COMMAND in recorder["output"][0][0]
    assert recorder["output"][0][1] == "stderr"
    assert recorder["terminated"] == [1]
    assert not adapter.is_active()
