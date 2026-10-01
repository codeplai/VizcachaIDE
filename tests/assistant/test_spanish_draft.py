"""The Spanish draft (docs/i18n/assistant.es.po) covers every catalog text and formats well."""

from pathlib import Path

from babel.messages.mofile import write_mo
from babel.messages.pofile import read_po

from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.i18n import install_language
from vizcacha.infrastructure.error_catalog import DEFAULT_ENTRIES, GoErrorExplainer
from vizcacha.ui.features.assistant.presentation import present

DRAFT = Path(__file__).resolve().parents[2] / "docs" / "i18n" / "assistant.es.po"


def _install_draft(locale_dir: Path) -> None:
    with DRAFT.open("rb") as handle:
        catalog = read_po(handle)
    target = locale_dir / "es" / "LC_MESSAGES"
    target.mkdir(parents=True)
    with (target / "vizcacha.mo").open("wb") as handle:
        write_mo(handle, catalog)
    install_language("es", locale_dir)


def test_draft_translates_every_catalog_text():
    with DRAFT.open("rb") as handle:
        translated = {message.id for message in read_po(handle) if message.string}
    texts = {text for item in DEFAULT_ENTRIES for text in (item.title, item.body, item.fix_hint)}

    assert not texts - translated


def test_explanation_is_shown_in_spanish_with_placeholders(tmp_path: Path):
    _install_draft(tmp_path)
    diagnostic = Diagnostic(
        None, Severity.ERROR, "index out of range [5] with length 3", "panic: ...", "panic"
    )

    view = present(diagnostic, GoErrorExplainer().explain(diagnostic))

    assert view.title == "El índice 5 está fuera de rango"
    assert "solo tiene 3 elemento(s)" in view.body
    assert view.original == "panic: ..."
