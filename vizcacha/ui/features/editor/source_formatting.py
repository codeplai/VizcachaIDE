"""gofmt: on save (optional, default on) and on demand ("Format Code")."""

from pathlib import Path

from vizcacha.application.errors import GoFormatError, GoToolchainNotFoundError
from vizcacha.i18n import _
from vizcacha.ui.editor.code_editor import CodeEditor
from vizcacha.ui.editor.text_replacement import replace_text_keeping_view
from vizcacha.ui.features.editor.editor_preferences import FORMAT_ON_SAVE_KEY, setting
from vizcacha.ui.workbench import Workbench

GO_SUFFIX = ".go"
STATUS_TIMEOUT_MS = 10000


class SourceFormatter:
    def __init__(self, workbench: Workbench) -> None:
        self.workbench = workbench
        self._missing_tool_reported = False

    def format_before_save(self, editor: CodeEditor, path: Path) -> None:
        """TabbedEditor save hook. Never blocks the save: errors are only reported."""
        if Path(path).suffix != GO_SUFFIX:
            return
        if not setting(self.workbench.services.settings, FORMAT_ON_SAVE_KEY):
            return
        self.format_editor(editor, report_missing_once=True)

    def format_current(self) -> None:
        editor = self.workbench.editor.current_editor()
        if editor is not None and self.format_editor(editor, report_missing_once=False):
            self.workbench.show_status_message(_("Code formatted."), STATUS_TIMEOUT_MS)

    def format_editor(self, editor: CodeEditor, report_missing_once: bool) -> bool:
        """Format with gofmt. Returns True when the text was formatted successfully."""
        try:
            formatted = self.workbench.services.toolchain.format_source(editor.toPlainText())
        except GoFormatError as error:
            self._report_syntax_error(str(error))
            return False
        except GoToolchainNotFoundError as error:
            self._report_missing_tool(str(error), report_missing_once)
            return False
        replace_text_keeping_view(editor, formatted)
        return True

    def _report_syntax_error(self, details: str) -> None:
        first_line = details.splitlines()[0] if details else ""
        message = _("The code was not formatted (gofmt): {details}").format(details=first_line)
        self.workbench.show_status_message(message, STATUS_TIMEOUT_MS)

    def _report_missing_tool(self, details: str, once: bool) -> None:
        if once and self._missing_tool_reported:
            return
        self._missing_tool_reported = True
        message = _("gofmt is not available, so the code is saved without formatting. {details}")
        text = message.format(details=details)
        self.workbench.show_status_message(text, STATUS_TIMEOUT_MS)
        self.workbench.console.append_error(text + "\n")
