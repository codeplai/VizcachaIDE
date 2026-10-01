"""Translate LSP ranges, highlights and document symbols into the domain."""

from pathlib import Path

from lsprotocol import types

from vizcacha.domain.code_structure import DocumentSymbol, SourceRange, SymbolKind
from vizcacha.infrastructure.gopls_lsp.positions import location_in

L = types.SymbolKind
SYMBOL_KINDS = {
    L.Function: SymbolKind.FUNCTION,
    L.Constructor: SymbolKind.FUNCTION,
    L.Method: SymbolKind.METHOD,
    L.Struct: SymbolKind.STRUCT,
    L.Interface: SymbolKind.INTERFACE,
    L.Class: SymbolKind.TYPE,
    L.Enum: SymbolKind.TYPE,
    L.TypeParameter: SymbolKind.TYPE,
    L.Variable: SymbolKind.VARIABLE,
    L.Constant: SymbolKind.CONSTANT,
    L.EnumMember: SymbolKind.CONSTANT,
    L.Field: SymbolKind.FIELD,
    L.Property: SymbolKind.FIELD,
    L.Package: SymbolKind.PACKAGE,
    L.Module: SymbolKind.PACKAGE,
    L.Namespace: SymbolKind.PACKAGE,
}


def to_range(lsp_range: types.Range, path: Path, text: str) -> SourceRange:
    return SourceRange(
        location_in(text, path, lsp_range.start), location_in(text, path, lsp_range.end)
    )


def to_ranges(highlights: object, path: Path, text: str) -> list[SourceRange]:
    """``textDocument/documentHighlight`` result -> ranges in ``path``."""
    return [to_range(item.range, path, text) for item in highlights or ()]


def to_symbol_kind(kind: object) -> SymbolKind:
    try:
        return SYMBOL_KINDS.get(types.SymbolKind(kind), SymbolKind.OTHER)
    except ValueError:
        return SymbolKind.OTHER


def to_document_symbols(symbols: object, path: Path, text: str) -> list[DocumentSymbol]:
    """``textDocument/documentSymbol`` result (nested or flat) -> domain symbols."""
    return [_document_symbol(symbol, path, text) for symbol in symbols or ()]


def _document_symbol(symbol: object, path: Path, text: str) -> DocumentSymbol:
    if isinstance(symbol, types.DocumentSymbol):
        children = tuple(_document_symbol(child, path, text) for child in symbol.children or ())
        whole = to_range(symbol.range, path, text)
        location = location_in(text, path, symbol.selection_range.start)
        return DocumentSymbol(
            symbol.name, to_symbol_kind(symbol.kind), location, whole, symbol.detail or "", children
        )
    whole = to_range(symbol.location.range, path, text)  # SymbolInformation (flat list)
    return DocumentSymbol(symbol.name, to_symbol_kind(symbol.kind), whole.start, whole)
