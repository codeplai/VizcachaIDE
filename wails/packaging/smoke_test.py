"""Start a packaged VizcachaIDE (Wails) for a few seconds; it must still be running.

    python wails/packaging/smoke_test.py path/to/VizcachaIDE.exe --seconds 8
    python wails/packaging/smoke_test.py VizcachaIDE.app/Contents/MacOS/vizcacha --go VizcachaIDE.app/Contents/MacOS/toolchain/go/bin/go
    xvfb-run -a python wails/packaging/smoke_test.py ./VizcachaIDE/VizcachaIDE --seconds 8

Pass criteria: the process is alive after --seconds; with --go, ``go version`` of the bundled
toolchain also works. Only the PID started here is ever killed.
"""

from __future__ import annotations

import argparse
import subprocess
import sys
import tempfile
from pathlib import Path


def run_for(executable: Path, seconds: float) -> tuple[bool, str]:
    """Return (still_running_after_seconds, combined output)."""
    with tempfile.TemporaryFile() as log:
        process = subprocess.Popen([str(executable)], stdout=log, stderr=subprocess.STDOUT)
        try:
            process.wait(timeout=seconds)
            running = False
        except subprocess.TimeoutExpired:
            running = True
            process.kill()  # by PID: this very child
            process.wait()
        log.seek(0)
        return running, log.read().decode("utf-8", errors="replace")


def bundled_go_version(go: Path) -> tuple[bool, str]:
    result = subprocess.run([str(go), "version"], capture_output=True, text=True, timeout=60)
    return result.returncode == 0, (result.stdout or result.stderr).strip()


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("executable", type=Path)
    parser.add_argument("--seconds", type=float, default=8.0)
    parser.add_argument("--go", type=Path, default=None, help="bundled go to check (full variant)")
    args = parser.parse_args(argv)
    if not args.executable.exists():
        print(f"FAIL: {args.executable} does not exist")
        return 1
    if args.go is not None:
        ok, text = bundled_go_version(args.go)
        print(f"{'OK' if ok else 'FAIL'}: bundled go -> {text}")
        if not ok:
            return 1
    running, output = run_for(args.executable, args.seconds)
    if output.strip():
        print(output)
    if not running:
        print(f"FAIL: the application exited before {args.seconds} s")
        return 1
    print(f"OK: still running after {args.seconds} s")
    return 0


if __name__ == "__main__":
    sys.exit(main())
