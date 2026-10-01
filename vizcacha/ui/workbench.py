"""Workbench: the API that features use to plug themselves into the IDE.

FROZEN CONTRACT (phase 0). A feature is a module with
``register(workbench: Workbench) -> None`` that adds actions, panels, settings
pages and event handlers. Features never import each other; they communicate
through ``workbench.events``.
"""

from collections.abc import Callable

from PyQt5.QtCore import QObject, Qt, pyqtSignal
from PyQt5.QtWidgets import QAction, QDockWidget, QMenu, QWidget

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import N_, _
from vizcacha.ui.main_window import MainWindow
from vizcacha.ui.services import Services

MENU_TITLES = {
    "file": N_("&File"),
    "edit": N_("&Edit"),
    "view": N_("&View"),
    "run": N_("&Run"),
    "debug": N_("&Debug"),
    "tools": N_("&Tools"),
    "help": N_("&Help"),
}
PANEL_AREAS = {
    "right": Qt.RightDockWidgetArea,
    "bottom": Qt.BottomDockWidgetArea,
    "left": Qt.LeftDockWidgetArea,
}


class WorkbenchEvents(QObject):
    """Application-wide events. Arguments are plain types or domain objects."""

    navigate_to = pyqtSignal(object)  # SourceLocation: open file and move cursor
    process_output = pyqtSignal(str, str)  # text, stream ("stdout" | "stderr")
    program_started = pyqtSignal(object)  # RunConfiguration
    program_finished = pyqtSignal(int)  # exit code
    diagnostics_changed = pyqtSignal(object, object)  # Path, list[Diagnostic]
    settings_changed = pyqtSignal()


SettingsPageFactory = Callable[[], QWidget]


class Workbench:
    def __init__(self, window: MainWindow, services: Services) -> None:
        self.window = window
        self.services = services
        self.events = WorkbenchEvents()
        self.editor = window.editor
        self.console = window.console
        self.toolbar = window.toolbar
        self._settings_pages: list[SettingsPageFactory] = []
        self._close_guards: list[Callable[[], bool]] = []
        self._menus = {key: self._create_menu(title) for key, title in MENU_TITLES.items()}
        window.close_guard = self._can_close
        self.events.settings_changed.connect(self.apply_window_settings)

    # --- menus, toolbar, panels ------------------------------------------
    def menu(self, menu_id: str) -> QMenu:
        return self._menus[menu_id]

    def add_action(
        self, menu_id: str, action: QAction, toolbar: bool = False, separator: bool = False
    ) -> QAction:
        menu = self._menus[menu_id]
        if separator and not menu.isEmpty():
            menu.addSeparator()
        menu.addAction(action)
        menu.menuAction().setVisible(True)
        if toolbar:
            self.toolbar.addAction(action)
        return action

    def add_toolbar_separator(self) -> None:
        self.toolbar.addSeparator()

    def add_panel(
        self, panel_id: str, title: str, widget: QWidget, area: str = "right"
    ) -> QDockWidget:
        dock = QDockWidget(title, self.window)
        dock.setObjectName(panel_id)
        dock.setWidget(widget)
        self.window.addDockWidget(PANEL_AREAS[area], dock)
        self.add_action("view", dock.toggleViewAction())
        return dock

    def add_status_widget(self, widget: QWidget, permanent: bool = False) -> None:
        bar = self.window.statusBar()
        if permanent:
            bar.addPermanentWidget(widget)
        else:
            bar.addWidget(widget)

    def show_status_message(self, text: str, timeout_ms: int = 5000) -> None:
        self.window.statusBar().showMessage(text, timeout_ms)

    # --- settings and lifecycle -------------------------------------------
    def add_settings_page(self, factory: SettingsPageFactory) -> None:
        """``factory()`` -> QWidget with ``title``, ``load(settings)`` and ``save(settings)``."""
        self._settings_pages.append(factory)

    def settings_page_factories(self) -> list[SettingsPageFactory]:
        return list(self._settings_pages)

    def add_close_guard(self, guard: Callable[[], bool]) -> None:
        """``guard()`` returns False to cancel closing the window."""
        self._close_guards.append(guard)

    def start(self) -> None:
        """Called once after every feature is registered."""
        if self.editor.count() == 0:
            self.editor.new_tab()
        self.apply_window_settings()

    def apply_window_settings(self) -> None:
        settings = self.services.settings
        self.toolbar.setVisible(settings.get(SettingsKeys.SHOW_TOOLBAR, True))
        self.window.statusBar().setVisible(settings.get(SettingsKeys.SHOW_STATUS_BAR, False))

    def _create_menu(self, title: str) -> QMenu:
        menu = self.window.menuBar().addMenu(_(title))
        menu.menuAction().setVisible(False)
        return menu

    def _can_close(self) -> bool:
        return all(guard() for guard in self._close_guards)
