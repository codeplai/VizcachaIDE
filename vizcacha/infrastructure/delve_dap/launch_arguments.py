"""Arguments of the DAP ``initialize`` / ``launch`` / ``setBreakpoints`` requests."""

import os
import tempfile
from collections.abc import Iterable, Mapping
from pathlib import Path

from vizcacha.domain.project import RunConfiguration

CLIENT_ID = "vizcacha"
DEBUG_BINARY_NAME = "vizcacha_debug_bin"

INITIALIZE_ARGUMENTS = {
    "clientID": CLIENT_ID,
    "clientName": "VizcachaIDE",
    "adapterID": "go",
    "linesStartAt1": True,
    "columnsStartAt1": True,
    "pathFormat": "path",
    "supportsVariableType": True,
}


def program_path(config: RunConfiguration) -> Path:
    """File or package folder that Delve builds (``go_target_argument`` from ``working_dir``)."""
    return (config.working_dir / config.go_target_argument()).resolve()


def debug_binary_path(process_id: int, windows: bool = os.name == "nt") -> str:
    """Where Delve writes the debug build, so it never litters the user's folder."""
    suffix = ".exe" if windows else ""
    return str(Path(tempfile.gettempdir()) / f"{DEBUG_BINARY_NAME}_{process_id}{suffix}")


def remove_stale_debug_binaries(current_pid: int, directory: Path | None = None) -> list[Path]:
    """Delete debug builds left by earlier sessions that crashed or were killed.

    A binary still used by another running VizcachaIDE is locked (Windows) or harmless
    to unlink (POSIX), so failures are skipped.
    """
    folder = directory or Path(tempfile.gettempdir())
    own_stem = f"{DEBUG_BINARY_NAME}_{current_pid}"
    removed: list[Path] = []
    for path in folder.glob(f"{DEBUG_BINARY_NAME}_*"):
        if path.stem == own_stem:
            continue
        try:
            path.unlink()
        except OSError:
            continue
        removed.append(path)
    return removed


def launch_arguments(config: RunConfiguration, environment: Mapping[str, str], output: str) -> dict:
    return {
        "request": "launch",
        "mode": "debug",
        "program": str(program_path(config)),
        "cwd": str(config.working_dir),
        "args": list(config.program_args),
        "env": dict(environment),
        "output": output,
        # "remote": the program's stdout/stderr arrive as DAP "output" events.
        "outputMode": "remote",
        "stopOnEntry": False,
    }


def set_breakpoints_arguments(file: Path, lines: Iterable[int]) -> dict:
    return {
        "source": {"name": file.name, "path": str(file)},
        "breakpoints": [{"line": line} for line in sorted(set(lines))],
    }
