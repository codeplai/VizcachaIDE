"""Internationalisation (English / Spanish).

Usage everywhere in the project::

    from vizcacha.i18n import _

    label = _("Run")

Never use ``_`` as a throwaway variable name (``name, _ = ...``): it would hide
the translation function. Use ``_unused`` or ``_filter`` instead.
"""

from vizcacha.i18n.translator import (
    DEFAULT_LANGUAGE,
    N_,
    SUPPORTED_LANGUAGES,
    _,
    active_language,
    install_language,
    ngettext,
    resolve_language,
)

__all__ = [
    "DEFAULT_LANGUAGE",
    "SUPPORTED_LANGUAGES",
    "N_",
    "_",
    "active_language",
    "install_language",
    "ngettext",
    "resolve_language",
]
