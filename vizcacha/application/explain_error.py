"""Use case: explain the problems found in Go output or reported by a language server."""

from collections.abc import Iterable
from pathlib import Path

from vizcacha.application.ports import ErrorExplainerPort
from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.domain.explanations import ErrorExplanation

ExplainedDiagnostic = tuple[Diagnostic, ErrorExplanation | None]

PROBLEM_SEVERITIES = (Severity.ERROR, Severity.WARNING)


class ExplainError:
    def __init__(self, explainer: ErrorExplainerPort) -> None:
        self._explainer = explainer

    def from_output(self, raw_output: str, working_dir: Path) -> list[ExplainedDiagnostic]:
        """Parse raw ``go build`` / ``go run`` / ``go vet`` output and explain each problem."""
        return self.from_diagnostics(self._explainer.parse(raw_output, working_dir))

    def from_diagnostics(self, diagnostics: Iterable[Diagnostic]) -> list[ExplainedDiagnostic]:
        """Explain errors and warnings (hints and infos are skipped), without duplicates."""
        problems = [item for item in diagnostics if item.severity in PROBLEM_SEVERITIES]
        unique = list(dict.fromkeys(problems))
        return [(diagnostic, self._explainer.explain(diagnostic)) for diagnostic in unique]
