"""Tool resolution order: configured -> bundled next to the app -> PATH -> missing."""

import os
import stat
from pathlib import Path

import pytest

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.infrastructure.go_toolchain import GoEnvironment, ToolOrigin
from vizcacha.infrastructure.go_toolchain.tool_locator import executable_file_name
from vizcacha.infrastructure.settings import InMemorySettingsRepository


def _fake_tool(directory: Path, tool: str) -> Path:
    directory.mkdir(parents=True, exist_ok=True)
    path = directory / executable_file_name(tool)
    path.write_text("#!/bin/sh\n", encoding="utf-8")
    path.chmod(path.stat().st_mode | stat.S_IEXEC)
    return path


@pytest.fixture
def app_dir(tmp_path: Path) -> Path:
    return tmp_path / "app"


@pytest.fixture
def path_dir(tmp_path: Path) -> Path:
    directory = tmp_path / "path_bin"
    directory.mkdir()
    return directory


def _environment(app_dir: Path, path_dir: Path, settings=None) -> GoEnvironment:
    base = {"PATH": str(path_dir), "PATHEXT": ".EXE"}
    return GoEnvironment(settings or InMemorySettingsRepository(), base, app_directory=app_dir)


def test_missing_tools_fall_back_to_bare_names(app_dir, path_dir):
    env = _environment(app_dir, path_dir)

    origins = env.tool_origins()

    assert {tool: location.origin for tool, location in origins.items()} == {
        "go": ToolOrigin.MISSING,
        "gofmt": ToolOrigin.MISSING,
        "dlv": ToolOrigin.MISSING,
        "gopls": ToolOrigin.MISSING,
    }
    assert env.go_executable() == "go"
    assert env.gopls_executable() == "gopls"


def test_tools_on_path_are_found(app_dir, path_dir):
    go = _fake_tool(path_dir, "go")
    env = _environment(app_dir, path_dir)

    location = env.locate("go")

    assert location.origin is ToolOrigin.PATH
    assert Path(location.path).samefile(go)
    assert env.locate("dlv").origin is ToolOrigin.MISSING


def test_bundled_toolchain_wins_over_path(app_dir, path_dir):
    _fake_tool(path_dir, "go")
    bundled_go = _fake_tool(app_dir / "toolchain" / "go" / "bin", "go")
    _fake_tool(app_dir / "toolchain" / "go" / "bin", "gofmt")
    bundled_dlv = _fake_tool(app_dir / "toolchain" / "bin", "dlv")
    env = _environment(app_dir, path_dir)

    assert env.locate("go").origin is ToolOrigin.BUNDLED
    assert env.go_executable() == str(bundled_go)
    assert env.locate("gofmt").origin is ToolOrigin.BUNDLED
    assert env.delve_executable() == str(bundled_dlv)
    assert env.locate("gopls").origin is ToolOrigin.MISSING


def test_bundled_go_sets_goroot_and_prepends_path(app_dir, path_dir):
    _fake_tool(app_dir / "toolchain" / "go" / "bin", "go")
    _fake_tool(app_dir / "toolchain" / "bin", "gopls")
    env = _environment(app_dir, path_dir)

    variables = env.variables()

    assert variables["GOROOT"] == str(app_dir / "toolchain" / "go")
    entries = variables["PATH"].split(os.pathsep)
    assert set(entries[:2]) == {
        str(app_dir / "toolchain" / "go" / "bin"),
        str(app_dir / "toolchain" / "bin"),
    }
    assert entries[-1] == str(path_dir)


def test_configured_goroot_overrides_bundled_one(app_dir, path_dir):
    _fake_tool(app_dir / "toolchain" / "go" / "bin", "go")
    settings = InMemorySettingsRepository({SettingsKeys.GOROOT: "/custom/go"})

    variables = _environment(app_dir, path_dir, settings).variables()

    assert variables["GOROOT"] == "/custom/go"


def test_configured_paths_win_over_everything(app_dir, path_dir):
    _fake_tool(app_dir / "toolchain" / "go" / "bin", "go")
    settings = InMemorySettingsRepository(
        {SettingsKeys.GO_PATH: "/opt/go/bin/go", SettingsKeys.GOPLS_PATH: "/opt/gopls"}
    )
    env = _environment(app_dir, path_dir, settings)

    assert env.locate("go").origin is ToolOrigin.CONFIGURED
    assert env.locate("gofmt").path == str(Path("/opt/go/bin/gofmt"))
    assert env.locate("gofmt").origin is ToolOrigin.CONFIGURED
    assert env.gopls_executable() == "/opt/gopls"
    assert "GOROOT" not in env.variables()  # configured go: no bundled GOROOT


def test_detect_again_sees_newly_installed_tools(app_dir, path_dir):
    env = _environment(app_dir, path_dir)
    assert env.locate("gopls").origin is ToolOrigin.MISSING
    _fake_tool(path_dir, "gopls")

    assert env.locate("gopls").origin is ToolOrigin.MISSING  # cached
    assert env.detect_again()["gopls"].origin is ToolOrigin.PATH
