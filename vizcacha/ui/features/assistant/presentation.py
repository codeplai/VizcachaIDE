"""Turns a (Diagnostic, ErrorExplanation | None) pair into the texts the panel shows."""

from dataclasses import dataclass
from urllib.parse import quote_plus

from vizcacha.domain.diagnostics import Diagnostic
from vizcacha.domain.explanations import ErrorExplanation
from vizcacha.i18n import _

SEARCH_URL = "https://www.google.com/search?q={query}"


class _KeepUnknown(dict):
    """Leaves ``{name}`` as it is when a translation uses an unknown placeholder."""

    def __missing__(self, key: str) -> str:
        return "{" + key + "}"


@dataclass(frozen=True)
class ExplanationView:
    title: str
    body: str
    fix_hint: str
    location: str
    original: str
    recognised: bool


def fill(message_id: str, placeholders: dict[str, str]) -> str:
    """Translate ``message_id`` and fill its ``{placeholders}``."""
    text = _(message_id)
    try:
        return text.format_map(_KeepUnknown(placeholders))
    except (ValueError, IndexError):  # broken braces in a translation: show it unfilled
        return text


def describe_location(diagnostic: Diagnostic) -> str:
    location = diagnostic.location
    if location is None:
        return ""
    return _("{file}, line {line}").format(file=location.file.name, line=location.line)


def present(diagnostic: Diagnostic, explanation: ErrorExplanation | None) -> ExplanationView:
    location = describe_location(diagnostic)
    original = diagnostic.raw_text or diagnostic.message
    if explanation is None:
        return ExplanationView(
            title=_("Go reported a problem"),
            body=_(
                "There is no beginner explanation for this message yet. The text below is "
                "exactly what Go printed: searching for it on the web usually helps."
            ),
            fix_hint="",
            location=location,
            original=original,
            recognised=False,
        )
    placeholders = explanation.placeholders
    return ExplanationView(
        title=fill(explanation.title, placeholders),
        body=fill(explanation.body, placeholders),
        fix_hint=fill(explanation.fix_hint, placeholders) if explanation.fix_hint else "",
        location=location,
        original=original,
        recognised=True,
    )


def search_url(diagnostic: Diagnostic) -> str:
    """Web search for the untranslated Go message (without the file path)."""
    return SEARCH_URL.format(query=quote_plus(f"golang {diagnostic.message}"))
