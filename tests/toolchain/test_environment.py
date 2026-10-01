from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.infrastructure.go_toolchain import GoEnvironment, parse_extra_variables
from vizcacha.infrastructure.settings import InMemorySettingsRepository


def test_parse_extra_variables_ignores_noise():
    text = "GOOS=linux\n\n  GOARCH = amd64 \nnot a variable\n=nokey\nA=b=c"

    assert parse_extra_variables(text) == {"GOOS": "linux", "GOARCH": "amd64", "A": "b=c"}


def test_missing_tools_fall_back_to_their_bare_names(tmp_path):
    empty_path = str(tmp_path)  # no tools here, whatever is installed on the machine
    env = GoEnvironment(
        InMemorySettingsRepository(),
        base_environment={"PATH": empty_path},
        app_directory=tmp_path,
    )

    assert env.go_executable() == "go"
    assert env.gofmt_executable() == "gofmt"
    assert env.delve_executable() == "dlv"
    assert env.gopls_executable() == "gopls"
    assert env.variables() == {"PATH": empty_path}


def test_configured_values_override_base_environment():
    settings = InMemorySettingsRepository(
        {
            SettingsKeys.GO_PATH: r"C:\Go\bin\go.exe",
            SettingsKeys.GOPATH: "/home/me/go",
            SettingsKeys.GOROOT: "  ",
            SettingsKeys.DELVE_PATH: "/opt/dlv",
            SettingsKeys.EXTRA_VARS: "CGO_ENABLED=0",
        }
    )
    env = GoEnvironment(settings, base_environment={"GOROOT": "/usr/lib/go"})

    assert env.gofmt_executable().endswith("gofmt.exe")
    assert env.delve_executable() == "/opt/dlv"
    assert env.variables() == {
        "GOROOT": "/usr/lib/go",
        "GOPATH": "/home/me/go",
        "CGO_ENABLED": "0",
    }
