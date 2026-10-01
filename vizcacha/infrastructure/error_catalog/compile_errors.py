"""Compiler errors (``go build`` / ``go run``) about names, packages and program structure.

Order matters: the first entry whose pattern matches wins, so E-UNEXPORTED comes
before the more general E-UNDEFINED.
"""

from vizcacha.i18n import N_
from vizcacha.infrastructure.error_catalog.catalog_entry import catalog_entry

COMPILE_ERRORS = (
    catalog_entry(
        "E-UNUSED-VAR",
        (r"declared and not used: (?P<name>\w+)", r"^(?P<name>\w+) declared (?:and|but) not used"),
        N_('Variable "{name}" is never used'),
        N_(
            'You created the variable "{name}" but never read it. Go refuses to compile code '
            "with unused variables, because they are usually a mistake or leftover code."
        ),
        N_(
            'Use "{name}" somewhere (for example, print it), delete it, or replace it '
            "with _ if you must receive a value you do not need."
        ),
    ),
    catalog_entry(
        "E-UNUSED-IMPORT",
        (r"\"(?P<package>[^\"]+)\" imported (?:as \w+ )?and not used",),
        N_('Package "{package}" is imported but never used'),
        N_(
            'The file imports the package "{package}", but no code uses it. Go does not '
            "allow unused imports, to keep programs clean and fast to compile."
        ),
        N_('Delete the line that imports "{package}", or use something from it.'),
    ),
    catalog_entry(
        "E-MISSING-RETURN",
        (r"^missing return",),
        N_("The function can end without returning a value"),
        N_(
            "This function says it returns a value, but there is a way to reach its closing "
            "brace without a return statement. For example, an if returns a value but the "
            "case where the condition is false does not."
        ),
        N_(
            "Add a return statement at the end of the function, or an else branch that "
            "also returns a value."
        ),
    ),
    catalog_entry(
        "E-UNEXPORTED",
        (
            r"undefined: (?P<package>\w+)\.(?P<name>\w+) \(but have \w+\)",
            r"name (?P<name>\w+) not exported by package (?P<package>\w+)",
            r"cannot refer to unexported (?:name|field|method) (?P<package>\w+)\.(?P<name>\w+)",
        ),
        N_('"{name}" is private to package {package}'),
        N_(
            "In Go, only names that start with a capital letter can be used from another "
            'package. "{name}" starts with a lowercase letter, so package {package} keeps '
            "it private, or the real name has different capital letters."
        ),
        N_(
            "Write the name exactly as the package documents it, starting with a capital "
            "letter (for example strings.ToUpper, not strings.toUpper)."
        ),
    ),
    catalog_entry(
        "E-UNDEFINED",
        (r"undefined: (?P<name>[\w.]+)",),
        N_("Unknown name: {name}"),
        N_(
            'Go does not know anything called "{name}". It was never declared, it is '
            "declared in a place this code cannot see (such as inside another function or "
            "block), or it is misspelled."
        ),
        N_(
            "Check the spelling and the capital letters, declare the variable before you use "
            "it (for example with :=), or import the package it belongs to."
        ),
    ),
    catalog_entry(
        "E-NO-NEW-VARS",
        (r"no new variables on left side of :=",),
        N_("The variable already exists: use = instead of :="),
        N_(
            ":= creates new variables. Every variable on its left side already exists in this "
            "block, so there is nothing new to create."
        ),
        N_("Use = to change the value of an existing variable. Use := only the first time."),
    ),
    catalog_entry(
        "E-NO-MAIN",
        (r"function main is undeclared in the main package",),
        N_("The program has no main function"),
        N_(
            "A Go program starts running in the function main of package main. This package "
            "is called main, but it does not contain a function named main."
        ),
        N_("Add a function with exactly this header: func main() {{ ... }}"),
    ),
    catalog_entry(
        "E-PACKAGE-NOT-MAIN",
        (r"is not a main package", r"cannot run non-main package"),
        N_("This file is not a runnable program"),
        N_(
            "Only code in package main can be run. The first line of this file declares a "
            "different package, which Go treats as a library."
        ),
        N_("Change the first line to: package main (and make sure there is a func main)."),
    ),
    catalog_entry(
        "E-IMPORT-NOT-FOUND",
        (
            r"package (?P<package>\S+) is not in (?:std|GOROOT)",
            r"no required module provides package (?P<package>[^\s;:]+)",
            r"cannot find package \"(?P<package>[^\"]+)\"",
        ),
        N_('Package "{package}" was not found'),
        N_(
            'Go cannot find the package "{package}". Standard packages have fixed names '
            "(fmt, math, strings, os...), and packages from the internet must be added to "
            "your module before you can import them."
        ),
        N_(
            "Check the spelling of the import path. For an external package, run "
            "go get {package} in a folder that has a go.mod file."
        ),
    ),
)
