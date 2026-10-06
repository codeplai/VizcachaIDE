"""Build the VizcachaIDE Wails release packages (lite and the full variants) for one platform.

    python wails/packaging/build_release.py --variant all
    python wails/packaging/build_release.py --variant full-python --target windows/amd64
    python wails/packaging/build_release.py --variant full-cpp --target windows/amd64 --no-installer
    python wails/packaging/build_release.py --variant full,lite --target darwin/arm64
    python wails/packaging/build_release.py --variant lite --cache-dir D:/shared/cache

Outputs go to ``wails/dist/release/VizcachaIDE-<version>-<os>-<arch>-<variant>...``:

    windows  -setup.exe (NSIS, per user, EN/ES, WebView2 bootstrapper) and -portable.zip
    macos    .dmg (VizcachaIDE.app inside)
    linux    .AppImage and .tar.gz

Variants (what goes under ``toolchain/``):

    lite         nothing, the IDE uses the Go / Python installed on the system
    full-go      Go, Delve and gopls (this was called "full" before Python was bundled)
    full-python  CPython + debugpy, python-lsp-server and ruff
    full-cpp     clang/clang++, lld, lldb-dap, clangd and clang-format (llvm-mingw, Windows only)
    full         everything: Go + Python + Rust, plus C++ on Windows (macOS and Linux have no bundled
                 C++ compiler, they use the one of the system, so there ``full`` = Go + Python + Rust).
                 There is no separate full-rust variant: Rust ships only inside ``full``.

The Go part uses the same code as the PyQt packaging: ``packaging/fetch_toolchain.py`` and
``packaging/go_tools.py`` are imported, not copied. The Python part is
``packaging/fetch_python.py``, the C++ part ``packaging/fetch_cpp.py`` (llvm-mingw pruned to x86_64,
placed in ``toolchain/cpp/``) and the Rust part ``packaging/fetch_rust.py`` (the official standalone
components installed into one sysroot, ``toolchain/rust/``). ``--variant both`` (lite + full-go) and ``--variant all`` (every
variant) are shortcuts, and a comma list such as ``lite,full-python`` also works. ``full-cpp`` is
skipped, with a note, for macOS and Linux targets.
The toolchain is placed next to the executable, which is where ``toolchain.Locator`` looks
(``filepath.Dir(os.Executable())/toolchain/go/bin`` and ``.../toolchain/bin``):

    Windows  <install>/toolchain/...                 next to VizcachaIDE.exe
    Linux    <AppDir>/usr/bin/toolchain/...          next to the binary (or tar.gz folder)
    macOS    VizcachaIDE.app/Contents/Resources/toolchain, plus a symlink
             Contents/MacOS/toolchain -> ../Resources/toolchain, because the locator
             resolves the executable's folder, i.e. Contents/MacOS. No Go change is needed.

Python 3.11+ (tomllib in packaging/versions.py). ``wails`` must be on PATH; on Windows
``makensis`` too (or in NSIS_HOME / %LOCALAPPDATA%\\Programs\\NSIS / a tools folder).
"""

from __future__ import annotations

import argparse
import contextlib
import hashlib
import json
import os
import platform
import shutil
import subprocess
import sys
import tarfile
import tempfile
import zipfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
WAILS_DIR = HERE.parent
REPO_ROOT = WAILS_DIR.parent
LEGACY_PACKAGING = REPO_ROOT / "packaging"
INSTALLER_DIR = WAILS_DIR / "build" / "windows" / "installer"
BIN_DIR = WAILS_DIR / "build" / "bin"
DIST_DIR = WAILS_DIR / "dist"
RELEASE_DIR = DIST_DIR / "release"
STAGE_DIR = DIST_DIR / "stage"
# Which toolchain parts each variant bundles. Order matters for ``--variant all``.
VARIANT_PARTS = {
    "lite": (),
    "full-go": ("go",),
    "full-python": ("python",),
    "full-cpp": ("cpp",),
    "full": ("go", "python", "cpp", "rust"),
}
# Parts that exist only for some operating systems; elsewhere the system tools are used.
PART_OPERATING_SYSTEMS = {"cpp": ("windows",)}
PARTS = ("go", "python", "cpp", "rust")
VARIANT_SHORTCUTS = {"both": ["lite", "full-go"], "all": list(VARIANT_PARTS)}
PRODUCT = "VizcachaIDE"

_OS = {"win32": "windows", "darwin": "darwin", "linux": "linux"}
_ARCH = {"amd64": "amd64", "x86_64": "amd64", "arm64": "arm64", "aarch64": "arm64"}


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    host_os = next((v for k, v in _OS.items() if sys.platform.startswith(k)), "linux")
    host_arch = _ARCH.get(platform.machine().lower(), "amd64")
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "--variant",
        default="both",
        help="lite, full-go, full-python, full-cpp (Windows), full (Go + Python + Rust, + C++ on Windows), "
        "both (lite + full-go), all, "
        "or a comma list; default both",
    )
    parser.add_argument(
        "--target", default=f"{host_os}/{host_arch}", help="os/arch (default: host)"
    )
    parser.add_argument(
        "--cache-dir",
        type=Path,
        default=None,
        help="toolchain cache (downloads/, tools/, gopath/); default packaging/cache",
    )
    parser.add_argument("--go", default="go", help="host go used to build dlv/gopls")
    parser.add_argument("--skip-app-build", action="store_true", help="reuse build/bin")
    parser.add_argument("--skip-fetch", action="store_true", help="reuse dist/stage/<os-arch>")
    parser.add_argument(
        "--no-installer", action="store_true", help="only the .app/exe/zip, no nsis/dmg"
    )
    return parser.parse_args(argv)


# ---------------------------------------------------------------- legacy packaging helpers


def load_legacy(cache_dir: Path | None):
    """Import packaging/versions.py, fetch_toolchain.py and go_tools.py (no copy of their code)."""
    sys.path.insert(0, str(LEGACY_PACKAGING))
    import versions  # noqa: PLC0415

    if cache_dir is not None:
        # fetch_toolchain and go_tools do "from versions import CACHE_DIR": patch it first.
        versions.CACHE_DIR = cache_dir.resolve()
    import fetch_toolchain  # noqa: PLC0415
    import go_tools  # noqa: PLC0415

    return versions, fetch_toolchain, go_tools


def parse_variants(text: str) -> list[str]:
    names: list[str] = []
    for item in text.split(","):
        names += VARIANT_SHORTCUTS.get(item.strip(), [item.strip()])
    unknown = [name for name in names if name not in VARIANT_PARTS]
    if unknown:
        known = ", ".join(VARIANT_PARTS)
        raise SystemExit(f"Unknown variant(s): {', '.join(unknown)}. Known: {known}")
    return list(dict.fromkeys(names))


def parts_for(variant: str, os_name: str) -> tuple[str, ...]:
    """The toolchain parts of ``variant`` that exist for ``os_name`` (C++ is bundled on Windows only)."""
    return tuple(
        part
        for part in VARIANT_PARTS[variant]
        if os_name in PART_OPERATING_SYSTEMS.get(part, (os_name,))
    )


def drop_unavailable_variants(variants: list[str], os_name: str) -> list[str]:
    """Remove the variants that would be empty on ``os_name`` but are not ``lite`` (full-cpp off Windows)."""
    kept = [name for name in variants if name == "lite" or parts_for(name, os_name)]
    for name in set(variants) - set(kept):
        print(f"[release] {name} does not exist for {os_name}: skipped (it uses the system compiler)")
    return kept


def stage_part(part: str, target: str, args: argparse.Namespace) -> Path:
    """Fetch one toolchain part ("go", "python", "cpp" or "rust"); return the folder that contains toolchain/."""
    versions, fetch_toolchain, _ = load_legacy(args.cache_dir)
    os_name, arch = target.split("/")
    stage = STAGE_DIR / f"{os_name}-{arch}" / part
    if args.skip_fetch and (stage / "toolchain" / "VERSIONS.txt").is_file():
        print(f"[stage] reusing {stage}")
        return stage
    if stage.exists():
        shutil.rmtree(stage)
    stage.mkdir(parents=True)
    if part == "go":
        options = argparse.Namespace(skip_go=False, skip_tools=False, go=args.go)
        fetch_toolchain.fetch_toolchain(versions.Target(os_name, arch), stage, options)
    elif part == "python":
        import fetch_python  # noqa: PLC0415

        fetch_python.fetch_python(versions.Target(os_name, arch), stage)
    elif part == "cpp":
        import fetch_cpp  # noqa: PLC0415

        fetch_cpp.fetch_cpp(versions.Target(os_name, arch), stage)
    else:
        import fetch_rust  # noqa: PLC0415

        fetch_rust.fetch_rust(versions.Target(os_name, arch), stage)
    return stage


def merge_stages(parts: list[Path], destination: Path) -> Path:
    """Copy several staged toolchains into one folder, merging licenses and VERSIONS.txt."""
    if destination.exists():
        shutil.rmtree(destination)
    lines: list[str] = []
    for part in parts:
        copy_tree(part / "toolchain", destination / "toolchain")
        text = (part / "toolchain" / "VERSIONS.txt").read_text(encoding="utf-8")
        lines += [line for line in text.splitlines() if line not in lines]
    manifest = destination / "toolchain" / "VERSIONS.txt"
    manifest.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return destination


def stage_variants(target: str, variants: list[str], args: argparse.Namespace) -> dict[str, Path]:
    """Return {variant: folder containing toolchain/} for every variant that bundles something."""
    os_name, arch = target.split("/")
    needed = {part for variant in variants for part in parts_for(variant, os_name)}
    parts = {part: stage_part(part, target, args) for part in PARTS if part in needed}
    stages: dict[str, Path] = {}
    for variant in variants:
        wanted = [parts[part] for part in parts_for(variant, os_name)]
        if len(wanted) == 1:
            stages[variant] = wanted[0]
        elif wanted:
            merged = STAGE_DIR / f"{os_name}-{arch}" / variant
            stages[variant] = merge_stages(wanted, merged)
    return stages


# ---------------------------------------------------------------- helpers


def run(command: list[str], **kwargs) -> None:
    print("+", " ".join(str(part) for part in command), flush=True)
    subprocess.run([str(part) for part in command], check=True, **kwargs)


def product_version() -> str:
    config = json.loads((WAILS_DIR / "wails.json").read_text(encoding="utf-8"))
    return config["info"]["productVersion"]


def output_filename() -> str:
    config = json.loads((WAILS_DIR / "wails.json").read_text(encoding="utf-8"))
    return config.get("outputfilename", config["name"])


def sha256_of(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            digest.update(block)
    return digest.hexdigest()


def find_makensis() -> str:
    found = shutil.which("makensis")
    if found:
        return found
    candidates = [
        os.environ.get("NSIS_HOME", ""),
        os.path.join(os.environ.get("LOCALAPPDATA", ""), "Programs", "NSIS"),
        r"C:\Program Files (x86)\NSIS",
        r"C:\Program Files\NSIS",
        str(Path.home() / "tools" / "nsis-3.10"),
    ]
    for folder in candidates:
        if folder and (Path(folder) / "makensis.exe").is_file():
            os.environ["PATH"] = folder + os.pathsep + os.environ["PATH"]  # wails looks it up too
            return str(Path(folder) / "makensis.exe")
    raise SystemExit(
        "makensis not found. Windows: 'choco install nsis' (admin) or unzip the official "
        "nsis-3.x.zip anywhere and set NSIS_HOME. See wails/packaging/README.md."
    )


def write_license_notice() -> Path:
    """NOTICE + MIT license as UTF-16 text for the NSIS license page."""
    text = (HERE / "NOTICE.md").read_text(encoding="utf-8")
    text += "\n\n" + (REPO_ROOT / "LICENSE").read_text(encoding="utf-8")
    target = INSTALLER_DIR / "LICENSE-NOTICE.txt"
    target.write_bytes(b"\xff\xfe" + text.replace("\n", "\r\n").encode("utf-16-le"))
    return target


@contextlib.contextmanager
def short_source_path(path: Path):
    """Yield a short path for ``path`` (NSIS ``File /r`` fails on paths over 260 characters).

    debugpy ships paths of about 150 characters; in a deep checkout the staged copy goes over
    MAX_PATH. On Windows a directory junction in the temp folder (no administrator needed) gives
    NSIS a short name; elsewhere, or when the path is already short, ``path`` is yielded as is."""
    if os.name != "nt" or len(str(path)) < 90:
        yield path
        return
    junction = Path(tempfile.gettempdir()) / f"vz-{hashlib.sha256(str(path).encode()).hexdigest()[:8]}"
    subprocess.run(["cmd", "/c", "mklink", "/J", str(junction), str(path)], check=True)
    try:
        yield junction
    finally:
        junction.rmdir()  # removes the junction only, never the target's content


def copy_tree(source: Path, destination: Path) -> None:
    shutil.copytree(source, destination, symlinks=True, dirs_exist_ok=True)


def zip_folder(folder: Path, archive: Path, root_name: str) -> None:
    with zipfile.ZipFile(archive, "w", zipfile.ZIP_DEFLATED, compresslevel=9) as bundle:
        for path in sorted(folder.rglob("*")):
            if path.is_file():
                bundle.write(path, Path(root_name) / path.relative_to(folder))


# ---------------------------------------------------------------- wails build


def build_app(target: str, args: argparse.Namespace) -> None:
    os_name = target.split("/")[0]
    command = ["wails", "build", "-clean", "-platform", target, "-trimpath"]
    if os_name == "linux":
        command += ["-tags", "webkit2_41"]  # WebKitGTK 4.1 binding (Ubuntu 22.04+)
    if os_name == "windows" and not args.no_installer:
        command += ["-nsis", "-installscope", "user"]
        find_makensis()
        write_license_notice()
    run(command, cwd=WAILS_DIR)


# ---------------------------------------------------------------- Windows


def package_windows(
    target: str, variants: list[str], args, version: str, stages: dict[str, Path]
) -> list[Path]:
    arch = target.split("/")[1]
    exe = BIN_DIR / f"{output_filename()}.exe"  # "outputfilename" in wails.json
    outputs: list[Path] = []
    for variant in variants:
        base = f"{PRODUCT}-{version}-windows-{arch}-{variant}"
        # Portable zip: VizcachaIDE/VizcachaIDE.exe (+ toolchain/ for full).
        with_tmp = DIST_DIR / "work" / base
        if with_tmp.exists():
            shutil.rmtree(with_tmp)
        with_tmp.mkdir(parents=True)
        shutil.copy2(exe, with_tmp / f"{PRODUCT}.exe")
        if variant in stages:
            copy_tree(stages[variant] / "toolchain", with_tmp / "toolchain")
        portable = RELEASE_DIR / f"{base}-portable.zip"
        zip_folder(with_tmp, portable, PRODUCT)
        shutil.rmtree(with_tmp)
        outputs.append(portable)
        if args.no_installer:
            continue
        setup = RELEASE_DIR / f"{base}-setup.exe"
        if variant == "lite":
            built = next(BIN_DIR.glob("*-installer.exe"))  # made by "wails build -nsis"
            shutil.copy2(built, setup)
        else:
            # Same template; re-run makensis with the toolchain folder. wails_tools.nsh and
            # tmp/MicrosoftEdgeWebview2Setup.exe were generated by the wails build above.
            with short_source_path(stages[variant]) as toolchain_source:
                run(
                    [
                        find_makensis(),
                        f"-DARG_WAILS_{arch.upper()}_BINARY={exe}",
                        "-DREQUEST_EXECUTION_LEVEL=user",
                        "-DWAILS_INSTALL_SCOPE=user",
                        f"-DARG_TOOLCHAIN_DIR={toolchain_source}",
                        f"-DARG_OUTFILE={setup}",
                        "project.nsi",
                    ],
                    cwd=INSTALLER_DIR,
                )
        outputs.append(setup)
    return outputs


# ---------------------------------------------------------------- macOS


def adhoc_sign(app: Path) -> None:
    run(["codesign", "--force", "--deep", "--sign", "-", app])


def package_macos(
    target: str, variants: list[str], args, version: str, stages: dict[str, Path]
) -> list[Path]:
    arch = target.split("/")[1]
    built = next(BIN_DIR.glob("*.app"))
    outputs: list[Path] = []
    for variant in variants:
        base = f"{PRODUCT}-{version}-macos-{arch}-{variant}"
        work = DIST_DIR / "work" / base
        if work.exists():
            shutil.rmtree(work)
        work.mkdir(parents=True)
        app = work / f"{PRODUCT}.app"
        run(["ditto", built, app])  # keeps symlinks, xattrs and signatures
        if variant in stages:
            resources = app / "Contents" / "Resources"
            copy_tree(stages[variant] / "toolchain", resources / "toolchain")
            # os.Executable() is Contents/MacOS/<exe>, so the locator looks in
            # Contents/MacOS/toolchain.
            link = app / "Contents" / "MacOS" / "toolchain"
            link.symlink_to(Path("..") / "Resources" / "toolchain")
        adhoc_sign(app)
        outputs.append(app)  # kept for the smoke test; the dmg is the deliverable
        if not args.no_installer:
            dmg = RELEASE_DIR / f"{base}.dmg"
            run(["bash", HERE / "macos" / "make_dmg.sh", app, dmg])
            outputs.append(dmg)
    return outputs


# ---------------------------------------------------------------- Linux


def package_linux(
    target: str, variants: list[str], args, version: str, stages: dict[str, Path]
) -> list[Path]:
    arch = target.split("/")[1]
    built = BIN_DIR / output_filename()
    outputs: list[Path] = []
    for variant in variants:
        base = f"{PRODUCT}-{version}-linux-{arch}-{variant}"
        work = DIST_DIR / "work" / base
        if work.exists():
            shutil.rmtree(work)
        app_dir = work / PRODUCT
        app_dir.mkdir(parents=True)
        shutil.copy2(built, app_dir / PRODUCT)
        if variant in stages:
            copy_tree(stages[variant] / "toolchain", app_dir / "toolchain")
        # tar.gz: unpack anywhere and run ./VizcachaIDE/VizcachaIDE.
        extras = work / "tar" / PRODUCT
        copy_tree(app_dir, extras)
        shutil.copy2(HERE / "linux" / "vizcacha.desktop", extras)
        shutil.copy2(WAILS_DIR / "build" / "appicon.png", extras / "vizcacha.png")
        shutil.copy2(HERE / "NOTICE.md", extras / "NOTICE.md")
        shutil.copy2(REPO_ROOT / "LICENSE", extras / "LICENSE")
        tarball = RELEASE_DIR / f"{base}.tar.gz"
        with tarfile.open(tarball, "w:gz", compresslevel=9) as bundle:
            bundle.add(extras, arcname=PRODUCT)
        outputs.append(tarball)
        if not args.no_installer:
            appimage = RELEASE_DIR / f"{base}.AppImage"
            run(["bash", HERE / "linux" / "make_appimage.sh", app_dir, appimage, version])
            outputs.append(appimage)
    return outputs


# ---------------------------------------------------------------- main


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    target = args.target
    os_name = target.split("/")[0]
    variants = drop_unavailable_variants(parse_variants(args.variant), os_name)
    version = product_version()
    print(f"[release] {PRODUCT} {version} for {target}, variants: {', '.join(variants)}")

    RELEASE_DIR.mkdir(parents=True, exist_ok=True)
    stages = stage_variants(target, variants, args)
    if not args.skip_app_build:
        build_app(target, args)
    packagers = {"windows": package_windows, "darwin": package_macos, "linux": package_linux}
    outputs = packagers[os_name](target, variants, args, version, stages)

    print("\n[release] artifacts")
    sums = []
    for path in outputs:
        if path.is_dir():
            continue
        digest = sha256_of(path)
        sums.append(f"{digest}  {path.name}")
        print(f"  {path.name}  {path.stat().st_size / 1e6:.1f} MB  sha256={digest}")
    (RELEASE_DIR / f"SHA256SUMS-{os_name}-{target.split('/')[1]}.txt").write_text(
        "\n".join(sums) + "\n", encoding="utf-8"
    )
    if (DIST_DIR / "work").exists() and os_name == "windows":
        shutil.rmtree(DIST_DIR / "work", ignore_errors=True)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
