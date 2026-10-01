"""UI of the run and project features (pytest-qt, no Go needed)."""

import stat
from pathlib import Path

from PyQt5.QtWidgets import QLineEdit

from vizcacha.ui.features.project import feature as project_feature
from vizcacha.ui.features.project.feature import LAST_FOLDER_KEY
from vizcacha.ui.features.project.files_panel import FilesPanel, is_hidden_entry


def _arguments_edit(workbench) -> QLineEdit:
    return workbench.window.findChild(QLineEdit, "program_arguments")


def _write(path: Path, text: str = "") -> Path:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(text, encoding="utf-8")
    return path


def _visible_names(panel: FilesPanel, qtbot, expected_count: int) -> set[str]:
    qtbot.waitUntil(
        lambda: panel.proxy.rowCount(panel.tree.rootIndex()) == expected_count, timeout=15000
    )
    root = panel.tree.rootIndex()  # proxy indexes are not persistent: read it after loading
    return {panel.proxy.index(row, 0, root).data() for row in range(panel.proxy.rowCount(root))}


def test_hidden_entry_rules():
    assert is_hidden_entry(".git", True, False)
    assert is_hidden_entry("__debug_bin12345", False, True)
    assert is_hidden_entry("hello.exe", False, True)
    assert is_hidden_entry("hello", False, True)  # Go binary on Linux/macOS
    assert not is_hidden_entry("main.go", False, False)
    assert not is_hidden_entry("cmd", True, True)


def test_files_panel_hides_git_and_binaries(qtbot, tmp_path: Path):
    _write(tmp_path / "main.go", "package main\n")
    _write(tmp_path / "go.mod", "module x\n")
    _write(tmp_path / ".git" / "HEAD")
    _write(tmp_path / "app.exe")
    _write(tmp_path / "__debug_bin3021")
    _write(tmp_path / "pkg" / "util.go")
    binary = _write(tmp_path / "app")
    binary.chmod(binary.stat().st_mode | stat.S_IEXEC)  # hidden on POSIX only
    panel = FilesPanel()
    qtbot.addWidget(panel)

    panel.set_folder(tmp_path)

    names = _visible_names(panel, qtbot, 3 if binary.stat().st_mode & stat.S_IXUSR else 4)
    assert {"main.go", "go.mod", "pkg"} <= names
    assert not names & {".git", "app.exe", "__debug_bin3021"}


def test_open_folder_shows_panel_and_double_click_navigates(
    workbench, qtbot, tmp_path: Path, monkeypatch
):
    source = _write(tmp_path / "main.go", "package main\n\nfunc main() {}\n")
    feature_panel = _open_folder(workbench, tmp_path, monkeypatch)
    qtbot.waitUntil(
        lambda: feature_panel.proxy.rowCount(feature_panel.tree.rootIndex()) == 1, timeout=15000
    )
    root = feature_panel.tree.rootIndex()

    feature_panel.tree.doubleClicked.emit(feature_panel.proxy.index(0, 0, root))

    assert workbench.editor.current_file_path() == source
    assert workbench.services.settings.get(LAST_FOLDER_KEY, "") == str(tmp_path)
    assert "Files" in [action.text() for action in workbench.menu("view").actions()]


def _open_folder(workbench, folder: Path, monkeypatch) -> FilesPanel:
    monkeypatch.setattr(
        project_feature.QFileDialog, "getExistingDirectory", lambda *args: str(folder)
    )
    action = next(a for a in workbench.menu("file").actions() if a.text() == "Open &Folder...")
    action.trigger()
    return workbench.window.findChild(FilesPanel)


def test_last_folder_is_restored(qtbot, settings, tmp_path: Path):
    from vizcacha.ui.app import build_services, build_workbench

    settings.set(LAST_FOLDER_KEY, str(tmp_path))
    restored = build_workbench(build_services(settings))
    qtbot.addWidget(restored.window)

    assert restored.window.findChild(FilesPanel).folder == tmp_path


def test_program_arguments_are_remembered_per_file(workbench, tmp_path: Path):
    first = _write(tmp_path / "first.go", "package main\n")
    second = _write(tmp_path / "second.go", "package main\n")
    edit = _arguments_edit(workbench)
    workbench.editor.open_file(first)
    edit.setText("uno dos")
    workbench.editor.open_file(second)
    assert edit.text() == ""
    edit.setText("tres")

    workbench.editor.open_file(first)

    assert edit.text() == "uno dos"
    workbench.editor.open_file(second)
    assert edit.text() == "tres"


def test_run_untitled_does_not_ask_to_save(workbench, monkeypatch):
    calls: list[tuple] = []
    toolchain = workbench.services.toolchain
    monkeypatch.setattr(toolchain, "run_untitled", lambda *args: calls.append(args))
    workbench.editor.current_editor().setPlainText("package main\n")
    _arguments_edit(workbench).setText('a "b c"')

    workbench.menu("run").actions()[0].trigger()

    assert calls == [("package main\n", ("a", "b c"))]


def test_environment_page_shows_tool_origins(workbench):
    from vizcacha.ui.settings_dialog import SettingsDialog

    dialog = SettingsDialog(workbench)
    page = next(page for page in dialog.pages if page.title == "Environment")
    box = page.tool_origins

    box.detect_button.click()

    texts = [label.text() for label in box.labels.values()]
    assert all(texts)
    known = ("configured", "bundled", "found on PATH", "not found")
    assert all(any(word in text for word in known) for text in texts)


def test_go_modules_dialog_sends_commands(workbench, tmp_path: Path, monkeypatch):
    from vizcacha.ui.features.project.modules_dialog import GoModulesDialog

    commands: list[tuple] = []
    toolchain = workbench.services.toolchain
    monkeypatch.setattr(toolchain, "run_go_command", lambda *args: commands.append(args) or True)
    _open_folder(workbench, tmp_path, monkeypatch)
    tools = workbench.menu("tools").actions()
    next(action for action in tools if action.text() == "Go &Modules...").trigger()
    dialog = workbench.window.findChild(GoModulesDialog)
    assert dialog.module_path.text() == tmp_path.name
    assert not dialog.tidy_button.isEnabled()  # no go.mod yet

    dialog.module_path.setText("example.com/demo")
    dialog.init_button.click()
    (tmp_path / "go.mod").write_text("module example.com/demo\n", encoding="utf-8")
    toolchain.execution_finished.emit(0)
    dialog.package.setText("github.com/google/uuid@latest")
    dialog.get_button.click()

    assert commands == [
        (tmp_path, ["mod", "init", "example.com/demo"]),
        (tmp_path, ["get", "github.com/google/uuid@latest"]),
    ]
    assert "[Command finished successfully]" in workbench.console.toPlainText()
    dialog.close()
