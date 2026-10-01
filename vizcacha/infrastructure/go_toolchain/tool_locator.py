"""Finds Go tools: configured path -> toolchain bundled with the app -> PATH.

Bundled layout (agreed with track F, relative to the application directory):

    toolchain/go/bin/go[.exe]       toolchain/go/bin/gofmt[.exe]
    toolchain/bin/dlv[.exe]         toolchain/bin/gopls[.exe]
"""

import os
import shutil
import sys
from dataclasses import dataclass
from enum import Enum
from pathlib import Path

GO = "go"
GOFMT = "gofmt"
DELVE = "dlv"
GOPLS = "gopls"
TOOLS = (GO, GOFMT, DELVE, GOPLS)

BUNDLE_DIRECTORY = "toolchain"
BUNDLED_GOROOT = (BUNDLE_DIRECTORY, "go")
BUNDLED_TOOL_DIRECTORIES = {
    GO: (*BUNDLED_GOROOT, "bin"),
    GOFMT: (*BUNDLED_GOROOT, "bin"),
    DELVE: (BUNDLE_DIRECTORY, "bin"),
    GOPLS: (BUNDLE_DIRECTORY, "bin"),
}


class ToolOrigin(Enum):
    CONFIGURED = "configured"
    BUNDLED = "bundled"
    PATH = "path"
    MISSING = "missing"


@dataclass(frozen=True)
class ToolLocation:
    """Where a tool comes from. ``path`` is empty when the tool is missing."""

    tool: str
    origin: ToolOrigin
    path: str = ""


def application_directory() -> Path:
    """Directory of the installed app (frozen) or the repository root (development)."""
    if getattr(sys, "frozen", False):
        return Path(sys.executable).resolve().parent
    return Path(__file__).resolve().parents[3]


def executable_file_name(tool: str, windows: bool | None = None) -> str:
    windows = os.name == "nt" if windows is None else windows
    return f"{tool}.exe" if windows else tool


class ToolLocator:
    """Auto-detection (bundled, then PATH), cached until ``forget()`` is called."""

    def __init__(self, app_directory: Path, search_path: str | None) -> None:
        self.app_directory = app_directory
        self._search_path = search_path
        self._cache: dict[str, ToolLocation] = {}

    def forget(self) -> None:
        self._cache.clear()

    def bundled_goroot(self) -> Path:
        return self.app_directory.joinpath(*BUNDLED_GOROOT)

    def bundled_bin_directories(self) -> list[Path]:
        directories = {
            self.app_directory.joinpath(*parts) for parts in BUNDLED_TOOL_DIRECTORIES.values()
        }
        return sorted(directory for directory in directories if directory.is_dir())

    def detect(self, tool: str) -> ToolLocation:
        if tool not in self._cache:
            self._cache[tool] = self._bundled(tool) or self._on_path(tool)
        return self._cache[tool]

    def _bundled(self, tool: str) -> ToolLocation | None:
        directory = self.app_directory.joinpath(*BUNDLED_TOOL_DIRECTORIES[tool])
        candidate = directory / executable_file_name(tool)
        if not candidate.is_file():
            return None
        return ToolLocation(tool, ToolOrigin.BUNDLED, str(candidate))

    def _on_path(self, tool: str) -> ToolLocation:
        found = shutil.which(tool, path=self._search_path)
        if found is None:
            return ToolLocation(tool, ToolOrigin.MISSING)
        return ToolLocation(tool, ToolOrigin.PATH, found)
