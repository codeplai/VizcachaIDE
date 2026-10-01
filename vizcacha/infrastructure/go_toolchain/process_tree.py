"""Stops ``go run`` together with the program it started.

``go run`` compiles the program and starts it as a child process. Killing only
``go`` would leave the user's program running, so the whole tree is signalled.
"""

import subprocess

SYSTEM_COMMAND_TIMEOUT_SECONDS = 5


def _run_system_command(arguments: list[str]) -> None:
    try:
        subprocess.run(
            arguments, capture_output=True, check=False, timeout=SYSTEM_COMMAND_TIMEOUT_SECONDS
        )
    except (OSError, subprocess.TimeoutExpired):
        # Best effort: the caller still kills the direct child through QProcess.
        return


def kill_tree_windows(pid: int) -> None:
    """Forcefully end ``pid`` and every descendant (``taskkill /T /F``)."""
    _run_system_command(["taskkill", "/PID", str(pid), "/T", "/F"])


def signal_children_posix(pid: int, signal_name: str) -> None:
    """Send ``SIGTERM``/``SIGKILL`` (``"TERM"``/``"KILL"``) to the children of ``pid``."""
    _run_system_command(["pkill", f"-{signal_name}", "-P", str(pid)])
