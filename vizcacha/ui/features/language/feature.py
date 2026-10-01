"""Code intelligence with gopls: completion, diagnostics, hover, definition, outline.

Without a language server (or if gopls is missing/crashes) every editor keeps its
static completion and a one-time notice is shown in the status bar.
"""

from pathlib import Path

from vizcacha.domain.diagnostics import Diagnostic
from vizcacha.i18n import N_, _
from vizcacha.ui.editor import CodeEditor
from vizcacha.ui.features.language.completion import language_server_provider
from vizcacha.ui.features.language.editor_session import EditorSession
from vizcacha.ui.features.language.interactions import EditorInteractions
from vizcacha.ui.features.language.locations import same_file
from vizcacha.ui.features.language.outline_view import OutlineView
from vizcacha.ui.workbench import Workbench

NOT_FOUND_NOTICE = N_("gopls not found: using basic completion")
CRASHED_NOTICE = N_("gopls stopped: using basic completion")
NOTICES = {"not_found": NOT_FOUND_NOTICE, "crashed": CRASHED_NOTICE}
NOTICE_TIMEOUT_MS = 10000


class LanguageFeature:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self.editor = workbench.editor
        self.server = workbench.services.language_server
        self.outline = OutlineView()
        self.sessions: list[EditorSession] = []
        self._notice_shown = False

    def register(self) -> None:
        self.workbench.add_panel("outline", _("Outline"), self.outline, "left")
        self.outline.symbol_activated.connect(self.workbench.events.navigate_to.emit)
        if self.server is None:
            self.show_notice("not_found")
            return
        self.server.diagnostics_published.connect(self._on_diagnostics)
        self.server.server_unavailable.connect(self.show_notice)
        self.editor.editor_created.connect(self.attach)
        self.editor.active_file_changed.connect(self._on_active_file_changed)
        self.editor.file_saved.connect(lambda _path: self._sync_paths())
        self.editor.tabCloseRequested.connect(lambda _index: self._release_closed_editors())
        self.workbench.add_close_guard(self._shutdown)
        for editor in list(self.editor.editors()):
            self.attach(editor)

    # --- editors ----------------------------------------------------------------
    def attach(self, editor: CodeEditor) -> None:
        session = EditorSession(editor, self.server)
        editor.completion_provider = language_server_provider(session, editor.completion_provider)
        EditorInteractions(session, self.workbench.events.navigate_to.emit)
        session.analyzed.connect(self._on_analyzed)
        self.sessions.append(session)
        session.sync_path()

    def session_for(self, editor: CodeEditor | None) -> EditorSession | None:
        return next((s for s in self.sessions if s.editor is editor), None)

    def _sync_paths(self) -> None:
        for session in self.sessions:
            session.sync_path()

    def _release_closed_editors(self) -> None:
        open_editors = list(self.editor.editors())
        for session in [s for s in self.sessions if s.editor not in open_editors]:
            session.release()
            self.sessions.remove(session)
        self.refresh_outline()

    def _on_active_file_changed(self, _path: str) -> None:
        self._sync_paths()
        self.refresh_outline()

    # --- results ----------------------------------------------------------------
    def _on_diagnostics(self, path: Path, diagnostics: list[Diagnostic]) -> None:
        for session in self.sessions:
            if same_file(session.path, path):
                session.show_diagnostics(diagnostics)
        self.workbench.events.diagnostics_changed.emit(path, diagnostics)

    def _on_analyzed(self, editor: CodeEditor) -> None:
        if editor is self.editor.current_editor():
            self.refresh_outline()

    def refresh_outline(self) -> None:
        session = self.session_for(self.editor.current_editor())
        if session is None or session.path is None or not session.is_analyzed:
            self.outline.clear()
            return
        self.outline.show_symbols(self.server.document_symbols(session.path))

    # --- lifecycle --------------------------------------------------------------
    def show_notice(self, reason: str = "not_found") -> None:
        if self._notice_shown:
            return
        self._notice_shown = True
        message = _(NOTICES.get(reason, NOT_FOUND_NOTICE))
        self.workbench.show_status_message(message, NOTICE_TIMEOUT_MS)

    def _shutdown(self) -> bool:
        self.server.shutdown()
        for session in self.sessions:
            session.forget()
        return True


def register(workbench: Workbench) -> None:
    LanguageFeature(workbench).register()
