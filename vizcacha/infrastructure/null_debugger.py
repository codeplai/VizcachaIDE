"""Placeholder DebuggerPort used until the Delve adapter (track A) exists.

Version 0.1 shipped a *simulated* debugger that jumped to random lines and showed
made-up variables. That was misleading for beginners, so it was removed: this
adapter only tells the user that debugging is not available yet.
"""

from collections.abc import Sequence
from pathlib import Path

from PyQt5.QtCore import QObject, pyqtSignal

from vizcacha.domain.debugging import Breakpoint, Variable
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _


class NullDebugger(QObject):
    stopped = pyqtSignal(object)
    output = pyqtSignal(str, str)
    terminated = pyqtSignal(int)

    def start(self, config: RunConfiguration, breakpoints: Sequence[Breakpoint]) -> None:
        self.output.emit(
            _("The debugger is not available yet. It will use Delve in a future version.") + "\n",
            "console",
        )
        self.terminated.emit(0)

    def set_breakpoints(self, file: Path, lines: Sequence[int]) -> None:
        return

    def step_over(self) -> None:
        return

    def step_into(self) -> None:
        return

    def step_out(self) -> None:
        return

    def resume(self) -> None:
        return

    def run_to(self, location: SourceLocation) -> None:
        return

    def variables(self, reference: int) -> list[Variable]:
        return []

    def stop(self) -> None:
        return

    def is_active(self) -> bool:
        return False
