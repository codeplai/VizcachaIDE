"""Options > Appearance: themes, custom colours, toolbar and status bar."""

from PyQt5.QtWidgets import (
    QCheckBox,
    QColorDialog,
    QComboBox,
    QFormLayout,
    QGroupBox,
    QHBoxLayout,
    QLabel,
    QPushButton,
    QVBoxLayout,
    QWidget,
)

from vizcacha.application.settings_keys import SettingsKeys as Keys
from vizcacha.i18n import N_, _
from vizcacha.ui.editor.themes import LIGHT
from vizcacha.ui.features.editor.editor_preferences import USE_CUSTOM_COLORS_KEY, setting

# Stored values stay in English (compatibility with 0.1); only labels are translated.
# They must match the names in vizcacha/ui/editor/themes.py.
EDITOR_THEMES = (N_("Light"), N_("Dark"), N_("Solarized Light"), N_("Solarized Dark"))
CONSOLE_THEMES = (N_("Dark"), N_("Light"))


class ColorPicker(QWidget):
    def __init__(self, default: str) -> None:
        super().__init__()
        self.color = default
        button = QPushButton(_("Choose Color"))
        button.clicked.connect(self._choose)
        self.preview = QLabel("      ")
        layout = QHBoxLayout(self)
        layout.setContentsMargins(0, 0, 0, 0)
        layout.addWidget(button)
        layout.addWidget(self.preview)
        layout.addStretch()

    def set_color(self, color: str) -> None:
        self.color = color
        self.preview.setStyleSheet(f"background-color: {color}; border: 1px solid #000;")

    def _choose(self) -> None:
        chosen = QColorDialog.getColor()
        if chosen.isValid():
            self.set_color(chosen.name())


class AppearancePage(QWidget):
    def __init__(self) -> None:
        super().__init__()
        self.title = _("Appearance")
        self.editor_theme = self._combo(EDITOR_THEMES)
        self.console_theme = self._combo(CONSOLE_THEMES)
        self.use_custom_colors = QCheckBox(_("Use custom colors instead of the theme colors"))
        self.background = ColorPicker(LIGHT.background)
        self.text_color = ColorPicker(LIGHT.text)
        self.use_custom_colors.toggled.connect(self._enable_pickers)
        self.show_toolbar = QCheckBox(_("Show toolbar"))
        self.show_status_bar = QCheckBox(_("Show status bar"))
        layout = QVBoxLayout(self)
        layout.addWidget(
            self._group(
                _("Color Theme"),
                [
                    (_("Editor theme:"), self.editor_theme),
                    (_("Console theme:"), self.console_theme),
                ],
            )
        )
        layout.addWidget(
            self._group(
                _("Custom Colors"),
                [
                    ("", self.use_custom_colors),
                    (_("Background:"), self.background),
                    (_("Text:"), self.text_color),
                ],
            )
        )
        layout.addWidget(
            self._group(_("UI Elements"), [("", self.show_toolbar), ("", self.show_status_bar)])
        )
        layout.addStretch()

    def load(self, settings) -> None:
        self._select(self.editor_theme, settings.get(Keys.EDITOR_THEME, "Light"))
        self._select(self.console_theme, settings.get(Keys.CONSOLE_THEME, "Dark"))
        self.background.set_color(settings.get(Keys.BACKGROUND_COLOR, "") or LIGHT.background)
        self.text_color.set_color(settings.get(Keys.TEXT_COLOR, "") or LIGHT.text)
        self.use_custom_colors.setChecked(setting(settings, USE_CUSTOM_COLORS_KEY))
        self._enable_pickers(self.use_custom_colors.isChecked())
        self.show_toolbar.setChecked(settings.get(Keys.SHOW_TOOLBAR, True))
        self.show_status_bar.setChecked(settings.get(Keys.SHOW_STATUS_BAR, False))

    def save(self, settings) -> None:
        settings.set(Keys.EDITOR_THEME, self.editor_theme.currentData())
        settings.set(Keys.CONSOLE_THEME, self.console_theme.currentData())
        settings.set(Keys.BACKGROUND_COLOR, self.background.color)
        settings.set(Keys.TEXT_COLOR, self.text_color.color)
        settings.set(USE_CUSTOM_COLORS_KEY, self.use_custom_colors.isChecked())
        settings.set(Keys.SHOW_TOOLBAR, self.show_toolbar.isChecked())
        settings.set(Keys.SHOW_STATUS_BAR, self.show_status_bar.isChecked())

    def _enable_pickers(self, enabled: bool) -> None:
        self.background.setEnabled(enabled)
        self.text_color.setEnabled(enabled)

    @staticmethod
    def _combo(values: tuple[str, ...]) -> QComboBox:
        combo = QComboBox()
        for value in values:
            combo.addItem(_(value), value)
        return combo

    @staticmethod
    def _select(combo: QComboBox, value: str) -> None:
        index = combo.findData(value)
        combo.setCurrentIndex(max(index, 0))

    @staticmethod
    def _group(title: str, rows) -> QGroupBox:
        form = QFormLayout()
        for label, widget in rows:
            form.addRow(label, widget)
        group = QGroupBox(title)
        group.setLayout(form)
        return group
