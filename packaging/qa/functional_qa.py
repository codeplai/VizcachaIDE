"""Headless functional QA of VizcachaIDE through the real workbench (no pytest needed).

    python packaging/qa/functional_qa.py --language en
    python packaging/qa/functional_qa.py --language es --bundle dist/full/VizcachaIDE

Builds ``vizcacha.ui.app.build_workbench`` with QT_QPA_PLATFORM=offscreen and drives it like a
user would: Run on examples/hello and examples/variables, five catalogued errors explained by the
Assistant, a real Delve session on examples/functions and gofmt on save. Every file is a copy
in a temporary folder. With ``--bundle`` the tools are taken from ``<bundle>/toolchain`` and Go
is removed from PATH, so the run proves the bundled toolchain is the one being used.
Exit code 0 = every check passed.
"""

from __future__ import annotations

import argparse
import os
import shutil
import sys
import tempfile
import time
from pathlib import Path

os.environ.setdefault("QT_QPA_PLATFORM", "offscreen")
REPO_ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(REPO_ROOT))
sys.path.insert(0, str(Path(__file__).resolve().parent))

from PyQt5.QtWidgets import QApplication  # noqa: E402
from qa_steps import QaSession  # noqa: E402

from vizcacha.i18n import install_language  # noqa: E402

ERROR_IDS = ("E-UNUSED-VAR", "E-UNDEFINED", "E-TYPE-MISMATCH", "P-INDEX-RANGE", "P-DIVIDE-ZERO")


def parse_args(argv: list[str] | None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--language", choices=["en", "es"], default="en")
    parser.add_argument("--bundle", type=Path, help="app folder containing toolchain/")
    parser.add_argument("--skip-debug", action="store_true", help="skip the Delve session")
    return parser.parse_args(argv)


def base_environment(bundle: Path | None) -> dict[str, str]:
    """Process environment; with a bundle, PATH entries that contain a go executable are dropped."""
    env = dict(os.environ)
    if bundle is None:
        return env
    name = next((key for key in env if key.upper() == "PATH"), "PATH")
    kept = [
        entry
        for entry in env.get(name, "").split(os.pathsep)
        if entry and shutil.which("go", path=entry) is None
    ]
    env[name] = os.pathsep.join(kept)
    return env


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")  # logs redirected on Windows
    app = QApplication.instance() or QApplication([])
    install_language(args.language)
    bundle = args.bundle.resolve() if args.bundle else None
    work = Path(tempfile.mkdtemp(prefix="vizcacha-qa-"))
    session = QaSession(app, work, base_environment(bundle), bundle)
    print(f"[qa] language={args.language} bundle={bundle} work={work}")
    print(f"[qa] tools: {session.tool_origins()}")
    print(f"[qa] build_workbench + show: {session.startup_seconds:.2f} s")
    started = time.monotonic()
    try:
        session.check_bundled_toolchain()
        session.check_ui_language()
        session.run_example("hello.go", "Welcome to beginner-friendly Go development!")
        session.run_example("variables.go", "Count: 42")
        for error_id in ERROR_IDS:
            session.explain_error(error_id)
        if not args.skip_debug:
            session.debug_functions(breakpoint_line=13, variable="result", value="35")
        session.format_on_save()
    finally:
        session.close()
        shutil.rmtree(work, ignore_errors=True)
    print(session.report(time.monotonic() - started))
    return 0 if session.all_passed() else 1


if __name__ == "__main__":
    raise SystemExit(main())
