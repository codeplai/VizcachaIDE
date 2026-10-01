"""Mouse and keyboard extras: hover tooltip, Ctrl+click definition and call-tips."""

import html
from collections.abc import Callable

from PyQt5.QtCore import QEvent, QObject, Qt, QTimer
from PyQt5.QtWidgets import QToolTip

from vizcacha.domain.completion import SignatureHelp
from vizcacha.domain.diagnostics import SourceLocation
from vizcacha.ui.features.language.editor_layers import diagnostics_on_line
from vizcacha.ui.features.language.editor_session import EditorSession

CALL_TIP_KEYS = ("(", ",")
CLOSE_TIP_KEYS = (")",)
Navigate = Callable[[SourceLocation], None]


def tooltip_html(parts: list[str]) -> str:
    paragraphs = (
        f"<p style='white-space:pre-wrap'>{html.escape(part)}</p>" for part in parts if part
    )
    return "<hr>".join(paragraphs)


def call_tip_html(signature: SignatureHelp) -> str:
    label = html.escape(signature.label)
    if 0 <= signature.active_parameter < len(signature.parameters):
        active = html.escape(signature.parameters[signature.active_parameter])
        label = label.replace(active, f"<b>{active}</b>", 1)
    documentation = signature.documentation.strip().split("\n")[0]
    if not documentation:
        return f"<code>{label}</code>"
    return f"<code>{label}</code><br>{html.escape(documentation)}"


class EditorInteractions(QObject):
    def __init__(self, session: EditorSession, navigate: Navigate) -> None:
        super().__init__(session)
        self.session = session
        self.editor = session.editor
        self.viewport = session.editor.viewport()
        self._navigate = navigate
        self.editor.installEventFilter(self)
        self.viewport.installEventFilter(self)

    def eventFilter(self, watched: QObject, event: QEvent) -> bool:  # noqa: N802 - Qt API
        if watched is self.viewport:
            return self._viewport_event(event)
        if watched is self.editor and event.type() == QEvent.KeyPress:
            self._key_pressed(event.text())
        return False

    def _viewport_event(self, event: QEvent) -> bool:
        if event.type() == QEvent.ToolTip:
            self._show_hover(event)
            return True
        is_click = event.type() == QEvent.MouseButtonPress and event.button() == Qt.LeftButton
        if is_click and event.modifiers() & Qt.ControlModifier:
            return self._go_to_definition(event.pos())
        return False

    # --- hover ------------------------------------------------------------------
    def _show_hover(self, event) -> None:
        cursor = self.editor.cursorForPosition(event.pos())
        location = self.session.location(cursor)
        over_text = abs(self.editor.cursorRect(cursor).x() - event.pos().x()) <= (
            self.editor.fontMetrics().averageCharWidth()
        )
        if location is None or not over_text:
            QToolTip.hideText()
            return
        parts = [d.message for d in diagnostics_on_line(self.session.diagnostics, location.line)]
        if self.session.is_analyzed:
            self.session.flush()
            parts.append(self.session.server.hover(location) or "")
        if not any(parts):
            QToolTip.hideText()
            return
        QToolTip.showText(event.globalPos(), tooltip_html(parts), self.viewport)

    # --- Ctrl+click -------------------------------------------------------------
    def _go_to_definition(self, position) -> bool:
        location = self.session.location(self.editor.cursorForPosition(position))
        if location is None:
            return False
        self.session.flush()
        target = self.session.server.definition(location)
        if target is not None:
            self._navigate(target)
        return True

    # --- call-tips --------------------------------------------------------------
    def _key_pressed(self, text: str) -> None:
        if text in CALL_TIP_KEYS:
            QTimer.singleShot(0, self.show_call_tip)
        elif text in CLOSE_TIP_KEYS:
            QToolTip.hideText()

    def show_call_tip(self) -> None:
        location = self.session.location()
        if location is None:
            return
        self.session.flush()
        signature = self.session.server.signature_help(location)
        if signature is None:
            return
        point = self.editor.mapToGlobal(self.editor.cursorRect().bottomLeft())
        QToolTip.showText(point, call_tip_html(signature), self.editor)
