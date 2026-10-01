"""Permanent status bar indicators: interface language and detected Go version."""

import re
from collections.abc import Mapping

from PyQt5.QtCore import QProcess, QProcessEnvironment, pyqtSignal
from PyQt5.QtWidgets import QLabel

from vizcacha.i18n import _, active_language

LANGUAGE_NAMES = {"en": "English", "es": "Español"}
INDICATOR_MARGINS = (8, 0, 8, 0)  # left, top, right, bottom: room between indicators
GO_VERSION_PATTERN = re.compile(r"\bgo(\d+(?:\.\d+)*\S*)\s")


def parse_go_version(output: str) -> str | None:
    """``go version go1.25.1 windows/amd64`` -> ``1.25.1``."""
    match = GO_VERSION_PATTERN.search(output + " ")
    return match.group(1) if match else None


class LanguageIndicator(QLabel):
    def __init__(self) -> None:
        super().__init__()
        self.setObjectName("status_language")
        self.setContentsMargins(*INDICATOR_MARGINS)
        language = active_language()
        self.setText(LANGUAGE_NAMES.get(language, language))
        self.setToolTip(_("Interface language. Change it in Tools > Options > General."))


class GoVersionIndicator(QLabel):
    """Runs ``go version`` in the background, so start-up is never blocked."""

    version_detected = pyqtSignal(object)  # str, or None when Go was not found

    def __init__(self) -> None:
        super().__init__(_("Go: detecting..."))
        self.setObjectName("status_go_version")
        self.setContentsMargins(*INDICATOR_MARGINS)
        self._process = QProcess(self)
        self._process.finished.connect(self._on_finished)
        self._process.errorOccurred.connect(self._on_error)

    def detect(self, go_executable: str, environment: Mapping[str, str]) -> None:
        if self._process.state() != QProcess.NotRunning:
            return
        process_environment = QProcessEnvironment()
        for name, value in environment.items():
            process_environment.insert(name, value)
        self._process.setProcessEnvironment(process_environment)
        self._process.start(go_executable, ["version"])

    def stop(self) -> bool:
        """Close guard: never leave ``go version`` running when the IDE closes."""
        if self._process.state() != QProcess.NotRunning:
            self._process.kill()
            self._process.waitForFinished(1000)
        return True

    def show_version(self, version: str | None) -> None:
        self.version_detected.emit(version)
        if version is None:
            self.setText(_("Go not found"))
            self.setToolTip(_("Install Go or set its path in Tools > Options > Environment."))
            return
        self.setText(_("Go {version}").format(version=version))
        self.setToolTip(_("Go version used to run and build your programs."))

    def _on_finished(self, exit_code: int, _status) -> None:
        output = bytes(self._process.readAllStandardOutput()).decode(errors="replace")
        self.show_version(parse_go_version(output) if exit_code == 0 else None)

    def _on_error(self, error) -> None:
        if error == QProcess.FailedToStart:
            self.show_version(None)
