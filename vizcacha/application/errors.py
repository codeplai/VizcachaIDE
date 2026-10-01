"""Typed errors raised by adapters. Catch these, never a bare ``Exception``."""


class VizcachaError(Exception):
    """Base class for every expected, user-reportable failure."""


class GoToolchainNotFoundError(VizcachaError):
    """The ``go`` executable (or a companion tool like ``gofmt``) could not be found."""


class GoFormatError(VizcachaError):
    """``gofmt`` rejected the source (usually a syntax error). Message = gofmt stderr."""


class DebugAdapterNotFoundError(VizcachaError):
    """Delve (``dlv``) could not be found."""


class DebugAdapterError(VizcachaError):
    """Delve reported an error or the DAP connection failed."""


class ProgramArgumentsError(VizcachaError):
    """The "Program arguments" text cannot be split (e.g. an unclosed quote)."""


class GoCommandArgumentError(VizcachaError):
    """A module path or package typed by the user cannot be passed to ``go``."""


class LanguageServerError(VizcachaError):
    """gopls could not be started or answered with an error."""
