"""Print the [rust.sha256] tables of versions.toml for a Rust release, from static.rust-lang.org.

    python packaging/print_rust_hashes.py 1.99.0

Source: https://static.rust-lang.org/dist/channel-rust-<version>.toml (the official channel
manifest, every component url and its xz sha256). Paste the output into packaging/versions.toml
and update rust.version and rust.date.
"""

from __future__ import annotations

import argparse
import urllib.error
import urllib.request

import tomllib
from fetch_rust import COMPONENTS, TRIPLES

MANIFEST_URL = "https://static.rust-lang.org/dist/channel-rust-{version}.toml"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("version", help="a stable version such as 1.99.0")
    version = parser.parse_args(argv).version
    try:
        with urllib.request.urlopen(MANIFEST_URL.format(version=version), timeout=60) as response:
            channel = tomllib.loads(response.read().decode("utf-8"))
    except urllib.error.URLError as error:
        raise SystemExit(f"Cannot read the channel manifest of {version}: {error}") from error
    print(f'version = "{version}"\ndate = "{channel["date"]}"\n')
    for key, triple in TRIPLES.items():
        print(f"[rust.sha256.{key}]")
        for name, package in COMPONENTS.items():
            targets = channel["pkg"][package]["target"]
            entry = targets.get(triple) or targets.get("*")
            if entry is None:  # rust-mingw exists for Windows only
                continue
            print(f'{name} = "{entry["xz_hash"]}"')
        print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
