"""Non-modal Find / Replace bar floating over the top-right corner of the editor."""

from collections.abc import Callable

from PyQt5.QtCore import QEvent, Qt
from PyQt5.QtWidgets import (
    QCheckBox,
    QFrame,
    QGridLayout,
    QLabel,
    QLineEdit,
    QPushButton,
    QWidget,
)

from vizcacha.i18n import _, ngettext
from vizcacha.ui.editor.code_editor import CodeEditor
from vizcacha.ui.editor.text_search import (
    SEARCH_LAYER,
    SearchQuery,
    find_all,
    find_next,
    match_selections,
    replace_all,
    replace_current,
)

EditorSource = Callable[[], CodeEditor | None]
MARGIN = 16


class FindReplaceBar(QFrame):
    def __init__(self, current_editor: EditorSource, parent: QWidget) -> None:
        super().__init__(parent)
        self._current_editor = current_editor
        self._highlighted: CodeEditor | None = None
        self.setFrameShape(QFrame.StyledPanel)
        self.setAutoFillBackground(True)
        self.find_field = QLineEdit()
        self.find_field.setPlaceholderText(_("Find"))
        self.replace_field = QLineEdit()
        self.replace_field.setPlaceholderText(_("Replace"))
        self.match_case = QCheckBox(_("Match case"))
        self.whole_word = QCheckBox(_("Whole word"))
        self.status = QLabel()
        self._build_layout()
        self.find_field.textChanged.connect(self._search_from_selection_start)
        self.find_field.returnPressed.connect(self.find_next)
        self.replace_field.returnPressed.connect(self.replace)
        self.match_case.toggled.connect(self.refresh_highlights)
        self.whole_word.toggled.connect(self.refresh_highlights)
        parent.installEventFilter(self)
        self.hide()

    def _build_layout(self) -> None:
        layout = QGridLayout(self)
        layout.setContentsMargins(6, 6, 6, 6)
        layout.addWidget(self.find_field, 0, 0)
        layout.addWidget(self._button(_("Previous"), self.find_previous), 0, 1)
        layout.addWidget(self._button(_("Next"), self.find_next), 0, 2)
        layout.addWidget(self._button("✕", self.close_bar), 0, 3)
        layout.addWidget(self.replace_field, 1, 0)
        self.replace_button = self._button(_("Replace"), self.replace)
        self.replace_all_button = self._button(_("Replace All"), self.replace_all)
        layout.addWidget(self.replace_button, 1, 1)
        layout.addWidget(self.replace_all_button, 1, 2)
        layout.addWidget(self.match_case, 2, 0)
        layout.addWidget(self.whole_word, 2, 1, 1, 2)
        layout.addWidget(self.status, 3, 0, 1, 4)

    def _button(self, text: str, slot) -> QPushButton:
        button = QPushButton(text)
        button.setFocusPolicy(Qt.NoFocus)
        button.clicked.connect(lambda _checked=False: slot())
        return button

    # --- public commands --------------------------------------------------
    def show_find(self) -> None:
        self._open(with_replace=False)

    def show_replace(self) -> None:
        self._open(with_replace=True)

    def query(self) -> SearchQuery:
        return SearchQuery(
            self.find_field.text(), self.match_case.isChecked(), self.whole_word.isChecked()
        )

    def find_next(self) -> bool:
        return self._find(backward=False)

    def find_previous(self) -> bool:
        return self._find(backward=True)

    def replace(self) -> None:
        editor = self._current_editor()
        if editor is not None:
            replace_current(editor, self.query(), self.replace_field.text())
            self.refresh_highlights()

    def replace_all(self) -> int:
        editor = self._current_editor()
        if editor is None:
            return 0
        count = replace_all(editor, self.query(), self.replace_field.text())
        self.refresh_highlights()
        self.status.setText(
            ngettext("{count} replacement", "{count} replacements", count).format(count=count)
        )
        return count

    def close_bar(self) -> None:
        self._clear_highlights()
        self.hide()
        editor = self._current_editor()
        if editor is not None:
            editor.setFocus()

    def refresh_highlights(self) -> None:
        self._clear_highlights()
        editor = self._current_editor()
        if editor is None or self.isHidden():
            return
        matches = find_all(editor.document(), self.query())
        editor.set_selection_layer(
            SEARCH_LAYER, match_selections(matches, editor.theme.search_match)
        )
        self._highlighted = editor
        self._show_count(len(matches))

    # --- internals ----------------------------------------------------------
    def _open(self, with_replace: bool) -> None:
        editor = self._current_editor()
        selected = editor.textCursor().selectedText() if editor else ""
        if selected and " " not in selected:
            self.find_field.setText(selected)
        for widget in (self.replace_field, self.replace_button, self.replace_all_button):
            widget.setVisible(with_replace)
        self.show()
        self.adjustSize()
        self._reposition()
        self.raise_()
        self.find_field.setFocus()
        self.find_field.selectAll()
        self.refresh_highlights()

    def _find(self, backward: bool) -> bool:
        editor = self._current_editor()
        if editor is None:
            return False
        found = find_next(editor, self.query(), backward)
        self.refresh_highlights()
        return found

    def _search_from_selection_start(self) -> None:
        editor = self._current_editor()
        if editor is not None:
            cursor = editor.textCursor()
            cursor.setPosition(cursor.selectionStart())
            editor.setTextCursor(cursor)
        self._find(backward=False)

    def _show_count(self, count: int) -> None:
        if not self.find_field.text():
            self.status.clear()
            return
        if count == 0:
            self.status.setText(_("No matches"))
            return
        self.status.setText(ngettext("{count} match", "{count} matches", count).format(count=count))

    def _clear_highlights(self) -> None:
        if self._highlighted is not None:
            self._highlighted.set_selection_layer(SEARCH_LAYER, [])
            self._highlighted = None

    def _reposition(self) -> None:
        parent = self.parentWidget()
        self.move(max(0, parent.width() - self.width() - MARGIN), MARGIN * 2)

    def keyPressEvent(self, event) -> None:  # noqa: N802 - Qt override
        if event.key() == Qt.Key_Escape:
            self.close_bar()
            return
        super().keyPressEvent(event)

    def eventFilter(self, watched, event) -> bool:  # noqa: N802 - Qt override
        if event.type() == QEvent.Resize and self.isVisible():
            self._reposition()
        return False
