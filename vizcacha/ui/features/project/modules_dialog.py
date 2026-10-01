"""Tools > Go Modules...: go mod init, go get <package>[@version] and go mod tidy."""

from pathlib import Path

from PyQt5.QtCore import pyqtSignal
from PyQt5.QtWidgets import (
    QDialog,
    QDialogButtonBox,
    QFormLayout,
    QGroupBox,
    QLabel,
    QLineEdit,
    QMessageBox,
    QPushButton,
    QVBoxLayout,
)

from vizcacha.application.run_program import find_go_module
from vizcacha.domain.project import GoModule
from vizcacha.i18n import _
from vizcacha.infrastructure.go_toolchain import (
    GoCommandArgumentError,
    mod_get_arguments,
    mod_init_arguments,
    mod_tidy_arguments,
)


class GoModulesDialog(QDialog):
    command_requested = pyqtSignal(object, object)  # working dir (Path), go arguments (list)

    def __init__(self, parent=None) -> None:
        super().__init__(parent)
        self.setWindowTitle(_("Go Modules"))
        self.setMinimumWidth(480)
        self.folder: Path | None = None
        self.module: GoModule | None = None
        self.status = QLabel()
        self.status.setWordWrap(True)
        self.module_path = QLineEdit()
        self.module_path.setPlaceholderText("example.com/hello")
        self.init_button = QPushButton(_("Create go.mod (go mod init)"))
        self.package = QLineEdit()
        self.package.setPlaceholderText("github.com/google/uuid@latest")
        self.get_button = QPushButton(_("Add package (go get)"))
        self.tidy_button = QPushButton(_("Clean up dependencies (go mod tidy)"))
        self.init_button.clicked.connect(self._request_init)
        self.get_button.clicked.connect(self._request_get)
        self.tidy_button.clicked.connect(self._request_tidy)
        self._build_layout()

    def _build_layout(self) -> None:
        init_form = QFormLayout()
        init_form.addRow(_("Module path:"), self.module_path)
        init_form.addRow(self.init_button)
        init_group = QGroupBox(_("New module"))
        init_group.setLayout(init_form)
        deps_form = QFormLayout()
        deps_form.addRow(_("Package[@version]:"), self.package)
        deps_form.addRow(self.get_button)
        deps_form.addRow(self.tidy_button)
        deps_group = QGroupBox(_("Dependencies"))
        deps_group.setLayout(deps_form)
        buttons = QDialogButtonBox(QDialogButtonBox.Close)
        buttons.rejected.connect(self.reject)
        layout = QVBoxLayout(self)
        layout.addWidget(self.status)
        layout.addWidget(init_group)
        layout.addWidget(deps_group)
        layout.addWidget(buttons)

    def set_folder(self, folder: Path) -> None:
        self.folder = Path(folder)
        if not self.module_path.text():
            self.module_path.setText(self.folder.name)
        self.refresh()

    def refresh(self, busy: bool = False) -> None:
        self.module = find_go_module(self.folder) if self.folder else None
        if self.module is None:
            self.status.setText(
                _("Folder: {folder}\nNo go.mod found. Create one to start a module.").format(
                    folder=self.folder
                )
            )
        else:
            self.status.setText(
                _("Module: {module}\ngo.mod in: {root}").format(
                    module=self.module.module_path, root=self.module.root
                )
            )
        has_module = self.module is not None
        self.init_button.setEnabled(not busy and not has_module and self.folder is not None)
        self.get_button.setEnabled(not busy and has_module)
        self.tidy_button.setEnabled(not busy and has_module)

    def _request_init(self) -> None:
        self._request(self.folder, mod_init_arguments, self.module_path.text())

    def _request_get(self) -> None:
        self._request(self.module.root, mod_get_arguments, self.package.text())

    def _request_tidy(self) -> None:
        self.command_requested.emit(self.module.root, mod_tidy_arguments())

    def _request(self, working_dir: Path, build_arguments, text: str) -> None:
        try:
            arguments = build_arguments(text)
        except GoCommandArgumentError as error:
            QMessageBox.warning(self, _("Go Modules"), str(error))
            return
        self.command_requested.emit(working_dir, arguments)
