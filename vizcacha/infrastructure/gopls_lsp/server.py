"""GoplsLanguageServer: LanguageServerPort implemented with ``gopls serve``.

Design decision: the port's queries (completion, hover…) are synchronous. They
are answered by waiting for gopls on the calling (UI) thread for at most
``request_timeout_ms`` (1.5 s by default; usually gopls answers in a few ms).
On timeout, error or a missing/crashed gopls they return an empty result, so the
UI degrades to the static completion instead of failing. After a timeout (gopls
is typically still loading packages) the next queries wait only
``BUSY_TIMEOUT_MS`` until one is answered again, so the UI never stutters for long.
"""

import logging
from pathlib import Path

from lsprotocol import types
from PyQt5.QtCore import QObject, Qt, pyqtSignal

from vizcacha.application.errors import LanguageServerError
from vizcacha.domain.completion import CompletionItem, SignatureHelp
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.infrastructure.go_toolchain import GoEnvironment
from vizcacha.infrastructure.gopls_lsp import client_messages as messages
from vizcacha.infrastructure.gopls_lsp import domain_mapping as mapping
from vizcacha.infrastructure.gopls_lsp.connection import GoplsConnection
from vizcacha.infrastructure.gopls_lsp.documents import OpenDocuments
from vizcacha.infrastructure.gopls_lsp.jsonrpc import LanguageServerTimeoutError
from vizcacha.infrastructure.gopls_lsp.positions import module_root, path_to_uri, to_lsp_position
from vizcacha.infrastructure.gopls_lsp.symbols import OutlineSymbol, TextRange

LOG = logging.getLogger(__name__)
REQUEST_TIMEOUT_MS = 1500
BUSY_TIMEOUT_MS = 150  # after a timeout (e.g. gopls still loading packages)
NOT_FOUND, CRASHED = "not_found", "crashed"  # reasons of server_unavailable
IDLE, STARTING, READY, UNAVAILABLE = "idle", "starting", "ready", "unavailable"


class GoplsLanguageServer(QObject):
    diagnostics_published = pyqtSignal(object, object)  # Path, list[Diagnostic]
    server_unavailable = pyqtSignal(str)  # NOT_FOUND | CRASHED, emitted at most once

    def __init__(
        self,
        environment: GoEnvironment,
        parent: QObject | None = None,
        request_timeout_ms: int = REQUEST_TIMEOUT_MS,
    ) -> None:
        super().__init__(parent)
        self._environment = environment
        self._timeout_ms = min(request_timeout_ms, REQUEST_TIMEOUT_MS)
        self._documents = OpenDocuments()
        self._connection: GoplsConnection | None = None
        self._state = IDLE
        self._init_id = 0
        self._folders: set[Path] = set()
        self._queue: list[tuple[str, object]] = []
        self._busy = False

    # --- document synchronisation -----------------------------------------
    def open_document(self, path: Path, text: str) -> None:
        self._ensure_started(Path(path))
        if self._state == UNAVAILABLE:
            return
        self._ensure_folder(module_root(path))
        uri = self._documents.open(path, text)
        self._notify(types.TEXT_DOCUMENT_DID_OPEN, messages.did_open_params(uri, text, 1))

    def change_document(self, path: Path, text: str, version: int) -> None:
        uri = self._documents.change(path, text, version)
        if uri is None:
            return
        version = self._documents.get(uri).version
        self._notify(types.TEXT_DOCUMENT_DID_CHANGE, messages.did_change_params(uri, text, version))

    def close_document(self, path: Path) -> None:
        uri = self._documents.close(path)
        if uri is not None:
            self._notify(types.TEXT_DOCUMENT_DID_CLOSE, messages.did_close_params(uri))

    # --- queries ------------------------------------------------------------
    def completion(self, location: SourceLocation) -> list[CompletionItem]:
        return mapping.to_completion_items(self._at(types.TEXT_DOCUMENT_COMPLETION, location))

    def hover(self, location: SourceLocation) -> str | None:
        return mapping.to_hover_text(self._at(types.TEXT_DOCUMENT_HOVER, location))

    def definition(self, location: SourceLocation) -> SourceLocation | None:
        result = self._at(types.TEXT_DOCUMENT_DEFINITION, location)
        return mapping.to_location(result, self._documents.text_of)

    def signature_help(self, location: SourceLocation) -> SignatureHelp | None:
        return mapping.to_signature_help(self._at(types.TEXT_DOCUMENT_SIGNATURE_HELP, location))

    def document_highlights(self, location: SourceLocation) -> list[TextRange]:
        result = self._at(types.TEXT_DOCUMENT_DOCUMENT_HIGHLIGHT, location)
        return mapping.to_ranges(result, location.file, self._documents.text_of(location.file))

    def document_symbols(self, path: Path) -> list[OutlineSymbol]:
        uri = path_to_uri(path)
        if self._documents.get(uri) is None:
            return []
        result = self._request(types.TEXT_DOCUMENT_DOCUMENT_SYMBOL, messages.document_params(uri))
        return mapping.to_outline(result, path, self._documents.text_of(path))

    def shutdown(self) -> None:
        connection, self._connection = self._connection, None
        self._state = IDLE
        self._queue.clear()
        self._documents = OpenDocuments()
        if connection is not None:
            connection.close()

    # --- internals ------------------------------------------------------------
    def _ensure_started(self, path: Path) -> None:
        if self._state != IDLE:
            return
        root = module_root(path)
        connection = GoplsConnection(self)
        # Queued: handlers (e.g. the UI reacting to diagnostics) never run nested in a read.
        connection.notification_received.connect(self._on_notification, Qt.QueuedConnection)
        connection.response_received.connect(self._on_response)
        connection.failed.connect(lambda _reason: self._disable(CRASHED))
        try:
            connection.start(
                self._environment.gopls_executable(), self._environment.variables(), root
            )
            self._connection = connection
            self._init_id = connection.send_request(
                types.INITIALIZE, messages.initialize_params(root)
            )
        except LanguageServerError as error:
            LOG.warning("Language server unavailable: %s", error)
            self._disable(NOT_FOUND)
            return
        self._folders = {root}
        self._state = STARTING

    def _ensure_folder(self, root: Path) -> None:
        if root in self._folders:
            return
        self._folders.add(root)
        self._notify(types.WORKSPACE_DID_CHANGE_WORKSPACE_FOLDERS, messages.add_folder_params(root))

    def _on_response(self, request_id: int) -> None:
        if request_id != self._init_id or self._state != STARTING:
            return
        try:
            self._connection.take_result(request_id, types.INITIALIZE)
        except LanguageServerError as error:
            LOG.warning("gopls rejected initialize: %s", error)
            self._disable(CRASHED)
            return
        self._state = READY
        self._connection.notify(types.INITIALIZED, types.InitializedParams())
        queued, self._queue = self._queue, []
        for method, params in queued:
            self._connection.notify(method, params)

    def _on_notification(self, method: str, params: object) -> None:
        if method == types.TEXT_DOCUMENT_PUBLISH_DIAGNOSTICS:
            path, diagnostics = mapping.published_diagnostics(params, self._documents)
            self.diagnostics_published.emit(path, diagnostics)

    def _notify(self, method: str, params: object) -> None:
        if self._state == STARTING:
            self._queue.append((method, params))
        elif self._state == READY:
            self._connection.notify(method, params)

    def _at(self, method: str, location: SourceLocation) -> object:
        uri = path_to_uri(location.file)
        document = self._documents.get(uri)
        if document is None:
            return None
        position = to_lsp_position(document.text, location.line, location.column)
        return self._request(method, messages.position_params(method, uri, position))

    def _request(self, method: str, params: object) -> object:
        if self._state not in (STARTING, READY):
            return None
        timeout_ms = BUSY_TIMEOUT_MS if self._busy else self._timeout_ms
        try:
            result = self._connection.request(
                method, params, timeout_ms, ready=lambda: self._state == READY
            )
        except LanguageServerTimeoutError:
            self._busy = True
            return None
        except LanguageServerError as error:
            LOG.info("%s failed: %s", method, error)
            return None
        self._busy = False
        return result

    def _disable(self, reason: str) -> None:
        if self._state == UNAVAILABLE:
            return
        self._state = UNAVAILABLE
        self._queue.clear()
        self.server_unavailable.emit(reason)
