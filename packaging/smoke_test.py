"""Start a packaged VizcachaIDE headless for a few seconds and fail on any crash.

    python packaging/smoke_test.py dist/lite/VizcachaIDE/VizcachaIDE.exe --seconds 6

Pass criteria: the process is still running after ``--seconds`` (QT_QPA_PLATFORM=offscreen),
printed no Python traceback and, on Windows, opened no native window. With the offscreen
platform Qt creates no native windows, so a visible one is PyInstaller's
"Unhandled exception" dialog of a windowed build (which would otherwise hang silently).
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys
import tempfile
from pathlib import Path


def visible_windows_of(pid: int) -> int:
    if sys.platform != "win32":
        return 0
    import ctypes
    from ctypes import wintypes

    user32 = ctypes.windll.user32
    found = []

    @ctypes.WINFUNCTYPE(wintypes.BOOL, wintypes.HWND, wintypes.LPARAM)
    def collect(hwnd, _lparam):
        owner = wintypes.DWORD()
        user32.GetWindowThreadProcessId(hwnd, ctypes.byref(owner))
        if owner.value == pid and user32.IsWindowVisible(hwnd):
            found.append(hwnd)
        return True

    user32.EnumWindows(collect, 0)
    return len(found)


def run_for(executable: Path, seconds: float) -> tuple[bool, int, str]:
    """Return (still_running, visible_windows, combined output)."""
    env = dict(os.environ, QT_QPA_PLATFORM="offscreen", PYTHONFAULTHANDLER="1")
    with tempfile.TemporaryFile() as log:
        process = subprocess.Popen([str(executable)], stdout=log, stderr=subprocess.STDOUT, env=env)
        try:
            process.wait(timeout=seconds)
            running, windows = False, 0
        except subprocess.TimeoutExpired:
            running, windows = True, visible_windows_of(process.pid)
            process.kill()
            process.wait()
        log.seek(0)
        output = log.read().decode("utf-8", errors="replace")
    return running, windows, output


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("executable", type=Path)
    parser.add_argument("--seconds", type=float, default=6.0)
    args = parser.parse_args(argv)
    if not args.executable.exists():
        print(f"FAIL: {args.executable} does not exist")
        return 1
    running, windows, output = run_for(args.executable, args.seconds)
    if output.strip():
        print(output)
    if not running:
        print(f"FAIL: the application exited before {args.seconds} s")
        return 1
    if "Traceback" in output or windows:
        print(f"FAIL: traceback in output or error dialog shown (windows={windows})")
        return 1
    print(f"OK: running after {args.seconds} s, no traceback")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
