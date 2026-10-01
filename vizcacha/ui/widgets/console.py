"""Console: program output (stdout/stderr/success colours) and stdin input."""

from PyQt5.QtCore import Qt, QTimer, pyqtSignal
from PyQt5.QtGui import QColor, QFont, QPalette, QTextCursor
from PyQt5.QtWidgets import QTextEdit

from vizcacha.i18n import _

OUTPUT_COLOR = "#D4D4D4"
ERROR_COLOR = "#F48771"
SUCCESS_COLOR = "#4EC9B0"
INPUT_ENABLE_DELAY_MS = 100


class ConsoleWidget(QTextEdit):
    input_submitted = pyqtSignal(str)

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.input_enabled = False
        self.input_start_pos = 0
        self.waiting_for_input = False
        font = QFont("Consolas", 10)
        if not font.exactMatch():
            font = QFont("Courier New", 10)
        self.setFont(font)
        palette = self.palette()
        palette.setColor(QPalette.Base, QColor("#1E1E1E"))
        palette.setColor(QPalette.Text, QColor(OUTPUT_COLOR))
        self.setPalette(palette)
        self.setReadOnly(True)
        self.append_output(_("VizcachaIDE Console - Ready") + "\n")

    def append_output(self, text: str) -> None:
        self._append(text, OUTPUT_COLOR)
        if self.waiting_for_input and not self.input_enabled:
            QTimer.singleShot(INPUT_ENABLE_DELAY_MS, self.enable_input)

    def append_error(self, text: str) -> None:
        self._append(text, ERROR_COLOR)

    def append_success(self, text: str) -> None:
        self._append(text, SUCCESS_COLOR)

    def clear(self) -> None:
        super().clear()
        self.waiting_for_input = False

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
        text_format = cursor.charFormat()
        text_format.setForeground(QColor(color))
        cursor.setCharFormat(text_format)
        cursor.insertText(text)
        self.setTextCursor(cursor)
        self.ensureCursorVisible()
