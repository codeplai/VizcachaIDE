"""Assistant panel: explains compiler errors, vet warnings and panics for beginners.

Listens to ``program_started`` / ``process_output`` / ``program_finished`` (run
output) and ``diagnostics_changed`` (live diagnostics, e.g. from gopls). It also
connects the console's clickable ``file.go:LINE`` links to ``navigate_to``.
"""

from pathlib import Path

from PyQt5.QtCore import QObject, QUrl
from PyQt5.QtGui import QDesktopServices

from vizcacha.application.explain_error import ExplainedDiagnostic, ExplainError
from vizcacha.application.ports import ErrorExplainerPort
from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.infrastructure.error_catalog import GoErrorExplainer
from vizcacha.ui.features.assistant.panel import AssistantPanel
from vizcacha.ui.features.assistant.presentation import search_url
from vizcacha.ui.workbench import Workbench


def open_search(diagnostic: Diagnostic) -> None:
    QDesktopServices.openUrl(QUrl(search_url(diagnostic)))


class AssistantFeature(QObject):
    """A QObject child of the main window, so Qt keeps it (and its slots) alive."""

    def __init__(self, workbench: Workbench, explainer: ErrorExplainerPort) -> None:
        super().__init__(workbench.window)
        self.workbench = workbench
        self.use_case = ExplainError(explainer)
        self.panel = AssistantPanel()
        self._stderr: list[str] = []
        self._working_dir = Path.cwd()
        self._live_file: Path | None = None

    def register(self) -> None:
        self.dock = self.workbench.add_panel("assistant", _("Assistant"), self.panel, "bottom")
        events = self.workbench.events
        events.program_started.connect(self._on_program_started)
        events.process_output.connect(self._on_process_output)
        events.program_finished.connect(self._on_program_finished)
        events.diagnostics_changed.connect(self._on_diagnostics_changed)
        self.panel.navigate_requested.connect(events.navigate_to)
        self.panel.search_requested.connect(open_search)
        self.workbench.console.location_activated.connect(events.navigate_to)

    def _on_program_started(self, config: RunConfiguration) -> None:
        self._stderr = []
        self._working_dir = config.working_dir
        self._live_file = None
        self.workbench.console.set_link_base_dir(config.working_dir)

    def _on_process_output(self, text: str, stream: str) -> None:
        if stream == "stderr":
            self._stderr.append(text)

    def _on_program_finished(self, _exit_code: int) -> None:
        results = self.use_case.from_output("".join(self._stderr), self._working_dir)
        self._stderr = []
        self.panel.show_results(results)
        if results:
            self.dock.show()
            self.dock.raise_()

    def _on_diagnostics_changed(self, file: Path, diagnostics: list[Diagnostic]) -> None:
        if file != self.workbench.editor.current_file_path():
            return  # live diagnostics of background files would hide the one being edited
        results = self.use_case.from_diagnostics(diagnostics)
        if not results and self._live_file != file:
            return  # keep showing the last run's explanation
        self._live_file = file if results else None
        self.panel.show_results(results)
        if _has_errors(results):
            self.dock.show()


def _has_errors(results: list[ExplainedDiagnostic]) -> bool:
    return any(diagnostic.severity is Severity.ERROR for diagnostic, _unused in results)


def register(workbench: Workbench, explainer: ErrorExplainerPort | None = None) -> None:
    chosen = explainer or workbench.services.explainer or GoErrorExplainer()
    AssistantFeature(workbench, chosen).register()
