"""AssistantPanel: list of explained problems, or a friendly message when there are none."""

from collections.abc import Sequence

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtWidgets import QLabel, QScrollArea, QVBoxLayout, QWidget

from vizcacha.application.explain_error import ExplainedDiagnostic
from vizcacha.i18n import _, ngettext
from vizcacha.ui.features.assistant.error_card import ErrorCard


class AssistantPanel(QWidget):
    navigate_requested = pyqtSignal(object)  # SourceLocation
    search_requested = pyqtSignal(object)  # Diagnostic

    def __init__(self, parent: QWidget | None = None) -> None:
        super().__init__(parent)
        self.setObjectName("assistant_panel")
        self.cards: list[ErrorCard] = []
        self.summary = QLabel()
        self.summary.setObjectName("assistant_summary")
        self.summary.setWordWrap(True)
        self._cards_layout = QVBoxLayout()
        self._cards_layout.setAlignment(Qt.AlignTop)
        container = QWidget()
        container.setLayout(self._cards_layout)
        scroll = QScrollArea()
        scroll.setWidgetResizable(True)
        scroll.setWidget(container)
        layout = QVBoxLayout(self)
        layout.addWidget(self.summary)
        layout.addWidget(scroll, 1)
        self.show_results([])

    def show_results(self, results: Sequence[ExplainedDiagnostic]) -> None:
        self._remove_cards()
        self.summary.setText(self._summary_text(len(results)))
        for diagnostic, explanation in results:
            card = ErrorCard(diagnostic, explanation)
            card.navigate_requested.connect(self.navigate_requested)
            card.search_requested.connect(self.search_requested)
            self._cards_layout.addWidget(card)
            self.cards.append(card)

    def _remove_cards(self) -> None:
        for card in self.cards:
            self._cards_layout.removeWidget(card)
            card.setParent(None)
            card.deleteLater()
        self.cards = []

    @staticmethod
    def _summary_text(count: int) -> str:
        if count == 0:
            return _("Nothing to explain. When Go reports an error, I will explain it here.")
        return ngettext(
            "Go reported {count} problem:", "Go reported {count} problems:", count
        ).format(count=count)
