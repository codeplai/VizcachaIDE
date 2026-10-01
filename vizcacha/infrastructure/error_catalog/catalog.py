"""ErrorCatalog: finds the explanation for a Go message (regex -> stable id)."""

from collections.abc import Sequence

from vizcacha.domain.explanations import ErrorExplanation
from vizcacha.infrastructure.error_catalog.catalog_entry import CatalogEntry
from vizcacha.infrastructure.error_catalog.compile_errors import COMPILE_ERRORS
from vizcacha.infrastructure.error_catalog.runtime_panics import RUNTIME_PANICS
from vizcacha.infrastructure.error_catalog.syntax_type_errors import SYNTAX_TYPE_ERRORS
from vizcacha.infrastructure.error_catalog.vet_warnings import VET_WARNINGS

DEFAULT_ENTRIES: tuple[CatalogEntry, ...] = (
    COMPILE_ERRORS + SYNTAX_TYPE_ERRORS + RUNTIME_PANICS + VET_WARNINGS
)


class ErrorCatalog:
    def __init__(self, entries: Sequence[CatalogEntry] = DEFAULT_ENTRIES) -> None:
        self._entries = tuple(entries)

    def entries(self) -> tuple[CatalogEntry, ...]:
        return self._entries

    def identify(self, message: str) -> str:
        """Stable id of the first matching entry, or "" when the message is unknown."""
        explanation = self.explain(message)
        return explanation.explanation_id if explanation is not None else ""

    def explain(self, message: str) -> ErrorExplanation | None:
        for entry in self._entries:
            match = entry.search(message)
            if match is None:
                continue
            placeholders = {
                name: value for name, value in match.groupdict().items() if value is not None
            }
            return ErrorExplanation(
                explanation_id=entry.explanation_id,
                title=entry.title,
                body=entry.body,
                fix_hint=entry.fix_hint,
                placeholders=placeholders,
            )
        return None
