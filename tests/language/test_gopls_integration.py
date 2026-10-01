"""End-to-end with a real ``gopls``: diagnostics, completion and definition.

gopls may need a long time to load packages the first time (cold caches,
antivirus…), so these tests wait generously for the first diagnostics and then
retry the (1.5 s bounded) queries until gopls answers.
"""

import time
from pathlib import Path

import pytest
from PyQt5.QtWidgets import QApplication

from vizcacha.domain.diagnostics import Severity, SourceLocation
from vizcacha.infrastructure.go_toolchain import GoEnvironment
from vizcacha.infrastructure.gopls_lsp import GoplsLanguageServer
from vizcacha.infrastructure.settings import InMemorySettingsRepository

pytestmark = pytest.mark.requires_gopls

LOAD_TIMEOUT_S = 240
QUERY_TIMEOUT_S = 60
SOURCE = """package main

import "fmt"

func greet(name string) string {
	return "hola " + name
}

func main() {
	unused := 1
	fmt.Println(greet("vizcacha"))
	fmt.Pr
}
"""


def wait_for(condition, timeout_s: float):
    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        value = condition()
        if value:
            return value
        QApplication.processEvents()
        time.sleep(0.05)
    return condition()


@pytest.fixture(scope="module")
def session(qapp, tmp_path_factory):
    folder = tmp_path_factory.mktemp("gomodule")
    (folder / "go.mod").write_text("module example.com/demo\n\ngo 1.21\n", encoding="utf-8")
    source = folder / "main.go"
    source.write_text(SOURCE, encoding="utf-8")
    server = GoplsLanguageServer(GoEnvironment(InMemorySettingsRepository()))
    published: dict[Path, list] = {}
    server.diagnostics_published.connect(lambda path, items: published.__setitem__(path, items))
    server.open_document(source, SOURCE)
    wait_for(lambda: published.get(source), LOAD_TIMEOUT_S)
    yield server, source, published
    server.shutdown()


def test_file_with_errors_publishes_diagnostics(session):
    _server, source, published = session

    diagnostics = published.get(source, [])

    unused = [d for d in diagnostics if "unused" in d.message]
    assert unused, diagnostics
    assert unused[0].location == SourceLocation(source, 10, 2)
    assert unused[0].severity == Severity.ERROR
    assert unused[0].source == "gopls"


def test_completion_of_fmt_pr_includes_println(session):
    server, source, _published = session
    location = SourceLocation(source, 12, len("\tfmt.Pr") + 1)

    items = wait_for(lambda: server.completion(location), QUERY_TIMEOUT_S)

    assert "Println" in {item.label for item in items}


def test_definition_of_a_local_function(session):
    server, source, _published = session
    call = SourceLocation(source, 11, SOURCE.splitlines()[10].index("greet") + 1)

    target = wait_for(lambda: server.definition(call), QUERY_TIMEOUT_S)

    assert target is not None
    assert (target.line, target.column) == (5, 6)
    assert Path(target.file).resolve() == source.resolve()
