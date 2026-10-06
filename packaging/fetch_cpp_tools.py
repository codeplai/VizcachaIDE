"""Stage CMake, Ninja, vcpkg and vcpkg's seed next to the bundled llvm-mingw (docs/PLAN_CPP_CMAKE.md section 6).

Result layout (where ``wails/internal/adapters/cpp/locator.go`` and ``vcpkg/locator.go`` look for it)::

    <cpp>/cmake/bin/cmake.exe, ninja.exe      CMake without documentation, and Ninja next to it
    <cpp>/cmake/share/cmake-4.4/              the modules CMake needs
    <cpp>/vcpkg/vcpkg.exe, ports/, scripts/, triplets/, .vcpkg-root, vcpkg.disable-metrics
    <cpp>/vcpkg-seed/                         the downloads vcpkg needs the first time (PowerShell, 7-Zip)
    <toolchain>/licenses/                     the licenses of all of them

Every download is cached in packaging/cache/downloads/ and verified with the sha256 of
``[cpp.cmake]``, ``[cpp.ninja]``, ``[cpp.vcpkg]`` and ``[[cpp.vcpkg_seed]]`` in versions.toml.
"""

from __future__ import annotations

import shutil
import tarfile
import zipfile
from fnmatch import fnmatchcase
from pathlib import Path

from fetch_toolchain import download
from versions import CACHE_DIR, PACKAGING_DIR

LICENSES = PACKAGING_DIR / "licenses"
LICENSE_FILES = (
    "cmake-LICENSE.txt", "ninja-LICENSE.txt", "vcpkg-LICENSE.txt", "powershell-LICENSE.txt", "7zip-License.txt",
)


def matches_any(relative: str, patterns: list[str]) -> bool:
    """True when ``relative`` equals a pattern, matches a glob, or lives below a folder pattern."""
    for pattern in patterns:
        if pattern.endswith("/"):
            if relative.startswith(pattern):
                return True
        elif fnmatchcase(relative, pattern):
            return True
    return False


def cached(url: str, sha256: str) -> Path:
    return download(url, CACHE_DIR / "downloads" / url.rsplit("/", 1)[1], sha256)


def write(destination: Path, source) -> None:
    destination.parent.mkdir(parents=True, exist_ok=True)
    with destination.open("wb") as target:
        shutil.copyfileobj(source, target)


def fetch_cmake(config: dict, cpp_dir: Path) -> None:
    """The official CMake zip without what the IDE does not use (doc/, man/, Help, the GUI)."""
    version = config["version"]
    archive = cached(config["url"].format(version=version), config["sha256"])
    root = cpp_dir / "cmake"
    with zipfile.ZipFile(archive) as bundle:
        for member in bundle.infolist():
            if member.is_dir():
                continue
            relative = member.filename.split("/", 1)[1]  # drop the "cmake-<version>-windows-x86_64/" root
            if not matches_any(relative, config["prune"]):
                with bundle.open(member) as source:
                    write(root / relative, source)
    if not (root / "bin" / "cmake.exe").is_file():
        raise SystemExit("cmake.exe is missing after pruning: check prune in versions.toml")


def fetch_ninja(config: dict, cpp_dir: Path) -> None:
    """ninja.exe goes next to cmake.exe, so one folder on PATH is enough."""
    archive = cached(config["url"].format(version=config["version"]), config["sha256"])
    with zipfile.ZipFile(archive) as bundle, bundle.open("ninja.exe") as source:
        write(cpp_dir / "cmake" / "bin" / "ninja.exe", source)


def fetch_vcpkg(config: dict, cpp_dir: Path) -> None:
    """The ports snapshot (kept: ports, scripts, triplets, licenses) and the vcpkg.exe that matches it."""
    snapshot = cached(config["snapshot_url"].format(commit=config["commit"]), config["snapshot_sha256"])
    root = cpp_dir / "vcpkg"
    with tarfile.open(snapshot, "r:gz") as bundle:
        for member in bundle:
            parts = member.name.split("/", 1)  # drop the "vcpkg-<commit>/" root
            relative = parts[1] if len(parts) > 1 else ""
            if member.isfile() and matches_any(relative, config["keep"]):
                source = bundle.extractfile(member)
                if source is not None:
                    write(root / relative, source)
    tool = cached(config["tool_url"].format(tool_release=config["tool_release"]), config["tool_sha256"])
    shutil.copy2(tool, root / "vcpkg.exe")
    (root / ".vcpkg-root").write_bytes(b"")  # without it vcpkg.cmake fails
    (root / "vcpkg.disable-metrics").write_bytes(b"")
    if not (root / "ports" / "fmt" / "vcpkg.json").is_file():
        raise SystemExit("the vcpkg snapshot has no ports/fmt: check keep in versions.toml")


def fetch_seed(files: list[dict], cpp_dir: Path) -> None:
    """The downloads vcpkg would fetch on its own the first time, so the IDE can hand them over."""
    for entry in files:
        archive = cached(entry["url"], entry["sha256"])
        destination = cpp_dir / "vcpkg-seed" / entry["path"]
        destination.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(archive, destination)


def copy_licenses(licenses_dir: Path) -> None:
    licenses_dir.mkdir(parents=True, exist_ok=True)
    for name in LICENSE_FILES:
        shutil.copy2(LICENSES / name, licenses_dir / name)


def fetch_cpp_tools(config: dict, cpp_dir: Path, licenses_dir: Path) -> None:
    """Stage everything of ``[cpp.*]`` that is not llvm-mingw under ``cpp_dir``."""
    for name in ("cmake", "vcpkg", "vcpkg-seed"):
        shutil.rmtree(cpp_dir / name, ignore_errors=True)
    fetch_cmake(config["cmake"], cpp_dir)
    fetch_ninja(config["ninja"], cpp_dir)
    fetch_vcpkg(config["vcpkg"], cpp_dir)
    fetch_seed(config["vcpkg_seed"], cpp_dir)
    copy_licenses(licenses_dir)


def manifest_lines(config: dict) -> list[str]:
    return [
        f"cmake = {config['cmake']['version']}",
        f"ninja = {config['ninja']['version']}",
        f"vcpkg = {config['vcpkg']['commit'][:10]} tool {config['vcpkg']['tool_release']}",
    ]
