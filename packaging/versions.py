"""Shared helpers for the packaging scripts: pinned versions, paths and platform names.

Packaging scripts need Python 3.11+ (``tomllib``). The application itself supports 3.10+.
"""

from __future__ import annotations

import platform
import re
import sys
from dataclasses import dataclass
from pathlib import Path

import tomllib

PACKAGING_DIR = Path(__file__).resolve().parent
REPO_ROOT = PACKAGING_DIR.parent
CACHE_DIR = PACKAGING_DIR / "cache"
VERSIONS_FILE = PACKAGING_DIR / "versions.toml"

_OS_NAMES = {"win32": "windows", "darwin": "darwin", "linux": "linux"}
_ARCH_NAMES = {"amd64": "amd64", "x86_64": "amd64", "arm64": "arm64", "aarch64": "arm64"}


@dataclass(frozen=True)
class Target:
    os: str
    arch: str

    @property
    def key(self) -> str:
        return f"{self.os}-{self.arch}"

    @property
    def exe_suffix(self) -> str:
        return ".exe" if self.os == "windows" else ""

    @property
    def archive_ext(self) -> str:
        return "zip" if self.os == "windows" else "tar.gz"


def host_target() -> Target:
    os_name = next((v for k, v in _OS_NAMES.items() if sys.platform.startswith(k)), None)
    arch = _ARCH_NAMES.get(platform.machine().lower())
    if os_name is None or arch is None:
        raise SystemExit(f"Unsupported host platform: {sys.platform}/{platform.machine()}")
    return Target(os_name, arch)


def load_versions(path: Path = VERSIONS_FILE) -> dict:
    with path.open("rb") as handle:
        return tomllib.load(handle)


def app_version() -> str:
    """Read ``__version__`` without importing the package (it would import nothing heavy,
    but the build machine may not have the runtime dependencies installed)."""
    text = (REPO_ROOT / "vizcacha" / "__init__.py").read_text(encoding="utf-8")
    match = re.search(r'^__version__\s*=\s*"([^"]+)"', text, re.MULTILINE)
    if match is None:
        raise SystemExit("Cannot find __version__ in vizcacha/__init__.py")
    return match.group(1)


def numeric_version(version: str) -> str:
    """'0.2.0.dev0' -> '0.2.0.0' (Windows version resources need four integers)."""
    numbers = re.findall(r"\d+", version.split(".dev")[0].split("rc")[0])[:4]
    return ".".join(numbers + ["0"] * (4 - len(numbers)))
