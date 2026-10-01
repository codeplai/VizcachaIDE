"""Variables panel: name / type / value tree."""

from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import QHeaderView, QTreeWidget, QTreeWidgetItem

from vizcacha.domain.debugging import Variable
from vizcacha.i18n import _


class VariablesView(QTreeWidget):
    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setHeaderLabels([_("Name"), _("Type"), _("Value")])
        header = self.header()
        header.setSectionResizeMode(0, QHeaderView.ResizeToContents)
        header.setSectionResizeMode(1, QHeaderView.ResizeToContents)
        header.setSectionResizeMode(2, QHeaderView.Stretch)
        font = QFont("Consolas", 9)
        if not font.exactMatch():
            font = QFont("Courier New", 9)
        self.setFont(font)
        self.setAlternatingRowColors(True)

    def show_variables(self, variables: tuple[Variable, ...] | list[Variable]) -> None:
        self.clear()
        for variable in variables:
            self._add(variable, self.invisibleRootItem())
        self.expandToDepth(0)

    def _add(self, variable: Variable, parent: QTreeWidgetItem) -> QTreeWidgetItem:
        item = QTreeWidgetItem(parent, [variable.name, variable.type_name, variable.value])
        for child in variable.children:
            self._add(child, item)
        return item
