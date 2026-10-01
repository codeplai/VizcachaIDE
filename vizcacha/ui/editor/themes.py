"""Colour themes for the editor and the console, defined as plain data.

Theme names are the values stored in QSettings (``appearance/editor_theme`` and
``appearance/console_theme``); they stay in English. Unknown names fall back to
the first theme of each table.
"""

from dataclasses import dataclass, field


@dataclass(frozen=True)
class SyntaxStyle:
    color: str
    bold: bool = False
    italic: bool = False


@dataclass(frozen=True)
class EditorTheme:
    name: str
    background: str
    text: str
    selection_background: str
    selection_text: str
    gutter_background: str
    gutter_text: str
    current_line: str  # debugger "current line" highlight and gutter marker
    breakpoint: str
    bracket_match: str
    search_match: str
    syntax: dict[str, SyntaxStyle] = field(default_factory=dict)


@dataclass(frozen=True)
class ConsoleTheme:
    name: str
    background: str
    text: str


SYNTAX_CATEGORIES = ("keyword", "type", "builtin", "string", "number", "comment")

LIGHT = EditorTheme(
    name="Light",
    background="#FFFFFF",
    text="#000000",
    selection_background="#ADD6FF",
    selection_text="#000000",
    gutter_background="#F0F0F0",
    gutter_text="#808080",
    current_line="#FFFF99",
    breakpoint="#E51400",
    bracket_match="#C8E6C9",
    search_match="#FFE082",
    syntax={
        "keyword": SyntaxStyle("#0000FF", bold=True),
        "type": SyntaxStyle("#008080", bold=True),
        "builtin": SyntaxStyle("#800080"),
        "string": SyntaxStyle("#008000"),
        "number": SyntaxStyle("#FF6600"),
        "comment": SyntaxStyle("#808080", italic=True),
    },
)

DARK = EditorTheme(
    name="Dark",
    background="#1E1E1E",
    text="#D4D4D4",
    selection_background="#264F78",
    selection_text="#FFFFFF",
    gutter_background="#252526",
    gutter_text="#858585",
    current_line="#4B4B18",
    breakpoint="#F14C4C",
    bracket_match="#3A5A3A",
    search_match="#613214",
    syntax={
        "keyword": SyntaxStyle("#569CD6", bold=True),
        "type": SyntaxStyle("#4EC9B0", bold=True),
        "builtin": SyntaxStyle("#DCDCAA"),
        "string": SyntaxStyle("#CE9178"),
        "number": SyntaxStyle("#B5CEA8"),
        "comment": SyntaxStyle("#6A9955", italic=True),
    },
)

# Solarized palette by Ethan Schoonover (https://ethanschoonover.com/solarized/).
SOLARIZED_LIGHT = EditorTheme(
    name="Solarized Light",
    background="#FDF6E3",
    text="#657B83",
    selection_background="#EEE8D5",
    selection_text="#586E75",
    gutter_background="#EEE8D5",
    gutter_text="#93A1A1",
    current_line="#F5E6B8",
    breakpoint="#DC322F",
    bracket_match="#D9E7C5",
    search_match="#F2D98C",
    syntax={
        "keyword": SyntaxStyle("#859900", bold=True),
        "type": SyntaxStyle("#B58900", bold=True),
        "builtin": SyntaxStyle("#268BD2"),
        "string": SyntaxStyle("#2AA198"),
        "number": SyntaxStyle("#D33682"),
        "comment": SyntaxStyle("#93A1A1", italic=True),
    },
)

SOLARIZED_DARK = EditorTheme(
    name="Solarized Dark",
    background="#002B36",
    text="#839496",
    selection_background="#073642",
    selection_text="#93A1A1",
    gutter_background="#073642",
    gutter_text="#586E75",
    current_line="#3B3A16",
    breakpoint="#DC322F",
    bracket_match="#0E4B3E",
    search_match="#5C4A00",
    syntax={
        "keyword": SyntaxStyle("#859900", bold=True),
        "type": SyntaxStyle("#B58900", bold=True),
        "builtin": SyntaxStyle("#268BD2"),
        "string": SyntaxStyle("#2AA198"),
        "number": SyntaxStyle("#D33682"),
        "comment": SyntaxStyle("#586E75", italic=True),
    },
)

EDITOR_THEMES: dict[str, EditorTheme] = {
    theme.name: theme for theme in (LIGHT, DARK, SOLARIZED_LIGHT, SOLARIZED_DARK)
}

CONSOLE_THEMES: dict[str, ConsoleTheme] = {
    "Dark": ConsoleTheme("Dark", background="#1E1E1E", text="#D4D4D4"),
    "Light": ConsoleTheme("Light", background="#FFFFFF", text="#1E1E1E"),
}


def editor_theme(name: str) -> EditorTheme:
    return EDITOR_THEMES.get(name, LIGHT)


def console_theme(name: str) -> ConsoleTheme:
    return CONSOLE_THEMES.get(name, CONSOLE_THEMES["Dark"])
