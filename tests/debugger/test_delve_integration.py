"""Real ``dlv dap`` session on examples/functions.go.

The first debug build compiles the standard library with ``-gcflags all=-N -l``;
on Windows with an antivirus this can take more than a minute, hence the long timeout.
"""

import shutil
from pathlib import Path

import pytest

from vizcacha.domain.debugging import Breakpoint, StopReason
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.infrastructure.delve_dap import DelveDapDebugger
from vizcacha.infrastructure.delve_dap.session import KILL_DELAY_MS
from vizcacha.infrastructure.go_toolchain import GoEnvironment

EXAMPLE = Path(__file__).resolve().parents[2] / "examples" / "functions.go"
BUILD_TIMEOUT_MS = 300_000
STEP_TIMEOUT_MS = 120_000  # generous: CI machines and antivirus scans are slow
BREAKPOINT_LINE = 13  # "return result" inside multiply(5, 7)
POINT_PROGRAM = """package main

import "fmt"

type Point struct{ X, Y int }

func main() {
\tp := Point{X: 3, Y: 4}
\tfmt.Println(p)
}
"""
POINT_LINE = 9  # fmt.Println(p)


@pytest.mark.requires_dlv
@pytest.mark.requires_go
def test_breakpoint_step_and_variables(qtbot, settings, tmp_path: Path):
    source = tmp_path / EXAMPLE.name
    shutil.copy(EXAMPLE, source)
    debugger = DelveDapDebugger(GoEnvironment(settings))
    states, outputs, exits = [], [], []
    debugger.stopped.connect(states.append)
    debugger.output.connect(lambda text, _category: outputs.append(text))
    debugger.terminated.connect(exits.append)

    breakpoint_at = [Breakpoint(SourceLocation(source, BREAKPOINT_LINE))]
    debugger.start(RunConfiguration.for_file(source), breakpoint_at)
    try:
        qtbot.waitUntil(lambda: bool(states or exits), timeout=BUILD_TIMEOUT_MS)
        assert not exits, "".join(outputs)
        first = states[0]
        assert first.reason is StopReason.BREAKPOINT
        assert first.current_location.line == BREAKPOINT_LINE
        assert {v.name: v.value for v in first.variables}["result"] == "35"
        assert first.goroutines
        assert first.goroutines[0].location is not None

        debugger.step_over()
        qtbot.waitUntil(lambda: len(states) > 1, timeout=STEP_TIMEOUT_MS)
        assert states[1].reason is StopReason.STEP
        assert states[1].frames[0].function == "main.main"

        debugger.resume()
        qtbot.waitUntil(lambda: bool(exits), timeout=STEP_TIMEOUT_MS)
    finally:
        debugger.stop()
        qtbot.wait(KILL_DELAY_MS + 500)  # let the session release dlv
    assert exits[0] == 0
    assert "5 * 7 = 35\n" in "".join(outputs)


@pytest.mark.requires_dlv
@pytest.mark.requires_go
def test_variable_children_load_asynchronously(qtbot, settings, tmp_path: Path):
    source = tmp_path / "point.go"
    source.write_text(POINT_PROGRAM, encoding="utf-8")
    debugger = DelveDapDebugger(GoEnvironment(settings))
    breakpoint_at = [Breakpoint(SourceLocation(source, POINT_LINE))]
    try:
        with qtbot.waitSignal(debugger.stopped, timeout=BUILD_TIMEOUT_MS) as stop:
            debugger.start(RunConfiguration.for_file(source), breakpoint_at)
        point = next(v for v in stop.args[0].variables if v.name == "p")
        assert point.reference != 0

        with qtbot.waitSignal(debugger.variables_loaded, timeout=STEP_TIMEOUT_MS) as loaded:
            debugger.request_variables(point.reference)
        reference, children = loaded.args
        assert reference == point.reference
        assert {child.name: child.value for child in children} == {"X": "3", "Y": "4"}
    finally:
        debugger.stop()
        qtbot.wait(KILL_DELAY_MS + 500)  # let the session release dlv
