"""One entry of the error catalog: regular expressions -> explanation texts."""

import re
from dataclasses import dataclass


@dataclass(frozen=True)
class CatalogEntry:
    """Recognises one kind of Go message.

    ``patterns`` are searched in the tool's message. Their named groups become the
    explanation placeholders, so every ``{placeholder}`` used in ``title``, ``body``
    or ``fix_hint`` must be a named group of *every* pattern.

    The texts are English gettext message ids (marked with ``N_``); the UI
    translates them and then calls ``str.format(**placeholders)``, so literal
    braces are written doubled: ``func main() {{ ... }}``.
    """

    explanation_id: str
    patterns: tuple[re.Pattern[str], ...]
    title: str
    body: str
    fix_hint: str

    def search(self, message: str) -> re.Match[str] | None:
        for pattern in self.patterns:
            match = pattern.search(message)
            if match is not None:
                return match
        return None


def catalog_entry(
    explanation_id: str, patterns: tuple[str, ...], title: str, body: str, fix_hint: str
) -> CatalogEntry:
    compiled = tuple(re.compile(pattern) for pattern in patterns)
    return CatalogEntry(explanation_id, compiled, title, body, fix_hint)
