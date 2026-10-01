"""Checks that adapters honour the frozen contracts in application/ports.py."""

import re
from pathlib import Path

import pytest
from importlinter.cli import lint_imports
from PyQt5.QtCore import pyqtBoundSignal

from vizcacha.application.ports import (
    DEBUGGER_SIGNALS,
    LANGUAGE_SERVER_SIGNALS,
    TOOLCHAIN_SIGNALS,
    DebuggerPort,
    ErrorExplainerPort,
    GoToolchainPort,
    LanguageServerPort,
    SettingsRepository,
)
from vizcacha.infrastructure.delve_dap import DelveDapDebugger
from vizcacha.infrastructure.error_catalog import GoErrorExplainer
from vizcacha.infrastructure.go_toolchain import GoEnvironment, GoToolchain
from vizcacha.infrastructure.gopls_lsp import GoplsLanguageServer
from vizcacha.infrastructure.null_debugger import NullDebugger
from vizcacha.infrastructure.settings import InMemorySettingsRepository, QSettingsRepository

PACKAGE_ROOT = Path(__file__).resolve().parents[1] / "vizcacha"
GUI_IMPORT = re.compile(r"^\s*(from|import)\s+PyQt5\.(QtWidgets|QtGui)", re.MULTILINE)


def _assert_signals(adapter, expected: dict) -> None:
    for name in expected:
        assert isinstance(getattr(adapter, name, None), pyqtBoundSignal), name


@pytest.mark.parametrize("factory", [lambda s: GoToolchain(GoEnvironment(s))])
def test_toolchain_adapters_follow_contract(factory, settings):
    adapter = factory(settings)
    assert isinstance(adapter, GoToolchainPort)
    _assert_signals(adapter, TOOLCHAIN_SIGNALS)


@pytest.mark.parametrize(
    "factory", [lambda s: NullDebugger(), lambda s: DelveDapDebugger(GoEnvironment(s))]
)
def test_debugger_adapters_follow_contract(factory, settings):
    adapter = factory(settings)
    assert isinstance(adapter, DebuggerPort)
    _assert_signals(adapter, DEBUGGER_SIGNALS)


def test_language_server_follows_contract(settings):
    adapter = GoplsLanguageServer(GoEnvironment(settings))
    assert isinstance(adapter, LanguageServerPort)
    _assert_signals(adapter, LANGUAGE_SERVER_SIGNALS)


def test_error_explainer_follows_contract():
    assert isinstance(GoErrorExplainer(), ErrorExplainerPort)


def test_settings_repositories_follow_contract():
    assert isinstance(InMemorySettingsRepository(), SettingsRepository)
    assert isinstance(QSettingsRepository.__new__(QSettingsRepository), SettingsRepository)


def test_infrastructure_does_not_use_widgets():
    offenders = [
        str(path.relative_to(PACKAGE_ROOT))
        for path in (PACKAGE_ROOT / "infrastructure").rglob("*.py")
        if GUI_IMPORT.search(path.read_text(encoding="utf-8"))
    ]
    assert offenders == []


def test_import_layers(capsys):
    assert lint_imports() == 0, capsys.readouterr().out
