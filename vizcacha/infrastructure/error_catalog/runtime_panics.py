"""Run-time panics and fatal errors printed while the program runs."""

from vizcacha.i18n import N_
from vizcacha.infrastructure.error_catalog.catalog_entry import catalog_entry

RUNTIME_PANICS = (
    catalog_entry(
        "P-INDEX-RANGE",
        (r"index out of range \[(?P<index>-?\d+)\] with length (?P<length>\d+)",),
        N_("Index {index} is out of range"),
        N_(
            "The program tried to use position {index}, but the slice, array or string only "
            "has {length} element(s). Valid positions go from 0 to {length} - 1."
        ),
        N_(
            "Check the index before using it (for example: if i < len(items)) and remember "
            "that counting starts at 0."
        ),
    ),
    catalog_entry(
        "P-NIL-MAP",
        (r"assignment to entry in nil map",),
        N_("Writing to a map that was never created"),
        N_(
            "The map variable was declared but never created, so it is nil. You can read from "
            "a nil map, but writing to it stops the program."
        ),
        N_(
            "Create the map before writing to it: ages := make(map[string]int) or "
            "ages := map[string]int{{}}."
        ),
    ),
    catalog_entry(
        "P-NIL-POINTER",
        (r"invalid memory address or nil pointer dereference",),
        N_("Using a pointer that is nil"),
        N_(
            "The program read or wrote through a pointer (or an interface, map or slice "
            "element holding one) whose value is nil, so it points to nothing."
        ),
        N_(
            "Make sure the pointer gets a value before you use it, for example "
            "p := &Person{{}} or p := new(Person), or check if p != nil first."
        ),
    ),
    catalog_entry(
        "P-DEADLOCK",
        (r"all goroutines are asleep - deadlock!",),
        N_("Deadlock: every goroutine is waiting forever"),
        N_(
            "All goroutines are blocked waiting for each other, so nothing can continue. "
            "Usually a channel is sent to with nobody receiving, or received from with "
            "nobody sending, or a WaitGroup or Mutex is never released."
        ),
        N_(
            "Make sure another goroutine receives what you send (go func() {{ ... }}()), "
            "use a buffered channel, close the channel when you are done, or call wg.Done()."
        ),
    ),
    catalog_entry(
        "P-DIVIDE-ZERO",
        (r"integer divide by zero",),
        N_("Division by zero"),
        N_(
            "The program divided an integer by zero (with / or %). That has no result, so "
            "Go stops the program."
        ),
        N_(
            "Check that the divisor is not 0 before dividing, for example: "
            "if people != 0 {{ ... }}."
        ),
    ),
    catalog_entry(
        "P-SLICE-BOUNDS",
        (r"slice bounds out of range \[(?P<bounds>[^\]]*)\]",),
        N_("Slice limits [{bounds}] are out of range"),
        N_(
            "The program tried to take the part [{bounds}] of a slice or string, but those "
            "limits go past its end, or the start is bigger than the end."
        ),
        N_(
            "Make sure that 0 <= start <= end <= len(s) before slicing, for example with "
            "min(end, len(s))."
        ),
    ),
    catalog_entry(
        "P-TYPE-ASSERTION",
        (r"interface conversion: (?P<interface>.+?) is (?P<actual>.+?), not (?P<type>.+)$",),
        N_("Type assertion failed: {actual} is not {type}"),
        N_(
            "The program asserted that a value holds a {type}, but it actually holds a "
            "{actual}. A failed assertion with a single result stops the program."
        ),
        N_(
            "Use the two-result form to check safely: n, ok := value.({type}). "
            "Or use a type switch."
        ),
    ),
)
