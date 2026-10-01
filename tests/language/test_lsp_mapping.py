"""lsprotocol -> domain mapping, fed with a transcript recorded from gopls v0.23."""

import json
from pathlib import Path

import pytest
from lsprotocol import types
from lsprotocol.converters import get_converter

from vizcacha.domain.completion import CompletionKind
from vizcacha.domain.diagnostics import Severity, SourceLocation
from vizcacha.infrastructure.gopls_lsp import domain_mapping as mapping
from vizcacha.infrastructure.gopls_lsp.client_messages import (
    did_change_params,
    initialize_params,
)
from vizcacha.infrastructure.gopls_lsp.documents import OpenDocuments
from vizcacha.infrastructure.gopls_lsp.positions import (
    from_lsp_position,
    module_root,
    path_to_uri,
    to_lsp_position,
    uri_to_path,
)

TRANSCRIPTS = Path(__file__).parent / "transcripts"
CONVERTER = get_converter()


@pytest.fixture
def recorded(tmp_path: Path):
    """The recorded session, re-rooted in ``tmp_path`` (where main.go is written)."""
    source = tmp_path / "main.go"
    source.write_text((TRANSCRIPTS / "session_main.go").read_text(encoding="utf-8"), "utf-8")
    raw = (TRANSCRIPTS / "gopls_session.json").read_text(encoding="utf-8")
    messages = json.loads(raw.replace("{ROOT_URI}", tmp_path.resolve().as_uri()))
    documents = OpenDocuments()
    documents.open(source, source.read_text(encoding="utf-8"))
    return messages, source, documents


def result_of(message: dict, method: str):
    return CONVERTER.structure(message, types.METHOD_TO_TYPES[method][1]).result


def test_publish_diagnostics_become_domain_diagnostics(recorded):
    messages, source, documents = recorded

    path, diagnostics = mapping.published_diagnostics(
        messages["publishDiagnostics"]["params"], documents
    )

    assert path == source
    [diagnostic] = diagnostics
    assert diagnostic.location == SourceLocation(source, 10, 2)
    assert diagnostic.severity == Severity.ERROR
    assert diagnostic.message == diagnostic.raw_text == "declared and not used: x"
    assert (diagnostic.source, diagnostic.code) == ("gopls", "UnusedVar")


@pytest.mark.parametrize(
    ("severity", "expected"),
    [(2, Severity.WARNING), (3, Severity.INFO), (4, Severity.HINT), (None, Severity.ERROR)],
)
def test_diagnostic_severity_and_numeric_code(severity, expected):
    raw = {"range": {"start": {"line": 0, "character": 3}, "end": {"line": 0, "character": 4}}}
    raw.update(message="m", code=42, severity=severity)
    item = CONVERTER.structure(raw, types.Diagnostic)

    diagnostic = mapping.to_diagnostic(item, Path("a.go"), "abcdef")

    assert diagnostic.severity == expected
    assert diagnostic.code == "42"
    assert diagnostic.location.column == 4


def test_completion_items(recorded):
    messages, _source, _documents = recorded

    items = mapping.to_completion_items(
        result_of(messages["completion"], types.TEXT_DOCUMENT_COMPLETION)
    )

    assert [item.label for item in items] == ["Print", "Printf", "Println", "Fprint"]
    assert {item.kind for item in items} == {CompletionKind.FUNCTION}
    assert items[2].detail.startswith("func(a ...any)")
    assert items[2].text_to_insert == "Println"


def test_hover_definition_signature_highlights_and_outline(recorded):
    messages, source, documents = recorded

    hover = mapping.to_hover_text(result_of(messages["hover"], types.TEXT_DOCUMENT_HOVER))
    definition = mapping.to_location(
        result_of(messages["definition"], types.TEXT_DOCUMENT_DEFINITION), documents.text_of
    )
    signature = mapping.to_signature_help(
        result_of(messages["signatureHelp"], types.TEXT_DOCUMENT_SIGNATURE_HELP)
    )
    text = documents.text_of(source)
    ranges = mapping.to_ranges(
        result_of(messages["documentHighlight"], types.TEXT_DOCUMENT_DOCUMENT_HIGHLIGHT),
        source,
        text,
    )
    outline = mapping.to_outline(
        result_of(messages["documentSymbol"], types.TEXT_DOCUMENT_DOCUMENT_SYMBOL), source, text
    )

    assert hover.startswith("func fmt.Println(a ...any)")
    assert definition == SourceLocation(source, 5, 6)
    assert signature.label == "greet(name string) string"
    assert signature.parameters == ("name string",)
    assert [(r.start.line, r.start.column, r.end.column) for r in ranges] == [
        (5, 6, 11),
        (11, 14, 19),
    ]
    assert [(s.name, s.kind, s.location.line) for s in outline] == [
        ("greet", "function", 5),
        ("main", "function", 9),
    ]


def test_empty_results_map_to_empty_values():
    assert mapping.to_completion_items(None) == []
    assert mapping.to_hover_text(None) is None
    assert mapping.to_location(None, lambda path: "") is None
    assert mapping.to_location([], lambda path: "") is None
    assert mapping.to_signature_help(None) is None
    assert mapping.to_outline(None, Path("a.go"), "") == []


def test_signature_parameter_given_as_utf16_offsets():
    help_result = types.SignatureHelp(
        signatures=[
            types.SignatureInformation(
                label="f(año int, b string)",
                parameters=[types.ParameterInformation(label=(2, 9))],
            )
        ],
        active_parameter=0,
    )

    assert mapping.to_signature_help(help_result).parameters == ("año int",)


def test_positions_count_utf16_units_but_columns_count_code_points():
    text = 'package main\ns := "😀" + x\n'

    position = to_lsp_position(text, 2, 12)  # the "x", after a 2-unit emoji

    assert (position.line, position.character) == (1, 12)
    assert from_lsp_position(text, types.Position(line=1, character=12)) == (2, 12)


def test_uri_round_trip_and_module_root(tmp_path: Path):
    (tmp_path / "go.mod").write_text("module demo\n", encoding="utf-8")
    nested = tmp_path / "cmd" / "tool"
    nested.mkdir(parents=True)
    loose = tmp_path.parent / "loose.go"

    assert uri_to_path(path_to_uri(nested / "main.go")) == (nested / "main.go").resolve()
    assert module_root(nested / "main.go") == tmp_path.resolve()
    assert module_root(loose) == loose.resolve().parent


def test_client_messages_serialize_to_lsp_json(tmp_path: Path):
    initialize = CONVERTER.unstructure(initialize_params(tmp_path))
    change = CONVERTER.unstructure(did_change_params("file:///a.go", "package a", 3))

    assert initialize["rootUri"] == tmp_path.resolve().as_uri()
    assert (
        initialize["capabilities"]["textDocument"]["completion"]["completionItem"]["snippetSupport"]
        is False
    )
    assert change["textDocument"] == {"uri": "file:///a.go", "version": 3}
    assert change["contentChanges"] == [{"text": "package a"}]
