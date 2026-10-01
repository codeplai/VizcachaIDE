"""Help > About."""

from PyQt5.QtCore import Qt
from PyQt5.QtGui import QPixmap
from PyQt5.QtWidgets import QAction, QMessageBox

from vizcacha import __version__
from vizcacha.i18n import _
from vizcacha.ui.resources import resource_path
from vizcacha.ui.workbench import Workbench


def show_about(workbench: Workbench) -> None:
    box = QMessageBox(workbench.window)
    box.setWindowTitle(_("About VizcachaIDE"))
    logo = resource_path("logo.png")
    if logo.exists():
        box.setIconPixmap(
            QPixmap(str(logo)).scaled(128, 128, Qt.KeepAspectRatio, Qt.SmoothTransformation)
        )
    box.setText(
        "<h3>VizcachaIDE</h3><p>"
        + _("Beginner-friendly Go IDE")
        + "</p>"
        + "<p>"
        + _("Version {version}").format(version=__version__)
        + "</p>"
    )
    box.setInformativeText(
        "<p><b>" + _("Author:") + "</b> Marks Calderon<br>"
        "<b>" + _("Contact:") + "</b> hola@codeplai.pe<br>"
        "CEO of Codeplai Games<br>Peru</p>"
        "<p><b>" + _("License:") + "</b> MIT License<br>"
        "Copyright (c) 2025 Marks Calderon - Codeplai Games</p>"
        "<p><small>"
        + _(
            "The installable package bundles PyQt5 (GPLv3), Go, Delve and gopls, so the "
            "distributed program is licensed under GPLv3. See NOTICE.md."
        )
        + "</small></p>"
    )
    box.exec_()


def register(workbench: Workbench) -> None:
    action = QAction(_("&About"), workbench.window)
    action.setMenuRole(QAction.AboutRole)
    action.triggered.connect(lambda _checked=False: show_about(workbench))
    workbench.add_action("help", action)
