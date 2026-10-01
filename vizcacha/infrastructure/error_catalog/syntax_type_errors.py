"""Compiler errors about syntax, types and function calls.

Order matters: E-MISSING-BRACE comes before the more general E-SYNTAX-UNEXPECTED.
"""

from vizcacha.i18n import N_
from vizcacha.infrastructure.error_catalog.catalog_entry import catalog_entry

SYNTAX_TYPE_ERRORS = (
    catalog_entry(
        "E-MISSING-BRACE",
        (
            r"syntax error: unexpected EOF, expect(?:ed|ing) \}",
            r"expected '\}', found 'EOF'",
            r"syntax error: unexpected .+, expect(?:ed|ing) \{ after",
        ),
        N_("A curly brace {{ }} is missing"),
        N_(
            "Every {{ needs a matching }}. Go reached the end of the file, or of a block, while "
            "it was still waiting for one. Go also needs the opening {{ on the same line as "
            "func, if, for or else."
        ),
        N_(
            "Count your {{ and }} and add the missing one. Keep each opening {{ at the end of "
            "the line that starts the block."
        ),
    ),
    catalog_entry(
        "E-SYNTAX-UNEXPECTED",
        (r"syntax error: unexpected (?P<token>.+?)(?:,|;| in | at | after |$)",),
        N_("Syntax error: unexpected {token}"),
        N_(
            "Go could not understand this line. It found {token} in a place where Go's "
            "grammar does not allow it. The real mistake is often just before that point."
        ),
        N_(
            "Look right before the marked position for a missing comma, parenthesis, quote "
            "or operator."
        ),
    ),
    catalog_entry(
        "E-TYPE-MISMATCH",
        (r"cannot use (?P<value>.+?) \([^()]*\) as (?P<type>.+?) value in",),
        N_("Wrong type: a {type} was expected"),
        N_(
            "The value {value} does not have the type {type}, which is what Go expects here. "
            "Go never converts between types such as int and string automatically."
        ),
        N_(
            "Use a value of type {type}, or convert it explicitly, for example with int(x), "
            "float64(x) or strconv.Itoa(n)."
        ),
    ),
    catalog_entry(
        "E-MISMATCHED-TYPES-OP",
        (
            r"invalid operation: (?P<expression>.+) "
            r"\(mismatched types (?P<left_type>\S+) and (?P<right_type>\S+)\)",
        ),
        N_("Cannot combine {left_type} and {right_type}"),
        N_(
            "In {expression} the two sides have different types ({left_type} and "
            "{right_type}). Operators such as +, - or == only work between values of the "
            "same type."
        ),
        N_(
            "Convert one side so both have the same type, for example strconv.Itoa(n) to "
            "turn an int into a string, or float64(n) to mix it with decimals."
        ),
    ),
    catalog_entry(
        "E-ASSIGN-MISMATCH",
        (
            r"assignment mismatch: (?P<left>\d+) variables? but "
            r"(?:\S+ returns )?(?P<right>\d+) values?",
        ),
        N_("The number of variables and values does not match"),
        N_(
            "The left side of = or := has {left} variable(s), but the right side gives "
            "{right} value(s). Both numbers must be equal. Many functions, such as "
            "strconv.Atoi, return two values: a result and an error."
        ),
        N_(
            "Add or remove variables until both sides match, for example: "
            "number, err := strconv.Atoi(text). Use _ for a value you do not need."
        ),
    ),
    catalog_entry(
        "E-NOT-ENOUGH-ARGS",
        (r"not enough arguments in call to (?P<function>\S+)",),
        N_("Not enough arguments for {function}"),
        N_(
            "The function {function} needs more values than you passed. In the original "
            'message, "have" lists what you passed and "want" lists what it expects.'
        ),
        N_("Pass one value for each parameter, in the same order and with the right types."),
    ),
    catalog_entry(
        "E-TOO-MANY-ARGS",
        (r"too many arguments in call to (?P<function>\S+)",),
        N_("Too many arguments for {function}"),
        N_(
            "The function {function} receives fewer values than you passed. In the original "
            'message, "have" lists what you passed and "want" lists what it expects.'
        ),
        N_("Remove the extra values so there is exactly one for each parameter."),
    ),
)
