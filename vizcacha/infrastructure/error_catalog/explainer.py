"""GoErrorExplainer: the ErrorExplainerPort implementation (parser + catalog)."""

from pathlib import Path

from vizcacha.domain.diagnostics import Diagnostic
from vizcacha.domain.explanations import ErrorExplanation
from vizcacha.infrastructure.error_catalog.catalog import ErrorCatalog
from vizcacha.infrastructure.error_catalog.go_output_parser import GoOutputParser


class GoErrorExplainer:
    def __init__(self, catalog: ErrorCatalog | None = None) -> None:
        self._catalog = catalog if catalog is not None else ErrorCatalog()
        self._parser = GoOutputParser(self._catalog.identify)

    def parse(self, raw_output: str, working_dir: Path) -> list[Diagnostic]:
        return self._parser.parse(raw_output, working_dir)

    def explain(self, diagnostic: Diagnostic) -> ErrorExplanation | None:
        return self._catalog.explain(diagnostic.message)
