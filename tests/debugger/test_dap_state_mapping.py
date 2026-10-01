from pathlib import Path

import pytest

from vizcacha.domain.debugging import StopReason, Variable
from vizcacha.infrastructure.delve_dap import state_mapping


def _response(messages: list[dict], command: str) -> dict:
    return next(m for m in messages if m.get("command") == command)["body"]


@pytest.mark.parametrize(
    ("reason", "expected"),
    [
        ("breakpoint", StopReason.BREAKPOINT),
        ("function breakpoint", StopReason.BREAKPOINT),
        ("step", StopReason.STEP),
        ("entry", StopReason.ENTRY),
        ("exception", StopReason.PANIC),
        ("pause", StopReason.PAUSE),
        ("something new", StopReason.PAUSE),
    ],
)
def test_stop_reasons(reason, expected):
    assert state_mapping.stop_reason(reason) is expected


def test_frames_from_recorded_stack_trace(functions_transcript, source_dir: Path):
    frames = state_mapping.map_frames(_response(functions_transcript, "stackTrace"))

    assert [frame.function for frame in frames] == ["main.multiply", "main.main", "runtime.main"]
    assert frames[0].frame_id == 1000
    assert frames[0].location.file == source_dir / "functions.go"
    assert (frames[0].location.line, frames[0].location.column) == (13, 1)


def test_panic_skips_leading_runtime_frames(panic_transcript, source_dir: Path):
    body = panic_transcript["panic"]["stackTrace"]["body"]

    frames = state_mapping.map_frames(body, skip_runtime=True)

    assert frames[0].function == "main.main"
    assert frames[0].location.file == source_dir / "panic.go"
    assert frames[-1].function == "runtime.main"


def test_panic_description(panic_transcript):
    body = panic_transcript["panic"]["stopped"]["body"]

    description = state_mapping.stop_description(body)

    assert description == "panic: runtime error: index out of range [3] with length 0"


def test_variables_keep_reference_for_lazy_expansion(functions_transcript, panic_transcript):
    simple = state_mapping.map_variables(_response(functions_transcript, "variables"))
    nested = state_mapping.map_variables(panic_transcript["panic"]["long_variable"]["body"])

    assert simple[0] == Variable("a", "int", "5")
    assert [variable.name for variable in simple] == ["a", "b", "~r0", "result"]
    assert nested[0].reference == 1001
    assert nested[0].has_children


def test_long_values_are_truncated(panic_transcript):
    raw = panic_transcript["panic"]["long_variable"]["body"]["variables"][0]

    value = state_mapping.map_variable(raw).value

    assert len(raw["value"]) > state_mapping.MAX_VALUE_LENGTH
    assert len(value) == state_mapping.MAX_VALUE_LENGTH
    assert value.endswith(state_mapping.ELLIPSIS)


def test_locals_scope_even_for_optimized_functions(functions_transcript, panic_transcript):
    optimized = panic_transcript["panic"]["optimized_scopes"]["body"]

    assert state_mapping.locals_reference(_response(functions_transcript, "scopes")) == 1000
    assert state_mapping.locals_reference(optimized) == 1000
    assert state_mapping.locals_reference({"scopes": [{"name": "Globals"}]}) == 0


def test_goroutines_from_threads(functions_transcript):
    goroutines = state_mapping.map_goroutines(_response(functions_transcript, "threads"))

    assert len(goroutines) == 6
    assert goroutines[0].goroutine_id == 1
    assert goroutines[0].name == "[Go 1] main.multiply (Thread 41936)"


@pytest.mark.parametrize(
    ("body", "expected"),
    [
        ({"category": "stdout", "output": "hi\n"}, ("hi\n", "stdout")),
        ({"category": "stderr", "output": "boom\n"}, ("boom\n", "stderr")),
        ({"category": "important", "output": "x\n"}, ("x\n", "console")),
        ({"category": "console", "output": "Type 'dlv help' for list of commands.\n"}, None),
        ({"category": "telemetry", "output": "{}"}, None),
    ],
)
def test_output_events(body, expected):
    assert state_mapping.output_event(body) == expected
