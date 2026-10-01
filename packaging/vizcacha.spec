# -*- mode: python ; coding: utf-8 -*-
"""PyInstaller spec for VizcachaIDE (onedir, windowed).

Normally driven by packaging/build.py. Direct use:

    pyinstaller packaging/vizcacha.spec -- --variant lite
    pyinstaller packaging/vizcacha.spec -- --variant full --toolchain packaging/cache/stage/windows-amd64/toolchain

Bundle layout (PyInstaller 6 puts data in _internal/, which is sys._MEIPASS):
    _internal/logo.png, logo.ico                        -> vizcacha.ui.resources.resource_path()
    _internal/vizcacha/i18n/locale/<lang>/LC_MESSAGES/*.mo -> vizcacha.i18n.translator.LOCALE_DIR
    _internal/examples/*.go, LICENSE, NOTICE.md
    toolchain/                                          (full variant, next to the executable)
macOS: the toolchain lives in Contents/Resources/toolchain and Contents/MacOS/toolchain is a
symlink to it, so "next to the executable" still holds and codesign stays happy.
"""

import argparse
import shutil
import subprocess
import sys
from pathlib import Path

parser = argparse.ArgumentParser(prog="vizcacha.spec")
parser.add_argument("--variant", choices=["full", "lite"], default="lite")
parser.add_argument("--toolchain", type=Path, help="prepared toolchain/ dir (full variant)")
options = parser.parse_args()
if options.variant == "full" and not (options.toolchain and options.toolchain.is_dir()):
    raise SystemExit("--variant full needs --toolchain <dir> (run packaging/fetch_toolchain.py)")

ROOT = Path(SPECPATH).resolve().parent  # noqa: F821 - injected by PyInstaller
APP_NAME = "VizcachaIDE"
sys.path.insert(0, str(ROOT / "packaging"))
from versions import app_version  # noqa: E402

VERSION = app_version()


def locale_datas():
    locale_dir = ROOT / "vizcacha" / "i18n" / "locale"
    return [(str(mo), str(mo.parent.relative_to(ROOT))) for mo in locale_dir.glob("*/LC_MESSAGES/*.mo")]


datas = locale_datas() + [
    (str(ROOT / "logo.png"), "."),
    (str(ROOT / "logo.ico"), "."),
    (str(ROOT / "examples"), "examples"),
    (str(ROOT / "LICENSE"), "."),
    (str(ROOT / "packaging" / "NOTICE.md"), "."),
    (str(ROOT / "packaging" / "licenses" / "GPL-3.0.txt"), "licenses"),
]

if sys.platform == "win32":
    icon = str(ROOT / "logo.ico")
elif sys.platform == "darwin":
    icon = str(ROOT / "logo.png")  # converted to .icns by PyInstaller (needs Pillow)
else:
    icon = None

a = Analysis(  # noqa: F821
    [str(ROOT / "main.py")],
    pathex=[str(ROOT)],
    datas=datas,
    excludes=["tkinter", "pytest", "PyQt5.QtWebEngineWidgets", "PyQt5.QtQml", "PyQt5.QtQuick"],
    noarchive=False,
)

# Qt pieces a QtWidgets editor never loads: the WebGL platform plugin drags in QtQuick/QML,
# and opengl32sw.dll is the software OpenGL fallback (~20 MB) only used by OpenGL widgets.
UNUSED_QT = ("qwebgl", "qt5quick", "qt5qml", "qt5websockets", "opengl32sw")


def is_used(entry):
    name = Path(entry[0]).name.lower()
    return not name.startswith(UNUSED_QT)


a.binaries = [entry for entry in a.binaries if is_used(entry)]
pyz = PYZ(a.pure)  # noqa: F821
exe = EXE(  # noqa: F821
    pyz,
    a.scripts,
    [],
    exclude_binaries=True,
    name=APP_NAME,
    console=False,
    icon=icon,
    upx=False,
)
coll = COLLECT(exe, a.binaries, a.datas, name=APP_NAME, upx=False)  # noqa: F821

DIST = Path(DISTPATH)  # noqa: F821
if sys.platform == "darwin":
    app = BUNDLE(  # noqa: F821
        coll,
        name=f"{APP_NAME}.app",
        icon=icon,
        bundle_identifier="pe.codeplai.vizcachaide",
        version=VERSION,
        info_plist={
            "CFBundleShortVersionString": VERSION,
            "NSHighResolutionCapable": True,
            "LSMinimumSystemVersion": "11.0",
            "CFBundleDocumentTypes": [
                {
                    "CFBundleTypeName": "Go source file",
                    "CFBundleTypeRole": "Editor",
                    "LSHandlerRank": "Alternate",
                    "LSItemContentTypes": ["public.source-code"],
                    "CFBundleTypeExtensions": ["go"],
                }
            ],
        },
    )


def install_toolchain(source):
    if sys.platform != "darwin":
        target = DIST / APP_NAME / "toolchain"
        shutil.rmtree(target, ignore_errors=True)
        shutil.copytree(source, target, symlinks=True)
        return
    contents = DIST / f"{APP_NAME}.app" / "Contents"
    target = contents / "Resources" / "toolchain"
    shutil.rmtree(target, ignore_errors=True)
    shutil.copytree(source, target, symlinks=True)
    link = contents / "MacOS" / "toolchain"
    link.unlink(missing_ok=True)
    link.symlink_to(Path("..") / "Resources" / "toolchain")
    # Adding files invalidates PyInstaller's ad-hoc signature: re-sign ad hoc.
    subprocess.run(["codesign", "--force", "--deep", "--sign", "-", str(contents.parent)], check=True)


if options.variant == "full":
    install_toolchain(options.toolchain.resolve())
