"""Completion provider backed by gopls, with the static GoAnalyzer as fallback."""

from pathlib import Path

from vizcacha.domain.completion import CompletionItem
from vizcacha.ui.editor.code_editor import CompletionProvider
from vizcacha.ui.features.language.editor_session import EditorSession
from vizcacha.ui.features.language.locations import offset_location


def language_server_provider(
    session: EditorSession, fallback: CompletionProvider
) -> CompletionProvider:
    """gopls suggestions; ``fallback`` when gopls is unavailable, slow or has none."""

    def provide(text: str, position: int, path: Path | None) -> list[CompletionItem]:
        location = offset_location(session.editor, position) if session.path else None
        items: list[CompletionItem] = []
        if location is not None:
            session.flush()
            items = session.server.completion(location)
        return items or fallback(text, position, path)

    return provide
