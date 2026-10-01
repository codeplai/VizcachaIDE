"""Offline completion from static Go language data (fallback when gopls is unavailable).

Mostly data tables, which is why this file is longer than the usual 200-line limit.
Track C (gopls) keeps this as the fallback provider; it also suggests the names
declared in the file itself (see ``local_symbols``).
"""

from vizcacha.domain.completion import CompletionItem, CompletionKind
from vizcacha.infrastructure.static_completion.local_symbols import local_completions

MAX_SUGGESTIONS = 20


class GoAnalyzer:
    """Suggests keywords, builtin types/functions and common stdlib functions."""

    def __init__(self):
        # Built-in Go keywords
        self.keywords = [
            "break",
            "case",
            "chan",
            "const",
            "continue",
            "default",
            "defer",
            "else",
            "fallthrough",
            "for",
            "func",
            "go",
            "goto",
            "if",
            "import",
            "interface",
            "map",
            "package",
            "range",
            "return",
            "select",
            "struct",
            "switch",
            "type",
            "var",
        ]

        # Built-in types
        self.types = [
            "bool",
            "byte",
            "complex64",
            "complex128",
            "error",
            "float32",
            "float64",
            "int",
            "int8",
            "int16",
            "int32",
            "int64",
            "rune",
            "string",
            "uint",
            "uint8",
            "uint16",
            "uint32",
            "uint64",
            "uintptr",
        ]

        # Built-in functions
        self.builtins = {
            "append": {
                "signature": "(slice, elements...)",
                "doc": "Appends elements to the end of a slice and returns the updated slice.",
            },
            "cap": {
                "signature": "(v Type)",
                "doc": "Returns the capacity of v, according to its type.",
            },
            "close": {
                "signature": "(c chan<- Type)",
                "doc": "Closes a channel, indicating that no more values will be sent.",
            },
            "complex": {
                "signature": "(r, i FloatType)",
                "doc": "Constructs a complex value from floating-point real and imaginary parts.",
            },
            "copy": {
                "signature": "(dst, src []Type)",
                "doc": "Copies elements from source slice to destination slice.",
            },
            "delete": {
                "signature": "(m map[Type]Type1, key Type)",
                "doc": "Deletes the element with the specified key from the map.",
            },
            "imag": {
                "signature": "(c ComplexType)",
                "doc": "Returns the imaginary part of the complex number.",
            },
            "len": {
                "signature": "(v Type)",
                "doc": "Returns the length of v, according to its type.",
            },
            "make": {
                "signature": "(Type, size ...IntegerType)",
                "doc": "Allocates and initializes an object of type slice, map, or channel.",
            },
            "new": {
                "signature": "(Type)",
                "doc": "Allocates memory for a variable of the given type and returns a pointer to it.",
            },
            "panic": {
                "signature": "(v interface{})",
                "doc": "Stops normal execution of the current goroutine.",
            },
            "print": {"signature": "(args ...Type)", "doc": "Prints arguments to standard error."},
            "println": {
                "signature": "(args ...Type)",
                "doc": "Prints arguments to standard error with spaces and newline.",
            },
            "real": {
                "signature": "(c ComplexType)",
                "doc": "Returns the real part of the complex number.",
            },
            "recover": {"signature": "()", "doc": "Regains control of a panicking goroutine."},
        }

        # Common standard library packages and their functions
        self.stdlib = {
            "fmt": {
                "Print": "Formats using the default formats and writes to standard output.",
                "Println": "Formats using the default formats, adds newline, writes to standard output.",
                "Printf": "Formats according to a format specifier and writes to standard output.",
                "Sprint": "Formats using the default formats and returns the resulting string.",
                "Sprintf": "Formats according to a format specifier and returns the resulting string.",
                "Scan": "Scans text read from standard input.",
                "Scanf": "Scans text read from standard input with format.",
                "Scanln": "Scans text read from standard input until newline.",
            },
            "strings": {
                "Contains": "Reports whether substr is within s.",
                "HasPrefix": "Tests whether string begins with prefix.",
                "HasSuffix": "Tests whether string ends with suffix.",
                "Index": "Returns the index of the first instance of substr in s.",
                "Join": "Concatenates the elements of a to create a single string.",
                "Replace": "Returns a copy of s with replacements.",
                "Split": "Slices s into all substrings separated by sep.",
                "ToLower": "Returns s with all Unicode letters mapped to lowercase.",
                "ToUpper": "Returns s with all Unicode letters mapped to uppercase.",
                "Trim": "Returns a slice of s with leading and trailing cutset removed.",
            },
            "strconv": {
                "Atoi": "Converts string to int.",
                "Itoa": "Converts int to string.",
                "ParseBool": "Parses a string to boolean.",
                "ParseFloat": "Parses a string to float64.",
                "ParseInt": "Parses a string to int64.",
                "FormatBool": "Formats a boolean to string.",
                "FormatFloat": "Formats a float to string.",
                "FormatInt": "Formats an int to string.",
            },
            "os": {
                "Create": "Creates or truncates the named file.",
                "Open": "Opens the named file for reading.",
                "Exit": "Causes the program to exit with the given status code.",
                "Getenv": "Retrieves the value of the environment variable.",
                "Setenv": "Sets the value of the environment variable.",
                "Remove": "Removes the named file or directory.",
                "Mkdir": "Creates a new directory with the specified name.",
            },
            "io": {
                "Copy": "Copies from src to dst until EOF.",
                "ReadAll": "Reads from r until EOF and returns the data.",
                "WriteString": "Writes the contents of the string s to w.",
            },
            "time": {
                "Now": "Returns the current local time.",
                "Sleep": "Pauses the current goroutine for the specified duration.",
                "Since": "Returns the time elapsed since t.",
                "Parse": "Parses a formatted string and returns the time value.",
                "Format": "Formats time according to a layout.",
            },
            "math": {
                "Abs": "Returns the absolute value of x.",
                "Ceil": "Returns the least integer value greater than or equal to x.",
                "Floor": "Returns the greatest integer value less than or equal to x.",
                "Max": "Returns the larger of x or y.",
                "Min": "Returns the smaller of x or y.",
                "Pow": "Returns x**y, the base-x exponential of y.",
                "Sqrt": "Returns the square root of x.",
                "Round": "Returns the nearest integer, rounding half away from zero.",
            },
        }

    def get_completions(self, code: str, cursor_position: int) -> list[CompletionItem]:
        """Suggestions for the word being typed at ``cursor_position``."""
        prefix = _word_before(code, cursor_position).lower()
        package_name = _package_before_dot(code, cursor_position - len(prefix))
        if package_name is not None:
            items = self._package_members(package_name, prefix)
        else:
            items = self._global_names(prefix)
            known = {item.label for item in items}
            items += local_completions(code, prefix, known)
        items.sort(key=lambda item: item.label.lower())
        return items[:MAX_SUGGESTIONS]

    def _package_members(self, package_name: str, prefix: str) -> list[CompletionItem]:
        members = self.stdlib.get(package_name, {})
        return [
            CompletionItem(name, CompletionKind.FUNCTION, f"{package_name}.{name}(...)", doc)
            for name, doc in members.items()
            if name.lower().startswith(prefix)
        ]

    def _global_names(self, prefix: str) -> list[CompletionItem]:
        items = [
            CompletionItem(word, CompletionKind.KEYWORD, documentation=f"Go keyword: {word}")
            for word in self.keywords
            if word.startswith(prefix)
        ]
        items += [
            CompletionItem(name, CompletionKind.TYPE, documentation=f"Built-in type: {name}")
            for name in self.types
            if name.startswith(prefix)
        ]
        items += [
            CompletionItem(name, CompletionKind.FUNCTION, info["signature"], info["doc"])
            for name, info in self.builtins.items()
            if name.startswith(prefix)
        ]
        items += [
            CompletionItem(name, CompletionKind.PACKAGE, documentation=f"Package: {name}")
            for name in self.stdlib
            if name.startswith(prefix)
        ]
        return items


def _word_before(code: str, position: int) -> str:
    start = position
    while start > 0 and (code[start - 1].isalnum() or code[start - 1] == "_"):
        start -= 1
    return code[start:position]


def _package_before_dot(code: str, position: int) -> str | None:
    """Return ``fmt`` when the text before ``position`` is ``fmt.``; otherwise None."""
    if position <= 0 or code[position - 1] != ".":
        return None
    package = _word_before(code, position - 1)
    return package or None
