"""QA of the Windows downloads of the Wails variant: NSIS installer and portable zip.

    python wails/packaging/qa/install_qa.py --setup wails/dist/release/VizcachaIDE-...-full-setup.exe
    python wails/packaging/qa/install_qa.py --zip   wails/dist/release/VizcachaIDE-...-full-portable.zip

Installer: silent per-user install (no admin) into a temporary folder, Start-menu shortcut and
uninstall entry (HKCU), the app stays alive ~8 s, bundled ``go version`` (full) and ``go run``,
silent uninstall, and nothing left behind (folder, shortcut, registry, WebView2 data folder).
Exit code 0 = every check passed. Adapted from packaging/qa/install_qa.py (Inno Setup).
"""

from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
import time
import winreg
import zipfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE.parent))
import smoke_test  # noqa: E402

REPO_ROOT = HERE.parents[2]
EXE = "VizcachaIDE.exe"
UNINSTALL_KEY = r"Software\Microsoft\Windows\CurrentVersion\Uninstall\CodeplaiGamesVizcachaIDE"
START_MENU = Path(os.environ["APPDATA"]) / r"Microsoft\Windows\Start Menu\Programs"
DESKTOP = Path(os.environ["USERPROFILE"]) / "Desktop"
WEBVIEW_DATA = [Path(os.environ["APPDATA"]) / EXE, Path(os.environ["LOCALAPPDATA"]) / EXE]
FAILURES: list[str] = []


def check(name: str, passed: bool, detail: str = "") -> None:
    print(f"[{'PASS' if passed else 'FAIL'}] {name} {detail}".rstrip())
    if not passed:
        FAILURES.append(name)


def has_registry_key(path: str) -> bool:
    try:
        winreg.CloseKey(winreg.OpenKey(winreg.HKEY_CURRENT_USER, path))
        return True
    except OSError:
        return False


def without_go_on_path() -> dict[str, str]:
    env = dict(os.environ)
    name = next((key for key in env if key.upper() == "PATH"), "PATH")
    kept = [p for p in env.get(name, "").split(os.pathsep) if p and not shutil.which("go", path=p)]
    env[name] = os.pathsep.join(kept)
    return env


def check_app(app_dir: Path, variant: str) -> None:
    exe = app_dir / EXE
    check("executable present", exe.is_file())
    started = time.monotonic()
    running, output = smoke_test.run_for(exe, 8)
    check("app alive after 8 s", running, f"({time.monotonic() - started:.1f} s) {output.strip()[:200]}")
    toolchain = app_dir / "toolchain"
    if variant == "lite":
        check("lite has no toolchain/", not toolchain.exists())
        return
    go = toolchain / "go" / "bin" / "go.exe"
    env = without_go_on_path()
    env.update(GOROOT=str(go.parents[1]), GOTOOLCHAIN="local")
    env["PATH"] = os.pathsep.join([str(go.parent), str(toolchain / "bin"), env.get("PATH", "")])
    version = subprocess.run([str(go), "version"], capture_output=True, text=True, env=env)
    check("bundled go version", version.returncode == 0, version.stdout.strip())
    check("bundled gofmt", (go.parent / "gofmt.exe").is_file())
    for tool in ("dlv.exe", "gopls.exe"):
        check(f"bundled {tool}", (toolchain / "bin" / tool).is_file())
    check("go license shipped", (toolchain / "go" / "LICENSE").is_file())
    with tempfile.TemporaryDirectory() as work:
        hello = Path(work) / "hello.go"
        hello.write_text('package main\n\nimport "fmt"\n\nfunc main() { fmt.Println("Hello") }\n')
        result = subprocess.run(
            [str(go), "run", "hello.go"], cwd=work, capture_output=True, text=True, env=env
        )
        check("go run hello.go with bundled go", "Hello" in result.stdout, result.stderr.strip()[:200])


def leftovers(folder: Path) -> list[str]:
    return [str(p.relative_to(folder)) for p in folder.rglob("*")] if folder.exists() else []


def installer_qa(setup: Path, variant: str) -> None:
    root = Path(tempfile.mkdtemp(prefix="vizcacha-install-"))
    app_dir = root / "VizcachaIDE"
    started = time.monotonic()
    # NSIS: /S silent; /D must be the last argument and unquoted. Per-user installer: no admin.
    code = subprocess.run([str(setup), "/S", f"/D={app_dir}"]).returncode
    check("silent per-user install", code == 0, f"(exit {code}, {time.monotonic() - started:.1f} s)")
    check("uninstaller written", (app_dir / "uninstall.exe").is_file())
    check("uninstall entry in HKCU", has_registry_key(UNINSTALL_KEY))
    check("start menu shortcut", (START_MENU / "VizcachaIDE.lnk").is_file())
    check("no desktop shortcut", not (DESKTOP / "VizcachaIDE.lnk").exists())
    check_app(app_dir, variant)
    uninstaller = app_dir / "uninstall.exe"
    # NSIS copies the uninstaller to %TEMP% and returns at once: wait for the folder to vanish.
    code = subprocess.run([str(uninstaller), "/S"]).returncode
    deadline = time.monotonic() + 120
    while app_dir.exists() and time.monotonic() < deadline:
        time.sleep(1)
    time.sleep(2)
    remaining = leftovers(app_dir)
    check("silent uninstall", code == 0 and not app_dir.exists(), f"(exit {code}) leftovers={remaining[:10]}")
    check("shortcut removed", not (START_MENU / "VizcachaIDE.lnk").exists())
    check("uninstall entry removed", not has_registry_key(UNINSTALL_KEY))
    check("WebView2 data removed", not any(p.exists() for p in WEBVIEW_DATA), str([str(p) for p in WEBVIEW_DATA if p.exists()]))
    shutil.rmtree(root, ignore_errors=True)


def zip_qa(archive: Path, variant: str) -> None:
    with tempfile.TemporaryDirectory(prefix="vizcacha-zip-") as work:
        with zipfile.ZipFile(archive) as bundle:
            bundle.extractall(work)
        check_app(Path(work) / "VizcachaIDE", variant)
        time.sleep(1)  # let the killed app release its files


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--setup", type=Path, action="append", default=[])
    parser.add_argument("--zip", type=Path, action="append", default=[])
    args = parser.parse_args(argv)
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    for path in [*args.setup, *args.zip]:
        variant = "lite" if "-lite-" in path.name else "full"
        print(f"== {path.name} ({variant})")
        (installer_qa if path.suffix == ".exe" else zip_qa)(path.resolve(), variant)
    print(f"[qa] {len(FAILURES)} failure(s): {FAILURES}" if FAILURES else "[qa] all checks passed")
    return 1 if FAILURES else 0


if __name__ == "__main__":
    raise SystemExit(main())
