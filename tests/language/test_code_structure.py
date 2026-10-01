"""Domain types for ranges and symbols, and their mapping from LSP results."""

from pathlib import Path

import pytest
from lsprotocol import types

from vizcacha.domain.code_structure import DocumentSymbol, SourceRange, SymbolKind
from vizcacha.domain.diagnostics import Diagnostic, Severity, SourceLocation
from vizcacha.infrastructure.gopls_lsp import symbols

PATH = Path("shapes.go")
TEXT = "package main\n\ntype Point struct {\n\tX int\n}\n\nvar año = 1\n"


def at(line: int, column: int) -> SourceLocation:
    return SourceLocation(PATH, line, column)


def lsp_range(start: tuple[int, int], end: tuple[int, int]) -> types.Range:
    return types.Range(types.Position(*start), types.Position(*end))


def test_source_range_knows_when_it_is_empty():
    assert SourceRange(at(1, 1), at(1, 1)).is_empty
    assert SourceRange(at(2, 5), at(1, 9)).is_empty
    assert not SourceRange(at(1, 1), at(1, 2)).is_empty
    assert not SourceRange(at(1, 9), at(2, 1)).is_empty


def test_document_symbol_defaults_and_diagnostic_end_is_optional():
    symbol = DocumentSymbol("main", SymbolKind.FUNCTION, at(5, 6))
    diagnostic = Diagnostic(at(1, 1), Severity.ERROR, "m", "m")

    assert (symbol.range, symbol.detail, symbol.children) == (None, "", ())
    assert diagnostic.end is None


def test_lsp_range_counts_code_points():
    mapped = symbols.to_range(lsp_range((6, 4), (6, 7)), PATH, TEXT)

    assert mapped == SourceRange(at(7, 5), at(7, 8))


@pytest.mark.parametrize(
    ("kind", "expected"),
    [
        (types.SymbolKind.Function, SymbolKind.FUNCTION),
        (types.SymbolKind.Method, SymbolKind.METHOD),
        (types.SymbolKind.Struct, SymbolKind.STRUCT),
        (types.SymbolKind.Class, SymbolKind.TYPE),
        (types.SymbolKind.Field, SymbolKind.FIELD),
        (types.SymbolKind.Event, SymbolKind.OTHER),
        (999, SymbolKind.OTHER),
    ],
)
def test_symbol_kinds(kind, expected):
    assert symbols.to_symbol_kind(kind) == expected


def test_nested_document_symbols_keep_ranges_and_children():
    field = types.DocumentSymbol(
        name="X",
        kind=types.SymbolKind.Field,
        range=lsp_range((3, 1), (3, 6)),
        selection_range=lsp_range((3, 1), (3, 2)),
    )
    point = types.DocumentSymbol(
        name="Point",
        kind=types.SymbolKind.Struct,
        range=lsp_range((2, 0), (4, 1)),
        selection_range=lsp_range((2, 5), (2, 10)),
        detail="struct{...}",
        children=[field],
    )

    [mapped] = symbols.to_document_symbols([point], PATH, TEXT)

    assert (mapped.name, mapped.kind, mapped.location) == ("Point", SymbolKind.STRUCT, at(3, 6))
    assert mapped.range == SourceRange(at(3, 1), at(5, 2))
    assert mapped.detail == "struct{...}"
    field_range = SourceRange(at(4, 2), at(4, 7))
    assert mapped.children == (DocumentSymbol("X", SymbolKind.FIELD, at(4, 2), field_range),)


def test_flat_symbol_information_is_mapped_too():
    information = types.SymbolInformation(
        name="año",
        kind=types.SymbolKind.Variable,
        location=types.Location(PATH.resolve().as_uri(), lsp_range((6, 4), (6, 7))),
    )

    [mapped] = symbols.to_document_symbols([information], PATH, TEXT)

    whole = SourceRange(at(7, 5), at(7, 8))
    assert mapped == DocumentSymbol("año", SymbolKind.VARIABLE, at(7, 5), whole)
