"""Module detection and RunConfiguration building (application/run_program.py)."""

from pathlib import Path

import pytest

from vizcacha.application.run_program import (
    ProgramArgumentsError,
    configuration_for_file,
    find_go_module,
    parse_module_path,
    split_program_arguments,
)
from vizcacha.domain.project import GoModule, RunTarget


@pytest.mark.parametrize(
    ("text", "expected"),
    [
        ("module example.com/hello\n\ngo 1.22\n", "example.com/hello"),
        ('// comment\nmodule "quoted/name" // trailing\n', "quoted/name"),
        ("go 1.22\n", ""),
    ],
)
def test_parse_module_path(text, expected):
    assert parse_module_path(text) == expected


def test_find_go_module_walks_up(tmp_path: Path):
    (tmp_path / "go.mod").write_text("module example.com/app\n", encoding="utf-8")
    nested = tmp_path / "internal" / "greet"
    nested.mkdir(parents=True)

    assert find_go_module(nested) == GoModule(tmp_path, "example.com/app")


def test_loose_file_runs_in_file_mode(tmp_path: Path):
    source = tmp_path / "hello.go"

    config = configuration_for_file(source, ("a",))

    assert config.mode is RunTarget.FILE
    assert config.go_target_argument() == "hello.go"
    assert config.working_dir == tmp_path
    assert config.program_args == ("a",)
    assert config.module is None


def test_file_in_module_runs_its_package(tmp_path: Path):
    (tmp_path / "go.mod").write_text("module example.com/app\n", encoding="utf-8")
    source = tmp_path / "main.go"

    config = configuration_for_file(source, ("x", "y"))

    assert config.mode is RunTarget.PACKAGE
    assert config.go_target_argument() == "."
    assert config.working_dir == tmp_path
    assert config.module == GoModule(tmp_path, "example.com/app")
    assert config.program_args == ("x", "y")


def test_file_in_subpackage_runs_that_package(tmp_path: Path):
    (tmp_path / "go.mod").write_text("module example.com/app\n", encoding="utf-8")
    tool = tmp_path / "cmd" / "tool"
    tool.mkdir(parents=True)

    config = configuration_for_file(tool / "main.go")

    assert config.mode is RunTarget.PACKAGE
    assert config.working_dir == tool
    assert config.module.root == tmp_path
    assert config.executable_name(windows=True) == "tool.exe"


@pytest.mark.parametrize(
    ("text", "expected"),
    [
        ("", ()),
        ("  one   two ", ("one", "two")),
        ('"hello world" x', ("hello world", "x")),
        (r"C:\Users\me\data.txt", (r"C:\Users\me\data.txt",)),
        ("say 'hi there'", ("say", "hi there")),
    ],
)
def test_split_program_arguments(text, expected):
    assert split_program_arguments(text) == expected


def test_unclosed_quote_is_a_typed_error():
    with pytest.raises(ProgramArgumentsError):
        split_program_arguments('"oops')
