"""GoOutputParser: raw output of go build / go run / go vet and panics -> list[Diagnostic].

Recognised shapes::

    # command-line-arguments                       (package header, ignored)
    # [command-line-arguments]                     (go vet header: what follows is vet)
    ./main.go:12:5: declared and not used: x       (relative, absolute, Windows or POSIX)
    \thave (number)                                (continuation of the previous message)
    panic: runtime error: index out of range [5] with length 3
    goroutine 1 [running]:
    main.main()
    \t/home/ana/hello/main.go:9 +0x1d               (first user frame = location)
"""

import re
from collections.abc import Callable
from dataclasses import replace
from pathlib import Path

from vizcacha.domain.diagnostics import Diagnostic, Severity, SourceLocation
from vizcacha.infrastructure.error_catalog.source_links import location_from_match

COMPILER_LINE = re.compile(
    r"^(?P<path>(?:[A-Za-z]:)?[^:\r\n]+?\.go):(?P<line>\d+)(?::(?P<column>\d+))?:\s*(?P<message>.*)$"
)
FRAME_LINE = re.compile(r"^\t(?P<path>.+?\.go):(?P<line>\d+)(?: \+0x[0-9a-fA-F]+)?\s*$")
PANIC_START = re.compile(r"^(?:panic: |fatal error: )")
VET_HEADER = re.compile(r"^# \[.*\]\s*$")
NOISE = re.compile(r"^(?:# .*|exit status \d+|too many errors)\s*$")
RUNTIME_FUNCTIONS = ("runtime.", "runtime/", "panic(", "internal/")

Identify = Callable[[str], str]


def _unknown(_message: str) -> str:
    return ""


class GoOutputParser:
    """``identify(message)`` returns the catalog id of a message ("" when unknown)."""

    def __init__(self, identify: Identify = _unknown) -> None:
        self._identify = identify

    def parse(self, raw_output: str, working_dir: Path) -> list[Diagnostic]:
        lines = raw_output.splitlines()
        diagnostics: list[Diagnostic] = []
        source = "compiler"
        for index, line in enumerate(lines):
            if PANIC_START.match(line):
                diagnostics.append(self._panic(lines[index:], working_dir))
                break
            if VET_HEADER.match(line):
                source = "vet"
                continue
            if line.startswith("\t") and diagnostics:
                previous = diagnostics[-1]
                diagnostics[-1] = replace(previous, raw_text=previous.raw_text + "\n" + line)
                continue
            diagnostic = self._line(line, working_dir, source)
            if diagnostic is not None:
                diagnostics.append(diagnostic)
        return diagnostics

    def _line(self, line: str, working_dir: Path, source: str) -> Diagnostic | None:
        match = COMPILER_LINE.match(line)
        if match is not None:
            location = location_from_match(match, working_dir)
            return self._diagnostic(location, match.group("message").strip(), line, source)
        text = line.strip()
        if not text or NOISE.match(text):
            return None
        if self._identify(text) or text.startswith("go: "):
            return self._diagnostic(None, text, line, source)
        return None

    def _diagnostic(
        self, location: SourceLocation | None, message: str, raw_text: str, source: str
    ) -> Diagnostic:
        code = self._identify(message)
        if code.startswith("V-"):
            source = "vet"
        severity = Severity.WARNING if source == "vet" else Severity.ERROR
        return Diagnostic(location, severity, message, raw_text.rstrip(), source, code)

    def _panic(self, lines: list[str], working_dir: Path) -> Diagnostic:
        block = [line for line in lines if not NOISE.match(line)]
        message = block[0].strip()
        return Diagnostic(
            location=_first_user_frame(block, working_dir),
            severity=Severity.ERROR,
            message=message,
            raw_text="\n".join(block).strip(),
            source="panic",
            code=self._identify(message),
        )


def _first_user_frame(block: list[str], working_dir: Path) -> SourceLocation | None:
    for index, line in enumerate(block):
        match = FRAME_LINE.match(line)
        if match is None or _is_runtime_frame(block[index - 1], match.group("path")):
            continue
        return location_from_match(match, working_dir)
    return None


def _is_runtime_frame(function_line: str, path: str) -> bool:
    return function_line.startswith(RUNTIME_FUNCTIONS) or "/src/runtime/" in path.replace("\\", "/")
