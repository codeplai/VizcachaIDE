"""Outline panel: functions, types, variables… of the current file (documentSymbol)."""

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtWidgets import QTreeWidget, QTreeWidgetItem

from vizcacha.i18n import N_, _
from vizcacha.infrastructure.gopls_lsp import OutlineSymbol

LOCATION_ROLE = Qt.UserRole
KIND_LABELS = {
    "function": N_("function"),
    "method": N_("method"),
    "struct": N_("struct"),
    "interface": N_("interface"),
    "class": N_("type"),
    "variable": N_("variable"),
    "constant": N_("constant"),
    "field": N_("field"),
    "package": N_("package"),
}
OTHER_KIND = N_("symbol")


def kind_label(kind: str) -> str:
    return _(KIND_LABELS.get(kind, OTHER_KIND))


class OutlineView(QTreeWidget):
    symbol_activated = pyqtSignal(object)  # SourceLocation

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setColumnCount(2)
        self.setHeaderLabels([_("Name"), _("Kind")])
        self.itemDoubleClicked.connect(self._activate)

    def show_symbols(self, symbols: list[OutlineSymbol]) -> None:
        self.clear()
        for symbol in symbols:
            self.addTopLevelItem(self._item(symbol))
        self.expandAll()
        self.resizeColumnToContents(0)

    def _item(self, symbol: OutlineSymbol) -> QTreeWidgetItem:
        item = QTreeWidgetItem([symbol.name, kind_label(symbol.kind)])
        item.setData(0, LOCATION_ROLE, symbol.location)
        if symbol.detail:
            item.setToolTip(0, symbol.detail)
        for child in symbol.children:
            item.addChild(self._item(child))
        return item

    def _activate(self, item: QTreeWidgetItem, _column: int) -> None:
        location = item.data(0, LOCATION_ROLE)
        if location is not None:
            self.symbol_activated.emit(location)
