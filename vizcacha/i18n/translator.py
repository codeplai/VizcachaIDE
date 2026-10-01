"""English/Spanish translation based on gettext catalogs.

Source strings are written in English and wrapped in ``_()``. Catalogs live in
``vizcacha/i18n/locale/<lang>/LC_MESSAGES/vizcacha.po`` and are maintained with
Babel (see ``babel.cfg`` and docs/PLAN_DESARROLLO.md section 3).

This module is a leaf: it must not import Qt or any other vizcacha package.
"""

import gettext
from pathlib import Path

DOMAIN = "vizcacha"
LOCALE_DIR = Path(__file__).resolve().parent / "locale"
SUPPORTED_LANGUAGES = ("en", "es")
DEFAULT_LANGUAGE = "en"

_active: gettext.NullTranslations = gettext.NullTranslations()
_active_language = DEFAULT_LANGUAGE


def _(message: str) -> str:
    """Translate ``message`` into the active language."""
    return _active.gettext(message)


def ngettext(singular: str, plural: str, count: int) -> str:
    return _active.ngettext(singular, plural, count)


def N_(message: str) -> str:  # noqa: N802 - gettext convention
    """Mark a string for extraction without translating it now (module-level constants)."""
    return message


def resolve_language(preferred: str | None, system_locale: str) -> str:
    """Pick the UI language: explicit preference, else system locale (es* -> es), else en."""
    if preferred in SUPPORTED_LANGUAGES:
        return preferred
    if system_locale.lower().startswith("es"):
        return "es"
    return DEFAULT_LANGUAGE


def install_language(language: str, locale_dir: Path = LOCALE_DIR) -> str:
    """Activate ``language``. Unknown languages and missing catalogs fall back to English."""
    global _active, _active_language
    if language not in SUPPORTED_LANGUAGES:
        language = DEFAULT_LANGUAGE
    _active = gettext.translation(
        DOMAIN, localedir=str(locale_dir), languages=[language], fallback=True
    )
    _active_language = language
    return language


def active_language() -> str:
    return _active_language
