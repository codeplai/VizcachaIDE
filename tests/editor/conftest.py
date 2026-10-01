"""Editor tests modify documents, so the window is closed only after they are marked clean.

(The shared ``workbench`` fixture lets pytest-qt close the window first, which would
open the modal "Unsaved Changes" question.)
"""

import pytest


@pytest.fixture
def workbench(qtbot, settings):
    from vizcacha.ui.app import build_services, build_workbench

    built = build_workbench(build_services(settings))
    yield built
    for editor in built.editor.editors():
        editor.document().setModified(False)
    built.window.close()
    built.window.deleteLater()
