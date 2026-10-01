"""Gutter with line numbers, breakpoint dots and the current-line marker."""

from PyQt5.QtCore import QSize, Qt
from PyQt5.QtGui import QColor, QPainter
from PyQt5.QtWidgets import QWidget

GUTTER_BACKGROUND = QColor("#F0F0F0")
LINE_NUMBER_COLOR = QColor("#808080")
BREAKPOINT_COLOR = QColor("#FF0000")
CURRENT_LINE_COLOR = QColor("#FFFF00")
BREAKPOINT_MARGIN = 18  # room for the breakpoint dot, left of the numbers


class LineNumberArea(QWidget):
    def __init__(self, editor) -> None:
        super().__init__(editor)
        self.editor = editor

    def sizeHint(self) -> QSize:  # noqa: N802 - Qt override
        return QSize(self.preferred_width(), 0)

    def preferred_width(self) -> int:
        digits = len(str(max(1, self.editor.blockCount())))
        return BREAKPOINT_MARGIN + 8 + self.editor.fontMetrics().horizontalAdvance("9") * digits

    def mousePressEvent(self, event) -> None:  # noqa: N802 - Qt override
        if event.button() != Qt.LeftButton:
            return
        line = self._line_at(event.pos().y())
        if line is not None:
            self.editor.toggle_breakpoint_at_line(line)

    def paintEvent(self, event) -> None:  # noqa: N802 - Qt override
        painter = QPainter(self)
        painter.fillRect(event.rect(), GUTTER_BACKGROUND)
        for line, top in self._visible_lines(event.rect().top(), event.rect().bottom()):
            self._paint_line(painter, line, top)

    def _paint_line(self, painter: QPainter, line: int, top: int) -> None:
        height = self.editor.fontMetrics().height()
        if line in self.editor.breakpoints:
            painter.setPen(Qt.NoPen)
            painter.setBrush(BREAKPOINT_COLOR)
            painter.drawEllipse(3, top + 2, 12, 12)
        if line == self.editor.current_line:
            painter.fillRect(0, top, self.width(), height, CURRENT_LINE_COLOR)
        painter.setPen(LINE_NUMBER_COLOR)
        painter.drawText(0, top, self.width() - 5, height, Qt.AlignRight, str(line))

    def _visible_lines(self, rect_top: int, rect_bottom: int):
        """Yield (1-based line, top y) for every visible block inside the rect."""
        block = self.editor.firstVisibleBlock()
        top = int(
            self.editor.blockBoundingGeometry(block).translated(self.editor.contentOffset()).top()
        )
        while block.isValid() and top <= rect_bottom:
            bottom = top + int(self.editor.blockBoundingRect(block).height())
            if block.isVisible() and bottom >= rect_top:
                yield block.blockNumber() + 1, top
            block = block.next()
            top = bottom

    def _line_at(self, y: int) -> int | None:
        # rect_top = y + 1 keeps only the block whose [top, bottom) range contains y
        for line, _top in self._visible_lines(y + 1, y):
            return line
        return None
