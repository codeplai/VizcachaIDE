"""Download the pinned CPython (python-build-standalone) and add debugpy, pylsp and ruff.

Result layout (where ``wails/internal/adapters/python/locator.go`` looks for it)::

    <dest>/toolchain/python/python.exe              Windows
    <dest>/toolchain/python/bin/python3             macOS and Linux
    <dest>/toolchain/python/<site-packages>/        debugpy, pylsp (+ deps), ruff
    <dest>/toolchain/licenses/                      licenses of the interpreter and the wheels
    <dest>/toolchain/VERSIONS.txt                   what was bundled (python lines are merged)

Usage: python packaging/fetch_python.py --os windows --arch amd64 --dest build/stage
Downloads are cached in packaging/cache/ (the interpreter in downloads/, wheels in wheels/<target>/)
and verified with the sha256 in versions.toml. The wheels are fetched for the *target* platform
with ``pip download --platform`` and installed with ``pip install --target``, so a Windows host
can stage a Linux or macOS bundle. The host needs Python 3.11+ with pip and network access.
"""

from __future__ import annotations

import argparse
import shutil
import subprocess
import sys
from pathlib import Path

from fetch_toolchain import download, extract_archive
from versions import CACHE_DIR, Target, host_target, load_versions

PYTHON_ABI = "cp312"
PYTHON_TAG = "3.12"
LICENSE_PATTERNS = ("LICENSE*", "LICENCE*", "COPYING*", "NOTICE*")


def interpreter_path(target: Target, python_dir: Path) -> Path:
    if target.os == "windows":
        return python_dir / "python.exe"
    return python_dir / "bin" / "python3"


def site_packages_path(target: Target, python_dir: Path) -> Path:
    if target.os == "windows":
        return python_dir / "Lib" / "site-packages"
    return python_dir / "lib" / f"python{PYTHON_TAG}" / "site-packages"


def fetch_interpreter(config: dict, target: Target, toolchain_dir: Path) -> Path:
    """Download, verify and extract CPython; return ``toolchain/python``."""
    hashes = config["sha256"]
    if target.key not in hashes:
        raise SystemExit(f"No pinned Python for {target.key}. Known: {', '.join(hashes)}")
    triple = config["triples"][target.key]
    url = config["url_template"].format(
        release=config["release"], version=config["version"], triple=triple
    )
    filename = url.rsplit("/", 1)[1].replace("%2B", "+")
    archive = download(url, CACHE_DIR / "downloads" / filename, hashes[target.key])
    python_dir = toolchain_dir / "python"
    if python_dir.exists():
        shutil.rmtree(python_dir)
    print(f"[extract] {filename} -> {python_dir}")
    extract_archive(archive, toolchain_dir)  # the archive has a top-level "python/" folder
    return python_dir


def prune_interpreter(python_dir: Path, prune: list[str]) -> None:
    for pattern in prune:  # entries are paths or globs relative to python_dir
        for match in python_dir.glob(pattern):
            if match.is_dir():
                shutil.rmtree(match, ignore_errors=True)
            else:
                match.unlink()
    for cache in python_dir.rglob("__pycache__"):
        shutil.rmtree(cache, ignore_errors=True)


def wheel_requirements(config: dict) -> list[str]:
    """``name[extra]==version`` lines for the pinned wheels."""
    extras = config.get("extras", {})
    lines = []
    for name, version in config["wheels"].items():
        suffix = f"[{','.join(extras[name])}]" if name in extras else ""
        lines.append(f"{name}{suffix}=={version}")
    return lines


def platform_options(config: dict, target: Target) -> list[str]:
    options = ["--python-version", PYTHON_TAG, "--implementation", "cp", "--abi", PYTHON_ABI]
    for tag in config["pip_platforms"][target.key]:
        options += ["--platform", tag]
    options += ["--only-binary=:all:"]
    return options


def run_pip(arguments: list[str]) -> None:
    command = [sys.executable, "-m", "pip", "--disable-pip-version-check", *arguments]
    print("+", " ".join(command), flush=True)
    try:
        subprocess.run(command, check=True)
    except subprocess.CalledProcessError as error:
        raise SystemExit(f"pip failed (exit {error.returncode})") from error


def download_wheels(config: dict, target: Target) -> Path:
    wheels = CACHE_DIR / "wheels" / target.key
    shutil.rmtree(wheels, ignore_errors=True)  # only the pinned set, never stale versions
    wheels.mkdir(parents=True)
    run_pip(
        ["download", "--dest", str(wheels), *platform_options(config, target)]
        + wheel_requirements(config)
    )
    return wheels


def install_wheels(config: dict, target: Target, wheels: Path, site_packages: Path) -> None:
    site_packages.mkdir(parents=True, exist_ok=True)
    run_pip(
        [
            "install",
            "--no-index",
            "--find-links",
            str(wheels),
            "--target",
            str(site_packages),
            "--no-compile",
            "--upgrade",
            *platform_options(config, target),
        ]
        + wheel_requirements(config)
    )


def copy_licenses(python_dir: Path, site_packages: Path, licenses_dir: Path) -> None:
    """CPython's LICENSE plus the license files inside every ``*.dist-info`` folder."""
    licenses_dir.mkdir(parents=True, exist_ok=True)
    for stale in licenses_dir.glob("python-*"):
        stale.unlink()
    interpreter_license = python_dir / "LICENSE.txt"
    if interpreter_license.is_file():
        shutil.copy2(interpreter_license, licenses_dir / "python-LICENSE.txt")
    for dist_info in sorted(site_packages.glob("*.dist-info")):
        package = dist_info.name.removesuffix(".dist-info")
        for pattern in LICENSE_PATTERNS:
            for found in dist_info.rglob(pattern):
                if found.is_file():
                    shutil.copy2(found, licenses_dir / f"python-{package}-{found.name}")


def merge_manifest(config: dict, target: Target, toolchain_dir: Path) -> None:
    """Add the python lines to VERSIONS.txt, keeping the lines the Go toolchain wrote."""
    manifest = toolchain_dir / "VERSIONS.txt"
    kept = []
    if manifest.is_file():
        kept = [
            line
            for line in manifest.read_text(encoding="utf-8").splitlines()
            if line and not line.startswith(("python", "debugpy", "python-lsp-server", "pyflakes", "ruff"))
        ]
    if not any(line.startswith("target =") for line in kept):
        kept.insert(0, f"target = {target.key}")
    python_lines = [f"python = {config['version']} (python-build-standalone {config['release']})"]
    python_lines += [f"{name} = {version}" for name, version in config["wheels"].items()]
    manifest.write_text("\n".join(kept + python_lines) + "\n", encoding="utf-8")


def fetch_python(target: Target, dest: Path) -> Path:
    """Stage the bundled Python under ``dest/toolchain`` and return that folder."""
    config = load_versions()["python"]
    toolchain_dir = dest.resolve() / "toolchain"
    toolchain_dir.mkdir(parents=True, exist_ok=True)
    python_dir = fetch_interpreter(config, target, toolchain_dir)
    prune_interpreter(python_dir, config.get("prune", []))
    site_packages = site_packages_path(target, python_dir)
    wheels = download_wheels(config, target)
    install_wheels(config, target, wheels, site_packages)
    prune_interpreter(python_dir, [])  # __pycache__ only
    copy_licenses(python_dir, site_packages, toolchain_dir / "licenses")
    merge_manifest(config, target, toolchain_dir)
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
    toolchain_dir = fetch_python(Target(args.os, args.arch), args.dest)
    print(f"[done] python ready in {toolchain_dir / 'python'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
