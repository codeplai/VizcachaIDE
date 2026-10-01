"""Locating bundled files (logo, icons) both from source and from a PyInstaller bundle."""

import sys
from pathlib import Path


def resource_root() -> Path:
    bundle_dir = getattr(sys, "_MEIPASS", None)
    if getattr(sys, "frozen", False) and bundle_dir:
        return Path(bundle_dir)
    return Path(__file__).resolve().parents[2]


def resource_path(name: str) -> Path:
    return resource_root() / name
