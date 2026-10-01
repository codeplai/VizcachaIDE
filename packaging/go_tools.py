"""Build Go tools (dlv, gopls) for a target OS/arch with the host ``go`` command.

``go install pkg@version`` resolves the pinned module through proxy.golang.org and checks it
against sum.golang.org. ``GOTOOLCHAIN`` is set to the pinned Go version so every machine
compiles the tools with the same compiler (the host go downloads it once if needed).
Cross-compiled binaries cannot be installed with GOBIN set, so a private GOPATH in
packaging/cache/ is used and the binary is picked from ``bin/`` or ``bin/<os>_<arch>/``.
"""

from __future__ import annotations

import os
import shutil
import subprocess
from pathlib import Path

from versions import CACHE_DIR, Target

GOPATH = CACHE_DIR / "gopath"
BUILT_DIR = CACHE_DIR / "tools"
LICENSE_NAMES = ("LICENSE", "LICENSE.txt", "LICENSE.md")


def tool_environment(target: Target, go_version: str) -> dict[str, str]:
    env = dict(os.environ)
    env.pop("GOBIN", None)
    env.update(
        GOPATH=str(GOPATH),
        GOMODCACHE=str(GOPATH / "pkg" / "mod"),
        GOOS=target.os,
        GOARCH=target.arch,
        CGO_ENABLED="0",
        GOTOOLCHAIN=f"go{go_version}",
    )
    return env


def installed_binary(binary: str, target: Target) -> Path | None:
    for folder in (GOPATH / "bin" / f"{target.os}_{target.arch}", GOPATH / "bin"):
        candidate = folder / binary
        if candidate.exists():
            return candidate
    return None


def compile_tool(spec: dict, binary: str, target: Target, go_version: str, go: str) -> Path:
    """Run ``go install`` and return the freshly built binary."""
    for stale in (GOPATH / "bin" / f"{target.os}_{target.arch}" / binary, GOPATH / "bin" / binary):
        stale.unlink(missing_ok=True)
    command = [go, "install", "-trimpath", "-ldflags=-s -w", f"{spec['package']}@{spec['version']}"]
    print(f"[build] {' '.join(command)}  (GOOS={target.os} GOARCH={target.arch})")
    GOPATH.mkdir(parents=True, exist_ok=True)
    try:
        subprocess.run(command, env=tool_environment(target, go_version), cwd=GOPATH, check=True)
    except FileNotFoundError as error:
        raise SystemExit(f"Host Go not found ({go}); install Go or pass --go") from error
    except subprocess.CalledProcessError as error:
        raise SystemExit(f"Building {spec['package']} failed (exit {error.returncode})") from error
    built = installed_binary(binary, target)
    if built is None:
        raise SystemExit(f"go install succeeded but {binary} was not found under {GOPATH / 'bin'}")
    return built


def module_license(spec: dict) -> Path | None:
    module_dir = GOPATH / "pkg" / "mod" / f"{spec['module']}@{spec['version']}"
    for license_name in LICENSE_NAMES:
        if (module_dir / license_name).exists():
            return module_dir / license_name
    print(f"[warn] no license file found in {module_dir}")
    return None


def build_into_cache(spec: dict, binary: str, target: Target, go_version: str, go: str) -> Path:
    """Compile once per (target, version); the cache keeps the binary and its LICENSE.txt."""
    cache_dir = BUILT_DIR / target.key / f"{Path(spec['package']).name}-{spec['version']}"
    if (cache_dir / binary).exists():
        print(f"[cache] {binary} {spec['version']} for {target.key}")
        return cache_dir
    built = compile_tool(spec, binary, target, go_version, go)
    cache_dir.mkdir(parents=True, exist_ok=True)
    shutil.copy2(built, cache_dir / binary)
    license_file = module_license(spec)
    if license_file is not None:
        shutil.copyfile(license_file, cache_dir / "LICENSE.txt")
    return cache_dir


def build_tool(
    name: str, spec: dict, target: Target, go_version: str, toolchain_dir: Path, go: str = "go"
) -> Path:
    """Place ``<toolchain_dir>/bin/<name>[.exe]`` built for ``target`` plus its license."""
    binary = Path(spec["package"]).name + target.exe_suffix
    cache_dir = build_into_cache(spec, binary, target, go_version, go)
    destination = toolchain_dir / "bin" / binary
    destination.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(cache_dir / binary, destination)
    if (cache_dir / "LICENSE.txt").exists():
        (toolchain_dir / "licenses").mkdir(exist_ok=True)
        shutil.copyfile(
            cache_dir / "LICENSE.txt", toolchain_dir / "licenses" / f"{name}-LICENSE.txt"
        )
    return destination
