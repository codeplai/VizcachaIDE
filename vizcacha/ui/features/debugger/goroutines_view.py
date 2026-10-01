"""Goroutines panel: every goroutine of the stopped program, the current one marked.

A click on a goroutine whose location is known asks the IDE to show that line.
"""

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import QListWidget, QListWidgetItem

from vizcacha.domain.debugging import Goroutine
from vizcacha.ui.features.debugger.variables_view import monospace_font

CURRENT_MARK = "▶ "
OTHER_MARK = "   "


def describe_goroutine(goroutine: Goroutine, current: int | None) -> str:
    mark = CURRENT_MARK if goroutine.goroutine_id == current else OTHER_MARK
    if goroutine.location is None:
        return mark + goroutine.name
    location = goroutine.location
    return f"{mark}{goroutine.name}  {location.file.name}:{location.line}"


class GoroutinesView(QListWidget):
    goroutine_activated = pyqtSignal(object)  # Goroutine

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setFont(monospace_font())
        self.setAlternatingRowColors(True)
        self.itemActivated.connect(self._on_activated)
        self.itemClicked.connect(self._on_activated)

    def show_goroutines(self, goroutines: tuple[Goroutine, ...], current: int | None) -> None:
        self.clear()
        for goroutine in goroutines:
            item = QListWidgetItem(describe_goroutine(goroutine, current))
            item.setData(Qt.UserRole, goroutine)
            if goroutine.location is not None:
                item.setToolTip(str(goroutine.location.file))
            if goroutine.goroutine_id == current:
                font = QFont(self.font())
                font.setBold(True)
                item.setFont(font)
            self.addItem(item)

    def _on_activated(self, item: QListWidgetItem) -> None:
        self.goroutine_activated.emit(item.data(Qt.UserRole))
