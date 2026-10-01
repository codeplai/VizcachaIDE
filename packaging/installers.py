"""Turn a PyInstaller output folder into distributable packages (dist/release/).

Windows: portable zip + Inno Setup installer (if ISCC is available).
macOS:   .dmg via packaging/macos/make_dmg.sh (create-dmg).
Linux:   AppImage via packaging/linux/make_appimage.sh (appimagetool).
"""

from __future__ import annotations

import os
import shutil
import subprocess
from pathlib import Path

from versions import PACKAGING_DIR, REPO_ROOT, Target, numeric_version

RELEASE_DIR = REPO_ROOT / "dist" / "release"
INSTALLER_WORK = REPO_ROOT / "build" / "installer"
ISCC_LOCATIONS = (
    Path(os.environ.get("PROGRAMFILES(X86)", r"C:\Program Files (x86)")) / "Inno Setup 6",
    Path(os.environ.get("PROGRAMFILES", r"C:\Program Files")) / "Inno Setup 6",
    Path(os.environ.get("LOCALAPPDATA", "")) / "Programs" / "Inno Setup 6",
)
PLATFORM_LABELS = {
    "windows-amd64": "windows-x64",
    "windows-arm64": "windows-arm64",
    "darwin-amd64": "macos-x86_64",
    "darwin-arm64": "macos-arm64",
    "linux-amd64": "linux-x86_64",
}


def artifact_name(target: Target, variant: str, version: str) -> str:
    return f"VizcachaIDE-{version}-{PLATFORM_LABELS.get(target.key, target.key)}-{variant}"


def find_iscc() -> Path | None:
    on_path = shutil.which("iscc")
    if on_path:
        return Path(on_path)
    for folder in ISCC_LOCATIONS:
        candidate = folder / "ISCC.exe"
        if candidate.exists():
            return candidate
    return None


def license_page() -> Path:
    """NOTICE (GPLv3 distribution notice) followed by the MIT license of the source code."""
    INSTALLER_WORK.mkdir(parents=True, exist_ok=True)
    notice = (PACKAGING_DIR / "NOTICE.md").read_text(encoding="utf-8")
    mit = (REPO_ROOT / "LICENSE").read_text(encoding="utf-8")
    page = INSTALLER_WORK / "LICENSE-NOTICE.txt"
    separator = "\n\n" + "=" * 72 + "\nVizcachaIDE source code license (MIT)\n" + "=" * 72 + "\n\n"
    # UTF-8 with BOM so Inno Setup shows accents correctly.
    page.write_text(notice + separator + mit, encoding="utf-8-sig")
    return page


def portable_zip(app_dir: Path, name: str) -> Path:
    RELEASE_DIR.mkdir(parents=True, exist_ok=True)
    base = RELEASE_DIR / f"{name}-portable"
    archive = shutil.make_archive(str(base), "zip", root_dir=app_dir.parent, base_dir=app_dir.name)
    return Path(archive)


def inno_installer(app_dir: Path, name: str, variant: str, version: str) -> Path | None:
    iscc = find_iscc()
    if iscc is None:
        print(
            "[warn] Inno Setup (ISCC.exe) not found: skipping the installer. "
            "Install it from https://jrsoftware.org/isdl.php or `choco install innosetup`."
        )
        return None
    command = [
        str(iscc), "/Q",
        f"/DAppVersion={version}", f"/DAppVersionNumeric={numeric_version(version)}",
        f"/DSourceDir={app_dir}", f"/DVariant={variant}", f"/DLicenseFile={license_page()}",
        f"/DOutputDir={RELEASE_DIR}", f"/DOutputBaseName={name}-setup",
        str(PACKAGING_DIR / "windows" / "vizcacha.iss"),
    ]  # fmt: skip
    run(command, "Inno Setup")
    return RELEASE_DIR / f"{name}-setup.exe"


def run(command: list[str], what: str) -> None:
    print(f"[{what}] " + " ".join(command))
    try:
        subprocess.run(command, check=True)
    except FileNotFoundError as error:
        raise SystemExit(f"{what}: command not found: {command[0]}") from error
    except subprocess.CalledProcessError as error:
        raise SystemExit(f"{what} failed (exit {error.returncode})") from error


def package(app_dir: Path, target: Target, variant: str, version: str) -> list[Path]:
    RELEASE_DIR.mkdir(parents=True, exist_ok=True)
    name = artifact_name(target, variant, version)
    if target.os == "windows":
        artifacts = [portable_zip(app_dir, name)]
        installer = inno_installer(app_dir, name, variant, version)
        return artifacts + ([installer] if installer else [])
    if target.os == "darwin":
        dmg = RELEASE_DIR / f"{name}.dmg"
        run(["bash", str(PACKAGING_DIR / "macos" / "make_dmg.sh"), str(app_dir), str(dmg)], "dmg")
        return [dmg]
    appimage = RELEASE_DIR / f"{name}.AppImage"
    script = PACKAGING_DIR / "linux" / "make_appimage.sh"
    run(["bash", str(script), str(app_dir), str(appimage), version], "AppImage")
    return [appimage]
