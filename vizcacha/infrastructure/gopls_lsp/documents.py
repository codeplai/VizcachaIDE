"""Texts of the documents open in gopls, needed to convert UTF-16 positions."""

from dataclasses import dataclass
from pathlib import Path

from vizcacha.infrastructure.gopls_lsp.positions import path_to_uri, uri_to_path


@dataclass
class OpenDocument:
    path: Path
    text: str
    version: int


class OpenDocuments:
    def __init__(self) -> None:
        self._by_uri: dict[str, OpenDocument] = {}

    def open(self, path: Path, text: str) -> str:
        uri = path_to_uri(path)
        self._by_uri[uri] = OpenDocument(Path(path), text, 1)
        return uri

    def change(self, path: Path, text: str, version: int) -> str | None:
        """Store the new text; returns the uri, or None if ``path`` is not open."""
        document = self._by_uri.get(path_to_uri(path))
        if document is None:
            return None
        document.text = text
        document.version = max(version, document.version + 1)
        return path_to_uri(path)

    def close(self, path: Path) -> str | None:
        uri = path_to_uri(path)
        return uri if self._by_uri.pop(uri, None) is not None else None

    def get(self, uri: str) -> OpenDocument | None:
        return self._by_uri.get(uri)

    def all(self) -> list[tuple[str, OpenDocument]]:
        return list(self._by_uri.items())

    def path_for(self, uri: str) -> Path:
        """The path the editor used for ``uri`` (keeps its spelling), else decoded."""
        document = self._by_uri.get(uri)
        return document.path if document is not None else uri_to_path(uri)

    def text_of(self, path: Path) -> str:
        """Current text of ``path``: the open document, else the file on disk."""
        document = self._by_uri.get(path_to_uri(path))
        if document is not None:
            return document.text
        try:
            return Path(path).read_text(encoding="utf-8", errors="replace")
        except OSError:
            return ""
