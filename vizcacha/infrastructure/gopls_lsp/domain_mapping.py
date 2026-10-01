"""Translate lsprotocol results into domain objects (pure functions, no process)."""

from collections.abc import Callable, Sequence
from pathlib import Path

from lsprotocol import types
from lsprotocol.converters import get_converter

from vizcacha.domain.completion import CompletionItem, CompletionKind, SignatureHelp
from vizcacha.domain.diagnostics import Diagnostic, Severity, SourceLocation
from vizcacha.infrastructure.gopls_lsp.documents import OpenDocuments
from vizcacha.infrastructure.gopls_lsp.positions import location_in, uri_to_path

TextLookup = Callable[[Path], str]  # full text of a file (open document or disk)
DIAGNOSTIC_SOURCE = "gopls"
MAX_COMPLETIONS = 60
CONVERTER = get_converter()

SEVERITIES = {
    types.DiagnosticSeverity.Error: Severity.ERROR,
    types.DiagnosticSeverity.Warning: Severity.WARNING,
    types.DiagnosticSeverity.Information: Severity.INFO,
    types.DiagnosticSeverity.Hint: Severity.HINT,
}
K = types.CompletionItemKind
COMPLETION_KINDS = {
    K.Method: CompletionKind.METHOD,
    K.Function: CompletionKind.FUNCTION,
    K.Constructor: CompletionKind.FUNCTION,
    K.Field: CompletionKind.FIELD,
    K.Property: CompletionKind.FIELD,
    K.Variable: CompletionKind.VARIABLE,
    K.Constant: CompletionKind.CONSTANT,
    K.EnumMember: CompletionKind.CONSTANT,
    K.Class: CompletionKind.TYPE,
    K.Interface: CompletionKind.TYPE,
    K.Struct: CompletionKind.TYPE,
    K.Enum: CompletionKind.TYPE,
    K.TypeParameter: CompletionKind.TYPE,
    K.Module: CompletionKind.PACKAGE,
    K.Keyword: CompletionKind.KEYWORD,
}


def markup_text(value: object) -> str:
    """Plain text of a ``str | MarkupContent | MarkedString`` value (or "")."""
    if value is None:
        return ""
    if isinstance(value, str):
        return value
    return getattr(value, "value", "") or ""


def published_diagnostics(
    raw_params: object, documents: OpenDocuments
) -> tuple[Path, list[Diagnostic]]:
    """``textDocument/publishDiagnostics`` params (JSON) -> editor path + diagnostics."""
    params = CONVERTER.structure(raw_params, types.PublishDiagnosticsParams)
    path = documents.path_for(params.uri)
    return path, [to_diagnostic(item, path, documents.text_of(path)) for item in params.diagnostics]


def to_diagnostics(
    params: types.PublishDiagnosticsParams, text: str
) -> tuple[Path, list[Diagnostic]]:
    path = uri_to_path(params.uri)
    return path, [to_diagnostic(item, path, text) for item in params.diagnostics]


def to_diagnostic(item: types.Diagnostic, path: Path, text: str) -> Diagnostic:
    return Diagnostic(
        location=location_in(text, path, item.range.start),
        severity=SEVERITIES.get(item.severity, Severity.ERROR),
        message=item.message,
        raw_text=item.message,
        source=DIAGNOSTIC_SOURCE,
        code="" if item.code is None else str(item.code),
        end=location_in(text, path, item.range.end),
    )


def to_completion_items(result: object) -> list[CompletionItem]:
    if result is None:
        return []
    items = result.items if isinstance(result, types.CompletionList) else result
    return [to_completion_item(item) for item in list(items)[:MAX_COMPLETIONS]]


def to_completion_item(item: types.CompletionItem) -> CompletionItem:
    edit = item.text_edit
    insert_text = getattr(edit, "new_text", None) or item.insert_text or ""
    return CompletionItem(
        label=item.label,
        kind=COMPLETION_KINDS.get(item.kind, CompletionKind.OTHER),
        detail=item.detail or "",
        documentation=markup_text(item.documentation),
        insert_text=insert_text if insert_text != item.label else "",
    )


def to_hover_text(result: types.Hover | None) -> str | None:
    if result is None:
        return None
    contents = result.contents
    if isinstance(contents, Sequence) and not isinstance(contents, str):
        text = "\n\n".join(markup_text(part) for part in contents)
    else:
        text = markup_text(contents)
    return text.strip() or None


def to_location(result: object, lookup: TextLookup) -> SourceLocation | None:
    """First target of a definition result (Location, list of Location or LocationLink)."""
    if result is None:
        return None
    targets = [result] if isinstance(result, types.Location) else list(result)
    if not targets:
        return None
    target = targets[0]
    uri = getattr(target, "target_uri", None) or target.uri
    selected = getattr(target, "target_selection_range", None) or target.range
    path = uri_to_path(uri)
    return location_in(lookup(path), path, selected.start)


def to_signature_help(result: types.SignatureHelp | None) -> SignatureHelp | None:
    if result is None or not result.signatures:
        return None
    index = result.active_signature or 0
    signature = result.signatures[min(index, len(result.signatures) - 1)]
    parameters = tuple(_parameter_label(signature.label, p) for p in signature.parameters or ())
    active = signature.active_parameter
    if active is None:
        active = result.active_parameter or 0
    return SignatureHelp(
        label=signature.label,
        documentation=markup_text(signature.documentation),
        parameters=parameters,
        active_parameter=active,
    )


def _parameter_label(signature_label: str, parameter: types.ParameterInformation) -> str:
    label = parameter.label
    if isinstance(label, str):
        return label
    start, end = label  # UTF-16 offsets inside the signature label
    encoded = signature_label.encode("utf-16-le")
    return encoded[start * 2 : end * 2].decode("utf-16-le")
