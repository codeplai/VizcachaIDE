"""Paints one completion row: kind symbol, name, signature and first doc line."""

from PyQt5.QtCore import QSize, Qt
from PyQt5.QtGui import QColor, QFont
from PyQt5.QtWidgets import QStyle, QStyledItemDelegate

from vizcacha.domain.completion import CompletionItem, CompletionKind

ROW_HEIGHT = 45
TEXT_LEFT = 25
SELECTED_BACKGROUND = QColor("#0078D4")
HOVER_BACKGROUND = QColor("#E5F3FF")

KIND_STYLE = {
    CompletionKind.FUNCTION: ("ƒ", "#795E26"),
    CompletionKind.METHOD: ("ƒ", "#795E26"),
    CompletionKind.VARIABLE: ("◆", "#001080"),
    CompletionKind.FIELD: ("◆", "#001080"),
    CompletionKind.CONSTANT: ("◇", "#0070C1"),
    CompletionKind.KEYWORD: ("◇", "#0070C1"),
    CompletionKind.TYPE: ("⊤", "#267F99"),
    CompletionKind.PACKAGE: ("□", "#AF00DB"),
}


class CompletionItemDelegate(QStyledItemDelegate):
    def paint(self, painter, option, index) -> None:
        item: CompletionItem | None = index.data(Qt.UserRole)
        if item is None:
            super().paint(painter, option, index)
            return
        selected = bool(option.state & QStyle.State_Selected)
        painter.save()
        self._paint_background(painter, option, selected)
        self._paint_kind(painter, option, item.kind)
        self._paint_texts(painter, option, item, selected)
        painter.restore()

    def sizeHint(self, option, index) -> QSize:  # noqa: N802 - Qt override
        return QSize(option.rect.width(), ROW_HEIGHT)

    def _paint_background(self, painter, option, selected: bool) -> None:
        if selected:
            painter.fillRect(option.rect, SELECTED_BACKGROUND)
        elif option.state & QStyle.State_MouseOver:
            painter.fillRect(option.rect, HOVER_BACKGROUND)

    def _paint_kind(self, painter, option, kind: CompletionKind) -> None:
        symbol, color = KIND_STYLE.get(kind, ("●", "#000000"))
        painter.setPen(QColor(color))
        painter.setFont(QFont("Arial", 10, QFont.Bold))
        painter.drawText(option.rect.adjusted(5, 0, 0, 0), Qt.AlignVCenter, symbol)

    def _paint_texts(self, painter, option, item: CompletionItem, selected: bool) -> None:
        x = option.rect.left() + TEXT_LEFT
        y = option.rect.top()
        painter.setFont(QFont("Consolas", 10, QFont.Bold))
        painter.setPen(QColor("#FFFFFF" if selected else "#000000"))
        painter.drawText(x, y + 15, item.label)
        if item.detail:
            name_width = painter.fontMetrics().horizontalAdvance(item.label)
            painter.setFont(QFont("Consolas", 9))
            painter.setPen(QColor("#E0E0E0" if selected else "#666666"))
            painter.drawText(x + name_width + 5, y + 15, item.detail)
        if item.documentation:
            painter.setFont(QFont("Arial", 8))
            painter.setPen(QColor("#D0D0D0" if selected else "#888888"))
            first_line = item.documentation.split("\n")[0]
            width = option.rect.width() - TEXT_LEFT - 10
            elided = painter.fontMetrics().elidedText(first_line, Qt.ElideRight, width)
            painter.drawText(x, y + 32, elided)
