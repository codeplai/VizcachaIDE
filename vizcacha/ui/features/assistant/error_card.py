"""ErrorCard: one explained problem, with "Go to line" and "Search this error" buttons."""

from PyQt5.QtCore import Qt, pyqtSignal
from PyQt5.QtGui import QFontDatabase
from PyQt5.QtWidgets import QFrame, QHBoxLayout, QLabel, QPushButton, QVBoxLayout

from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.domain.explanations import ErrorExplanation
from vizcacha.i18n import _
from vizcacha.ui.features.assistant.presentation import present

ERROR_ACCENT = "#E06C5A"
WARNING_ACCENT = "#D7A23A"


def _label(text: str, name: str, bold: bool = False) -> QLabel:
    label = QLabel(text)
    label.setObjectName(name)
    label.setWordWrap(True)
    label.setTextFormat(Qt.PlainText)
    label.setTextInteractionFlags(Qt.TextSelectableByMouse)
    if bold:
        font = label.font()
        font.setBold(True)
        label.setFont(font)
    label.setVisible(bool(text))
    return label


class ErrorCard(QFrame):
    navigate_requested = pyqtSignal(object)  # SourceLocation
    search_requested = pyqtSignal(object)  # Diagnostic

    def __init__(self, diagnostic: Diagnostic, explanation: ErrorExplanation | None) -> None:
        super().__init__()
        self.diagnostic = diagnostic
        self.explanation = explanation
        self.setObjectName("assistant_card")
        self.setFrameShape(QFrame.StyledPanel)
        accent = WARNING_ACCENT if diagnostic.severity is Severity.WARNING else ERROR_ACCENT
        self.setStyleSheet(f"#assistant_card {{ border-left: 4px solid {accent}; }}")
        view = present(diagnostic, explanation)
        layout = QVBoxLayout(self)
        layout.addWidget(_label(view.title, "assistant_title", bold=True))
        layout.addWidget(_label(view.location, "assistant_location"))
        layout.addWidget(_label(view.body, "assistant_body"))
        hint = _("Try this: {hint}").format(hint=view.fix_hint) if view.fix_hint else ""
        layout.addWidget(_label(hint, "assistant_hint"))
        layout.addWidget(_label(_("Original message from Go:"), "assistant_original_caption"))
        layout.addWidget(self._original_label(view.original))
        layout.addLayout(self._buttons())

    def _original_label(self, text: str) -> QLabel:
        label = _label(text, "assistant_original")
        label.setWordWrap(False)
        label.setFont(QFontDatabase.systemFont(QFontDatabase.FixedFont))
        return label

    def _buttons(self) -> QHBoxLayout:
        row = QHBoxLayout()
        go_to_line = QPushButton(_("Go to line"))
        go_to_line.setObjectName("assistant_goto")
        go_to_line.setEnabled(self.diagnostic.location is not None)
        go_to_line.clicked.connect(self._navigate)
        search = QPushButton(_("Search this error"))
        search.setObjectName("assistant_search")
        search.clicked.connect(lambda _checked=False: self.search_requested.emit(self.diagnostic))
        row.addWidget(go_to_line)
        row.addWidget(search)
        row.addStretch(1)
        return row

    def _navigate(self) -> None:
        if self.diagnostic.location is not None:
            self.navigate_requested.emit(self.diagnostic.location)
