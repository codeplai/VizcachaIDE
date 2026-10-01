"""Goroutines panel: every goroutine of the stopped program, the current one marked."""

from PyQt5.QtCore import Qt
from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import QListWidget, QListWidgetItem

from vizcacha.domain.debugging import Goroutine
from vizcacha.ui.features.debugger.variables_view import monospace_font

CURRENT_MARK = "▶ "
OTHER_MARK = "   "


def describe_goroutine(goroutine: Goroutine, current: int | None) -> str:
    mark = CURRENT_MARK if goroutine.goroutine_id == current else OTHER_MARK
    return mark + goroutine.name


class GoroutinesView(QListWidget):
    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setFont(monospace_font())
        self.setAlternatingRowColors(True)

    def show_goroutines(self, goroutines: tuple[Goroutine, ...], current: int | None) -> None:
        self.clear()
        for goroutine in goroutines:
            item = QListWidgetItem(describe_goroutine(goroutine, current))
            item.setData(Qt.UserRole, goroutine)
            if goroutine.goroutine_id == current:
                font = QFont(self.font())
                font.setBold(True)
                item.setFont(font)
            self.addItem(item)
