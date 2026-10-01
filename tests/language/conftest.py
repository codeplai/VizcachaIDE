"""Fixtures for the language tests: a fake language server and a Workbench using it."""

from pathlib import Path

import pytest
from PyQt5.QtCore import QObject, pyqtSignal

from vizcacha.domain.completion import CompletionItem, SignatureHelp
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.infrastructure.gopls_lsp import OutlineSymbol
from vizcacha.ui.app import build_services, build_workbench

SOURCE = 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Pr\n}\n'


class FakeLanguageServer(QObject):
    diagnostics_published = pyqtSignal(object, object)
    server_unavailable = pyqtSignal(str)

    def __init__(self) -> None:
        super().__init__()
        self.calls: list[tuple] = []
        self.completions: list[CompletionItem] = []
        self.target: SourceLocation | None = None
        self.symbols: list[OutlineSymbol] = []

    def open_document(self, path, text):
        self.calls.append(("open", Path(path), text))

    def change_document(self, path, text, version):
        self.calls.append(("change", Path(path), text, version))

    def close_document(self, path):
        self.calls.append(("close", Path(path)))

    def completion(self, location):
        self.calls.append(("completion", location))
        return list(self.completions)

    def hover(self, location):
        return "func fake()"

    def definition(self, location):
        self.calls.append(("definition", location))
        return self.target

    def signature_help(self, location):
        self.calls.append(("signature_help", location))
        return SignatureHelp("f(a int)", parameters=("a int",))

    def document_highlights(self, location):
        return []

    def document_symbols(self, path):
        return list(self.symbols)

    def shutdown(self):
        self.calls.append(("shutdown",))

    def names(self) -> list[str]:
        return [call[0] for call in self.calls]


def build_with_server(qtbot, settings, server):
    services = build_services(settings)
    services.language_server = server
    workbench = build_workbench(services)
    qtbot.addWidget(workbench.window, before_close_func=lambda _window: mark_saved(workbench))
    return workbench


def mark_saved(workbench) -> None:
    """Avoid the modal "save changes?" dialog when pytest-qt closes the window."""
    for editor in workbench.editor.editors():
        editor.document().setModified(False)


@pytest.fixture
def source(tmp_path: Path) -> Path:
    path = tmp_path / "main.go"
    path.write_text(SOURCE, encoding="utf-8")
    return path


@pytest.fixture
def fake():
    return FakeLanguageServer()


@pytest.fixture
def make_workbench(qtbot, settings):
    """``make_workbench(server)`` -> Workbench whose language server is ``server``."""
    return lambda server: build_with_server(qtbot, settings, server)


@pytest.fixture
def opened(make_workbench, fake, source):
    workbench = make_workbench(fake)
    workbench.editor.open_file(source)
    return workbench, workbench.editor.current_editor()
