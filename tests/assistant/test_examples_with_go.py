"""Regenerates real Go output for some examples and checks the ids (Go version drift)."""

import shutil
import subprocess
from pathlib import Path

import pytest

from vizcacha.infrastructure.error_catalog import GoErrorExplainer

EXAMPLES = Path(__file__).resolve().parents[2] / "examples" / "errors"
SAMPLE_IDS = [
    "E-UNUSED-VAR",
    "E-UNDEFINED",
    "E-TYPE-MISMATCH",
    "E-MISSING-BRACE",
    "P-INDEX-RANGE",
    "P-NIL-MAP",
    "P-DEADLOCK",
    "V-PRINTF-ARGS",
]


@pytest.mark.requires_go
@pytest.mark.parametrize("error_id", SAMPLE_IDS)
def test_real_go_output_produces_the_expected_id(error_id, tmp_path: Path):
    source = tmp_path / f"{error_id}.go"
    shutil.copy(EXAMPLES / error_id / source.name, source)
    command = ["go", "vet" if error_id.startswith("V-") else "run", source.name]

    result = subprocess.run(
        command, cwd=tmp_path, capture_output=True, text=True, encoding="utf-8", timeout=180
    )
    diagnostics = GoErrorExplainer().parse(result.stdout + result.stderr, tmp_path)

    assert result.returncode != 0
    assert [diagnostic.code for diagnostic in diagnostics] == [error_id]
    assert diagnostics[0].location.file.resolve() == source.resolve()
