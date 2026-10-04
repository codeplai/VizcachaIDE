"""Start a packaged VizcachaIDE (Wails) for a few seconds; it must still be running.

    python wails/packaging/smoke_test.py path/to/VizcachaIDE.exe --seconds 8
    python wails/packaging/smoke_test.py VizcachaIDE.app/Contents/MacOS/vizcacha \
        --go VizcachaIDE.app/Contents/MacOS/toolchain/go/bin/go
    xvfb-run -a python wails/packaging/smoke_test.py ./VizcachaIDE/VizcachaIDE --seconds 8
    python wails/packaging/smoke_test.py --python <stage>/toolchain/python/python.exe
    python wails/packaging/smoke_test.py --cxx <stage>/toolchain/cpp/bin/clang++.exe

Pass criteria: the process is alive after --seconds; with --go, ``go version`` of the bundled
toolchain also works; with --python, the bundled interpreter must import debugpy, pylsp,
pyflakes and ruff and ``python -m ruff --version`` must run (the executable may then be omitted, which
checks only the interpreter); with --cxx, the bundled clang++ must compile and run a statically linked
hello world from a folder whose name has spaces and accents, and lldb-dap, clangd and clang-format
(next to it) must answer ``--version``. Only the PID started here is ever killed.
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


PYTHON_IMPORT_CHECK = "import debugpy, pylsp, pyflakes, ruff; print('ok')"


def bundled_python_check(python: Path) -> tuple[bool, str]:
    """The bundled interpreter imports debugpy, pylsp and ruff, and ruff's binary runs."""
    imports = subprocess.run(
        [str(python), "-c", PYTHON_IMPORT_CHECK], capture_output=True, text=True, timeout=120
    )
    if imports.returncode != 0 or imports.stdout.strip() != "ok":
        return False, (imports.stderr or imports.stdout).strip()
    ruff = subprocess.run(
        [str(python), "-m", "ruff", "--version"], capture_output=True, text=True, timeout=60
    )
    return ruff.returncode == 0, (ruff.stdout or ruff.stderr).strip()


HELLO_CPP = """#include <iostream>
int main() {
    std::cout << "hola desde C++" << std::endl;
    return 0;
}
"""
HELLO_OUTPUT = "hola desde C++"
CXX_TOOLS = ("lldb-dap", "clangd", "clang-format")


def bundled_cxx_check(cxx: Path) -> tuple[bool, str]:
    """Compile and run hola.cpp with -static from "<tmp>/prueba con espacios y acentos ñandú"."""
    with tempfile.TemporaryDirectory() as tmp:
        folder = Path(tmp) / "prueba con espacios y acentos ñandú"
        folder.mkdir()
        source, program = folder / "hola.cpp", folder / "hola.exe"
        source.write_text(HELLO_CPP, encoding="utf-8")
        build = subprocess.run(
            [str(cxx), "-std=c++17", "-static", str(source), "-o", str(program)],
            capture_output=True, text=True, timeout=180, cwd=folder,
        )
        if build.returncode != 0:
            return False, f"compile failed: {build.stderr.strip()}"
        run = subprocess.run([str(program)], capture_output=True, text=True, timeout=30, cwd=folder)
    if run.stdout.strip() != HELLO_OUTPUT:
        return False, f"unexpected output {run.stdout!r} (exit {run.returncode})"
    return True, f"compiled and ran: {run.stdout.strip()}"


def bundled_cxx_tools_check(cxx: Path) -> tuple[bool, str]:
    """lldb-dap, clangd and clang-format sit next to clang++ and answer --version."""
    lines = []
    for name in CXX_TOOLS:
        tool = cxx.with_name(name + cxx.suffix)
        result = subprocess.run([str(tool), "--version"], capture_output=True, text=True, timeout=60)
        text = (result.stdout or result.stderr).strip()
        if result.returncode != 0 or not text:
            return False, f"{name} --version failed (exit {result.returncode}): {text}"
        lines.append(f"{name}: {text.splitlines()[0]}")
    return True, "; ".join(lines)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("executable", type=Path, nargs="?", default=None)
    parser.add_argument("--seconds", type=float, default=8.0)
    parser.add_argument("--go", type=Path, default=None, help="bundled go to check (full variant)")
    parser.add_argument(
        "--python", type=Path, default=None, help="bundled python to check (Python variants)"
    )
    parser.add_argument(
        "--cxx", type=Path, default=None, help="bundled clang++ to check (C++ variants)"
    )
    args = parser.parse_args(argv)
    if args.executable is None and args.python is None and args.cxx is None:
        parser.error("give the executable, --python, --cxx, or a combination")
    if args.python is not None:
        ok, text = bundled_python_check(args.python)
        print(f"{'OK' if ok else 'FAIL'}: bundled python -> {text}")
        if not ok:
            return 1
    if args.cxx is not None:
        for check in (bundled_cxx_check, bundled_cxx_tools_check):
            ok, text = check(args.cxx)
            print(f"{'OK' if ok else 'FAIL'}: bundled c++ -> {text}")
            if not ok:
                return 1
    if args.executable is None:
        return 0
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
