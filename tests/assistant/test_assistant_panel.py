from pathlib import Path

from PyQt5.QtWidgets import QDockWidget, QLabel, QPushButton

from vizcacha.domain.diagnostics import Diagnostic, Severity, SourceLocation
from vizcacha.domain.project import RunConfiguration
from vizcacha.ui.features.assistant import feature as assistant_feature

UNUSED_VAR_OUTPUT = "# command-line-arguments\n.\\main.go:5:2: declared and not used: count\n"


def _dock(workbench) -> QDockWidget:
    return workbench.window.findChild(QDockWidget, "assistant")


def _texts(workbench, name: str) -> list[str]:
    panel = _dock(workbench).widget()
    return [label.text() for label in panel.findChildren(QLabel, name)]


def _run_failing_program(workbench, tmp_path: Path, stderr: str) -> None:
    events = workbench.events
    events.program_started.emit(RunConfiguration.for_file(tmp_path / "main.go"))
    events.process_output.emit("Running...\n", "stdout")
    events.process_output.emit(stderr, "stderr")
    events.program_finished.emit(1)


def test_panel_is_visible_from_start_up(workbench):
    assert _dock(workbench).isVisibleTo(workbench.window)
    assert _texts(workbench, "assistant_summary")[0].startswith("Nothing to explain")
    assert "Assistant" in [action.text() for action in workbench.menu("view").actions()]


def test_run_error_opens_panel_with_explanation_and_original_text(workbench, tmp_path: Path):
    _run_failing_program(workbench, tmp_path, UNUSED_VAR_OUTPUT)

    assert not _dock(workbench).isHidden()
    assert _texts(workbench, "assistant_title") == ['Variable "count" is never used']
    assert "count" in _texts(workbench, "assistant_body")[0]
    assert _texts(workbench, "assistant_hint")[0].startswith("Try this: ")
    assert _texts(workbench, "assistant_original") == [
        ".\\main.go:5:2: declared and not used: count"
    ]
    assert _texts(workbench, "assistant_location") == ["main.go, line 5"]


def test_go_to_line_emits_navigate_to(workbench, qtbot, tmp_path: Path):
    (tmp_path / "main.go").write_text(
        "package main\n\nfunc main() {\n\n\tcount := 10\n}\n", encoding="utf-8"
    )
    _run_failing_program(workbench, tmp_path, UNUSED_VAR_OUTPUT)
    button = _dock(workbench).widget().findChild(QPushButton, "assistant_goto")

    with qtbot.waitSignal(workbench.events.navigate_to) as blocker:
        button.click()

    assert blocker.args == [SourceLocation(tmp_path / "main.go", 5, 2)]
    cursor = workbench.editor.current_editor().textCursor()
    assert (cursor.blockNumber(), cursor.positionInBlock()) == (4, 1)


def test_unknown_error_shows_original_text_and_search(workbench, monkeypatch, tmp_path: Path):
    opened = []
    monkeypatch.setattr(
        assistant_feature.QDesktopServices, "openUrl", lambda url: opened.append(url.toString())
    )
    _run_failing_program(workbench, tmp_path, "./main.go:3:1: some brand new compiler error\n")

    assert _texts(workbench, "assistant_title") == ["Go reported a problem"]
    assert "some brand new compiler error" in _texts(workbench, "assistant_original")[0]
    _dock(workbench).widget().findChild(QPushButton, "assistant_search").click()

    assert len(opened) == 1
    assert opened[0].startswith("https://www.google.com/search?q=golang")
    assert "brand+new+compiler+error" in opened[0]


def test_panic_without_stderr_noise_from_stdout(workbench, tmp_path: Path):
    panic = (
        "panic: runtime error: index out of range [5] with length 3\n\n"
        "goroutine 1 [running]:\nmain.main()\n"
        f"\t{tmp_path.as_posix()}/main.go:9 +0x17\nexit status 2\n"
    )
    _run_failing_program(workbench, tmp_path, panic)

    assert _texts(workbench, "assistant_title") == ["Index 5 is out of range"]
    assert _texts(workbench, "assistant_location") == ["main.go, line 9"]


def test_successful_run_shows_nothing_to_explain(workbench, tmp_path: Path):
    workbench.events.program_started.emit(RunConfiguration.for_file(tmp_path / "main.go"))
    workbench.events.program_finished.emit(0)

    assert _texts(workbench, "assistant_title") == []
    assert _texts(workbench, "assistant_summary")[0].startswith("Nothing to explain")


def test_live_diagnostics_are_explained(workbench, tmp_path: Path):
    source = tmp_path / "main.go"
    source.write_text("package main\n\nfunc main() {\n\ttotal++\n}\n", encoding="utf-8")
    workbench.editor.open_file(source)
    diagnostic = Diagnostic(
        SourceLocation(source, 4, 2), Severity.ERROR, "undefined: total", "undefined: total"
    )

    workbench.events.diagnostics_changed.emit(tmp_path / "other.go", [diagnostic])
    assert _texts(workbench, "assistant_title") == []  # background files do not take over
    workbench.events.diagnostics_changed.emit(source, [diagnostic])

    assert not _dock(workbench).isHidden()
    assert _texts(workbench, "assistant_title") == ["Unknown name: total"]
    workbench.events.diagnostics_changed.emit(source, [])
    assert _texts(workbench, "assistant_title") == []
