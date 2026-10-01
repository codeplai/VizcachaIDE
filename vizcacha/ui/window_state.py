"""Remembers the window geometry and the dock layout between sessions.

Both are stored as base64 text, so every SettingsRepository (QSettings, in-memory)
keeps them the same way. Bump ``LAYOUT_VERSION`` when the default layout changes
in a way that old saved layouts should be ignored.
"""

from PyQt5.QtCore import QByteArray
from PyQt5.QtWidgets import QMainWindow

from vizcacha.application.ports import SettingsRepository
from vizcacha.application.settings_keys import SettingsKeys

LAYOUT_VERSION = 2


def _encode(data: QByteArray) -> str:
    return bytes(data.toBase64()).decode("ascii")


def _decode(text: object) -> QByteArray | None:
    if not isinstance(text, str) or not text:
        return None
    return QByteArray.fromBase64(text.encode("ascii"))


def save_window_state(window: QMainWindow, settings: SettingsRepository) -> None:
    settings.set(SettingsKeys.WINDOW_GEOMETRY, _encode(window.saveGeometry()))
    settings.set(SettingsKeys.WINDOW_STATE, _encode(window.saveState(LAYOUT_VERSION)))


def restore_window_state(window: QMainWindow, settings: SettingsRepository) -> bool:
    """True when a saved dock layout was restored (geometry alone does not count)."""
    geometry = _decode(settings.get(SettingsKeys.WINDOW_GEOMETRY, ""))
    if geometry is not None:
        window.restoreGeometry(geometry)
    state = _decode(settings.get(SettingsKeys.WINDOW_STATE, ""))
    if state is None:
        return False
    return window.restoreState(state, LAYOUT_VERSION)
