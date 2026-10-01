"""GoToolchainPort implementation: ``go run`` / ``go build`` / ``go mod`` through QProcess."""

import os
import tempfile
from collections.abc import Sequence
from pathlib import Path

from PyQt5.QtCore import QObject, QProcess, QProcessEnvironment, QTimer, pyqtSignal

from vizcacha.domain.project import RunConfiguration
from vizcacha.i18n import _
from vizcacha.infrastructure.go_toolchain.environment import GoEnvironment
from vizcacha.infrastructure.go_toolchain.formatter import format_go_source
from vizcacha.infrastructure.go_toolchain.process_tree import (
    kill_tree_windows,
    signal_children_posix,
)

SEPARATOR = "-" * 50 + "\n"
STOP_GRACE_MS = 2000
UNTITLED_FILE_NAME = "main.go"
UNTITLED_DIRECTORY_PREFIX = "vizcacha-untitled-"
IS_WINDOWS = os.name == "nt"


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
        self._scratch: tempfile.TemporaryDirectory | None = None

    def environment(self) -> dict[str, str]:
        return self._environment.variables()

    def is_running(self) -> bool:
        return self._process is not None and self._process.state() != QProcess.NotRunning

    def run(self, config: RunConfiguration) -> None:
        self._run(config, str(config.target))

    def run_untitled(
        self, source: str, program_args: Sequence[str] = ()
    ) -> RunConfiguration | None:
        """Run an unsaved tab: write it to a temporary directory and ``go run`` it."""
        if self._refuse_if_running():
            return None
        self._discard_scratch()
        self._scratch = tempfile.TemporaryDirectory(
            prefix=UNTITLED_DIRECTORY_PREFIX, ignore_cleanup_errors=True
        )
        path = Path(self._scratch.name) / UNTITLED_FILE_NAME
        try:
            path.write_text(source, encoding="utf-8")
        except OSError as error:
            self.error_received.emit(_("Could not save file: {error}").format(error=error) + "\n")
            return None
        config = RunConfiguration.for_file(path, tuple(program_args))
        self._run(config, _("Untitled"))
        return config

    def build(self, config: RunConfiguration) -> None:
        output = config.executable_name(windows=IS_WINDOWS)
        args = ["build", "-o", output, config.go_target_argument()]
        header = _("Building: {target}").format(target=config.target)
        self._start(config.working_dir, args, header)

    def run_go_command(self, working_dir: Path, args: Sequence[str]) -> bool:
        """Run ``go <args>`` (e.g. ``mod tidy``) asynchronously with the same signals.

        Returns False if another process is already running.
        """
        header = _("Running: {command}").format(command=" ".join(["go", *args]))
        return self._start(working_dir, list(args), header)

    def stop(self) -> None:
        if not self.is_running():
            return
        process = self._process
        if IS_WINDOWS:
            kill_tree_windows(int(process.processId()))
            process.kill()
        else:
            signal_children_posix(int(process.processId()), "TERM")
            process.terminate()
            QTimer.singleShot(STOP_GRACE_MS, lambda: self._kill_if_alive(process))
        self.output_received.emit("\n" + _("[Process terminated]") + "\n")

    def write_input(self, text: str) -> None:
        if not self.is_running():
            return
        self._process.write((text.rstrip("\n") + "\n").encode("utf-8"))

    def format_source(self, text: str) -> str:
        return format_go_source(
            text, self._environment.gofmt_executable(), self._environment.variables()
        )

    # --- internals ------------------------------------------------------------
    def _run(self, config: RunConfiguration, label: str) -> None:
        args = ["run", config.go_target_argument(), *config.program_args]
        header = _("Running: {target}").format(target=label)
        self._start(config.working_dir, args, header)

    def _refuse_if_running(self) -> bool:
        if self.is_running():
            self.error_received.emit(_("A process is already running.") + "\n")
        return self.is_running()

    def _start(self, working_dir: Path, args: list[str], header: str) -> bool:
        if self._refuse_if_running():
            return False
        if self._process is not None:
            self._process.deleteLater()
        self._process = self._create_process(working_dir)
        self.output_received.emit(header + "\n" + SEPARATOR)
        self._process.start(self._environment.go_executable(), args)
        return True

    def _create_process(self, working_dir: Path) -> QProcess:
        process = QProcess(self)
        process.setWorkingDirectory(str(working_dir))
        env = QProcessEnvironment()
        for name, value in self._environment.variables().items():
            env.insert(name, value)
        process.setProcessEnvironment(env)
        process.readyReadStandardOutput.connect(self._read_stdout)
        process.readyReadStandardError.connect(self._read_stderr)
        process.finished.connect(self._on_finished)
        process.errorOccurred.connect(self._on_error)
        return process

    def _kill_if_alive(self, process: QProcess) -> None:
        if process.state() == QProcess.NotRunning:
            return
        signal_children_posix(int(process.processId()), "KILL")
        process.kill()

    def _discard_scratch(self) -> None:
        if self._scratch is not None:
            self._scratch.cleanup()
            self._scratch = None

    def _read_stdout(self) -> None:
        data = bytes(self._process.readAllStandardOutput())
        self.output_received.emit(data.decode("utf-8", errors="replace"))

    def _read_stderr(self) -> None:
        data = bytes(self._process.readAllStandardError())
        self.error_received.emit(data.decode("utf-8", errors="replace"))

    def _on_finished(self, exit_code: int, _exit_status: QProcess.ExitStatus) -> None:
        self._discard_scratch()
        self.execution_finished.emit(exit_code)

    def _on_error(self, error: QProcess.ProcessError) -> None:
        self.error_received.emit(
            "\n" + _("Error: {message}").format(message=_process_error_message(error)) + "\n"
        )
        if error == QProcess.FailedToStart:
            self._discard_scratch()
            self.execution_finished.emit(-1)
