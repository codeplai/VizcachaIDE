import re
from pathlib import Path

import pytest
from babel.messages.pofile import read_po

from vizcacha.i18n import _, active_language, install_language, resolve_language
from vizcacha.i18n.translator import LOCALE_DIR

PLACEHOLDER = re.compile(r"\{[a-z_]+\}")


@pytest.mark.parametrize(
    ("preferred", "system", "expected"),
    [
        ("es", "en_US", "es"),
        ("en", "es_PE", "en"),
        ("", "es_PE", "es"),
        ("", "es", "es"),
        ("", "pt_BR", "en"),
        (None, "de_DE", "en"),
        ("fr", "fr_FR", "en"),
    ],
)
def test_resolve_language(preferred, system, expected):
    assert resolve_language(preferred, system) == expected


def test_spanish_catalog_translates_ui_strings():
    install_language("es")

    assert active_language() == "es"
    assert _("&File") == "&Archivo"
    assert _("Untitled") == "Sin título"


def test_unknown_language_and_missing_strings_fall_back_to_english():
    assert install_language("fr") == "en"
    install_language("es")
    assert _("A sentence that is not in any catalog") == "A sentence that is not in any catalog"


def test_missing_catalog_directory_falls_back_to_english(tmp_path: Path):
    install_language("es", locale_dir=tmp_path)
    assert _("&File") == "&File"


def _spanish_messages():
    with (LOCALE_DIR / "es" / "LC_MESSAGES" / "vizcacha.po").open("rb") as handle:
        return [message for message in read_po(handle) if message.id]


def _as_tuple(text: str | tuple[str, ...]) -> tuple[str, ...]:
    return text if isinstance(text, tuple) else (text,)


def test_spanish_translations_keep_placeholders():
    for message in _spanish_messages():
        expected = set(PLACEHOLDER.findall(_as_tuple(message.id)[0]))
        for translation in _as_tuple(message.string):
            if translation:
                assert set(PLACEHOLDER.findall(translation)) == expected, message.id


def test_spanish_catalog_is_complete():
    untranslated = [
        message.id for message in _spanish_messages() if not all(_as_tuple(message.string))
    ]
    assert untranslated == []


def test_menu_accelerators_are_unique_in_spanish():
    menus = ["&File", "&Edit", "&View", "&Run", "&Debug", "&Tools", "&Help"]
    install_language("es")
    keys = [_(title).split("&")[1][0].lower() for title in menus]
    assert len(keys) == len(set(keys))
