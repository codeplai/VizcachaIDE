"""Shows where go, gofmt, dlv and gopls come from, with a "Detect again" button."""

from PyQt5.QtCore import Qt
from PyQt5.QtWidgets import QFormLayout, QGroupBox, QLabel, QPushButton, QVBoxLayout

from vizcacha.i18n import N_, _
from vizcacha.infrastructure.go_toolchain import GoEnvironment, ToolLocation, ToolOrigin

ORIGIN_TEXTS = {
    ToolOrigin.CONFIGURED: N_("configured in Options"),
    ToolOrigin.BUNDLED: N_("bundled with VizcachaIDE"),
    ToolOrigin.PATH: N_("found on PATH"),
    ToolOrigin.MISSING: N_("not found"),
}
TOOL_LABELS = {"go": "Go:", "gofmt": "gofmt:", "dlv": "Delve:", "gopls": "gopls:"}


def describe_location(location: ToolLocation) -> str:
    origin = _(ORIGIN_TEXTS[location.origin])
    if not location.path:
        return origin
    return _("{origin}: {path}").format(origin=origin, path=location.path)


class ToolOriginsBox(QGroupBox):
    def __init__(self, environment: GoEnvironment) -> None:
        super().__init__(_("Detected Tools"))
        self._environment = environment
        self.labels = {tool: QLabel() for tool in TOOL_LABELS}
        form = QFormLayout()
        for tool, label in self.labels.items():
            label.setTextInteractionFlags(Qt.TextSelectableByMouse)
            label.setWordWrap(True)
            form.addRow(TOOL_LABELS[tool], label)
        self.detect_button = QPushButton(_("Detect again"))
        self.detect_button.clicked.connect(self.detect_again)
        layout = QVBoxLayout(self)
        layout.addLayout(form)
        layout.addWidget(self.detect_button)
        self.show_locations(environment.tool_origins())

    def detect_again(self) -> None:
        self.show_locations(self._environment.detect_again())

    def show_locations(self, locations: dict[str, ToolLocation]) -> None:
        for tool, label in self.labels.items():
            label.setText(describe_location(locations[tool]))
