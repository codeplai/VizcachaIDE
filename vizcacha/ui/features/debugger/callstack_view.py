"""Call stack panel. A click on a frame asks the IDE to show its source line."""

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtWidgets import QListWidget, QListWidgetItem

from vizcacha.domain.debugging import StackFrame
from vizcacha.ui.features.debugger.variables_view import monospace_font


def describe_frame(frame: StackFrame) -> str:
    if frame.location is None:
        return frame.function
    return f"{frame.function}  {frame.location.file.name}:{frame.location.line}"


class CallStackView(QListWidget):
    frame_activated = pyqtSignal(object)  # StackFrame

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setFont(monospace_font())
        self.setAlternatingRowColors(True)
        self.itemActivated.connect(self._on_activated)
        self.itemClicked.connect(self._on_activated)

    def show_frames(self, frames: tuple[StackFrame, ...] | list[StackFrame]) -> None:
        self.clear()
        for frame in frames:
            item = QListWidgetItem(describe_frame(frame))
            item.setData(Qt.UserRole, frame)
            self.addItem(item)

    def _on_activated(self, item: QListWidgetItem) -> None:
        self.frame_activated.emit(item.data(Qt.UserRole))
