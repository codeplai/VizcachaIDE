"""The offline fallback also suggests names declared in the file being edited."""

from vizcacha.domain.completion import CompletionKind
from vizcacha.infrastructure.static_completion import GoAnalyzer
from vizcacha.infrastructure.static_completion.local_symbols import declared_names

SOURCE = """package main

import "fmt"

type Point struct{ X, Y int }

const limit = 10

var (
    total int
    name, _ = "a", 1
)

func (p Point) Norm() int { return p.X }

func sumAll(values []int) int {
    count := 0
    first, second := 1, 2
    for index, value := range values {
        count += value + index
    }
    return count + first + second
}

func main() {
    fmt.Println(sumAll(nil), total, limit, name)
}
"""


def test_declared_names_cover_func_type_var_const_and_short_declarations():
    names = declared_names(SOURCE)

    assert names["sumAll"] == CompletionKind.FUNCTION
    assert names["Norm"] == CompletionKind.FUNCTION
    assert names["Point"] == CompletionKind.TYPE
    assert names["limit"] == CompletionKind.CONSTANT
    for variable in ("total", "name", "count", "first", "second", "index", "value"):
        assert names[variable] == CompletionKind.VARIABLE, variable
    assert "_" not in names


def test_analyzer_suggests_local_names_without_duplicating_builtins():
    analyzer = GoAnalyzer()
    code = SOURCE + "\nfunc later() { su"

    labels = [item.label for item in analyzer.get_completions(code, len(code))]

    assert "sumAll" in labels
    assert labels.count("sumAll") == 1


def test_package_members_are_unchanged():
    code = "fmt.Pr"

    labels = {item.label for item in GoAnalyzer().get_completions(code, len(code))}

    assert {"Print", "Printf", "Println"} <= labels
