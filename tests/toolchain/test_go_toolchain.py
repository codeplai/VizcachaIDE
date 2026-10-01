"""Integration tests that need a real Go installation."""

from pathlib import Path

import pytest

from vizcacha.application.errors import GoFormatError
from vizcacha.domain.project import RunConfiguration
from vizcacha.infrastructure.go_toolchain import GoEnvironment, GoToolchain

# Generous: linking is slow on Windows machines with antivirus scanning.
GO_TIMEOUT_MS = 240_000
HELLO = 'package main\n\nimport "fmt"\n\nfunc main() {\n\tfmt.Println("hola")\n}\n'


@pytest.fixture
def toolchain(settings):
    return GoToolchain(GoEnvironment(settings))


@pytest.mark.requires_go
def test_run_streams_output_and_exit_code(qtbot, toolchain, tmp_path: Path):
    source = tmp_path / "hello.go"
    source.write_text(HELLO, encoding="utf-8")
    output: list[str] = []
    toolchain.output_received.connect(output.append)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS) as finished:
        toolchain.run(RunConfiguration.for_file(source))

    assert finished.args == [0]
    assert "hola" in "".join(output)


@pytest.mark.requires_go
def test_compile_error_goes_to_stderr(qtbot, toolchain, tmp_path: Path):
    source = tmp_path / "broken.go"
    source.write_text("package main\n\nfunc main() {\n\tx := 1\n}\n", encoding="utf-8")
    errors: list[str] = []
    toolchain.error_received.connect(errors.append)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS) as finished:
        toolchain.run(RunConfiguration.for_file(source))

    assert finished.args[0] != 0
    assert "declared and not used" in "".join(errors)


@pytest.mark.requires_go
def test_program_reads_stdin(qtbot, toolchain, tmp_path: Path):
    source = tmp_path / "echo.go"
    source.write_text(
        'package main\n\nimport "fmt"\n\nfunc main() {\n\tvar s string\n'
        '\tfmt.Scanln(&s)\n\tfmt.Println("eco:", s)\n}\n',
        encoding="utf-8",
    )
    output: list[str] = []
    toolchain.output_received.connect(output.append)

    with qtbot.waitSignal(toolchain.execution_finished, timeout=GO_TIMEOUT_MS):
        toolchain.run(RunConfiguration.for_file(source))
        toolchain.write_input("vizcacha")  # QProcess buffers it until the program reads

    assert "eco: vizcacha" in "".join(output)


@pytest.mark.requires_go
def test_format_source(toolchain):
    assert toolchain.format_source("package main\nfunc main(){\nprintln(1)}\n") == (
        "package main\n\nfunc main() {\n\tprintln(1)\n}\n"
    )


@pytest.mark.requires_go
def test_format_source_rejects_invalid_code(toolchain):
    with pytest.raises(GoFormatError):
        toolchain.format_source("package main\nfunc main( {\n")
