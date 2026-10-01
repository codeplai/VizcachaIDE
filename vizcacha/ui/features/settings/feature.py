"""Tools > Options... opens the SettingsDialog built from every feature's pages."""

from PyQt5.QtWidgets import QAction

from vizcacha.i18n import _
from vizcacha.ui.settings_dialog import SettingsDialog
from vizcacha.ui.workbench import Workbench


def register(workbench: Workbench) -> None:
    action = QAction(_("&Options..."), workbench.window)
    action.setMenuRole(QAction.PreferencesRole)
    action.triggered.connect(
        lambda _checked=False: SettingsDialog(workbench, workbench.window).exec_()
    )
    workbench.add_action("tools", action)
