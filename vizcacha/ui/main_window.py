"""Main window: layout only. Behaviour is added by features through the Workbench."""

from collections.abc import Callable

from PyQt5.QtCore import Qt
from PyQt5.QtGui import QIcon
from PyQt5.QtWidgets import QMainWindow, QSplitter, QToolBar

from vizcacha.i18n import _
from vizcacha.ui.editor import TabbedEditor
from vizcacha.ui.resources import resource_path
from vizcacha.ui.widgets import ConsoleWidget


class MainWindow(QMainWindow):
    def __init__(self) -> None:
        super().__init__()
        self.setWindowTitle(_("VizcachaIDE - Beginner-friendly Go IDE"))
        self.setGeometry(100, 100, 1200, 800)
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
        splitter = QSplitter(Qt.Vertical)
        splitter.addWidget(self.editor)
        splitter.addWidget(self.console)
        splitter.setSizes([600, 200])
        self.setCentralWidget(splitter)

    def closeEvent(self, event) -> None:  # noqa: N802 - Qt override
        if self.close_guard():
            event.accept()
        else:
            event.ignore()
