"""Options > General: the interface language (applied on the next start)."""

from collections.abc import Callable

from PyQt5.QtWidgets import QComboBox, QFormLayout, QLabel, QVBoxLayout, QWidget

from vizcacha.application.ports import SettingsRepository
from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import N_, _

AUTOMATIC = ""
# Language names are written in their own language, so anyone can find theirs.
LANGUAGE_CHOICES = (
    (N_("Automatic (system language)"), AUTOMATIC),
    ("English", "en"),
    ("Español", "es"),
)


class GeneralPage(QWidget):
    def __init__(self, on_language_changed: Callable[[], None] | None = None) -> None:
        super().__init__()
        self.title = _("General")
        self._on_language_changed = on_language_changed
        self._loaded_language = AUTOMATIC
        self.language = QComboBox()
        self.language.setObjectName("general_language")
        for text, code in LANGUAGE_CHOICES:
            self.language.addItem(_(text), code)
        self.restart_note = QLabel(_("Restart VizcachaIDE to apply the new language."))
        self.restart_note.setObjectName("general_restart_note")
        self.restart_note.setWordWrap(True)
        self.restart_note.setVisible(False)
        self.language.currentIndexChanged.connect(self._update_note)
        form = QFormLayout()
        form.addRow(_("Language:"), self.language)
        layout = QVBoxLayout(self)
        layout.addLayout(form)
        layout.addWidget(self.restart_note)
        layout.addStretch(1)

    def selected_language(self) -> str:
        return self.language.currentData()

    def load(self, settings: SettingsRepository) -> None:
        stored = settings.get(SettingsKeys.LANGUAGE, AUTOMATIC)
        index = self.language.findData(stored)
        self.language.setCurrentIndex(max(index, 0))
        self._loaded_language = self.selected_language()
        self._update_note()

    def save(self, settings: SettingsRepository) -> None:
        chosen = self.selected_language()
        settings.set(SettingsKeys.LANGUAGE, chosen)
        if chosen == self._loaded_language:
            return
        self._loaded_language = chosen
        if self._on_language_changed is not None:
            self._on_language_changed()

    def _update_note(self) -> None:
        self.restart_note.setVisible(self.selected_language() != self._loaded_language)
