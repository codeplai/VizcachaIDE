"""SettingsRepository backed by QSettings (Windows registry, macOS plist, Linux .ini)."""

from typing import TypeVar

from PyQt5.QtCore import QSettings

T = TypeVar("T")

_TYPED_DEFAULTS = (bool, int, float, str)


class QSettingsRepository:
    def __init__(self, settings: QSettings | None = None) -> None:
        self._settings = settings if settings is not None else QSettings()

    def get(self, key: str, default: T) -> T:
        if default is None:
            return self._settings.value(key)
        if isinstance(default, _TYPED_DEFAULTS):
            return self._settings.value(key, default, type=type(default))
        return self._settings.value(key, default)

    def set(self, key: str, value: object) -> None:
        self._settings.setValue(key, value)
