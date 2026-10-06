"""Download the official standalone Rust components and stage them as one sysroot (the ``full`` variant).

Result layout (where ``wails/internal/adapters/rust/locator.go`` looks for it)::

    <dest>/toolchain/rust/bin/rustc, cargo, clippy-driver, cargo-clippy, rustfmt, cargo-fmt, rust-analyzer
    <dest>/toolchain/rust/lib/rustlib/<triple>/       the standard library (and MinGW on windows-gnu)
    <dest>/toolchain/rust/lib/rustlib/etc/            lldb_lookup.py and friends (the debugger formatters)
    <dest>/toolchain/rust/lib/rustlib/src/            rust-src, which rust-analyzer needs
    <dest>/toolchain/licenses/                        the Rust licenses (MIT and Apache-2.0)
    <dest>/toolchain/VERSIONS.txt                     what was bundled

Usage: python packaging/fetch_rust.py --os windows --arch amd64 --dest build/stage
The tarballs come from https://static.rust-lang.org/dist/<date>/ (the same files rustup downloads), are
cached in packaging/cache/downloads/ and verified with the sha256 of ``[rust.sha256.<os>-<arch>]`` in
versions.toml. rust-docs is never fetched. Windows uses the GNU target (x86_64-pc-windows-gnu): it links
with the MinGW that ships in the ``rust-mingw`` component, so no Visual Studio is needed. There is no
bash on Windows, so each component is installed by copying the files its ``manifest.in`` lists, which is
exactly what the ``install.sh`` of the tarball does.
"""

from __future__ import annotations

import argparse
import shutil
import tarfile
import tempfile
from pathlib import Path

from fetch_toolchain import download
from versions import CACHE_DIR, Target, host_target, load_versions

TRIPLES = {
    "windows-amd64": "x86_64-pc-windows-gnu",
    "linux-amd64": "x86_64-unknown-linux-gnu",
    "darwin-arm64": "aarch64-apple-darwin",
    "darwin-amd64": "x86_64-apple-darwin",
}
# Component name (the file prefix of its tarball) -> package key of the channel manifest.
COMPONENTS = {
    "rustc": "rustc",
    "cargo": "cargo",
    "rust-std": "rust-std",
    "rust-mingw": "rust-mingw",
    "clippy": "clippy-preview",
    "rustfmt": "rustfmt-preview",
    "rust-analyzer": "rust-analyzer-preview",
    "rust-src": "rust-src",
}
TARGET_INDEPENDENT = ("rust-src",)  # the tarball has no triple in its name
WINDOWS_ONLY = ("rust-mingw",)
LICENSE_FILES = ("LICENSE-MIT", "LICENSE-APACHE", "COPYRIGHT", "LICENSE-THIRD-PARTY")
REQUIRED_TOOLS = ("rustc", "cargo", "rustfmt", "rust-analyzer", "clippy-driver", "cargo-clippy")


def components_for(target: Target) -> list[str]:
    return [name for name in COMPONENTS if name not in WINDOWS_ONLY or target.os == "windows"]


def component_url(config: dict, name: str, triple: str) -> tuple[str, str]:
    """Return (url, filename) of the xz tarball of a component."""
    suffix = "" if name in TARGET_INDEPENDENT else f"-{triple}"
    filename = f"{name}-{config['version']}{suffix}.tar.xz"
    return config["url_template"].format(date=config["date"], filename=filename), filename


def extract_component(archive: Path, scratch: Path) -> Path:
    """Unpack a tarball into ``scratch`` and return its single top-level folder."""
    with tarfile.open(archive, "r:xz") as bundle:
        if hasattr(tarfile, "data_filter"):
            bundle.extractall(scratch, filter="data")
        else:
            bundle.extractall(scratch)  # noqa: S202 - pinned, sha256-verified archive
    (root,) = [path for path in scratch.iterdir() if path.is_dir()]
    return root


def install_component(root: Path, sysroot: Path) -> int:
    """Copy what the manifest.in of each component of ``root`` lists into ``sysroot``; return the file count."""
    copied = 0
    for component in (root / "components").read_text(encoding="utf-8").split():
        for line in (root / component / "manifest.in").read_text(encoding="utf-8").splitlines():
            kind, _, relative = line.partition(":")
            source = root / component / relative
            destination = sysroot / relative
            if kind == "file":
                destination.parent.mkdir(parents=True, exist_ok=True)
                shutil.copy2(source, destination)
                copied += 1
            elif kind == "dir":
                shutil.copytree(source, destination, symlinks=True, dirs_exist_ok=True)
                copied += sum(1 for path in source.rglob("*") if path.is_file())
    return copied


def copy_licenses(root: Path, licenses_dir: Path) -> None:
    licenses_dir.mkdir(parents=True, exist_ok=True)
    for name in LICENSE_FILES:
        if (root / name).is_file():
            shutil.copy2(root / name, licenses_dir / f"rust-{name}.txt")


def write_manifest(config: dict, target: Target, toolchain_dir: Path) -> None:
    """Add the rust line to VERSIONS.txt, keeping the lines other fetchers wrote."""
    manifest = toolchain_dir / "VERSIONS.txt"
    kept = []
    if manifest.is_file():
        text = manifest.read_text(encoding="utf-8")
        kept = [line for line in text.splitlines() if line and not line.startswith("rust")]
    if not any(line.startswith("target =") for line in kept):
        kept.insert(0, f"target = {target.key}")
    kept.append(f"rust = {config['version']} ({config['date']}, {TRIPLES[target.key]})")
    manifest.write_text("\n".join(kept) + "\n", encoding="utf-8")


def folder_size(folder: Path) -> int:
    return sum(path.stat().st_size for path in folder.rglob("*") if path.is_file())


def fetch_rust(target: Target, dest: Path) -> Path:
    """Stage the bundled Rust sysroot under ``dest/toolchain`` and return that folder."""
    config = load_versions()["rust"]
    if target.key not in TRIPLES:
        raise SystemExit(f"No bundled Rust for {target.key}. Known: {', '.join(TRIPLES)}")
    hashes = config["sha256"][target.key]
    toolchain_dir = dest.resolve() / "toolchain"
    sysroot = toolchain_dir / "rust"
    if sysroot.exists():
        shutil.rmtree(sysroot)
    sysroot.mkdir(parents=True)
    for name in components_for(target):
        url, filename = component_url(config, name, TRIPLES[target.key])
        archive = download(url, CACHE_DIR / "downloads" / filename, hashes[name])
        with tempfile.TemporaryDirectory(dir=toolchain_dir, prefix=".rust-") as scratch:
            root = extract_component(archive, Path(scratch))
            copied = install_component(root, sysroot)
            if name == "rustc":
                copy_licenses(root, toolchain_dir / "licenses")
        print(f"[install] {name}: {copied} files")
    for tool in REQUIRED_TOOLS:
        if not (sysroot / "bin" / f"{tool}{target.exe_suffix}").is_file():
            raise SystemExit(f"{tool} is missing from {sysroot / 'bin'}: check the component list")
    write_manifest(config, target, toolchain_dir)
    print(f"[size] rust: {folder_size(sysroot) / 1e6:.0f} MB")
    return toolchain_dir


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    host = host_target()
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--os", default=host.os, choices=["windows", "darwin", "linux"])
    parser.add_argument("--arch", default=host.arch, choices=["amd64", "arm64"])
    parser.add_argument("--dest", type=Path, required=True, help="directory that gets toolchain/")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    toolchain_dir = fetch_rust(Target(args.os, args.arch), args.dest)
    print(f"[done] Rust ready in {toolchain_dir / 'rust'}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
