"""Toolbar field "Program arguments", remembered per file while the IDE is open."""

from PyQt5.QtWidgets import QHBoxLayout, QLabel, QLineEdit, QWidget

from vizcacha.application.run_program import split_program_arguments
from vizcacha.i18n import _

UNTITLED_KEY_PREFIX = "untitled:"
FIELD_WIDTH = 220


class ProgramArgumentsField(QWidget):
    def __init__(self, tabbed_editor) -> None:
        super().__init__()
        self._tabs = tabbed_editor
        self._memory: dict[str, str] = {}
        self.edit = QLineEdit()
        self.edit.setObjectName("program_arguments")
        self.edit.setPlaceholderText(_("Program arguments"))
        self.edit.setToolTip(
            _(
                "Arguments passed to your program (os.Args), separated by spaces. "
                "Use quotes to keep words together."
            )
        )
        self.edit.setClearButtonEnabled(True)
        self.edit.setMinimumWidth(FIELD_WIDTH)
        layout = QHBoxLayout(self)
        layout.setContentsMargins(4, 0, 4, 0)
        layout.addWidget(QLabel(_("Arguments:")))
        layout.addWidget(self.edit)
        self._current_key = self._key()
        self.edit.textChanged.connect(self._remember)
        tabbed_editor.active_file_changed.connect(self._on_active_file_changed)

    def text(self) -> str:
        return self.edit.text()

    def arguments(self) -> tuple[str, ...]:
        """Split arguments. Raises ProgramArgumentsError on an unclosed quote."""
        return split_program_arguments(self.edit.text())

    def _remember(self, text: str) -> None:
        self._memory[self._current_key] = text

    def _on_active_file_changed(self, _path: str) -> None:
        previous = self._current_key
        self._current_key = self._key()
        if previous == self._untitled_key() and previous in self._memory:
            # An Untitled tab was just saved: keep its arguments under the new name.
            self._memory.setdefault(self._current_key, self._memory.pop(previous))
        self.edit.blockSignals(True)
        self.edit.setText(self._memory.get(self._current_key, ""))
        self.edit.blockSignals(False)

    def _key(self) -> str:
        editor = self._tabs.current_editor()
        if editor is not None and editor.file_path is not None:
            return str(editor.file_path)
        return self._untitled_key()

    def _untitled_key(self) -> str:
        return f"{UNTITLED_KEY_PREFIX}{id(self._tabs.current_editor())}"
