"""Variables panel: name / type / value tree with lazy, asynchronous expansion.

Variables whose ``reference`` is non-zero get an expand arrow. Opening one shows a
temporary "Loading..." child and asks ``request_children(reference)`` (the debugger);
the answer comes back through ``show_children(reference, children)``.

What the user expanded is remembered by name path (``("p", "address")``), so the same
variables open again after every step even though Delve renumbers the references.
"""

from collections.abc import Callable

from PyQt5.QtCore import Qt
from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import QHeaderView, QTreeWidget, QTreeWidgetItem

from vizcacha.domain.debugging import Variable
from vizcacha.i18n import N_, _

ChildrenRequest = Callable[[int], None]
REFERENCE_ROLE = Qt.UserRole
LOADING_TEXT = N_("Loading...")
NamePath = tuple[str, ...]


def monospace_font(size: int = 9) -> QFont:
    font = QFont("Consolas", size)
    if not font.exactMatch():
        font = QFont("Courier New", size)
    return font


def name_path(item: QTreeWidgetItem) -> NamePath:
    names = []
    while item is not None:
        names.append(item.text(0))
        item = item.parent()
    return tuple(reversed(names))


class VariablesView(QTreeWidget):
    def __init__(self, request_children: ChildrenRequest | None = None, parent=None) -> None:
        super().__init__(parent)
        self.request_children = request_children
        self.expanded_paths: set[NamePath] = set()
        self._pending: dict[int, QTreeWidgetItem] = {}
        self.setHeaderLabels([_("Name"), _("Type"), _("Value")])
        header = self.header()
        header.setSectionResizeMode(0, QHeaderView.ResizeToContents)
        header.setSectionResizeMode(1, QHeaderView.ResizeToContents)
        header.setSectionResizeMode(2, QHeaderView.Stretch)
        self.setFont(monospace_font())
        self.setAlternatingRowColors(True)
        self.itemExpanded.connect(self._on_expanded)
        self.itemCollapsed.connect(lambda item: self.expanded_paths.discard(name_path(item)))

    def clear(self) -> None:
        self._pending.clear()  # their items are about to be deleted
        super().clear()

    def forget_expansion(self) -> None:
        """A new session starts: nothing is expanded yet."""
        self.expanded_paths.clear()

    def show_variables(self, variables: tuple[Variable, ...] | list[Variable]) -> None:
        self.clear()
        self._add_all(variables, self.invisibleRootItem())

    def show_children(self, reference: int, children: list[Variable]) -> None:
        """Answer to ``request_children``; unknown or outdated references are ignored."""
        item = self._pending.pop(reference, None)
        if item is None:
            return
        item.takeChildren()  # the "Loading..." placeholder
        self._add_all(children, item)
        if not children:
            item.setChildIndicatorPolicy(QTreeWidgetItem.DontShowIndicatorWhenChildless)

    def _add_all(self, variables, parent: QTreeWidgetItem) -> None:
        for variable in variables:
            item = self._add(variable, parent)
            if variable.children or name_path(item) in self.expanded_paths:
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

    def _on_expanded(self, item: QTreeWidgetItem) -> None:
        self.expanded_paths.add(name_path(item))
        reference = item.data(0, REFERENCE_ROLE)
        if not reference or self.request_children is None:
            return
        item.setData(0, REFERENCE_ROLE, 0)  # load once
        placeholder = QTreeWidgetItem(item, [_(LOADING_TEXT)])
        placeholder.setDisabled(True)
        self._pending[reference] = item
        self.request_children(reference)  # may answer before returning
