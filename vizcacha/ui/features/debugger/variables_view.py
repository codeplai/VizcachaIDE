"""Variables panel: name / type / value tree with lazy expansion.

Variables whose ``reference`` is non-zero get an expand arrow; their children are
asked to ``children_provider(reference)`` (the debugger) only when the user opens them.
"""

from collections.abc import Callable

from PyQt5.QtCore import Qt
from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import QHeaderView, QTreeWidget, QTreeWidgetItem

from vizcacha.domain.debugging import Variable
from vizcacha.i18n import _

ChildrenProvider = Callable[[int], list[Variable]]
REFERENCE_ROLE = Qt.UserRole


def monospace_font(size: int = 9) -> QFont:
    font = QFont("Consolas", size)
    if not font.exactMatch():
        font = QFont("Courier New", size)
    return font


class VariablesView(QTreeWidget):
    def __init__(self, children_provider: ChildrenProvider | None = None, parent=None) -> None:
        super().__init__(parent)
        self.children_provider = children_provider
        self.setHeaderLabels([_("Name"), _("Type"), _("Value")])
        header = self.header()
        header.setSectionResizeMode(0, QHeaderView.ResizeToContents)
        header.setSectionResizeMode(1, QHeaderView.ResizeToContents)
        header.setSectionResizeMode(2, QHeaderView.Stretch)
        self.setFont(monospace_font())
        self.setAlternatingRowColors(True)
        self.itemExpanded.connect(self._load_children)

    def show_variables(self, variables: tuple[Variable, ...] | list[Variable]) -> None:
        self.clear()
        for variable in variables:
            item = self._add(variable, self.invisibleRootItem())
            if variable.children:
                item.setExpanded(True)

    def _add(self, variable: Variable, parent: QTreeWidgetItem) -> QTreeWidgetItem:
        item = QTreeWidgetItem(parent, [variable.name, variable.type_name, variable.value])
        item.setToolTip(2, variable.value)
        for child in variable.children:
            self._add(child, item)
        if variable.reference and not variable.children:
            item.setData(0, REFERENCE_ROLE, variable.reference)
            item.setChildIndicatorPolicy(QTreeWidgetItem.ShowIndicator)
        return item

    def _load_children(self, item: QTreeWidgetItem) -> None:
        reference = item.data(0, REFERENCE_ROLE)
        if not reference or self.children_provider is None:
            return
        item.setData(0, REFERENCE_ROLE, 0)  # load once
        children = self.children_provider(reference)
        for child in children:
            self._add(child, item)
        if not children:
            item.setChildIndicatorPolicy(QTreeWidgetItem.DontShowIndicatorWhenChildless)
