"""Code completion and call-tip data."""

from dataclasses import dataclass
from enum import Enum


class CompletionKind(Enum):
    KEYWORD = "keyword"
    FUNCTION = "function"
    METHOD = "method"
    VARIABLE = "variable"
    CONSTANT = "constant"
    FIELD = "field"
    TYPE = "type"
    PACKAGE = "package"
    OTHER = "other"


@dataclass(frozen=True)
class CompletionItem:
    label: str
    kind: CompletionKind
    detail: str = ""
    documentation: str = ""
    insert_text: str = ""

    @property
    def text_to_insert(self) -> str:
        return self.insert_text or self.label


@dataclass(frozen=True)
class SignatureHelp:
    label: str
    documentation: str = ""
    parameters: tuple[str, ...] = ()
    active_parameter: int = 0
