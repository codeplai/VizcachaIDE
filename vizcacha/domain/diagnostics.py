"""Problems found in the user's Go code (compiler errors, vet warnings, panics)."""

from dataclasses import dataclass
from enum import Enum
from pathlib import Path


class Severity(Enum):
    ERROR = "error"
    WARNING = "warning"
    INFO = "info"
    HINT = "hint"


@dataclass(frozen=True)
class SourceLocation:
    """A position in a source file. ``line`` and ``column`` are 1-based."""

    file: Path
    line: int
    column: int = 1


@dataclass(frozen=True)
class Diagnostic:
    """One problem reported by a Go tool.

    ``message`` is the tool's message for that problem and ``raw_text`` the exact
    text the tool printed. Both stay untranslated, so they can be searched on the web.
    """

    location: SourceLocation | None
    severity: Severity
    message: str
    raw_text: str
    source: str = "go"
    code: str = ""
