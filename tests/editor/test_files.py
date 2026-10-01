from pathlib import Path

import pytest
from PyQt5.QtGui import QTextCursor

from vizcacha.ui.editor import TabbedEditor
from vizcacha.ui.features.editor.editor_preferences import FORMAT_ON_SAVE_KEY
from vizcacha.ui.features.files.external_changes import ExternalChangeWatcher
from vizcacha.ui.features.files.recent_files import MAX_RECENT_FILES, RecentFiles


def _write(path: Path, text: str) -> Path:
    path.write_text(text, encoding="utf-8")
    return path


def _recent_menu(workbench):
    for action in workbench.menu("file").actions():
        if action.menu() is not None:
            return action.menu()
    raise AssertionError("Open Recent submenu not found")


# --- recent files -------------------------------------------------------------
def test_recent_files_are_unique_most_recent_first_and_limited(settings):
    recent = RecentFiles(settings)
    for number in range(MAX_RECENT_FILES + 2):
        recent.add(f"/tmp/{number}.go")
    recent.add("/tmp/5.go")

    paths = recent.paths()

    assert len(paths) == MAX_RECENT_FILES
    assert paths[0] == str(Path("/tmp/5.go"))
    assert paths.count(str(Path("/tmp/5.go"))) == 1
    recent.clear()
    assert recent.paths() == []


def test_open_recent_menu_lists_opened_files_and_can_be_cleared(workbench, tmp_path: Path):
    first = _write(tmp_path / "first.go", "package main\n")
    second = _write(tmp_path / "second.go", "package main\n")
    workbench.editor.open_file(first)
    workbench.editor.open_file(second)
    menu = _recent_menu(workbench)
    menu.rebuild()

    entries = [a.data() for a in menu.actions() if a.data()]
    assert entries == [str(second), str(first)]
    assert menu.title() == "Open &Recent"

    menu.clear_recent()
    texts = [a.text() for a in menu.actions() if a.text()]
    assert texts == ["No recent files", "&Clear Recent Files"]


def test_missing_recent_file_is_dropped(workbench, tmp_path: Path, monkeypatch):
    gone = _write(tmp_path / "gone.go", "package main\n")
    workbench.editor.open_file(gone)
    gone.unlink()
    menu = _recent_menu(workbench)
    monkeypatch.setattr(menu, "_open_path", lambda path: False)

    menu.open_recent(str(gone))

    assert str(gone) not in [a.data() for a in menu.actions()]


# --- save all -----------------------------------------------------------------
def test_save_all_saves_every_modified_tab(workbench, settings, tmp_path: Path):
    settings.set(FORMAT_ON_SAVE_KEY, False)
    paths = [_write(tmp_path / name, "package main\n") for name in ("a.go", "b.go")]
    for path in paths:
        workbench.editor.open_file(path)
        workbench.editor.current_editor().moveCursor(QTextCursor.End)
        workbench.editor.current_editor().insertPlainText("// edited\n")

    file_actions = {a.text(): a for a in workbench.menu("file").actions()}
    file_actions["Save A&ll"].trigger()

    for path in paths:
        assert path.read_text(encoding="utf-8") == "package main\n// edited\n"


# --- external changes ---------------------------------------------------------
@pytest.fixture
def tabs(qtbot):
    widget = TabbedEditor()
    qtbot.addWidget(widget)
    return widget


def test_unmodified_editor_reloads_when_the_file_changes(qtbot, tabs, tmp_path: Path):
    source = _write(tmp_path / "main.go", "package main\n")
    watcher = ExternalChangeWatcher(tabs, ask_reload=lambda path: pytest.fail("no question"))
    tabs.open_file(source)
    editor = tabs.current_editor()

    assert str(source) in [str(Path(p)) for p in watcher.watched_files()]
    _write(source, "package main\n\nfunc main() {}\n")

    qtbot.waitUntil(
        lambda: editor.toPlainText() == "package main\n\nfunc main() {}\n", timeout=5000
    )
    assert not editor.document().isModified()


def test_modified_editor_asks_before_reloading(qtbot, tabs, tmp_path: Path):
    source = _write(tmp_path / "main.go", "package main\n")
    answers: list[Path] = []
    watcher = ExternalChangeWatcher(tabs, ask_reload=lambda path: answers.append(path) or False)
    tabs.open_file(source)
    editor = tabs.current_editor()
    editor.insertPlainText("// local\n")
    _write(source, "package other\n")

    watcher.check(str(source))
    watcher.check(str(source))  # same content on disk: not asked again
    assert answers == [source]
    assert editor.toPlainText() == "// local\npackage main\n"

    watcher.ask_reload = lambda path: True
    _write(source, "package third\n")
    watcher.check(str(source))
    assert editor.toPlainText() == "package third\n"
    assert not editor.document().isModified()


def test_own_save_does_not_trigger_a_reload_question(qtbot, tabs, tmp_path: Path):
    source = _write(tmp_path / "main.go", "package main\n")
    asked: list[Path] = []
    ExternalChangeWatcher(tabs, ask_reload=lambda path: asked.append(path) or False)
    tabs.open_file(source)
    editor = tabs.current_editor()
    editor.insertPlainText("// mine\n")

    tabs.save_current_tab()
    editor.insertPlainText("// not saved yet\n")
    qtbot.wait(500)

    assert asked == []
