"""Popup list shown by Ctrl+Space."""

from PyQt5.QtCore import QPoint, Qt, pyqtSignal
from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import QListWidget, QListWidgetItem

from vizcacha.domain.completion import CompletionItem
from vizcacha.ui.editor.completion_delegate import CompletionItemDelegate

POPUP_STYLE = """
    QListWidget { background-color: #F3F3F3; border: 1px solid #CCCCCC; outline: none; }
    QListWidget::item { padding: 2px; border: none; }
    QListWidget::item:selected { background-color: #0078D4; color: white; }
"""


class AutocompletePopup(QListWidget):
    completion_selected = pyqtSignal(object)  # CompletionItem

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setWindowFlags(Qt.Popup | Qt.FramelessWindowHint)
        self.setFocusPolicy(Qt.NoFocus)
        self.setItemDelegate(CompletionItemDelegate(self))
        self.setFont(QFont("Consolas", 10))
        self.setMaximumHeight(300)
        self.setMinimumWidth(400)
        self.setStyleSheet(POPUP_STYLE)
        self.itemClicked.connect(self._accept_item)

    def show_completions(self, completions: list[CompletionItem], position: QPoint) -> None:
        self.clear()
        if not completions:
            self.hide()
            return
        for completion in completions:
            item = QListWidgetItem(completion.label)
            item.setData(Qt.UserRole, completion)
            self.addItem(item)
        self.setCurrentRow(0)
        self.move(position)
        self.show()
        self.setFocus()

    def keyPressEvent(self, event) -> None:  # noqa: N802 - Qt override
        key = event.key()
        if key in (Qt.Key_Return, Qt.Key_Enter):
            self._accept_item(self.currentItem())
            return
        if key == Qt.Key_Escape:
            self.hide()
            return
        if key in (Qt.Key_Up, Qt.Key_Down):
            super().keyPressEvent(event)
            return
        self.hide()
        if self.parent():
            self.parent().keyPressEvent(event)

    def focusOutEvent(self, event) -> None:  # noqa: N802 - Qt override
        self.hide()
        super().focusOutEvent(event)

    def _accept_item(self, item: QListWidgetItem | None) -> None:
        if item is not None and item.data(Qt.UserRole) is not None:
            self.completion_selected.emit(item.data(Qt.UserRole))
        self.hide()
