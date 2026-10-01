"""Parameters of the messages VizcachaIDE sends to gopls (lsprotocol types)."""

import os
from pathlib import Path

from lsprotocol import types

from vizcacha import __version__
from vizcacha.infrastructure.gopls_lsp.positions import path_to_uri

CLIENT_NAME = "VizcachaIDE"
LANGUAGE_ID = "go"
PLAIN = [types.MarkupKind.PlainText]


def _text_document_capabilities() -> types.TextDocumentClientCapabilities:
    return types.TextDocumentClientCapabilities(
        synchronization=types.TextDocumentSyncClientCapabilities(did_save=False),
        completion=types.CompletionClientCapabilities(
            completion_item=types.ClientCompletionItemOptions(
                snippet_support=False, documentation_format=PLAIN
            )
        ),
        hover=types.HoverClientCapabilities(content_format=PLAIN),
        signature_help=types.SignatureHelpClientCapabilities(
            signature_information=types.ClientSignatureInformationOptions(
                documentation_format=PLAIN,
                parameter_information=types.ClientSignatureParameterInformationOptions(
                    label_offset_support=True
                ),
                active_parameter_support=True,
            )
        ),
        definition=types.DefinitionClientCapabilities(link_support=False),
        document_highlight=types.DocumentHighlightClientCapabilities(),
        document_symbol=types.DocumentSymbolClientCapabilities(
            hierarchical_document_symbol_support=True
        ),
        publish_diagnostics=types.PublishDiagnosticsClientCapabilities(),
    )


def initialize_params(root: Path) -> types.InitializeParams:
    root_uri = path_to_uri(root)
    return types.InitializeParams(
        process_id=os.getpid(),
        client_info=types.ClientInfo(name=CLIENT_NAME, version=__version__),
        root_uri=root_uri,
        workspace_folders=[types.WorkspaceFolder(uri=root_uri, name=root.name or root_uri)],
        capabilities=types.ClientCapabilities(
            workspace=types.WorkspaceClientCapabilities(workspace_folders=True),
            text_document=_text_document_capabilities(),
            general=types.GeneralClientCapabilities(
                position_encodings=[types.PositionEncodingKind.Utf16]
            ),
        ),
    )


def add_folder_params(root: Path) -> types.DidChangeWorkspaceFoldersParams:
    folder = types.WorkspaceFolder(uri=path_to_uri(root), name=root.name or str(root))
    return types.DidChangeWorkspaceFoldersParams(
        event=types.WorkspaceFoldersChangeEvent(added=[folder], removed=[])
    )


def did_open_params(uri: str, text: str, version: int) -> types.DidOpenTextDocumentParams:
    return types.DidOpenTextDocumentParams(
        text_document=types.TextDocumentItem(
            uri=uri, language_id=LANGUAGE_ID, version=version, text=text
        )
    )


def did_change_params(uri: str, text: str, version: int) -> types.DidChangeTextDocumentParams:
    return types.DidChangeTextDocumentParams(
        text_document=types.VersionedTextDocumentIdentifier(uri=uri, version=version),
        content_changes=[types.TextDocumentContentChangeWholeDocument(text=text)],
    )


def did_close_params(uri: str) -> types.DidCloseTextDocumentParams:
    return types.DidCloseTextDocumentParams(text_document=types.TextDocumentIdentifier(uri=uri))


def position_params(method: str, uri: str, position: types.Position) -> object:
    """Params of ``method`` (completion, hover, definition…): document + position."""
    params_cls = types.METHOD_TO_TYPES[method][2]
    return params_cls(text_document=types.TextDocumentIdentifier(uri=uri), position=position)


def document_params(uri: str) -> types.DocumentSymbolParams:
    return types.DocumentSymbolParams(text_document=types.TextDocumentIdentifier(uri=uri))
