"""Dialogs used when saving: target path and the unsaved-changes question."""

from pathlib import Path

from PyQt5.QtWidgets import QFileDialog, QMessageBox, QWidget

from vizcacha.i18n import _

GO_SUFFIX = ".go"


def ask_save_path(parent: QWidget) -> Path | None:
    filename, _filter = QFileDialog.getSaveFileName(
        parent, _("Save Go File"), "", _("Go Files (*.go);;All Files (*)")
    )
    if not filename:
        return None
    path = Path(filename)
    return path if path.suffix == GO_SUFFIX else path.with_name(path.name + GO_SUFFIX)


def ask_save_changes(parent: QWidget, name: str) -> int:
    """Returns QMessageBox.Save, Discard or Cancel."""
    return QMessageBox.question(
        parent,
        _("Unsaved Changes"),
        _("Do you want to save changes to '{name}'?").format(name=name),
        QMessageBox.Save | QMessageBox.Discard | QMessageBox.Cancel,
    )


def show_file_error(parent: QWidget, message: str) -> None:
    QMessageBox.critical(parent, _("Error"), message)
