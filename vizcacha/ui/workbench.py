"""Workbench: the API that features use to plug themselves into the IDE.

FROZEN CONTRACT (phase 0). A feature is a module with
``register(workbench: Workbench) -> None`` that adds actions, panels, settings
pages and event handlers. Features never import each other; they communicate
through ``workbench.events``.
"""

from collections.abc import Callable

from PyQt5.QtCore import QObject, pyqtSignal
from PyQt5.QtWidgets import QAction, QDockWidget, QMenu, QWidget

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import N_, _
from vizcacha.ui.main_window import MainWindow
from vizcacha.ui.panel_layout import PanelLayout
from vizcacha.ui.services import Services
from vizcacha.ui.window_state import restore_window_state, save_window_state

MENU_TITLES = {
    "file": N_("&File"),
    "edit": N_("&Edit"),
    "view": N_("&View"),
    "run": N_("&Run"),
    "debug": N_("&Debug"),
    "tools": N_("&Tools"),
    "help": N_("&Help"),
}
CONSOLE_PANEL_ID = "console"


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
        self.panel_layout = PanelLayout(window)
        self._started = False
        self._layout_restored = False
        window.close_guard = self._can_close
        self.console_dock = self.add_panel(CONSOLE_PANEL_ID, _("Console"), self.console, "bottom")
        self.console_dock.setFeatures(
            QDockWidget.DockWidgetMovable | QDockWidget.DockWidgetFloatable
        )
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
        self,
        panel_id: str,
        title: str,
        widget: QWidget,
        area: str = "right",
        group: str | None = None,
    ) -> QDockWidget:
        """Dock ``widget`` in ``area`` ("left", "right" or "bottom").

        Left panels are tabs of one group, right panels are stacked from top to
        bottom and bottom panels sit side by side (see ``panel_layout``). Panels
        with the same ``group`` become tabs of each other.
        """
        dock = QDockWidget(title, self.window)
        dock.setObjectName(panel_id)
        dock.setWidget(widget)
        restored = self._started and self._layout_restored and self.window.restoreDockWidget(dock)
        self.panel_layout.place(dock, area, group, dock_now=not restored)
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
    def add_settings_page(self, factory: SettingsPageFactory, first: bool = False) -> None:
        """``factory()`` -> QWidget with ``title``, ``load(settings)`` and ``save(settings)``.

        ``first=True`` puts the page before the others (used by "General").
        """
        if first:
            self._settings_pages.insert(0, factory)
        else:
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
        self.panel_layout.remember_visibility()
        self._add_reset_layout_action()
        self._layout_restored = restore_window_state(self.window, self.services.settings)
        if not self._layout_restored:
            self.panel_layout.apply_sizes()
            self.window.first_shown.connect(self.panel_layout.apply_sizes)  # real size known
        self._started = True

    def reset_layout(self) -> None:
        """View > Reset Layout: every panel back to its default place, size and visibility."""
        self.panel_layout.reset()

    def apply_window_settings(self) -> None:
        settings = self.services.settings
        self.toolbar.setVisible(settings.get(SettingsKeys.SHOW_TOOLBAR, True))
        self.window.statusBar().setVisible(settings.get(SettingsKeys.SHOW_STATUS_BAR, True))

    def _add_reset_layout_action(self) -> None:
        action = QAction(_("Reset &Layout"), self.window)
        action.setObjectName("reset_layout")
        action.triggered.connect(lambda _checked=False: self.reset_layout())
        self.add_action("view", action, separator=True)

    def _create_menu(self, title: str) -> QMenu:
        menu = self.window.menuBar().addMenu(_(title))
        menu.menuAction().setVisible(False)
        return menu

    def _can_close(self) -> bool:
        if not all(guard() for guard in self._close_guards):
            return False
        save_window_state(self.window, self.services.settings)
        return True
