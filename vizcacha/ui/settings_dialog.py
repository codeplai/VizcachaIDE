"""Options dialog. Each tab is a page contributed by a feature.

A page is a QWidget with:
    title: str                      (already translated)
    load(settings) -> None          fill widgets from the SettingsRepository
    save(settings) -> None          write widgets back to the SettingsRepository
"""

from PyQt5.QtWidgets import QDialog, QDialogButtonBox, QTabWidget, QVBoxLayout

from vizcacha.i18n import _


class SettingsDialog(QDialog):
    def __init__(self, workbench, parent=None) -> None:
        super().__init__(parent)
        self._workbench = workbench
        self._settings = workbench.services.settings
        self.setWindowTitle(_("VizcachaIDE - Options"))
        self.setMinimumSize(600, 500)
        self.pages = [factory() for factory in workbench.settings_page_factories()]
        tabs = QTabWidget()
        for page in self.pages:
            page.load(self._settings)
            tabs.addTab(page, page.title)
        buttons = QDialogButtonBox(
            QDialogButtonBox.Ok | QDialogButtonBox.Cancel | QDialogButtonBox.Apply
        )
        buttons.accepted.connect(self.accept)
        buttons.rejected.connect(self.reject)
        buttons.button(QDialogButtonBox.Apply).clicked.connect(self.apply)
        layout = QVBoxLayout(self)
        layout.addWidget(tabs)
        layout.addWidget(buttons)

    def apply(self) -> None:
        for page in self.pages:
            page.save(self._settings)
        self._workbench.events.settings_changed.emit()

    def accept(self) -> None:
        self.apply()
        super().accept()
