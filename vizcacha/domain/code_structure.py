"""The structure of a Go source file: text ranges and the symbols it declares."""

from dataclasses import dataclass
from enum import Enum

from vizcacha.domain.diagnostics import SourceLocation


@dataclass(frozen=True)
class SourceRange:
    """From ``start`` (inclusive) to ``end`` (exclusive), both 1-based."""

    start: SourceLocation
    end: SourceLocation

    @property
    def is_empty(self) -> bool:
        return (self.start.line, self.start.column) >= (self.end.line, self.end.column)


class SymbolKind(Enum):
    FUNCTION = "function"
    METHOD = "method"
    STRUCT = "struct"
    INTERFACE = "interface"
    TYPE = "type"
    VARIABLE = "variable"
    CONSTANT = "constant"
    FIELD = "field"
    PACKAGE = "package"
    OTHER = "other"


@dataclass(frozen=True)
class DocumentSymbol:
    """One declaration of a file (for the Outline), with its nested declarations.

    ``location`` is where the symbol's name is (the navigation target) and
    ``range`` the whole declaration, when the language server reports it.
    """

    name: str
    kind: SymbolKind
    location: SourceLocation
    range: SourceRange | None = None
    detail: str = ""
    children: tuple["DocumentSymbol", ...] = ()
