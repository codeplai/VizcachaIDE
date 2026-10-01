from pathlib import Path

from vizcacha.infrastructure.delve_dap.launch_arguments import (
    DEBUG_BINARY_NAME,
    remove_stale_debug_binaries,
)


def test_stale_debug_binaries_are_removed_but_not_our_own(tmp_path: Path):
    own = tmp_path / f"{DEBUG_BINARY_NAME}_42.exe"
    stale = tmp_path / f"{DEBUG_BINARY_NAME}_7.exe"
    unrelated = tmp_path / "other.exe"
    for path in (own, stale, unrelated):
        path.write_bytes(b"x")

    removed = remove_stale_debug_binaries(42, tmp_path)

    assert removed == [stale]
    assert own.exists() and unrelated.exists() and not stale.exists()
