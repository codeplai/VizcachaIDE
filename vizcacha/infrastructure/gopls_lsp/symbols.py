"""Code-navigation results that the frozen domain does not model yet.

See "Contract change request" in the track C report: these could move to
``vizcacha/domain`` together with ``document_highlights``/``document_symbols`` in
``LanguageServerPort``.
"""

from dataclasses import dataclass, field

from vizcacha.domain.diagnostics import SourceLocation


@dataclass(frozen=True)
class TextRange:
    """From ``start`` (inclusive) to ``end`` (exclusive), both 1-based."""

    start: SourceLocation
    end: SourceLocation


@dataclass(frozen=True)
class OutlineSymbol:
    """One entry of the Outline: a function, type, variable… with its children."""

    name: str
    kind: str  # lowercase LSP SymbolKind name: "function", "struct", "method"…
    location: SourceLocation
    detail: str = ""
    children: tuple["OutlineSymbol", ...] = field(default_factory=tuple)
