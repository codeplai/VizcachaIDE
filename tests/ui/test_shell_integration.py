"""macOS file-open events, Options > General and the status bar indicators."""

from pathlib import Path

from PyQt5.QtCore import QCoreApplication, QEvent, QUrl

from vizcacha.application.settings_keys import SettingsKeys
from vizcacha.ui.features.settings.status_indicators import GoVersionIndicator, parse_go_version


class SyntheticFileOpenEvent(QEvent):
    """PyQt5 cannot instantiate QFileOpenEvent, so this QEvent of type FileOpen
    offers the same ``file()`` / ``url()`` API that macOS events have."""

    def __init__(self, path: Path | None = None, url: QUrl | None = None) -> None:
        super().__init__(QEvent.FileOpen)
        self._path, self._url = path, url

    def file(self) -> str:
        return str(self._path) if self._path else ""

    def url(self) -> QUrl:
        return self._url or QUrl.fromLocalFile(self.file())


def test_file_open_event_opens_go_file(workbench, tmp_path: Path):
    from vizcacha.ui.app import install_file_open_handler

    source = tmp_path / "dock.go"
    source.write_text("package main\n", encoding="utf-8")
    app = QCoreApplication.instance()
    handler = install_file_open_handler(app, workbench)
    try:
        assert handler.eventFilter(app, SyntheticFileOpenEvent(tmp_path / "notes.txt"))
        assert workbench.editor.current_file_path() is None
        assert handler.eventFilter(app, SyntheticFileOpenEvent(url=QUrl.fromLocalFile(str(source))))
        assert not handler.eventFilter(app, QEvent(QEvent.Show))
    finally:
        app.removeEventFilter(handler)

    assert workbench.editor.current_file_path() == source.resolve()
    assert workbench.editor.count() == 1


def test_command_line_files_still_open(workbench, tmp_path: Path):
    from vizcacha.ui.app import open_files_from_arguments

    source = tmp_path / "args.go"
    source.write_text("package main\n", encoding="utf-8")

    open_files_from_arguments(workbench, [str(source), str(tmp_path / "missing.go")])

    assert workbench.editor.current_file_path() == source.resolve()


def _general_page(workbench):
    from vizcacha.ui.settings_dialog import SettingsDialog

    dialog = SettingsDialog(workbench)
    return dialog, dialog.pages[0]


def test_general_page_saves_language_and_asks_to_restart(workbench, settings):
    dialog, page = _general_page(workbench)
    assert page.title == "General"
    assert page.selected_language() == ""  # Automatic
    assert not page.restart_note.isVisibleTo(page)

    page.language.setCurrentIndex(page.language.findData("es"))
    assert page.restart_note.isVisibleTo(page)
    dialog.apply()

    assert settings.get(SettingsKeys.LANGUAGE, "") == "es"
    assert "Restart" in workbench.window.statusBar().currentMessage()


def test_general_page_loads_stored_language(workbench, settings):
    settings.set(SettingsKeys.LANGUAGE, "en")

    _dialog, page = _general_page(workbench)

    codes = [page.language.itemData(index) for index in range(page.language.count())]
    assert codes == ["", "en", "es"]
    assert page.language.currentText() == "English"


def test_go_version_parsing_and_display(qtbot):
    assert parse_go_version("go version go1.25.1 windows/amd64\n") == "1.25.1"
    assert parse_go_version("go version go1.26rc1 linux/arm64") == "1.26rc1"
    assert parse_go_version("garbage") is None
    indicator = GoVersionIndicator()
    qtbot.addWidget(indicator)

    indicator.show_version("1.25.1")
    assert indicator.text() == "Go 1.25.1"
    indicator.show_version(None)
    assert indicator.text() == "Go not found"


def test_missing_go_is_reported(qtbot, tmp_path: Path):
    indicator = GoVersionIndicator()
    qtbot.addWidget(indicator)

    with qtbot.waitSignal(indicator.version_detected, timeout=5000):
        indicator.detect(str(tmp_path / "go-missing"), {})

    assert indicator.text() == "Go not found"
