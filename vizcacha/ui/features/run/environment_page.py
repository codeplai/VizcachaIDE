"""Options > Environment: Go, GOPATH, GOROOT, Delve, extra variables and detected tools."""

from PyQt5.QtWidgets import (
    QFileDialog,
    QFormLayout,
    QGroupBox,
    QHBoxLayout,
    QLabel,
    QLineEdit,
    QPlainTextEdit,
    QPushButton,
    QVBoxLayout,
    QWidget,
)

from vizcacha.application.settings_keys import SettingsKeys as Keys
from vizcacha.i18n import _
from vizcacha.infrastructure.go_toolchain import GoEnvironment
from vizcacha.ui.features.run.tool_origins_box import ToolOriginsBox


class PathField(QWidget):
    def __init__(self, placeholder: str, pick_directory: bool) -> None:
        super().__init__()
        self.edit = QLineEdit()
        self.edit.setPlaceholderText(placeholder)
        self._pick_directory = pick_directory
        browse = QPushButton(_("Browse..."))
        browse.clicked.connect(self._browse)
        layout = QHBoxLayout(self)
        layout.setContentsMargins(0, 0, 0, 0)
        layout.addWidget(self.edit)
        layout.addWidget(browse)

    def _browse(self) -> None:
        if self._pick_directory:
            chosen = QFileDialog.getExistingDirectory(self, _("Select Directory"))
        else:
            chosen, _filter = QFileDialog.getOpenFileName(self, _("Select Executable"))
        if chosen:
            self.edit.setText(chosen)


class EnvironmentPage(QWidget):
    def __init__(self, environment: GoEnvironment) -> None:
        super().__init__()
        self.title = _("Environment")
        self.fields = {
            Keys.GO_PATH: PathField(_("Path to the go executable (empty = PATH)"), False),
            Keys.GOPATH: PathField(_("GOPATH (empty = default)"), True),
            Keys.GOROOT: PathField(_("GOROOT (empty = default)"), True),
            Keys.DELVE_PATH: PathField(_("Path to dlv (empty = PATH)"), False),
            Keys.GOPLS_PATH: PathField(_("Path to gopls (empty = PATH)"), False),
        }
        labels = {
            Keys.GO_PATH: _("Go:"),
            Keys.GOPATH: "GOPATH:",
            Keys.GOROOT: "GOROOT:",
            Keys.DELVE_PATH: _("Delve:"),
            Keys.GOPLS_PATH: "gopls:",
        }
        tools_form = QFormLayout()
        for key, field in self.fields.items():
            tools_form.addRow(labels[key], field)
        self.extra_vars = QPlainTextEdit()
        self.extra_vars.setMaximumHeight(100)
        self.extra_vars.setPlaceholderText("GOOS=linux\nGOARCH=amd64")
        info = QLabel(_("Additional environment variables, one per line: NAME=VALUE"))
        info.setWordWrap(True)
        layout = QVBoxLayout(self)
        tools_group = QGroupBox(_("Go Tools"))
        tools_group.setLayout(tools_form)
        layout.addWidget(tools_group)
        layout.addWidget(info)
        layout.addWidget(self.extra_vars)
        self.tool_origins = ToolOriginsBox(environment)
        layout.addWidget(self.tool_origins)
        layout.addStretch()

    def load(self, settings) -> None:
        for key, field in self.fields.items():
            field.edit.setText(settings.get(key, ""))
        self.extra_vars.setPlainText(settings.get(Keys.EXTRA_VARS, ""))

    def save(self, settings) -> None:
        for key, field in self.fields.items():
            settings.set(key, field.edit.text().strip())
        settings.set(Keys.EXTRA_VARS, self.extra_vars.toPlainText())
        self.tool_origins.detect_again()
