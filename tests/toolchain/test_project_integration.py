"""Integration with a real Go: modules, Untitled tabs, program arguments, go mod."""

from pathlib import Path

import pytest

from vizcacha.application.run_program import configuration_for_file, find_go_module
from vizcacha.infrastructure.go_toolchain import (
    GoEnvironment,
    GoToolchain,
    mod_init_arguments,
    mod_tidy_arguments,
)

GO_TIMEOUT_MS = 240_000  # linking is slow on Windows machines with antivirus scanning
MAIN = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println(greeting())\n}\n'
GREETING = 'package main\n\nfunc greeting() string {\n\treturn "hola desde otro archivo"\n}\n'
ARGS = (
    'package main\n\nimport (\n\t"fmt"\n\t"os"\n)\n\n'
    "func main() {\n\tfmt.Println(len(os.Args)-1, os.Args[1:])\n}\n"
)


@pytest.fixture
def toolchain(settings):
    return GoToolchain(GoEnvironment(settings))


def _collect(toolchain) -> list[str]:
    chunks: list[str] = []
    toolchain.output_received.connect(chunks.append)
    toolchain.error_received.connect(chunks.append)
    return chunks


@pytest.mark.requires_go
def test_module_with_two_files_runs_as_package(qtbot, toolchain, tmp_path: Path):
    (tmp_path / "go.mod").write_text("module example.com/twofiles\n\ngo 1.21\n", encoding="utf-8")
    (tmp_path / "main.go").write_text(MAIN, encoding="utf-8")
    (tmp_path / "greeting.go").write_text(GREETING, encoding="utf-8")
    output = _collect(toolchain)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS) as finished:
        toolchain.run(configuration_for_file(tmp_path / "main.go"))

    assert finished.args == [0], "".join(output)
    assert "hola desde otro archivo" in "".join(output)


@pytest.mark.requires_go
def test_untitled_source_runs_without_saving(qtbot, toolchain):
    output = _collect(toolchain)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS) as finished:
        config = toolchain.run_untitled(MAIN.replace("greeting()", '"sin guardar"'))

    assert finished.args == [0], "".join(output)
    assert "sin guardar" in "".join(output)
    assert not config.working_dir.exists()


@pytest.mark.requires_go
def test_program_arguments_reach_the_program(qtbot, toolchain, tmp_path: Path):
    source = tmp_path / "args.go"
    source.write_text(ARGS, encoding="utf-8")
    output = _collect(toolchain)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS):
        toolchain.run(configuration_for_file(source, ("uno", "dos tres")))

    assert "2 [uno dos tres]" in "".join(output)


@pytest.mark.requires_go
def test_go_mod_init_and_tidy(qtbot, toolchain, tmp_path: Path):
    (tmp_path / "main.go").write_text(MAIN.replace("greeting()", '"ok"'), encoding="utf-8")
    output = _collect(toolchain)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS) as finished:
        toolchain.run_go_command(tmp_path, mod_init_arguments("example.com/nuevo"))
    assert finished.args == [0], "".join(output)
    assert find_go_module(tmp_path).module_path == "example.com/nuevo"

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS) as finished:
        toolchain.run_go_command(tmp_path, mod_tidy_arguments())
    assert finished.args == [0], "".join(output)


HEARTBEAT = (
    'package main\n\nimport (\n\t"fmt"\n\t"os"\n\t"time"\n)\n\n'
    'func main() {\n\tfmt.Println("started")\n'
    '\tfor i := 0; ; i++ {\n\t\tos.WriteFile("beat.txt", []byte(fmt.Sprint(i)), 0o644)\n'
    "\t\ttime.Sleep(50 * time.Millisecond)\n\t}\n}\n"
)


@pytest.mark.requires_go
def test_stop_ends_the_program_not_only_go_run(qtbot, toolchain, tmp_path: Path):
    source = tmp_path / "beat.go"
    source.write_text(HEARTBEAT, encoding="utf-8")
    output = _collect(toolchain)
    toolchain.run(configuration_for_file(source))
    qtbot.waitUntil(lambda: "started" in "".join(output), timeout=GO_TIMEOUT_MS)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=10_000):
        toolchain.stop()

    beat = tmp_path / "beat.txt"
    qtbot.wait(300)
    last = beat.read_text(encoding="utf-8")
    qtbot.wait(500)
    assert beat.read_text(encoding="utf-8") == last
