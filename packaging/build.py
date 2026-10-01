"""Build VizcachaIDE with PyInstaller for the host platform.

    python packaging/build.py --variant lite            # app only, uses the system Go
    python packaging/build.py --variant full            # + Go, dlv and gopls in toolchain/
    python packaging/build.py --variant full --package  # + installer / dmg / AppImage / zip

Output: dist/<variant>/VizcachaIDE/ (Windows, Linux) or dist/<variant>/VizcachaIDE.app (macOS).
Packages are written to dist/release/. PyInstaller cannot cross-compile, so the app is always
built for the machine running this script.
"""

from __future__ import annotations

import argparse
import subprocess
import sys
from pathlib import Path

import installers
from fetch_toolchain import fetch_toolchain
from fetch_toolchain import parse_args as toolchain_args
from versions import CACHE_DIR, PACKAGING_DIR, REPO_ROOT, Target, app_version, host_target

SPEC = PACKAGING_DIR / "vizcacha.spec"
DIST_DIR = REPO_ROOT / "dist"
WORK_DIR = REPO_ROOT / "build" / "pyinstaller"


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--variant", choices=["full", "lite"], required=True)
    parser.add_argument("--package", action="store_true", help="also build installers")
    parser.add_argument(
        "--skip-fetch", action="store_true", help="reuse an already prepared toolchain"
    )
    parser.add_argument("--go", default="go", help="host go used to build dlv and gopls")
    return parser.parse_args(argv)


def prepare_toolchain(target: Target, skip_fetch: bool, go: str) -> Path:
    stage = CACHE_DIR / "stage" / target.key
    toolchain_dir = stage / "toolchain"
    if skip_fetch and (toolchain_dir / "go").is_dir():
        print(f"[skip] reusing {toolchain_dir}")
        return toolchain_dir
    args = toolchain_args(["--os", target.os, "--arch", target.arch, "--dest", str(stage)])
    args.go = go
    return fetch_toolchain(target, stage, args)


def run_pyinstaller(variant: str, toolchain_dir: Path | None) -> Path:
    dist = DIST_DIR / variant
    command = [
        sys.executable, "-m", "PyInstaller", str(SPEC), "--noconfirm",
        "--distpath", str(dist), "--workpath", str(WORK_DIR / variant),
        "--", "--variant", variant,
    ]  # fmt: skip
    if toolchain_dir is not None:
        command += ["--toolchain", str(toolchain_dir)]
    print("[pyinstaller] " + " ".join(command))
    try:
        subprocess.run(command, check=True, cwd=REPO_ROOT)
    except subprocess.CalledProcessError as error:
        raise SystemExit(f"PyInstaller failed (exit {error.returncode})") from error
    return dist


def folder_size(path: Path) -> int:
    return sum(item.stat().st_size for item in path.rglob("*") if item.is_file())


def app_folder(dist: Path) -> Path:
    if sys.platform == "darwin":
        return dist / "VizcachaIDE.app"
    return dist / "VizcachaIDE"


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    target = host_target()
    version = app_version()
    print(f"[build] VizcachaIDE {version} {args.variant} for {target.key}")
    toolchain_dir = None
    if args.variant == "full":
        toolchain_dir = prepare_toolchain(target, args.skip_fetch, args.go)
    app_dir = app_folder(run_pyinstaller(args.variant, toolchain_dir))
    size_mb = folder_size(app_dir) / 1e6
    print(f"[size] {app_dir}: {size_mb:.1f} MB")
    if args.package:
        for artifact in installers.package(app_dir, target, args.variant, version):
            print(f"[package] {artifact} ({artifact.stat().st_size / 1e6:.1f} MB)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
