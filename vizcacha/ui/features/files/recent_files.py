"""Recently opened files (most recent first, at most 10) and the "Open Recent" submenu."""

from collections.abc import Callable
from pathlib import Path

from PyQt5.QtWidgets import QAction, QMenu, QWidget

from vizcacha.application.ports import SettingsRepository
from vizcacha.i18n import _

# New key owned by track D (Contract change request: move to SettingsKeys).
RECENT_FILES_KEY = "files/recent"
MAX_RECENT_FILES = 10
SEPARATOR = "\n"  # stored as one string: QSettings round-trips lists inconsistently


class RecentFiles:
    def __init__(self, settings: SettingsRepository) -> None:
        self._settings = settings

    def paths(self) -> list[str]:
        stored = self._settings.get(RECENT_FILES_KEY, "")
        return [path for path in str(stored or "").split(SEPARATOR) if path]

    def add(self, path: str) -> None:
        normalized = str(Path(path))
        paths = [normalized] + [p for p in self.paths() if p != normalized]
        self._store(paths[:MAX_RECENT_FILES])

    def remove(self, path: str) -> None:
        self._store([p for p in self.paths() if p != path])

    def clear(self) -> None:
        self._store([])

    def _store(self, paths: list[str]) -> None:
        self._settings.set(RECENT_FILES_KEY, SEPARATOR.join(paths))


class RecentFilesMenu(QMenu):
    """Rebuilt every time it is about to be shown, so it is always up to date."""

    def __init__(
        self, recent: RecentFiles, open_path: Callable[[Path], bool], parent: QWidget
    ) -> None:
        super().__init__(_("Open &Recent"), parent)
        self._recent = recent
        self._open_path = open_path
        self.aboutToShow.connect(self.rebuild)
        self.rebuild()

    def rebuild(self) -> None:
        self.clear()
        paths = self._recent.paths()
        for number, path in enumerate(paths, start=1):
            action = self.addAction(f"&{number % 10}  {path}")
            action.setData(path)
            action.triggered.connect(lambda _checked=False, p=path: self.open_recent(p))
        if not paths:
            empty = self.addAction(_("No recent files"))
            empty.setEnabled(False)
        self.addSeparator()
        clear: QAction = self.addAction(_("&Clear Recent Files"))
        clear.setEnabled(bool(paths))
        clear.triggered.connect(lambda _checked=False: self.clear_recent())

    def open_recent(self, path: str) -> None:
        if not Path(path).is_file():
            self._recent.remove(path)
            self.rebuild()
        self._open_path(Path(path))

    def clear_recent(self) -> None:
        self._recent.clear()
        self.rebuild()
