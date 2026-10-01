from pathlib import Path

import pytest
from PyQt5.QtGui import QTextCursor

from vizcacha.application.errors import GoFormatError, GoToolchainNotFoundError
from vizcacha.ui.editor.text_replacement import map_cursor_offset
from vizcacha.ui.features.editor.editor_preferences import FORMAT_ON_SAVE_KEY

MESSY = "package main\nfunc main(){\nx:=1\n_ = x\n}\n"
TIDY = "package main\n\nfunc main() {\n\tx := 1\n\t_ = x\n}\n"


class FakeToolchain:
    def __init__(self, result: str = TIDY, error: Exception | None = None) -> None:
        self.result = result
        self.error = error
        self.calls: list[str] = []

    def format_source(self, text: str) -> str:
        self.calls.append(text)
        if self.error is not None:
            raise self.error
        return self.result


@pytest.fixture
def fake(workbench) -> FakeToolchain:
    toolchain = FakeToolchain()
    workbench.services.toolchain = toolchain
    return toolchain


def _editor_with(workbench, text: str):
    editor = workbench.editor.current_editor()
    editor.setPlainText(text)
    return editor


def test_save_formats_with_gofmt_and_keeps_the_cursor(workbench, fake, tmp_path: Path):
    editor = _editor_with(workbench, MESSY)
    cursor = editor.textCursor()
    cursor.setPosition(MESSY.index("x:=1") + 1)  # right after "x"
    editor.setTextCursor(cursor)
    target = tmp_path / "main.go"

    assert workbench.editor.save_to_file(editor, target)

    assert target.read_text(encoding="utf-8") == TIDY
    assert editor.toPlainText() == TIDY
    assert editor.textCursor().position() == TIDY.index("x := 1") + 1
    assert not editor.document().isModified()


def test_formatting_is_one_undo_step(workbench, fake, tmp_path: Path):
    editor = _editor_with(workbench, MESSY)
    workbench.editor.save_to_file(editor, tmp_path / "main.go")

    editor.undo()

    assert editor.toPlainText() == MESSY


def test_syntax_error_saves_unformatted_and_reports_it(
    workbench, fake, tmp_path: Path, monkeypatch
):
    fake.error = GoFormatError("<standard input>:2:13: expected '('")
    editor = _editor_with(workbench, "package main\nfunc main{\n")
    target = tmp_path / "main.go"
    # Other features (e.g. "gopls not found" on machines without gopls) may write to the
    # status bar right after us, so record what was shown instead of reading it back.
    shown: list[str] = []
    original = workbench.show_status_message
    monkeypatch.setattr(
        workbench,
        "show_status_message",
        lambda text, timeout_ms=5000: (shown.append(text), original(text, timeout_ms)),
    )

    assert workbench.editor.save_to_file(editor, target)

    assert target.read_text(encoding="utf-8") == "package main\nfunc main{\n"
    assert any("expected '('" in message for message in shown)


def test_missing_gofmt_saves_unformatted_and_warns_once(workbench, fake, tmp_path: Path):
    fake.error = GoToolchainNotFoundError("gofmt was not found: gofmt")
    editor = _editor_with(workbench, MESSY)

    workbench.editor.save_to_file(editor, tmp_path / "a.go")
    editor.insertPlainText("// more\n")
    workbench.editor.save_to_file(editor, tmp_path / "a.go")

    assert (tmp_path / "a.go").read_text(encoding="utf-8").endswith("// more\n" + MESSY)
    assert workbench.console.toPlainText().count("gofmt is not available") == 1


def test_format_on_save_can_be_disabled(workbench, fake, settings, tmp_path: Path):
    settings.set(FORMAT_ON_SAVE_KEY, False)
    editor = _editor_with(workbench, MESSY)

    workbench.editor.save_to_file(editor, tmp_path / "main.go")

    assert fake.calls == []
    assert (tmp_path / "main.go").read_text(encoding="utf-8") == MESSY


def test_only_go_files_are_formatted(workbench, fake, tmp_path: Path):
    editor = _editor_with(workbench, "notes")
    workbench.editor.save_to_file(editor, tmp_path / "notes.txt")
    assert fake.calls == []


def test_format_code_action_formats_without_saving(workbench, fake):
    editor = _editor_with(workbench, MESSY)
    edit = {action.text(): action for action in workbench.menu("edit").actions()}

    edit["F&ormat Code"].trigger()

    assert editor.toPlainText() == TIDY
    assert editor.document().isModified()
    assert edit["F&ormat Code"].shortcut().toString() == "Ctrl+Shift+F"


@pytest.mark.parametrize(
    ("old", "new", "offset", "expected"),
    [
        ("a:=1", "a := 1", 2, 3),  # after ":" -> after ":"
        ("x\n  y", "x\n\ty", 3, 3),  # inside indentation -> before "y"
        ("x  \ny", "x\ny", 3, 1),  # trailing spaces -> end of "x"
        ("", "package main\n", 0, 0),
    ],
)
def test_map_cursor_offset(old, new, offset, expected):
    assert map_cursor_offset(old, new, offset) == expected


@pytest.mark.requires_go
def test_real_gofmt_formats_on_save(workbench, tmp_path: Path):
    editor = _editor_with(workbench, MESSY)
    editor.moveCursor(QTextCursor.End)
    target = tmp_path / "main.go"

    assert workbench.editor.save_to_file(editor, target)

    assert target.read_text(encoding="utf-8") == TIDY
    assert editor.toPlainText() == TIDY
