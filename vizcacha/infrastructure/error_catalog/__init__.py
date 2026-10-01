"""Beginner-friendly explanations of Go compiler errors, vet warnings and panics."""

from vizcacha.infrastructure.error_catalog.catalog import DEFAULT_ENTRIES, ErrorCatalog
from vizcacha.infrastructure.error_catalog.catalog_entry import CatalogEntry
from vizcacha.infrastructure.error_catalog.explainer import GoErrorExplainer
from vizcacha.infrastructure.error_catalog.go_output_parser import GoOutputParser
from vizcacha.infrastructure.error_catalog.source_links import (
    SourceLink,
    find_source_links,
    resolve_go_path,
)

__all__ = [
    "DEFAULT_ENTRIES",
    "CatalogEntry",
    "ErrorCatalog",
    "GoErrorExplainer",
    "GoOutputParser",
    "SourceLink",
    "find_source_links",
    "resolve_go_path",
]
