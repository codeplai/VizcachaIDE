"""Outline panel: functions, types, variables… of the current file (documentSymbol)."""

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtWidgets import QTreeWidget, QTreeWidgetItem

from vizcacha.domain.code_structure import DocumentSymbol, SymbolKind
from vizcacha.i18n import N_, _

LOCATION_ROLE = Qt.UserRole
KIND_LABELS = {
    SymbolKind.FUNCTION: N_("function"),
    SymbolKind.METHOD: N_("method"),
    SymbolKind.STRUCT: N_("struct"),
    SymbolKind.INTERFACE: N_("interface"),
    SymbolKind.TYPE: N_("type"),
    SymbolKind.VARIABLE: N_("variable"),
    SymbolKind.CONSTANT: N_("constant"),
    SymbolKind.FIELD: N_("field"),
    SymbolKind.PACKAGE: N_("package"),
}
OTHER_KIND = N_("symbol")


def kind_label(kind: SymbolKind) -> str:
    return _(KIND_LABELS.get(kind, OTHER_KIND))


class OutlineView(QTreeWidget):
    symbol_activated = pyqtSignal(object)  # SourceLocation

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setColumnCount(2)
        self.setHeaderLabels([_("Name"), _("Kind")])
        self.itemDoubleClicked.connect(self._activate)

    def show_symbols(self, symbols: list[DocumentSymbol]) -> None:
        self.clear()
        for symbol in symbols:
            self.addTopLevelItem(self._item(symbol))
        self.expandAll()
        self.resizeColumnToContents(0)

    def _item(self, symbol: DocumentSymbol) -> QTreeWidgetItem:
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
