from pathlib import Path

import pytest

from vizcacha.application.errors import DebugAdapterNotFoundError
from vizcacha.domain.debugging import Breakpoint
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration, RunTarget
from vizcacha.infrastructure.delve_dap import launch_arguments
from vizcacha.infrastructure.delve_dap.adapter_process import (
    INSTALL_COMMAND,
    parse_listening_port,
    resolve_executable,
)
from vizcacha.infrastructure.delve_dap.breakpoint_registry import BreakpointRegistry


def test_launch_for_a_single_file(tmp_path: Path):
    source = tmp_path / "main.go"
    config = RunConfiguration.for_file(source, program_args=("-n", "3"))

    arguments = launch_arguments.launch_arguments(config, {"GOPATH": "/g"}, "/tmp/bin")

    assert arguments["mode"] == "debug"
    assert Path(arguments["program"]) == source.resolve()
    assert arguments["cwd"] == str(tmp_path)
    assert arguments["args"] == ["-n", "3"]
    assert arguments["env"] == {"GOPATH": "/g"}
    assert arguments["outputMode"] == "remote"
    assert arguments["output"] == "/tmp/bin"


def test_launch_for_a_package_uses_the_folder(tmp_path: Path):
    config = RunConfiguration(tmp_path / "main.go", tmp_path, mode=RunTarget.PACKAGE)

    arguments = launch_arguments.launch_arguments(config, {}, "out")

    assert Path(arguments["program"]) == tmp_path.resolve()


def test_debug_binary_goes_to_the_temp_folder():
    assert launch_arguments.debug_binary_path(42, windows=True).endswith("_42.exe")
    assert launch_arguments.debug_binary_path(42, windows=False).endswith("_42")


def test_set_breakpoints_arguments(tmp_path: Path):
    source = tmp_path / "main.go"

    arguments = launch_arguments.set_breakpoints_arguments(source, [9, 3, 9])

    assert arguments["source"]["path"] == str(source)
    assert arguments["breakpoints"] == [{"line": 3}, {"line": 9}]


def test_listening_line_from_dlv():
    assert parse_listening_port("DAP server listening at: 127.0.0.1:41353") == ("127.0.0.1", 41353)
    assert parse_listening_port("API server listening at: 127.0.0.1:1") is None


def test_missing_dlv_explains_how_to_install(tmp_path: Path):
    with pytest.raises(DebugAdapterNotFoundError) as raised:
        resolve_executable(str(tmp_path / "nowhere" / "dlv"))

    assert INSTALL_COMMAND in str(raised.value)


def test_registry_adds_and_removes_the_run_to_cursor_breakpoint(tmp_path: Path):
    source = tmp_path / "main.go"
    registry = BreakpointRegistry()
    registry.reset([Breakpoint(SourceLocation(source, 5))])

    file = registry.set_temporary(SourceLocation(source, 12))

    assert registry.lines_for(source) == {5, 12}
    assert registry.clear_temporary() == file
    assert registry.lines_for(source) == {5}
    assert registry.clear_temporary() is None


def test_registry_replaces_lines_of_one_file(tmp_path: Path):
    registry = BreakpointRegistry()
    registry.reset([Breakpoint(SourceLocation(tmp_path / "a.go", 1))])

    registry.replace(tmp_path / "b.go", [4, 2])

    assert registry.lines_for(tmp_path / "a.go") == {1}
    assert registry.lines_for(tmp_path / "b.go") == {2, 4}
    assert len(registry.files()) == 2
