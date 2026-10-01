"""Composition root: the only place that instantiates concrete adapters.

Parallel tracks may add ONE line to FEATURES and change the ONE service they own
in ``build_services`` (track A: debugger). Nothing else.
"""

import sys
from pathlib import Path

from PyQt5.QtCore import QLibraryInfo, QLocale, QTranslator
from PyQt5.QtWidgets import QApplication

from vizcacha.application.ports import SettingsRepository
from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.i18n import install_language, resolve_language
from vizcacha.infrastructure.delve_dap import DelveDapDebugger
from vizcacha.infrastructure.error_catalog import GoErrorExplainer
from vizcacha.infrastructure.go_toolchain import GoEnvironment, GoToolchain
from vizcacha.infrastructure.gopls_lsp import GoplsLanguageServer
from vizcacha.infrastructure.settings import QSettingsRepository
from vizcacha.ui.features.assistant import register as register_assistant
from vizcacha.ui.features.debugger import register as register_debugger
from vizcacha.ui.features.editor import register as register_editor
from vizcacha.ui.features.files import register as register_files
from vizcacha.ui.features.help import register as register_help
from vizcacha.ui.features.language import register as register_language
from vizcacha.ui.features.project import register as register_project
from vizcacha.ui.features.run import register as register_run
from vizcacha.ui.features.settings import register as register_settings
from vizcacha.ui.main_window import MainWindow
from vizcacha.ui.services import Services
from vizcacha.ui.workbench import Workbench

APPLICATION_NAME = "VizcachaIDE"

# Order = order of menus' entries and toolbar buttons.
FEATURES = (
    register_files,
    register_editor,
    register_language,
    register_run,
    register_project,
    register_debugger,
    register_assistant,
    register_settings,
    register_help,
)


def build_services(settings_repository: SettingsRepository) -> Services:
    environment = GoEnvironment(settings_repository)
    return Services(
        settings=settings_repository,
        environment=environment,
        toolchain=GoToolchain(environment),
        debugger=DelveDapDebugger(environment),
        language_server=GoplsLanguageServer(environment),
        explainer=GoErrorExplainer(),
    )


def build_workbench(services: Services) -> Workbench:
    workbench = Workbench(MainWindow(), services)
    for register in FEATURES:
        register(workbench)
    workbench.start()
    return workbench


def configure_language(settings_repository: SettingsRepository) -> str:
    preferred = settings_repository.get(SettingsKeys.LANGUAGE, "")
    return install_language(resolve_language(preferred, QLocale.system().name()))


def install_qt_translations(app: QApplication, language: str) -> None:
    """Standard Qt buttons/dialogs (OK, Cancel, Save...) in the UI language."""
    translator = QTranslator(app)
    directory = QLibraryInfo.location(QLibraryInfo.TranslationsPath)
    if translator.load(f"qtbase_{language}", directory):
        app.installTranslator(translator)


def open_files_from_arguments(workbench: Workbench, arguments: list[str]) -> None:
    """Files passed on the command line (e.g. double-click on a .go file in the OS)."""
    for argument in arguments:
        path = Path(argument)
        if path.suffix == ".go" and path.is_file():
            workbench.editor.open_file(path.resolve())


def main(argv: list[str] | None = None) -> int:
    arguments = argv if argv is not None else sys.argv
    app = QApplication(arguments)
    app.setApplicationName(APPLICATION_NAME)
    app.setOrganizationName(APPLICATION_NAME)
    settings_repository = QSettingsRepository()
    install_qt_translations(app, configure_language(settings_repository))
    workbench = build_workbench(build_services(settings_repository))
    open_files_from_arguments(workbench, arguments[1:])
    workbench.window.show()
    return app.exec_()
