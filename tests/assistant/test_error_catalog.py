"""Every catalog id has an example program; recorded Go output maps to that id."""

import string
from pathlib import Path

import pytest

from vizcacha.application.ports import ErrorExplainerPort
from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.infrastructure.error_catalog import DEFAULT_ENTRIES, ErrorCatalog, GoErrorExplainer

REPO = Path(__file__).resolve().parents[2]
EXAMPLES = REPO / "examples" / "errors"
FIXTURES = Path(__file__).resolve().parent / "fixtures" / "go_output"
REQUIRED_IDS = {
    "E-UNUSED-VAR", "E-UNUSED-IMPORT", "E-MISSING-RETURN", "E-UNDEFINED", "E-TYPE-MISMATCH",
    "E-NO-MAIN", "E-PACKAGE-NOT-MAIN", "E-SYNTAX-UNEXPECTED", "E-MISSING-BRACE",
    "E-ASSIGN-MISMATCH", "E-NOT-ENOUGH-ARGS", "E-TOO-MANY-ARGS", "E-NO-NEW-VARS",
    "E-UNEXPORTED", "E-IMPORT-NOT-FOUND", "E-MISMATCHED-TYPES-OP", "P-INDEX-RANGE",
    "P-NIL-MAP", "P-NIL-POINTER", "P-DEADLOCK", "P-DIVIDE-ZERO", "P-SLICE-BOUNDS",
    "P-TYPE-ASSERTION", "V-PRINTF-ARGS", "V-UNREACHABLE",
}  # fmt: skip
CATALOG_IDS = [entry.explanation_id for entry in DEFAULT_ENTRIES]


def _fields(text: str) -> set[str]:
    return {name for _literal, name, _spec, _conv in string.Formatter().parse(text) if name}


def test_explainer_follows_contract():
    assert isinstance(GoErrorExplainer(), ErrorExplainerPort)


def test_catalog_has_the_required_stable_ids():
    assert len(CATALOG_IDS) == len(set(CATALOG_IDS))
    assert not REQUIRED_IDS - set(CATALOG_IDS)


@pytest.mark.parametrize("entry", DEFAULT_ENTRIES, ids=CATALOG_IDS)
def test_placeholders_are_named_groups_of_every_pattern(entry):
    used = set().union(*(_fields(text) for text in (entry.title, entry.body, entry.fix_hint)))
    for pattern in entry.patterns:
        assert used <= set(pattern.groupindex), pattern.pattern
    values = {name: "X" for name in used}
    for text in (entry.title, entry.body, entry.fix_hint):
        assert text.format(**values)


@pytest.mark.parametrize("error_id", CATALOG_IDS)
def test_example_output_produces_the_expected_id(error_id, tmp_path: Path):
    assert (EXAMPLES / f"{error_id}.go").is_file()
    raw = (FIXTURES / f"{error_id}.txt").read_text(encoding="utf-8")
    raw = raw.replace("$WORKDIR", tmp_path.as_posix())
    explainer = GoErrorExplainer()

    diagnostics = explainer.parse(raw, tmp_path)

    assert [diagnostic.code for diagnostic in diagnostics] == [error_id]
    explanation = explainer.explain(diagnostics[0])
    assert explanation is not None and explanation.explanation_id == error_id
    location = diagnostics[0].location
    if location is not None:
        assert location.file == tmp_path / f"{error_id}.go"


def test_placeholders_come_from_the_message():
    catalog = ErrorCatalog()

    explanation = catalog.explain("index out of range [7] with length 2")

    assert explanation.placeholders == {"index": "7", "length": "2"}
    assert catalog.identify("something Go never says") == ""
    assert catalog.explain("something Go never says") is None


def test_explain_uses_message_of_gopls_diagnostics():
    diagnostic = Diagnostic(None, Severity.ERROR, "declared and not used: n", "", "compiler")

    explanation = GoErrorExplainer().explain(diagnostic)

    assert explanation.explanation_id == "E-UNUSED-VAR"
    assert explanation.placeholders == {"name": "n"}
