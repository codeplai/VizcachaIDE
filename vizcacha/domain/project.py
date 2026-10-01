"""What the user wants to run: a single file or a Go module/package."""

from dataclasses import dataclass
from enum import Enum
from pathlib import Path


@dataclass(frozen=True)
class GoModule:
    root: Path
    module_path: str


class RunTarget(Enum):
    FILE = "file"
    PACKAGE = "package"


@dataclass(frozen=True)
class RunConfiguration:
    target: Path
    working_dir: Path
    mode: RunTarget = RunTarget.FILE
    program_args: tuple[str, ...] = ()
    module: GoModule | None = None

    @classmethod
    def for_file(cls, path: Path, program_args: tuple[str, ...] = ()) -> "RunConfiguration":
        return cls(target=path, working_dir=path.parent, program_args=program_args)

    def go_target_argument(self) -> str:
        """Argument passed to ``go run`` / ``go build`` from ``working_dir``."""
        if self.mode is RunTarget.PACKAGE:
            return "."
        return self.target.name

    def executable_name(self, windows: bool) -> str:
        base = self.target.stem if self.mode is RunTarget.FILE else self.working_dir.name
        return f"{base}.exe" if windows else base
