"""Print the [go.sha256] table of versions.toml for a Go release, from go.dev.

    python packaging/print_go_hashes.py 1.25.14

Source: https://go.dev/dl/?mode=json&include=all (official list of Go downloads).
Paste the output into packaging/versions.toml and update go.version.
"""

from __future__ import annotations

import argparse
import json
import urllib.error
import urllib.request

INDEX_URL = "https://go.dev/dl/?mode=json&include=all"
TARGETS = (
    "darwin-amd64",
    "darwin-arm64",
    "linux-amd64",
    "linux-arm64",
    "windows-amd64",
    "windows-arm64",
)


def fetch_release(version: str) -> dict:
    try:
        with urllib.request.urlopen(INDEX_URL, timeout=60) as response:
            releases = json.load(response)
    except (urllib.error.URLError, TimeoutError) as error:
        raise SystemExit(f"Cannot read {INDEX_URL}: {error}") from error
    for release in releases:
        if release["version"] == f"go{version}":
            return release
    raise SystemExit(f"go{version} not found in {INDEX_URL}")


def archive_hashes(release: dict) -> dict[str, str]:
    hashes = {}
    for item in release["files"]:
        key = f"{item['os']}-{item['arch']}"
        if item["kind"] == "archive" and key in TARGETS:
            hashes[key] = item["sha256"]
    return hashes


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("version", help="for example 1.25.14")
    version = parser.parse_args(argv).version.removeprefix("go")
    hashes = archive_hashes(fetch_release(version))
    print(f'# go.version = "{version}"')
    print("[go.sha256]")
    for key in sorted(hashes):
        print(f'{key} = "{hashes[key]}"')
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
