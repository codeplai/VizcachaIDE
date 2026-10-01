"""State of a debugging session as seen by the user."""

from dataclasses import dataclass
from enum import Enum

from vizcacha.domain.diagnostics import SourceLocation


@dataclass(frozen=True)
class Breakpoint:
    location: SourceLocation
    condition: str = ""


@dataclass(frozen=True)
class Variable:
    """A variable value.

    Children are either already loaded (``children``) or available lazily through
    ``reference`` (non-zero means "ask the debugger for the children").
    """

    name: str
    type_name: str
    value: str
    reference: int = 0
    children: tuple["Variable", ...] = ()

    @property
    def has_children(self) -> bool:
        return self.reference != 0 or bool(self.children)


@dataclass(frozen=True)
class StackFrame:
    frame_id: int
    function: str
    location: SourceLocation | None


@dataclass(frozen=True)
class Goroutine:
    goroutine_id: int
    name: str
    location: SourceLocation | None = None


class StopReason(Enum):
    ENTRY = "entry"
    BREAKPOINT = "breakpoint"
    STEP = "step"
    PAUSE = "pause"
    PANIC = "panic"


@dataclass(frozen=True)
class DebugState:
    """Snapshot taken every time the program stops."""

    reason: StopReason
    frames: tuple[StackFrame, ...]
    variables: tuple[Variable, ...] = ()
    goroutines: tuple[Goroutine, ...] = ()
    current_goroutine: int | None = None
    description: str = ""

    @property
    def current_location(self) -> SourceLocation | None:
        if not self.frames:
            return None
        return self.frames[0].location
