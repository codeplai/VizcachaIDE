"""GoplsLanguageServer honours LanguageServerPort and degrades silently without gopls."""

from pathlib import Path

from PyQt5.QtCore import pyqtBoundSignal

from vizcacha.application.ports import LANGUAGE_SERVER_SIGNALS, LanguageServerPort
from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.infrastructure.go_toolchain import GoEnvironment
from vizcacha.infrastructure.gopls_lsp import NOT_FOUND, GoplsLanguageServer


def test_gopls_adapter_follows_contract(qapp, settings):
    adapter = GoplsLanguageServer(GoEnvironment(settings))

    assert isinstance(adapter, LanguageServerPort)
    for name in LANGUAGE_SERVER_SIGNALS:
        assert isinstance(getattr(adapter, name, None), pyqtBoundSignal), name


def test_missing_gopls_degrades_silently_and_warns_once(qtbot, settings, tmp_path: Path):
    settings.set(SettingsKeys.GOPLS_PATH, str(tmp_path / "no-such-gopls"))
    adapter = GoplsLanguageServer(GoEnvironment(settings))
    source = tmp_path / "main.go"
    location = SourceLocation(source, 1, 1)
    reasons: list[str] = []
    adapter.server_unavailable.connect(reasons.append)

    adapter.open_document(source, "package main\n")
    adapter.change_document(source, "package main\n\n", 2)
    adapter.open_document(tmp_path / "other.go", "package main\n")

    assert reasons == [NOT_FOUND]
    assert adapter.completion(location) == []
    assert adapter.hover(location) is None
    assert adapter.definition(location) is None
    assert adapter.signature_help(location) is None
    assert adapter.document_highlights(location) == []
    assert adapter.document_symbols(source) == []
    adapter.close_document(source)
    adapter.shutdown()


def test_queries_on_documents_that_are_not_open_return_nothing(qapp, settings):
    adapter = GoplsLanguageServer(GoEnvironment(settings))
    location = SourceLocation(Path("never_opened.go"), 1, 1)

    assert adapter.completion(location) == []
    assert adapter.hover(location) is None
    adapter.shutdown()
