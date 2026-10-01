"""Single place that knows where Go tools live and which environment they get.

Tools are resolved in this order: path configured in Options -> toolchain bundled
next to the application (see ``tool_locator``) -> PATH.
"""

import os
from collections.abc import Mapping
from pathlib import Path

from vizcacha.application.ports import SettingsRepository
from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.infrastructure.go_toolchain.tool_locator import (
    DELVE,
    GO,
    GOFMT,
    GOPLS,
    TOOLS,
    ToolLocation,
    ToolLocator,
    ToolOrigin,
    application_directory,
)

CONFIGURED_TOOL_KEYS = {
    GO: SettingsKeys.GO_PATH,
    DELVE: SettingsKeys.DELVE_PATH,
    GOPLS: SettingsKeys.GOPLS_PATH,
}


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


def _path_variable_name(env: Mapping[str, str]) -> str:
    """``PATH`` as spelled in ``env`` (Windows may use ``Path``)."""
    return next((name for name in env if name.upper() == "PATH"), "PATH")


class GoEnvironment:
    def __init__(
        self,
        settings: SettingsRepository,
        base_environment: Mapping[str, str] | None = None,
        app_directory: Path | None = None,
    ) -> None:
        self._settings = settings
        self._base = dict(os.environ if base_environment is None else base_environment)
        search_path = self._base.get(_path_variable_name(self._base))
        self._locator = ToolLocator(app_directory or application_directory(), search_path)

    # --- executables ----------------------------------------------------------
    def go_executable(self) -> str:
        return self._executable(GO)

    def gofmt_executable(self) -> str:
        return self._executable(GOFMT)

    def delve_executable(self) -> str:
        return self._executable(DELVE)

    def gopls_executable(self) -> str:
        return self._executable(GOPLS)

    # --- where tools come from (for Options and a future first-run wizard) -----
    def locate(self, tool: str) -> ToolLocation:
        """Origin of ``tool`` ("go", "gofmt", "dlv" or "gopls")."""
        if tool == GOFMT:
            return self._configured_gofmt() or self._locator.detect(GOFMT)
        configured = self._configured(CONFIGURED_TOOL_KEYS[tool])
        if configured:
            return ToolLocation(tool, ToolOrigin.CONFIGURED, configured)
        return self._locator.detect(tool)

    def tool_origins(self) -> dict[str, ToolLocation]:
        return {tool: self.locate(tool) for tool in TOOLS}

    def detect_again(self) -> dict[str, ToolLocation]:
        """Forget cached detections (e.g. after installing Go) and detect again."""
        self._locator.forget()
        return self.tool_origins()

    # --- process environment --------------------------------------------------
    def variables(self) -> dict[str, str]:
        """Full process environment for Go tools and the user's program."""
        env = dict(self._base)
        if self.locate(GO).origin is ToolOrigin.BUNDLED:
            self._apply_bundled_toolchain(env)
        for name, key in (("GOPATH", SettingsKeys.GOPATH), ("GOROOT", SettingsKeys.GOROOT)):
            value = self._configured(key)
            if value:
                env[name] = value
        env.update(parse_extra_variables(self._settings.get(SettingsKeys.EXTRA_VARS, "") or ""))
        return env

    def _apply_bundled_toolchain(self, env: dict[str, str]) -> None:
        env["GOROOT"] = str(self._locator.bundled_goroot())
        env.setdefault("GOTOOLCHAIN", "local")
        path_name = _path_variable_name(env)
        directories = [str(directory) for directory in self._locator.bundled_bin_directories()]
        current = env.get(path_name, "")
        env[path_name] = os.pathsep.join([*directories, current] if current else directories)

    # --- internals ------------------------------------------------------------
    def _executable(self, tool: str) -> str:
        location = self.locate(tool)
        return location.path or tool

    def _configured_gofmt(self) -> ToolLocation | None:
        go = self._configured(SettingsKeys.GO_PATH)
        if not go:
            return None
        go_path = Path(go)
        gofmt = go_path.with_name(GOFMT + go_path.suffix)
        return ToolLocation(GOFMT, ToolOrigin.CONFIGURED, str(gofmt))

    def _configured(self, key: str) -> str:
        return (self._settings.get(key, "") or "").strip()
