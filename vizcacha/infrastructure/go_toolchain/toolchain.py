"""GoToolchainPort implementation: ``go run`` / ``go build`` through QProcess."""

import os

from PyQt5.QtCore import QObject, QProcess, QProcessEnvironment, pyqtSignal

from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.infrastructure.go_toolchain.environment import GoEnvironment
from vizcacha.infrastructure.go_toolchain.formatter import format_go_source

SEPARATOR = "-" * 50 + "\n"


def _process_error_message(error: QProcess.ProcessError) -> str:
    messages = {
        QProcess.FailedToStart: _(
            "Failed to start Go. Make sure Go is installed and in your PATH."
        ),
        QProcess.Crashed: _("The process crashed."),
        QProcess.Timedout: _("The process timed out."),
        QProcess.WriteError: _("Could not write to the process."),
        QProcess.ReadError: _("Could not read from the process."),
    }
    return messages.get(error, _("Unknown process error."))


class GoToolchain(QObject):
    output_received = pyqtSignal(str)
    error_received = pyqtSignal(str)
    execution_finished = pyqtSignal(int)

    def __init__(self, environment: GoEnvironment, parent: QObject | None = None) -> None:
        super().__init__(parent)
        self._environment = environment
        self._process: QProcess | None = None

    def environment(self) -> dict[str, str]:
        return self._environment.variables()

    def is_running(self) -> bool:
        return self._process is not None and self._process.state() != QProcess.NotRunning

    def run(self, config: RunConfiguration) -> None:
        args = ["run", config.go_target_argument(), *config.program_args]
        header = _("Running: {target}").format(target=config.target)
        self._start(config, args, header)

    def build(self, config: RunConfiguration) -> None:
        output = config.executable_name(windows=os.name == "nt")
        args = ["build", "-o", output, config.go_target_argument()]
        header = _("Building: {target}").format(target=config.target)
        self._start(config, args, header)

    def stop(self) -> None:
        if not self.is_running():
            return
        self._process.kill()
        self.output_received.emit("\n" + _("[Process terminated]") + "\n")

    def write_input(self, text: str) -> None:
        if not self.is_running():
            return
        self._process.write((text.rstrip("\n") + "\n").encode("utf-8"))

    def format_source(self, text: str) -> str:
        return format_go_source(
            text, self._environment.gofmt_executable(), self._environment.variables()
        )

    def _start(self, config: RunConfiguration, args: list[str], header: str) -> None:
        if self.is_running():
            self.error_received.emit(_("A process is already running.") + "\n")
            return
        self._process = self._create_process(config)
        self.output_received.emit(header + "\n" + SEPARATOR)
        self._process.start(self._environment.go_executable(), args)

    def _create_process(self, config: RunConfiguration) -> QProcess:
        process = QProcess(self)
        process.setWorkingDirectory(str(config.working_dir))
        env = QProcessEnvironment()
        for name, value in self._environment.variables().items():
            env.insert(name, value)
        process.setProcessEnvironment(env)
        process.readyReadStandardOutput.connect(self._read_stdout)
        process.readyReadStandardError.connect(self._read_stderr)
        process.finished.connect(self._on_finished)
        process.errorOccurred.connect(self._on_error)
        return process

    def _read_stdout(self) -> None:
        data = bytes(self._process.readAllStandardOutput())
        self.output_received.emit(data.decode("utf-8", errors="replace"))

    def _read_stderr(self) -> None:
        data = bytes(self._process.readAllStandardError())
        self.error_received.emit(data.decode("utf-8", errors="replace"))

    def _on_finished(self, exit_code: int, _exit_status: QProcess.ExitStatus) -> None:
        self.execution_finished.emit(exit_code)

    def _on_error(self, error: QProcess.ProcessError) -> None:
        self.error_received.emit(
            "\n" + _("Error: {message}").format(message=_process_error_message(error)) + "\n"
        )
        if error == QProcess.FailedToStart:
            self.execution_finished.emit(-1)
