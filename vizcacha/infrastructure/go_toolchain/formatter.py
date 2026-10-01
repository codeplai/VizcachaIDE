"""gofmt wrapper (synchronous; gofmt is fast enough for a single file)."""

import subprocess
from collections.abc import Mapping

from vizcacha.application.errors import GoFormatError, GoToolchainNotFoundError
from vizcacha.i18n import _

FORMAT_TIMEOUT_SECONDS = 10


def format_go_source(text: str, gofmt: str, environment: Mapping[str, str]) -> str:
    try:
        result = subprocess.run(
            [gofmt],
            input=text,
            capture_output=True,
            text=True,
            encoding="utf-8",
            env=dict(environment),
            timeout=FORMAT_TIMEOUT_SECONDS,
            check=False,
        )
    except FileNotFoundError as error:
        message = _("gofmt was not found: {path}").format(path=gofmt)
        raise GoToolchainNotFoundError(message) from error
    except subprocess.TimeoutExpired as error:
        raise GoFormatError(_("gofmt took too long and was stopped.")) from error
    if result.returncode != 0:
        raise GoFormatError(result.stderr.strip())
    return result.stdout
