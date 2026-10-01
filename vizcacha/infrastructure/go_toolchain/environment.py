"""Single place that knows where Go tools live and which environment they get.

Before phase 0 this logic was duplicated in core/runner.py and MainWindow.build_code().
Track E extends it (bundled toolchain next to the executable, then PATH).
"""

import os
from collections.abc import Mapping
from pathlib import Path

from vizcacha.application.ports import SettingsRepository
from vizcacha.application.settings_keys import SettingsKeys


def parse_extra_variables(text: str) -> dict[str, str]:
    """Parse ``NAME=VALUE`` lines; blank lines and lines without ``=`` are ignored."""
    variables: dict[str, str] = {}
    for line in text.splitlines():
        line = line.strip()
        if "=" not in line:
            continue
        key, value = line.split("=", 1)
        if key.strip():
            variables[key.strip()] = value.strip()
    return variables


class GoEnvironment:
    def __init__(
        self, settings: SettingsRepository, base_environment: Mapping[str, str] | None = None
    ) -> None:
        self._settings = settings
        self._base = dict(os.environ if base_environment is None else base_environment)

    def go_executable(self) -> str:
        return self._configured(SettingsKeys.GO_PATH) or "go"

    def gofmt_executable(self) -> str:
        go = self.go_executable()
        if go == "go":
            return "gofmt"
        go_path = Path(go)
        return str(go_path.with_name("gofmt" + go_path.suffix))

    def delve_executable(self) -> str:
        return self._configured(SettingsKeys.DELVE_PATH) or "dlv"

    def gopls_executable(self) -> str:
        return self._configured(SettingsKeys.GOPLS_PATH) or "gopls"

    def variables(self) -> dict[str, str]:
        """Full process environment for Go tools and the user's program."""
        env = dict(self._base)
        for name, key in (("GOPATH", SettingsKeys.GOPATH), ("GOROOT", SettingsKeys.GOROOT)):
            value = self._configured(key)
            if value:
                env[name] = value
        env.update(parse_extra_variables(self._settings.get(SettingsKeys.EXTRA_VARS, "") or ""))
        return env

    def _configured(self, key: str) -> str:
        return (self._settings.get(key, "") or "").strip()
