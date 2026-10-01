"""Tools > Options... (with the General page) and the language / Go status indicators."""

from PyQt5.QtWidgets import QAction

from vizcacha.i18n import _
from vizcacha.ui.features.settings.general_page import GeneralPage
from vizcacha.ui.features.settings.status_indicators import GoVersionIndicator, LanguageIndicator
from vizcacha.ui.settings_dialog import SettingsDialog
from vizcacha.ui.workbench import Workbench

RESTART_NOTICE_MS = 10000


def register(workbench: Workbench) -> None:
    action = QAction(_("&Options..."), workbench.window)
    action.setMenuRole(QAction.PreferencesRole)
    action.triggered.connect(
        lambda _checked=False: SettingsDialog(workbench, workbench.window).exec_()
    )
    workbench.add_action("tools", action)
    workbench.add_settings_page(lambda: GeneralPage(lambda: _notify_restart(workbench)), first=True)
    _add_status_indicators(workbench)


def _notify_restart(workbench: Workbench) -> None:
    workbench.show_status_message(
        _("Restart VizcachaIDE to apply the new language."), RESTART_NOTICE_MS
    )


def _add_status_indicators(workbench: Workbench) -> None:
    go_version = GoVersionIndicator()
    workbench.add_status_widget(go_version, permanent=True)
    workbench.add_status_widget(LanguageIndicator(), permanent=True)
    environment = workbench.services.environment

    def detect() -> None:
        go_version.detect(environment.go_executable(), workbench.services.toolchain.environment())

    workbench.events.settings_changed.connect(detect)
    workbench.add_close_guard(go_version.stop)
    detect()
