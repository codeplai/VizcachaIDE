"""Services handed to every feature. Built only in ``vizcacha.ui.app`` (composition root)."""

from dataclasses import dataclass

from vizcacha.application.ports import (
    DebuggerPort,
    GoToolchainPort,
    LanguageServerPort,
    SettingsRepository,
)
from vizcacha.infrastructure.go_toolchain import GoEnvironment


@dataclass
class Services:
    settings: SettingsRepository
    environment: GoEnvironment
    toolchain: GoToolchainPort
    debugger: DebuggerPort
    language_server: LanguageServerPort | None = None  # track C (gopls)
