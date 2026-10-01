"""Conversions between editor positions and LSP positions.

The domain uses 1-based lines and columns counted in Unicode code points (Python
string indices). LSP uses 0-based lines and characters counted in UTF-16 code units.
"""

from pathlib import Path
from urllib.parse import unquote, urlparse
from urllib.request import url2pathname

from lsprotocol import types

from vizcacha.domain.diagnostics import SourceLocation

GO_MODULE_FILE = "go.mod"


def path_to_uri(path: Path) -> str:
    return Path(path).resolve().as_uri()


def uri_to_path(uri: str) -> Path:
    parsed = urlparse(uri)
    return Path(url2pathname(unquote(parsed.path)))


def module_root(file: Path) -> Path:
    """Folder with the nearest ``go.mod`` above ``file``, or the file's own folder."""
    folder = Path(file).resolve().parent
    for candidate in (folder, *folder.parents):
        if (candidate / GO_MODULE_FILE).is_file():
            return candidate
    return folder


def utf16_offset(line_text: str, column_index: int) -> int:
    """UTF-16 code units before the code point index ``column_index``."""
    prefix = line_text[: max(column_index, 0)]
    return len(prefix.encode("utf-16-le")) // 2


def code_point_index(line_text: str, utf16_character: int) -> int:
    """Inverse of :func:`utf16_offset`."""
    units = 0
    for index, char in enumerate(line_text):
        if units >= utf16_character:
            return index
        units += 2 if ord(char) > 0xFFFF else 1
    return len(line_text)


def line_of(text: str, line_index: int) -> str:
    lines = text.split("\n")
    if 0 <= line_index < len(lines):
        return lines[line_index].rstrip("\r")
    return ""


def to_lsp_position(text: str, line: int, column: int) -> types.Position:
    """1-based (line, column) in ``text`` -> LSP Position."""
    line_index = max(line - 1, 0)
    character = utf16_offset(line_of(text, line_index), column - 1)
    return types.Position(line=line_index, character=character)


def from_lsp_position(text: str, position: types.Position) -> tuple[int, int]:
    """LSP Position -> 1-based (line, column) in ``text``."""
    column = code_point_index(line_of(text, position.line), position.character)
    return position.line + 1, column + 1


def location_in(text: str, path: Path, position: types.Position) -> SourceLocation:
    """LSP Position in ``text`` (the content of ``path``) -> domain SourceLocation."""
    line, column = from_lsp_position(text, position)
    return SourceLocation(path, line, column)
