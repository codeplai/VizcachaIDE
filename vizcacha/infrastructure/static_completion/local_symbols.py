"""Names declared in the file being edited, found with simple regular expressions.

Good enough for the offline fallback: it does not understand scopes, it just
suggests every ``func``, ``type``, ``var``, ``const`` and ``:=`` name it sees.
"""

import re

from vizcacha.domain.completion import CompletionItem, CompletionKind
from vizcacha.i18n import _

IDENTIFIER = r"[A-Za-z_]\w*"
NAME_LIST = rf"{IDENTIFIER}(?:\s*,\s*{IDENTIFIER})*"
PATTERNS = (
    (re.compile(rf"^\s*func\s+(?:\([^)]*\)\s*)?({IDENTIFIER})", re.M), CompletionKind.FUNCTION),
    (re.compile(rf"\btype\s+({IDENTIFIER})"), CompletionKind.TYPE),
    (re.compile(rf"\bvar\s+({NAME_LIST})"), CompletionKind.VARIABLE),
    (re.compile(rf"\bconst\s+({NAME_LIST})"), CompletionKind.CONSTANT),
    (re.compile(rf"({NAME_LIST})\s*:="), CompletionKind.VARIABLE),
    (re.compile(rf"\bfor\s+({NAME_LIST})\s*:?=\s*range\b"), CompletionKind.VARIABLE),
)
GROUP_BLOCK = re.compile(r"\b(var|const)\s*\(([^)]*)\)", re.S)
GROUP_KINDS = {"var": CompletionKind.VARIABLE, "const": CompletionKind.CONSTANT}
GROUP_ENTRY = re.compile(rf"^\s*({NAME_LIST})", re.M)
BLANK = "_"


def _split_names(name_list: str) -> list[str]:
    return [name.strip() for name in name_list.split(",") if name.strip() != BLANK]


def declared_names(code: str) -> dict[str, CompletionKind]:
    """Every declared identifier in ``code`` with its kind (first declaration wins)."""
    names: dict[str, CompletionKind] = {}
    for pattern, kind in PATTERNS:
        for match in pattern.finditer(code):
            for name in _split_names(match.group(1)):
                names.setdefault(name, kind)
    for block in GROUP_BLOCK.finditer(code):
        kind = GROUP_KINDS[block.group(1)]
        for entry in GROUP_ENTRY.finditer(block.group(2)):
            for name in _split_names(entry.group(1)):
                names.setdefault(name, kind)
    return names


def local_completions(code: str, prefix: str, exclude: set[str]) -> list[CompletionItem]:
    """Declared names starting with ``prefix`` (case-insensitive), except ``exclude``."""
    lowered = prefix.lower()
    detail = _("declared in this file")
    return [
        CompletionItem(name, kind, detail=detail)
        for name, kind in declared_names(code).items()
        if name.lower().startswith(lowered) and name not in exclude
    ]
