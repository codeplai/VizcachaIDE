"""Shared fixtures. Qt runs offscreen so tests work in CI without a display."""

import os
import shutil

os.environ.setdefault("QT_QPA_PLATFORM", "offscreen")

import pytest  # noqa: E402

from vizcacha.i18n import install_language  # noqa: E402
from vizcacha.infrastructure.settings import InMemorySettingsRepository  # noqa: E402

TOOL_MARKERS = {"requires_go": "go", "requires_dlv": "dlv", "requires_gopls": "gopls"}


def pytest_collection_modifyitems(config, items):
    for item in items:
        for marker, executable in TOOL_MARKERS.items():
            if marker in item.keywords and shutil.which(executable) is None:
                item.add_marker(pytest.mark.skip(reason=f"{executable} not found on PATH"))


@pytest.fixture(autouse=True)
def english_ui():
    """Every test starts in English; tests that need Spanish switch explicitly."""
    install_language("en")
    yield
    install_language("en")


@pytest.fixture
def settings():
    return InMemorySettingsRepository()


@pytest.fixture
def workbench(qtbot, settings):
    from vizcacha.ui.app import build_services, build_workbench

    built = build_workbench(build_services(settings))
    qtbot.addWidget(built.window)
    yield built
    for editor in built.editor.editors():
        editor.document().setModified(False)
