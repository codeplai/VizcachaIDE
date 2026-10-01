"""Arguments for ``go mod init``, ``go get`` and ``go mod tidy`` (validated)."""

from vizcacha.application.errors import VizcachaError
from vizcacha.i18n import _


class GoCommandArgumentError(VizcachaError):
    """A module path or package typed by the user cannot be passed to ``go``."""


def _single_word(value: str, empty_message: str) -> str:
    value = value.strip()
    if not value:
        raise GoCommandArgumentError(empty_message)
    if value.startswith("-") or any(character.isspace() for character in value):
        raise GoCommandArgumentError(_("'{value}' is not a valid name.").format(value=value))
    return value


def mod_init_arguments(module_path: str) -> list[str]:
    name = _single_word(
        module_path, _("Please enter a module path, for example example.com/hello.")
    )
    return ["mod", "init", name]


def mod_get_arguments(package: str) -> list[str]:
    name = _single_word(package, _("Please enter a package, for example github.com/google/uuid."))
    return ["get", name]


def mod_tidy_arguments() -> list[str]:
    return ["mod", "tidy"]
