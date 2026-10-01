"""QA of the Windows downloads: Inno Setup installer and portable zip (per user, no admin).

    python packaging/qa/install_qa.py --setup dist/release/VizcachaIDE-...-full-setup.exe
    python packaging/qa/install_qa.py --zip dist/release/VizcachaIDE-...-full-portable.zip

Installer: silent per-user install into a temporary folder, Start-menu shortcut present, no
desktop shortcut and no .go association by default (then again with /TASKS=associatego), headless
launch (packaging/smoke_test.py), bundled ``go version`` + ``go run examples/hello.go`` with the
app's environment, silent uninstall, and nothing left behind. Zip: extract, launch, toolchain.
Exit code 0 = every check passed.
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
import tempfile
import time
import winreg
import zipfile
from pathlib import Path

PACKAGING = Path(__file__).resolve().parents[1]
REPO_ROOT = PACKAGING.parent
sys.path.insert(0, str(PACKAGING))
import smoke_test  # noqa: E402

PROGID_KEY = r"Software\Classes\VizcachaIDE.go"
START_MENU = Path(os.environ["APPDATA"]) / r"Microsoft\Windows\Start Menu\Programs"
DESKTOP = Path(os.environ["USERPROFILE"]) / "Desktop"
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
    import shutil

    env = dict(os.environ)
    name = next((key for key in env if key.upper() == "PATH"), "PATH")
    kept = [p for p in env.get(name, "").split(os.pathsep) if p and not shutil.which("go", path=p)]
    env[name] = os.pathsep.join(kept)
    return env


def check_app(app_dir: Path, variant: str) -> None:
    exe = app_dir / "VizcachaIDE.exe"
    started = time.monotonic()
    running, windows, output = smoke_test.run_for(exe, 6)
    ok = running and not windows and "Traceback" not in output
    check("launch offscreen 6 s", ok, f"(windows={windows}, {time.monotonic() - started:.1f} s)")
    go = app_dir / "toolchain" / "go" / "bin" / "go.exe"
    if variant == "lite":
        check("lite has no toolchain/", not (app_dir / "toolchain").exists())
        return
    env = without_go_on_path()
    env.update(GOROOT=str(go.parents[1]), GOTOOLCHAIN="local")  # what GoEnvironment sets
    env["PATH"] = os.pathsep.join(
        [str(go.parent), str(app_dir / "toolchain" / "bin"), env.get("PATH", "")]
    )
    version = subprocess.run([str(go), "version"], capture_output=True, text=True, env=env)
    check("bundled go version", version.returncode == 0, version.stdout.strip())
    for tool in ("gofmt.exe",):
        check(f"bundled {tool}", (go.parent / tool).is_file())
    for tool in ("dlv.exe", "gopls.exe"):
        check(f"bundled {tool}", (app_dir / "toolchain" / "bin" / tool).is_file())
    with tempfile.TemporaryDirectory() as work:
        hello = Path(work) / "hello.go"
        hello.write_bytes((REPO_ROOT / "examples" / "hello.go").read_bytes())
        result = subprocess.run(
            [str(go), "run", "hello.go"], cwd=work, capture_output=True, text=True, env=env
        )
        check("go run hello.go with bundled go", "Hello" in result.stdout, result.stderr.strip())


def leftovers(folder: Path) -> list[str]:
    return [str(p.relative_to(folder)) for p in folder.rglob("*")] if folder.exists() else []


def installer_qa(setup: Path, variant: str) -> None:
    root = Path(tempfile.mkdtemp(prefix="vizcacha-install-"))
    app_dir = root / "VizcachaIDE"
    log = root.parent / f"{root.name}-setup.log"
    base = [str(setup), "/VERYSILENT", "/SUPPRESSMSGBOXES", "/CURRENTUSER", "/NORESTART"]
    started = time.monotonic()
    code = subprocess.run([*base, f"/DIR={app_dir}", f"/LOG={log}"]).returncode
    check(
        "silent per-user install", code == 0, f"(exit {code}, {time.monotonic() - started:.1f} s)"
    )
    check("start menu shortcut", (START_MENU / "VizcachaIDE.lnk").is_file())
    check("no desktop shortcut by default", not (DESKTOP / "VizcachaIDE.lnk").exists())
    check("no .go association by default", not has_registry_key(PROGID_KEY))
    check_app(app_dir, variant)
    code = subprocess.run([*base, f"/DIR={app_dir}", "/TASKS=associatego"]).returncode
    check(".go association when the task is chosen", code == 0 and has_registry_key(PROGID_KEY))
    uninstaller = app_dir / "unins000.exe"
    code = subprocess.run(
        [str(uninstaller), "/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART"]
    ).returncode
    time.sleep(3)  # the uninstaller deletes itself from a temporary copy
    remaining = leftovers(app_dir)
    check("silent uninstall", code == 0 and not remaining, f"leftovers={remaining[:10]}")
    check("shortcut removed", not (START_MENU / "VizcachaIDE.lnk").exists())
    check(".go association removed", not has_registry_key(PROGID_KEY))


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
