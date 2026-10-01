"""Main window: layout only. Behaviour is added by features through the Workbench.

The editor is the central widget. The console is a dock panel (added by the
Workbench), so it can sit next to the Assistant at the bottom.
"""

from collections.abc import Callable

from PyQt5.QtCore import QTimer, pyqtSignal
from PyQt5.QtGui import QIcon
from PyQt5.QtWidgets import QApplication, QMainWindow, QToolBar

from vizcacha.i18n import _
from vizcacha.ui.editor import TabbedEditor
from vizcacha.ui.resources import resource_path
from vizcacha.ui.widgets import ConsoleWidget

DEFAULT_SIZE = (1280, 800)
SCREEN_FRACTION = 0.9


class MainWindow(QMainWindow):
    first_shown = pyqtSignal()  # once, after the window got its real size on screen

    def __init__(self) -> None:
        super().__init__()
        self._was_shown = False
        self.setObjectName("main_window")
        self.setWindowTitle(_("VizcachaIDE - Beginner-friendly Go IDE"))
        self._set_default_geometry()
        icon = resource_path("logo.png")
        if icon.exists():
            self.setWindowIcon(QIcon(str(icon)))
        self.close_guard: Callable[[], bool] = lambda: True
        self.editor = TabbedEditor()
        self.console = ConsoleWidget()
        self.toolbar = QToolBar(_("Main Toolbar"))
        self.toolbar.setObjectName("main_toolbar")
        self.toolbar.setMovable(False)
        self.addToolBar(self.toolbar)
        self.setCentralWidget(self.editor)

    def _set_default_geometry(self) -> None:
        width, height = DEFAULT_SIZE
        screen = QApplication.primaryScreen()
        if screen is not None:
            available = screen.availableGeometry()
            width = min(width, int(available.width() * SCREEN_FRACTION))
            height = min(height, int(available.height() * SCREEN_FRACTION))
        self.resize(width, height)

    def showEvent(self, event) -> None:  # noqa: N802 - Qt override
        super().showEvent(event)
        if not self._was_shown:
            self._was_shown = True
            QTimer.singleShot(0, self.first_shown.emit)  # after the first layout pass

    def closeEvent(self, event) -> None:  # noqa: N802 - Qt override
        if self.close_guard():
            event.accept()
        else:
            event.ignore()
