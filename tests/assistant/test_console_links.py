from pathlib import Path

from PyQt5.QtCore import Qt
from PyQt5.QtGui import QColor, QTextCursor

from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.infrastructure.error_catalog import find_source_links
from vizcacha.ui.widgets.console import ERROR_COLOR, ConsoleWidget

WORKDIR = Path("/home/ana/hello")


def _format_at(console: ConsoleWidget, text: str):
    position = console.toPlainText().index(text) + 1
    cursor = console.textCursor()
    cursor.setPosition(position)
    return cursor.charFormat()


def test_find_source_links_handles_compiler_lines_frames_and_inline_text():
    text = (
        "./main.go:5:2: declared and not used: x\n"
        "\tC:/Users/Ana Perez/main.go:9 +0x17\n"
        "see util.go:3 for details\n"
    )

    links = find_source_links(text, WORKDIR)

    assert [text[link.start : link.end] for link in links] == [
        "./main.go:5:2",
        "C:/Users/Ana Perez/main.go:9",
        "util.go:3",
    ]
    assert links[0].location == SourceLocation(WORKDIR / "main.go", 5, 2)
    assert links[2].location == SourceLocation(WORKDIR / "util.go", 3, 1)


def test_locations_become_links_that_keep_the_stream_colour(qtbot):
    console = ConsoleWidget()
    qtbot.addWidget(console)
    console.set_link_base_dir(WORKDIR)

    console.append_error("./main.go:5:2: declared and not used: x\n")

    link_format = _format_at(console, "main.go:5:2")
    assert link_format.isAnchor()
    assert link_format.fontUnderline()
    assert link_format.foreground().color() == QColor(ERROR_COLOR)
    assert not _format_at(console, "declared").isAnchor()
    assert _format_at(console, "declared").foreground().color() == QColor(ERROR_COLOR)


def test_activating_a_link_emits_its_location(qtbot):
    console = ConsoleWidget()
    qtbot.addWidget(console)
    console.set_link_base_dir(WORKDIR)
    console.append_error("./main.go:5:2: declared and not used: x\n")
    href = _format_at(console, "main.go:5:2").anchorHref()

    with qtbot.waitSignal(console.location_activated) as blocker:
        console.activate_link(href)

    assert blocker.args == [SourceLocation(WORKDIR / "main.go", 5, 2)]
    console.activate_link("https://example.com")  # ignored
    console.clear()
    assert console.link_targets() == []


def test_stdin_input_still_works_after_a_link(qtbot):
    console = ConsoleWidget()
    qtbot.addWidget(console)
    console.append_error("main.go:1:1: oops\n")
    console.set_waiting_for_input(True)
    console.enable_input()
    console.moveCursor(QTextCursor.End)

    qtbot.keyClicks(console, "42")
    with qtbot.waitSignal(console.input_submitted) as blocker:
        qtbot.keyClick(console, Qt.Key_Return)

    assert blocker.args == ["42"]
    assert not _format_at(console, "42").isAnchor()
