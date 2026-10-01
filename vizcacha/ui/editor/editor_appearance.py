"""Appearance of a CodeEditor: theme, font and zoom, tab width and gutter geometry.

A mixin for ``CodeEditor`` (it must come before ``QPlainTextEdit`` in the bases).
"""

from PyQt5.QtCore import QEvent, QRect
from PyQt5.QtGui import QColor, QFont, QPalette

from vizcacha.ui.editor.themes import EditorTheme

MIN_FONT_SIZE = 6


def default_editor_font() -> QFont:
    font = QFont("Consolas", 11)
    return font if font.exactMatch() else QFont("Courier New", 11)


class EditorAppearance:
    def apply_theme(self, theme: EditorTheme, background: str = "", text: str = "") -> None:
        """Apply ``theme``; ``background``/``text`` optionally override its two main colours."""
        self.theme = theme
        palette = self.palette()
        palette.setColor(QPalette.Base, QColor(background or theme.background))
        palette.setColor(QPalette.Text, QColor(text or theme.text))
        palette.setColor(QPalette.Highlight, QColor(theme.selection_background))
        palette.setColor(QPalette.HighlightedText, QColor(theme.selection_text))
        self.setPalette(palette)
        self.highlighter.set_theme(theme)
        self.bracket_matcher.refresh()
        if self.current_line is not None:
            self._show_debug_line(self.current_line)
        self.line_number_area.update()

    def set_base_font(self, font: QFont) -> None:
        self._base_font = QFont(font)
        self.set_zoom(self.zoom_steps)

    def set_zoom(self, steps: int) -> None:
        self.zoom_steps = steps
        font = QFont(self._base_font)
        font.setPointSize(max(MIN_FONT_SIZE, self._base_font.pointSize() + steps))
        self.setFont(font)

    def set_tab_size(self, tab_size: int) -> None:
        self.tab_size = max(1, tab_size)
        self._refresh_metrics()

    def set_line_numbers_visible(self, visible: bool) -> None:
        self.line_numbers_visible = visible
        self._update_margins()
        self.line_number_area.update()

    # --- Qt overrides -----------------------------------------------------
    def changeEvent(self, event) -> None:  # noqa: N802 - Qt override
        super().changeEvent(event)
        if event.type() == QEvent.FontChange and hasattr(self, "line_number_area"):
            self._refresh_metrics()

    def resizeEvent(self, event) -> None:  # noqa: N802 - Qt override
        super().resizeEvent(event)
        self._place_gutter()

    # --- gutter geometry ------------------------------------------------------
    def _refresh_metrics(self) -> None:
        self.setTabStopDistance(self.fontMetrics().horizontalAdvance(" ") * self.tab_size)
        self._update_margins()

    def _update_margins(self, _block_count: int = 0) -> None:
        self.setViewportMargins(self.line_number_area.preferred_width(), 0, 0, 0)
        self._place_gutter()

    def _place_gutter(self) -> None:
        rect = self.contentsRect()
        width = self.line_number_area.preferred_width()
        self.line_number_area.setGeometry(QRect(rect.left(), rect.top(), width, rect.height()))

    def _update_line_number_area(self, rect, dy: int) -> None:
        if dy:
            self.line_number_area.scroll(0, dy)
        else:
            self.line_number_area.update(0, rect.y(), self.line_number_area.width(), rect.height())
        if rect.contains(self.viewport().rect()):
            self._update_margins()
