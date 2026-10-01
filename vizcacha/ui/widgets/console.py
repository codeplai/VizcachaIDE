"""Console: program output (stdout/stderr/success colours), stdin input and clickable
``file.go:LINE[:COL]`` links (``location_activated``)."""

from pathlib import Path

from PyQt5.QtCore import Qt, QTimer, pyqtSignal
from PyQt5.QtGui import QColor, QFont, QPalette, QTextCharFormat, QTextCursor
from PyQt5.QtWidgets import QTextEdit

from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.i18n import _
from vizcacha.infrastructure.error_catalog import find_source_links

OUTPUT_COLOR = "#D4D4D4"
ERROR_COLOR = "#F48771"
SUCCESS_COLOR = "#4EC9B0"
INPUT_ENABLE_DELAY_MS = 100
LINK_PREFIX = "location:"


class ConsoleWidget(QTextEdit):
    input_submitted = pyqtSignal(str)
    location_activated = pyqtSignal(object)  # SourceLocation

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.input_enabled = False
        self.input_start_pos = 0
        self.waiting_for_input = False
        self._link_base_dir = Path.cwd()
        self._colors = {"output": OUTPUT_COLOR, "error": ERROR_COLOR, "success": SUCCESS_COLOR}
        self._link_targets: list[SourceLocation] = []
        font = QFont("Consolas", 10)
        if not font.exactMatch():
            font = QFont("Courier New", 10)
        self.setFont(font)
        palette = self.palette()
        palette.setColor(QPalette.Base, QColor("#1E1E1E"))
        palette.setColor(QPalette.Text, QColor(OUTPUT_COLOR))
        self.setPalette(palette)
        self.setReadOnly(True)
        self.viewport().setMouseTracking(True)
        self.append_output(_("VizcachaIDE Console - Ready") + "\n")

    def append_output(self, text: str) -> None:
        self._append(text, self._colors["output"])
        if self.waiting_for_input and not self.input_enabled:
            QTimer.singleShot(INPUT_ENABLE_DELAY_MS, self.enable_input)

    def append_error(self, text: str) -> None:
        self._append(text, self._colors["error"])

    def append_success(self, text: str) -> None:
        self._append(text, self._colors["success"])

    def clear(self) -> None:
        super().clear()
        self.waiting_for_input = False
        self._link_targets = []

    def set_colors(self, background: str, output: str, error: str, success: str) -> None:
        """Theme colours for new text (existing text keeps its colours)."""
        palette = self.palette()
        palette.setColor(QPalette.Base, QColor(background))
        palette.setColor(QPalette.Text, QColor(output))
        self.setPalette(palette)
        self._colors = {"output": output, "error": error, "success": success}

    def set_link_base_dir(self, directory: Path) -> None:
        """Directory that relative ``file.go:LINE`` links are resolved against."""
        self._link_base_dir = directory

    def link_targets(self) -> list[SourceLocation]:
        return list(self._link_targets)

    def set_waiting_for_input(self, waiting: bool) -> None:
        self.waiting_for_input = waiting

    def enable_input(self) -> None:
        if self.input_enabled:
            return
        self.input_enabled = True
        self.setReadOnly(False)
        self.input_start_pos = self.textCursor().position()
        self.setFocus()

    def disable_input(self) -> None:
        self.input_enabled = False
        self.setReadOnly(True)

    def keyPressEvent(self, event) -> None:  # noqa: N802 - Qt override
        if not self.input_enabled:
            super().keyPressEvent(event)
            return
        if event.key() in (Qt.Key_Return, Qt.Key_Enter):
            self._submit_input()
            return
        cursor = self.textCursor()
        if event.key() == Qt.Key_Backspace and cursor.position() <= self.input_start_pos:
            return
        if cursor.position() < self.input_start_pos:
            cursor.setPosition(self.input_start_pos)
            self.setTextCursor(cursor)
        super().keyPressEvent(event)

    def mouseMoveEvent(self, event) -> None:  # noqa: N802 - Qt override
        over_link = self.anchorAt(event.pos()).startswith(LINK_PREFIX)
        self.viewport().setCursor(Qt.PointingHandCursor if over_link else Qt.IBeamCursor)
        super().mouseMoveEvent(event)

    def mouseReleaseEvent(self, event) -> None:  # noqa: N802 - Qt override
        super().mouseReleaseEvent(event)
        if event.button() != Qt.LeftButton or self.textCursor().hasSelection():
            return
        self.activate_link(self.anchorAt(event.pos()))

    def activate_link(self, href: str) -> None:
        """Emit ``location_activated`` for a ``location:<n>`` anchor (no-op otherwise)."""
        if not href.startswith(LINK_PREFIX):
            return
        index = int(href[len(LINK_PREFIX) :])
        if 0 <= index < len(self._link_targets):
            self.location_activated.emit(self._link_targets[index])

    def _submit_input(self) -> None:
        cursor = self.textCursor()
        cursor.movePosition(QTextCursor.End)
        self.setTextCursor(cursor)
        cursor.setPosition(self.input_start_pos)
        cursor.movePosition(QTextCursor.End, QTextCursor.KeepAnchor)
        text = cursor.selectedText()
        self.append("")
        self.disable_input()
        self.input_submitted.emit(text)

    def _append(self, text: str, color: str) -> None:
        cursor = self.textCursor()
        cursor.movePosition(QTextCursor.End)
        position = 0
        for link in find_source_links(text, self._link_base_dir):
            cursor.insertText(text[position : link.start], self._plain_format(color))
            cursor.insertText(text[link.start : link.end], self._link_format(color, link.location))
            position = link.end
        plain = self._plain_format(color)
        cursor.insertText(text[position:], plain)
        cursor.setCharFormat(plain)  # typed stdin must not continue a link
        self.setTextCursor(cursor)
        self.ensureCursorVisible()

    @staticmethod
    def _plain_format(color: str) -> QTextCharFormat:
        text_format = QTextCharFormat()
        text_format.setForeground(QColor(color))
        return text_format

    def _link_format(self, color: str, location: SourceLocation) -> QTextCharFormat:
        self._link_targets.append(location)
        text_format = self._plain_format(color)
        text_format.setAnchor(True)
        text_format.setAnchorHref(f"{LINK_PREFIX}{len(self._link_targets) - 1}")
        text_format.setFontUnderline(True)
        text_format.setToolTip(_("Click to go to this line"))
        return text_format
