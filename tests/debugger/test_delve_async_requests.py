"""Asynchronous variables and goroutine locations, replayed through the fake DAP session."""

from pathlib import Path

from vizcacha.domain.debugging import Breakpoint
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.infrastructure.delve_dap.stop_inspector import GOROUTINE_LOCATION_LIMIT

PARKED_LINE_OFFSET = 100  # fake goroutine N is parked at line 100 + N


def _start(adapter, sessions, transcript, source: Path):
    adapter.start(RunConfiguration.for_file(source), [Breakpoint(SourceLocation(source, 13))])
    connection = sessions[-1].connection
    connection.load(transcript)
    connection.connected.emit()
    return connection


def _stopped(transcript) -> dict:
    return next(message for message in transcript if message.get("event") == "stopped")


def _top_frame_responder(connection, parked_file: Path):
    """levels=1 → a parked frame per goroutine; the full stack → the recorded one."""
    recorded = connection.responses["stackTrace"]

    def respond(arguments: dict) -> dict:
        if arguments["levels"] != 1:
            return recorded
        line = PARKED_LINE_OFFSET + arguments["threadId"]
        frame = {"id": 9, "name": "runtime.gopark", "source": {"path": str(parked_file)}}
        return {"success": True, "body": {"stackFrames": [{**frame, "line": line}]}}

    return respond


def _threads(count: int):
    threads = [{"id": i, "name": f"[Go {i}] runtime.gopark"} for i in range(1, count + 1)]
    return lambda _arguments: {"success": True, "body": {"threads": threads}}


def test_variables_arrive_through_a_signal(
    adapter, recorder, fake_sessions, functions_transcript, panic_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "panic.go")
    connection.responses["variables"] = panic_transcript["panic"]["long_variable"]

    adapter.request_variables(1001)

    assert connection.sent[-1] == ("variables", {"variablesReference": 1001})
    [(reference, children)] = recorder["variables_loaded"]
    assert reference == 1001 and children[0].name == "msgs"


def test_requests_without_an_answer_from_delve_load_nothing(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    adapter.request_variables(7)  # no session yet
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "main.go")
    sent = len(connection.sent)
    adapter.request_variables(0)
    connection.responses["variables"] = {"success": False, "command": "variables"}
    adapter.request_variables(1001)

    assert recorder["variables_loaded"] == [(7, []), (0, []), (1001, [])]
    assert connection.commands()[sent:] == ["variables"]


def test_an_answer_that_arrives_after_a_step_is_dropped(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "main.go")
    connection.event_received.emit(_stopped(functions_transcript))
    connection.held = {"variables"}

    adapter.request_variables(1001)
    adapter.step_over()
    connection.release()

    assert recorder["variables_loaded"] == []


def test_goroutines_get_the_location_of_their_top_frame(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    source, parked = source_dir / "functions.go", source_dir / "proc.go"
    connection = _start(adapter, fake_sessions, functions_transcript, source)
    connection.responders["stackTrace"] = _top_frame_responder(connection, parked)

    connection.event_received.emit(_stopped(functions_transcript))

    goroutines = recorder["stopped"][0].goroutines
    assert goroutines[0].location == SourceLocation(source, 13)  # current: the user frame
    assert goroutines[1].location == SourceLocation(parked, PARKED_LINE_OFFSET + 2)
    assert all(goroutine.location is not None for goroutine in goroutines)


def test_goroutine_locations_are_limited(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "main.go")
    connection.responders["threads"] = _threads(GOROUTINE_LOCATION_LIMIT + 10)
    connection.responders["stackTrace"] = _top_frame_responder(connection, source_dir / "p.go")

    connection.event_received.emit(_stopped(functions_transcript))

    top_frames = [args for command, args in connection.sent if args and args.get("levels") == 1]
    goroutines = recorder["stopped"][0].goroutines
    assert len(top_frames) == GOROUTINE_LOCATION_LIMIT - 1  # the current one has its stack
    assert goroutines[GOROUTINE_LOCATION_LIMIT - 1].location is not None
    assert goroutines[GOROUTINE_LOCATION_LIMIT].location is None


def test_the_state_waits_for_every_goroutine_location(
    adapter, recorder, fake_sessions, functions_transcript, source_dir
):
    connection = _start(adapter, fake_sessions, functions_transcript, source_dir / "main.go")
    connection.held = {"stackTrace"}

    connection.event_received.emit(_stopped(functions_transcript))
    waiting = list(recorder["stopped"])
    connection.release()

    assert waiting == []
    assert len(recorder["stopped"]) == 1
