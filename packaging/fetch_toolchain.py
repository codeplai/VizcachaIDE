"""Download the pinned Go distribution and build dlv/gopls for a target platform.

Result layout (agreed with track E, relative to the packaged app directory)::

    <dest>/toolchain/go/bin/go[.exe], gofmt[.exe]   Go distribution (GOROOT)
    <dest>/toolchain/bin/dlv[.exe], gopls[.exe]      cross-compiled tools
    <dest>/toolchain/licenses/                       Delve and gopls licenses
    <dest>/toolchain/VERSIONS.txt                    what was bundled

Usage: python packaging/fetch_toolchain.py --os windows --arch amd64 --dest build/stage
Downloads are cached in packaging/cache/ and verified with the sha256 in versions.toml.
"""

from __future__ import annotations

import argparse
import hashlib
import shutil
import tarfile
import urllib.error
import urllib.request
import zipfile
from pathlib import Path

from go_tools import build_tool
from versions import CACHE_DIR, Target, host_target, load_versions

CHUNK = 1 << 20


def sha256_of(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(CHUNK), b""):
            digest.update(block)
    return digest.hexdigest()


def download(url: str, target: Path, expected_sha256: str) -> Path:
    """Download ``url`` into ``target`` unless a verified copy is already cached."""
    if target.exists() and sha256_of(target) == expected_sha256:
        print(f"[cache] {target.name}")
        return target
    target.parent.mkdir(parents=True, exist_ok=True)
    partial = target.with_name(target.name + ".part")
    print(f"[download] {url}")
    try:
        with urllib.request.urlopen(url, timeout=60) as response, partial.open("wb") as out:
            shutil.copyfileobj(response, out, CHUNK)
    except (urllib.error.URLError, TimeoutError) as error:
        partial.unlink(missing_ok=True)
        raise SystemExit(f"Download failed: {url}: {error}") from error
    actual = sha256_of(partial)
    if actual != expected_sha256:
        partial.unlink()
        raise SystemExit(
            f"sha256 mismatch for {url}\n  expected {expected_sha256}\n  got      {actual}"
        )
    partial.replace(target)
    return target


def extract_archive(archive: Path, destination: Path) -> None:
    if archive.suffix == ".zip":
        with zipfile.ZipFile(archive) as bundle:
            bundle.extractall(destination)
        return
    with tarfile.open(archive, "r:gz") as bundle:
        if hasattr(tarfile, "data_filter"):
            bundle.extractall(destination, filter="data")
            return
        bundle.extractall(destination)  # noqa: S202 - pinned, sha256-verified archive


def fetch_go(go_config: dict, target: Target, toolchain_dir: Path) -> None:
    hashes = go_config["sha256"]
    if target.key not in hashes:
        raise SystemExit(f"No pinned Go archive for {target.key}. Known: {', '.join(hashes)}")
    filename = f"go{go_config['version']}.{target.key}.{target.archive_ext}"
    url = go_config["url_template"].format(filename=filename)
    archive = download(url, CACHE_DIR / "downloads" / filename, hashes[target.key])
    goroot = toolchain_dir / "go"
    if goroot.exists():
        shutil.rmtree(goroot)
    print(f"[extract] {filename} -> {goroot}")
    extract_archive(archive, toolchain_dir)  # archives contain a top-level "go/" folder
    for name in go_config.get("prune", []):
        shutil.rmtree(goroot / name, ignore_errors=True)


def write_manifest(config: dict, target: Target, toolchain_dir: Path) -> None:
    lines = [f"target = {target.key}", f"go = {config['go']['version']}"]
    lines += [f"{name} = {spec['version']}" for name, spec in config["tools"].items()]
    (toolchain_dir / "VERSIONS.txt").write_text("\n".join(lines) + "\n", encoding="utf-8")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    host = host_target()
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--os", default=host.os, choices=["windows", "darwin", "linux"])
    parser.add_argument("--arch", default=host.arch, choices=["amd64", "arm64"])
    parser.add_argument("--dest", type=Path, required=True, help="directory that gets toolchain/")
    parser.add_argument("--skip-go", action="store_true", help="only build dlv and gopls")
    parser.add_argument("--skip-tools", action="store_true", help="only fetch Go")
    parser.add_argument("--go", default="go", help="host go used to build the tools")
    return parser.parse_args(argv)


def fetch_toolchain(target: Target, dest: Path, args: argparse.Namespace) -> Path:
    config = load_versions()
    toolchain_dir = dest.resolve() / "toolchain"
    toolchain_dir.mkdir(parents=True, exist_ok=True)
    if not args.skip_go:
        fetch_go(config["go"], target, toolchain_dir)
    if not args.skip_tools:
        for name, spec in config["tools"].items():
            build_tool(name, spec, target, config["go"]["version"], toolchain_dir, args.go)
    write_manifest(config, target, toolchain_dir)
    return toolchain_dir


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    toolchain_dir = fetch_toolchain(Target(args.os, args.arch), args.dest, args)
    print(f"[done] toolchain ready in {toolchain_dir}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
