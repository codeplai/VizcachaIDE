#!/usr/bin/env bash
# Build an AppImage from the Wails Linux binary (+ toolchain/ in the full variant).
#
#   bash wails/packaging/linux/make_appimage.sh <app-dir> <output.AppImage> [version]
#
# <app-dir> contains VizcachaIDE (the executable) and, for "full", toolchain/.
#
# The IDE locates its bundled Go with os.Executable(): inside an AppImage that is
# <mount>/usr/bin/VizcachaIDE, so toolchain/ goes to usr/bin/toolchain (next to the binary).
#
# WebKitGTK is NOT bundled: its helper processes (WebKitWebProcess, WebKitNetworkProcess)
# and GStreamer/GIO modules use absolute paths and do not survive relocation. The AppImage
# therefore needs libwebkit2gtk-4.1 + libgtk-3 from the host (Ubuntu 22.04+, Debian 12+,
# Fedora 38+). AppRun explains this if the library is missing.
#
# appimagetool: $APPIMAGETOOL, else downloaded once to wails/packaging/cache/.
# APPIMAGE_EXTRACT_AND_RUN=1 lets it run without FUSE (CI containers).
set -euo pipefail

SRC="${1:?usage: make_appimage.sh <app-dir> <output.AppImage> [version]}"
OUT="${2:?usage: make_appimage.sh <app-dir> <output.AppImage> [version]}"
VERSION="${3:-dev}"
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../../.." && pwd)"
ARCH="$(uname -m)"
TOOL_URL="https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-${ARCH}.AppImage"

TOOL="${APPIMAGETOOL:-$HERE/../cache/appimagetool-${ARCH}.AppImage}"
if [[ ! -x "$TOOL" ]]; then
  mkdir -p "$(dirname "$TOOL")"
  echo "Downloading appimagetool: $TOOL_URL"
  curl -fsSL -o "$TOOL" "$TOOL_URL"
  chmod +x "$TOOL"
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
APPDIR="$WORK/VizcachaIDE.AppDir"
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" \
         "$APPDIR/usr/share/icons/hicolor/256x256/apps"
cp -a "$SRC/." "$APPDIR/usr/bin/"

cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
if command -v ldconfig >/dev/null 2>&1 && ! ldconfig -p 2>/dev/null | grep -q 'libwebkit2gtk-4.1'; then
  MSG="VizcachaIDE needs WebKitGTK 4.1 (libwebkit2gtk-4.1-0).
Install it, e.g.: sudo apt install libwebkit2gtk-4.1-0
VizcachaIDE necesita WebKitGTK 4.1 (libwebkit2gtk-4.1-0)."
  echo "$MSG" >&2
  if command -v zenity >/dev/null 2>&1; then zenity --error --text="$MSG" 2>/dev/null || true; fi
  exit 1
fi
exec "$HERE/usr/bin/VizcachaIDE" "$@"
EOF
chmod +x "$APPDIR/AppRun"

cp "$HERE/vizcacha.desktop" "$APPDIR/vizcacha.desktop"
cp "$HERE/vizcacha.desktop" "$APPDIR/usr/share/applications/vizcacha.desktop"
cp "$REPO/wails/build/appicon.png" "$APPDIR/vizcacha.png"
cp "$REPO/wails/build/appicon.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/vizcacha.png"
ln -s vizcacha.png "$APPDIR/.DirIcon"

mkdir -p "$(dirname "$OUT")"
rm -f "$OUT"
ARCH="$ARCH" VERSION="$VERSION" APPIMAGE_EXTRACT_AND_RUN=1 "$TOOL" -n "$APPDIR" "$OUT"
echo "AppImage: $OUT"
