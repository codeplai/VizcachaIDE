"""Ports: the interfaces that the core needs from the outside world.

FROZEN CONTRACT (phase 0). Parallel tracks implement these, they do not edit them.
If a track needs a change, it writes a "Contract change request" in its final
report and the integrator decides.

Events: adapters that run asynchronously are ``QObject`` subclasses (that is
allowed in ``infrastructure``). They must expose the Qt signals listed in the
``*_SIGNALS`` constants below, with exactly those names and argument types.
``tests/test_contracts.py`` checks this for every registered adapter.
"""

from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Protocol, TypeVar, runtime_checkable

from vizcacha.domain.code_structure import DocumentSymbol, SourceRange
from vizcacha.domain.completion import CompletionItem, SignatureHelp
from vizcacha.domain.debugging import Breakpoint, Variable
from vizcacha.domain.diagnostics import Diagnostic, SourceLocation
from vizcacha.domain.explanations import ErrorExplanation
from vizcacha.domain.project import RunConfiguration

T = TypeVar("T")

# signal name -> argument types (documentation + contract test)
TOOLCHAIN_SIGNALS = {
    "output_received": (str,),  # program stdout chunk
    "error_received": (str,),  # program/compiler stderr chunk
    "execution_finished": (int,),  # exit code (also after build)
}
DEBUGGER_SIGNALS = {
    "stopped": (object,),  # DebugState
    "output": (str, str),  # text, category: "stdout" | "stderr" | "console"
    "terminated": (int,),  # exit code
}
LANGUAGE_SERVER_SIGNALS = {
    "diagnostics_published": (object, object),  # Path, list[Diagnostic]
    "server_unavailable": (str,),  # "not_found" | "crashed"
}


@runtime_checkable
class SettingsRepository(Protocol):
    def get(self, key: str, default: T) -> T: ...

    def set(self, key: str, value: object) -> None: ...


@runtime_checkable
class GoToolchainPort(Protocol):
    """Runs the user's program with the Go toolchain. Emits TOOLCHAIN_SIGNALS."""

    def environment(self) -> Mapping[str, str]: ...

    def run(self, config: RunConfiguration) -> None: ...

    def build(self, config: RunConfiguration) -> None: ...

    def stop(self) -> None: ...

    def is_running(self) -> bool: ...

    def write_input(self, text: str) -> None: ...

    def run_untitled(
        self, source: str, program_args: Sequence[str] = ()
    ) -> RunConfiguration | None:
        """Run unsaved source from a temporary directory. None if busy."""
        ...

    def run_go_command(self, working_dir: Path, args: Sequence[str]) -> bool:
        """Run ``go <args>`` (e.g. ``mod tidy``) with the same signals. False if busy."""
        ...

    def format_source(self, text: str) -> str:
        """Return gofmt-formatted text. Raises GoFormatError / GoToolchainNotFoundError."""
        ...


@runtime_checkable
class DebuggerPort(Protocol):
    """Interactive debugger (Delve). Emits DEBUGGER_SIGNALS."""

    def start(self, config: RunConfiguration, breakpoints: Sequence[Breakpoint]) -> None: ...

    def set_breakpoints(self, file: Path, lines: Sequence[int]) -> None: ...

    def step_over(self) -> None: ...

    def step_into(self) -> None: ...

    def step_out(self) -> None: ...

    def resume(self) -> None: ...

    def run_to(self, location: SourceLocation) -> None: ...

    def variables(self, reference: int) -> list[Variable]:
        """Children of a variable with a non-zero ``reference`` (lazy expansion)."""
        ...

    def stop(self) -> None: ...

    def is_active(self) -> bool: ...


@runtime_checkable
class LanguageServerPort(Protocol):
    """Code intelligence (gopls). Emits LANGUAGE_SERVER_SIGNALS."""

    def open_document(self, path: Path, text: str) -> None: ...

    def change_document(self, path: Path, text: str, version: int) -> None: ...

    def close_document(self, path: Path) -> None: ...

    def completion(self, location: SourceLocation) -> list[CompletionItem]: ...

    def hover(self, location: SourceLocation) -> str | None: ...

    def definition(self, location: SourceLocation) -> SourceLocation | None: ...

    def signature_help(self, location: SourceLocation) -> SignatureHelp | None: ...

    def document_highlights(self, location: SourceLocation) -> list[SourceRange]:
        """Occurrences of the symbol at ``location`` in the same file ([] if none)."""
        ...

    def document_symbols(self, path: Path) -> list[DocumentSymbol]:
        """Declarations of an open document, nested (for the Outline). [] if unknown."""
        ...

    def shutdown(self) -> None: ...


@runtime_checkable
class ErrorExplainerPort(Protocol):
    """Turns raw Go output into diagnostics and beginner-friendly explanations."""

    def parse(self, raw_output: str, working_dir: Path) -> list[Diagnostic]: ...

    def explain(self, diagnostic: Diagnostic) -> ErrorExplanation | None: ...
