"""Download the pinned llvm-mingw, keep only its x86_64 target and stage it as the bundled C++ toolchain.

Result layout (where ``wails/internal/adapters/cpp/locator.go`` looks for it)::

    <dest>/toolchain/cpp/bin/clang++.exe, clang.exe, ld.lld.exe, lldb-dap.exe, clangd.exe, clang-format.exe
    <dest>/toolchain/cpp/x86_64-w64-mingw32/        MinGW-w64 runtime, libc++, libunwind
    <dest>/toolchain/cpp/lib/clang/, include/        compiler headers and runtimes
    <dest>/toolchain/licenses/                      llvm-mingw and MinGW-w64 licenses
    <dest>/toolchain/VERSIONS.txt                   what was bundled (cpp lines are merged)

Usage: python packaging/fetch_cpp.py --dest build/stage      (Windows x86_64 only)
The zip is cached in packaging/cache/downloads/ and verified with the sha256 in versions.toml. The
release has four targets (735 MB unpacked); the ``keep`` and ``drop`` lists of ``[cpp.windows]`` decide
which members are extracted, so the other targets are never written to disk. macOS and Linux bundle no
C++ compiler: they use the one of the system.
"""

from __future__ import annotations

import argparse
import shutil
import zipfile
from fnmatch import fnmatchcase
from pathlib import Path

from fetch_toolchain import download
from versions import CACHE_DIR, Target, host_target, load_versions

LICENSE_GLOBS = ("LICENSE.TXT", "x86_64-w64-mingw32/share/mingw32/COPYING*")
MANIFEST_PREFIXES = ("cpp", "llvm")


def matches_any(relative: str, patterns: list[str]) -> bool:
    """True when ``relative`` equals a pattern, matches a glob, or lives below a folder pattern."""
    for pattern in patterns:
        if pattern.endswith("/"):
            if relative.startswith(pattern):
                return True
        elif fnmatchcase(relative, pattern):
            return True
    return False


def is_kept(relative: str, keep: list[str], drop: list[str]) -> bool:
    return matches_any(relative, keep) and not matches_any(relative, drop)


def download_archive(config: dict) -> Path:
    url = config["url"].format(release=config["release"])
    return download(url, CACHE_DIR / "downloads" / url.rsplit("/", 1)[1], config["sha256"])


def extract_pruned(archive: Path, config: dict, cpp_dir: Path) -> tuple[int, int]:
    """Extract only the kept members into ``cpp_dir``; return (extracted, skipped) file counts."""
    keep, drop = config["keep"], config.get("drop", [])
    extracted = skipped = 0
    with zipfile.ZipFile(archive) as bundle:
        for member in bundle.infolist():
            if member.is_dir():
                continue
            relative = member.filename.split("/", 1)[1]  # drop the "llvm-mingw-<release>-.../" root
            if not is_kept(relative, keep, drop):
                skipped += 1
                continue
            destination = cpp_dir / relative
            destination.parent.mkdir(parents=True, exist_ok=True)
            with bundle.open(member) as source, destination.open("wb") as target:
                shutil.copyfileobj(source, target)
            extracted += 1
    return extracted, skipped


def copy_licenses(cpp_dir: Path, licenses_dir: Path) -> None:
    licenses_dir.mkdir(parents=True, exist_ok=True)
    for stale in licenses_dir.glob("llvm-mingw-*"):
        stale.unlink()
    for pattern in LICENSE_GLOBS:
        for found in cpp_dir.glob(pattern):
            shutil.copy2(found, licenses_dir / f"llvm-mingw-{found.name}")


def merge_manifest(config: dict, target: Target, toolchain_dir: Path) -> None:
    """Add the cpp lines to VERSIONS.txt, keeping the lines other fetchers wrote."""
    manifest = toolchain_dir / "VERSIONS.txt"
    kept = []
    if manifest.is_file():
        kept = [
            line
            for line in manifest.read_text(encoding="utf-8").splitlines()
            if line and not line.startswith(MANIFEST_PREFIXES)
        ]
    if not any(line.startswith("target =") for line in kept):
        kept.insert(0, f"target = {target.key}")
    cpp_lines = [
        f"cpp = {config['distribution']} {config['release']}",
        f"llvm = {config['llvm_version']}",
    ]
    manifest.write_text("\n".join(kept + cpp_lines) + "\n", encoding="utf-8")


def folder_size(folder: Path) -> int:
    return sum(path.stat().st_size for path in folder.rglob("*") if path.is_file())


def fetch_cpp(target: Target, dest: Path) -> Path:
    """Stage the bundled C++ toolchain under ``dest/toolchain`` and return that folder."""
    if target.key != "windows-amd64":
        raise SystemExit(f"No bundled C++ toolchain for {target.key}: only windows-amd64 has one")
    config = load_versions()["cpp"]["windows"]
    toolchain_dir = dest.resolve() / "toolchain"
    cpp_dir = toolchain_dir / "cpp"
    if cpp_dir.exists():
        shutil.rmtree(cpp_dir)
    archive = download_archive(config)
    print(f"[extract] {archive.name} -> {cpp_dir} (x86_64 only)")
    extracted, skipped = extract_pruned(archive, config, cpp_dir)
    if not (cpp_dir / "bin" / "clang++.exe").is_file():
        raise SystemExit("clang++.exe is missing after pruning: check keep in versions.toml")
    copy_licenses(cpp_dir, toolchain_dir / "licenses")
    merge_manifest(config, target, toolchain_dir)
    print(f"[prune] {extracted} files kept, {skipped} skipped, {folder_size(cpp_dir) / 1e6:.0f} MB")
    return toolchain_dir


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    host = host_target()
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--os", default=host.os, choices=["windows", "darwin", "linux"])
    parser.add_argument("--arch", default=host.arch, choices=["amd64", "arm64"])
    parser.add_argument("--dest", type=Path, required=True, help="directory that gets toolchain/")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    toolchain_dir = fetch_cpp(Target(args.os, args.arch), args.dest)
    print(f"[done] C++ ready in {toolchain_dir / 'cpp'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
