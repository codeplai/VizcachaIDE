from pathlib import Path

from vizcacha.application.explain_error import ExplainError
from vizcacha.domain.diagnostics import Diagnostic, Severity
from vizcacha.infrastructure.error_catalog import GoErrorExplainer


def test_from_output_pairs_each_diagnostic_with_its_explanation(tmp_path: Path):
    raw = "./main.go:5:2: declared and not used: x\n./main.go:6:2: a brand new message\n"

    results = ExplainError(GoErrorExplainer()).from_output(raw, tmp_path)

    (known, explanation), (unknown, nothing) = results
    assert explanation.explanation_id == "E-UNUSED-VAR"
    assert known.location.file == tmp_path / "main.go"
    assert unknown.message == "a brand new message"
    assert nothing is None


def test_from_diagnostics_skips_hints_and_duplicates():
    error = Diagnostic(None, Severity.ERROR, "missing return", "missing return")
    hint = Diagnostic(None, Severity.HINT, "could simplify", "could simplify")

    results = ExplainError(GoErrorExplainer()).from_diagnostics([error, hint, error])

    assert [diagnostic for diagnostic, _explanation in results] == [error]
