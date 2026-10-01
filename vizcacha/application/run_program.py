"""Use case: decide how to run the active file (single file or Go module package)."""

import re
import shlex
from pathlib import Path

from vizcacha.application.errors import ProgramArgumentsError
from vizcacha.domain.project import GoModule, RunConfiguration, RunTarget
from vizcacha.i18n import _

GO_MOD_FILE = "go.mod"
_MODULE_DIRECTIVE = re.compile(r'^\s*module\s+"?([^"\s]+)"?', re.MULTILINE)

__all__ = ["ProgramArgumentsError", "configuration_for_file", "find_go_module"]


def parse_module_path(go_mod_text: str) -> str:
    """Module path declared in a go.mod file ("" when there is no ``module`` line)."""
    match = _MODULE_DIRECTIVE.search(go_mod_text)
    return match.group(1) if match else ""


def find_go_module(start_directory: Path) -> GoModule | None:
    """The nearest go.mod in ``start_directory`` or any of its parents."""
    for directory in (start_directory, *start_directory.parents):
        go_mod = directory / GO_MOD_FILE
        if not go_mod.is_file():
            continue
        try:
            text = go_mod.read_text(encoding="utf-8")
        except OSError:
            text = ""
        return GoModule(root=directory, module_path=parse_module_path(text))
    return None


def split_program_arguments(text: str) -> tuple[str, ...]:
    """Split like a shell: spaces separate, quotes group. Backslashes are kept
    literally so Windows paths work."""
    lexer = shlex.shlex(text, posix=True)
    lexer.whitespace_split = True
    lexer.escape = ""
    try:
        return tuple(lexer)
    except ValueError as error:
        raise ProgramArgumentsError(_("The program arguments have an unclosed quote.")) from error


def configuration_for_file(path: Path, program_args: tuple[str, ...] = ()) -> RunConfiguration:
    """``go run .`` in the file's folder inside a module, ``go run file.go`` otherwise."""
    path = Path(path).absolute()
    module = find_go_module(path.parent)
    if module is None:
        return RunConfiguration.for_file(path, program_args)
    return RunConfiguration(
        target=path,
        working_dir=path.parent,
        mode=RunTarget.PACKAGE,
        program_args=program_args,
        module=module,
    )
