"""Finding ``file.go:LINE[:COL]`` references in tool output and resolving their paths."""

import re
from dataclasses import dataclass
from pathlib import Path

from vizcacha.domain.diagnostics import SourceLocation

WINDOWS_ABSOLUTE = re.compile(r"^[A-Za-z]:[\\/]")
# At the start of a line (compiler messages, panic frames): absolute paths may contain spaces.
LINE_START_REFERENCE = re.compile(
    r"[ \t]*(?P<path>(?:[A-Za-z]:[\\/]|/)[^:\r\n]*?\.go|[^\s:]+?\.go)"
    r":(?P<line>\d+)(?::(?P<column>\d+))?"
)
INLINE_REFERENCE = re.compile(
    r"(?P<path>(?:[A-Za-z]:)?[\w.\\/~-]+\.go):(?P<line>\d+)(?::(?P<column>\d+))?"
)


@dataclass(frozen=True)
class SourceLink:
    """A reference found in a text: ``text[start:end]`` points to ``location``."""

    start: int
    end: int
    location: SourceLocation


def resolve_go_path(raw_path: str, working_dir: Path) -> Path:
    """Absolute paths (Windows or POSIX) are kept; relative ones are joined to ``working_dir``."""
    if WINDOWS_ABSOLUTE.match(raw_path) or raw_path.startswith(("/", "\\")):
        return Path(raw_path)
    return working_dir / raw_path.replace("\\", "/")


def location_from_match(match: re.Match[str], working_dir: Path) -> SourceLocation:
    column = match.groupdict().get("column")
    return SourceLocation(
        file=resolve_go_path(match.group("path"), working_dir),
        line=int(match.group("line")),
        column=int(column) if column else 1,
    )


def find_source_links(text: str, working_dir: Path) -> list[SourceLink]:
    links: list[SourceLink] = []
    offset = 0
    for line in text.splitlines(keepends=True):
        links.extend(_line_links(line, offset, working_dir))
        offset += len(line)
    return links


def _line_links(line: str, offset: int, working_dir: Path) -> list[SourceLink]:
    first = LINE_START_REFERENCE.match(line)
    matches = [first] if first is not None else list(INLINE_REFERENCE.finditer(line))
    return [
        SourceLink(
            start=offset + match.start("path"),
            end=offset + match.end(),
            location=location_from_match(match, working_dir),
        )
        for match in matches
    ]
