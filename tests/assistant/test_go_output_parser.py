from pathlib import Path, PureWindowsPath

from vizcacha.domain.diagnostics import Severity, SourceLocation
from vizcacha.infrastructure.error_catalog import ErrorCatalog, GoOutputParser, resolve_go_path

WORKDIR = Path("/home/ana/hello")
parser = GoOutputParser(ErrorCatalog().identify)


def test_relative_posix_and_windows_paths_resolve_against_working_dir():
    raw = (
        "# command-line-arguments\n"
        "./main.go:5:2: declared and not used: x\n"
        ".\\util\\calc.go:7:14: undefined: total\n"
    )

    first, second = parser.parse(raw, WORKDIR)

    assert first.location == SourceLocation(WORKDIR / "main.go", 5, 2)
    assert first.message == "declared and not used: x"
    assert first.raw_text == "./main.go:5:2: declared and not used: x"
    assert (first.code, first.source) == ("E-UNUSED-VAR", "compiler")
    assert first.severity is Severity.ERROR
    assert second.location == SourceLocation(WORKDIR / "util" / "calc.go", 7, 14)
    assert second.code == "E-UNDEFINED"


def test_absolute_windows_path_with_drive_and_spaces():
    raw = 'C:\\Users\\Ana Perez\\go\\main.go:3:8: "os" imported and not used\r\n'

    (diagnostic,) = parser.parse(raw, WORKDIR)

    assert PureWindowsPath(str(diagnostic.location.file)) == PureWindowsPath(
        "C:/Users/Ana Perez/go/main.go"
    )
    assert (diagnostic.location.line, diagnostic.location.column) == (3, 8)
    assert diagnostic.code == "E-UNUSED-IMPORT"


def test_absolute_posix_path_is_kept():
    (diagnostic,) = parser.parse("/tmp/x/main.go:10:1: missing return\n", WORKDIR)

    assert diagnostic.location == SourceLocation(Path("/tmp/x/main.go"), 10, 1)
    assert diagnostic.code == "E-MISSING-RETURN"


def test_tab_lines_continue_the_previous_message():
    raw = "./main.go:11:18: not enough arguments in call to add\n\thave (number)\n\twant (int)\n"

    (diagnostic,) = parser.parse(raw, WORKDIR)

    assert diagnostic.message == "not enough arguments in call to add"
    assert diagnostic.raw_text.endswith("\thave (number)\n\twant (int)")


def test_vet_header_marks_warnings():
    raw = (
        "# command-line-arguments\n# [command-line-arguments]\n"
        "./main.go:9:2: unreachable code\n./main.go:7:2: something vet found\n"
    )

    first, second = parser.parse(raw, WORKDIR)

    assert (first.source, first.severity, first.code) == ("vet", Severity.WARNING, "V-UNREACHABLE")
    assert (second.source, second.code) == ("vet", "")


def test_lines_without_location_are_kept_when_recognised():
    raw = (
        "# command-line-arguments\n"
        "runtime.main_main·f: function main is undeclared in the main package\n"
    )

    (diagnostic,) = parser.parse(raw, WORKDIR)

    assert diagnostic.location is None
    assert diagnostic.code == "E-NO-MAIN"


def test_unrecognised_output_is_ignored_but_go_command_errors_are_kept():
    raw = "hello from the program\ngo: cannot find main module\n"

    (diagnostic,) = parser.parse(raw, WORKDIR)

    assert diagnostic.message == "go: cannot find main module"
    assert diagnostic.code == ""


def test_panic_location_is_first_user_frame_not_runtime():
    raw = (
        "panic: runtime error: index out of range [5] with length 3\n"
        "\n"
        "goroutine 1 [running]:\n"
        "panic({0x4a1f20?, 0xc000012345?})\n"
        "\t/usr/local/go/src/runtime/panic.go:787 +0x132\n"
        "main.pick(...)\n"
        "\t/home/ana/hello/main.go:12 +0x1d\n"
        "main.main()\n"
        "\t/home/ana/hello/main.go:20 +0x2a\n"
        "exit status 2\n"
    )

    (diagnostic,) = parser.parse(raw, WORKDIR)

    assert diagnostic.location == SourceLocation(Path("/home/ana/hello/main.go"), 12, 1)
    assert diagnostic.message == "panic: runtime error: index out of range [5] with length 3"
    assert (diagnostic.source, diagnostic.code) == ("panic", "P-INDEX-RANGE")
    assert "goroutine 1 [running]:" in diagnostic.raw_text
    assert "exit status" not in diagnostic.raw_text


def test_windows_panic_frame_with_forward_slashes():
    raw = (
        "fatal error: all goroutines are asleep - deadlock!\n\n"
        "goroutine 1 [chan send]:\nmain.main()\n"
        "\tC:/Users/ana/hello/main.go:6 +0x28\nexit status 2\n"
    )

    (diagnostic,) = parser.parse(raw, WORKDIR)

    assert PureWindowsPath(str(diagnostic.location.file)) == PureWindowsPath(
        "C:/Users/ana/hello/main.go"
    )
    assert diagnostic.location.line == 6
    assert diagnostic.code == "P-DEADLOCK"


def test_resolve_go_path():
    assert resolve_go_path("./a.go", WORKDIR) == WORKDIR / "a.go"
    assert resolve_go_path("..\\b.go", WORKDIR) == WORKDIR / ".." / "b.go"
    assert resolve_go_path("/abs/c.go", WORKDIR) == Path("/abs/c.go")
