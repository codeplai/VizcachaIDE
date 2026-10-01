"""Warnings reported by ``go vet``."""

from vizcacha.i18n import N_
from vizcacha.infrastructure.error_catalog.catalog_entry import catalog_entry

VET_WARNINGS = (
    catalog_entry(
        "V-PRINTF-ARGS",
        (
            r"(?P<function>[\w.]+) format %\S+ reads arg #\d+, but call has \d+ args?",
            r"(?P<function>[\w.]+) call needs \d+ args? but has \d+ args?",
            r"(?P<function>[\w.]+) call has arguments but no formatting directives",
            r"(?P<function>[\w.]+) format %\S+ has arg .+ of wrong type",
        ),
        N_("The format of {function} does not match its values"),
        N_(
            "The format text given to {function} does not match the values after it. Each "
            "verb such as %d, %s or %v needs exactly one value, of a suitable type."
        ),
        N_(
            "Add or remove values so there is one per verb, or use fmt.Println if you do not "
            "need formatting."
        ),
    ),
    catalog_entry(
        "V-UNREACHABLE",
        (r"^unreachable code",),
        N_("This code can never run"),
        N_(
            "These statements come after a return, panic, break or infinite loop, so the "
            "program can never reach them."
        ),
        N_("Delete the unreachable lines, or move them before the return."),
    ),
)
