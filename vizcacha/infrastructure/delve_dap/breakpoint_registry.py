"""User breakpoints per file plus the temporary one used by "Run to cursor"."""

from collections.abc import Iterable, Sequence
from pathlib import Path

from vizcacha.domain.debugging import Breakpoint
from vizcacha.domain.diagnostics import SourceLocation


def _key(file: Path) -> Path:
    return Path(file).resolve()


class BreakpointRegistry:
    def __init__(self) -> None:
        self._lines: dict[Path, set[int]] = {}
        self._temporary: SourceLocation | None = None

    def reset(self, breakpoints: Iterable[Breakpoint]) -> None:
        self._lines.clear()
        self._temporary = None
        for point in breakpoints:
            key = _key(point.location.file)
            self._lines.setdefault(key, set()).add(point.location.line)

    def replace(self, file: Path, lines: Sequence[int]) -> Path:
        key = _key(file)
        self._lines[key] = set(lines)
        return key

    def set_temporary(self, location: SourceLocation) -> Path:
        self._temporary = SourceLocation(_key(location.file), location.line)
        return self._temporary.file

    def clear_temporary(self) -> Path | None:
        """Forget the "Run to cursor" breakpoint; returns the file that must be re-sent."""
        if self._temporary is None:
            return None
        file = self._temporary.file
        self._temporary = None
        return file

    def files(self) -> list[Path]:
        return list(self._lines)

    def lines_for(self, file: Path) -> set[int]:
        key = _key(file)
        lines = set(self._lines.get(key, set()))
        if self._temporary is not None and self._temporary.file == key:
            lines.add(self._temporary.line)
        return lines
