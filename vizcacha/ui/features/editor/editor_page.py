"""Options > Editor: font and editing behaviour."""

from PyQt5.QtGui import QFont
from PyQt5.QtWidgets import (
    QCheckBox,
    QComboBox,
    QFormLayout,
    QGroupBox,
    QLabel,
    QSpinBox,
    QVBoxLayout,
    QWidget,
)

from vizcacha.application.settings_keys import SettingsKeys as Keys
from vizcacha.i18n import _

MONOSPACE_FONTS = (
    "Consolas",
    "Courier New",
    "Monaco",
    "Menlo",
    "DejaVu Sans Mono",
    "Liberation Mono",
    "Source Code Pro",
)
PREVIEW_CODE = 'func main() { fmt.Println("Hello") }'


class EditorPage(QWidget):
    def __init__(self) -> None:
        super().__init__()
        self.title = _("Editor")
        self.font_family = QComboBox()
        self.font_family.addItems(MONOSPACE_FONTS)
        self.font_size = QSpinBox()
        self.font_size.setRange(8, 24)
        self.preview = QLabel(PREVIEW_CODE)
        self.preview.setStyleSheet("padding: 10px; background-color: #F0F0F0;")
        self.tab_size = QSpinBox()
        self.tab_size.setRange(2, 8)
        self.auto_indent = QCheckBox(_("Enable auto-indentation"))
        self.show_line_numbers = QCheckBox(_("Show line numbers"))
        self.word_wrap = QCheckBox(_("Enable word wrap"))
        self.font_family.currentTextChanged.connect(self._update_preview)
        self.font_size.valueChanged.connect(self._update_preview)
        self._build_layout()

    def _build_layout(self) -> None:
        font_form = QFormLayout()
        font_form.addRow(_("Font family:"), self.font_family)
        font_form.addRow(_("Font size:"), self.font_size)
        font_form.addRow(_("Preview:"), self.preview)
        behaviour_form = QFormLayout()
        behaviour_form.addRow(_("Tab size:"), self.tab_size)
        for checkbox in (self.auto_indent, self.show_line_numbers, self.word_wrap):
            behaviour_form.addRow("", checkbox)
        layout = QVBoxLayout(self)
        for title, form in ((_("Editor Font"), font_form), (_("Editor Behavior"), behaviour_form)):
            group = QGroupBox(title)
            group.setLayout(form)
            layout.addWidget(group)
        layout.addStretch()

    def load(self, settings) -> None:
        self.font_family.setCurrentText(settings.get(Keys.FONT_FAMILY, "Consolas"))
        self.font_size.setValue(settings.get(Keys.FONT_SIZE, 11))
        self.tab_size.setValue(settings.get(Keys.TAB_SIZE, 4))
        self.auto_indent.setChecked(settings.get(Keys.AUTO_INDENT, True))
        self.show_line_numbers.setChecked(settings.get(Keys.SHOW_LINE_NUMBERS, True))
        self.word_wrap.setChecked(settings.get(Keys.WORD_WRAP, False))
        self._update_preview()

    def save(self, settings) -> None:
        settings.set(Keys.FONT_FAMILY, self.font_family.currentText())
        settings.set(Keys.FONT_SIZE, self.font_size.value())
        settings.set(Keys.TAB_SIZE, self.tab_size.value())
        settings.set(Keys.AUTO_INDENT, self.auto_indent.isChecked())
        settings.set(Keys.SHOW_LINE_NUMBERS, self.show_line_numbers.isChecked())
        settings.set(Keys.WORD_WRAP, self.word_wrap.isChecked())

    def _update_preview(self) -> None:
        self.preview.setFont(QFont(self.font_family.currentText(), self.font_size.value()))
