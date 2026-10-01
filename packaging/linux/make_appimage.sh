#!/usr/bin/env bash
# Build an AppImage from the PyInstaller onedir output.
#
#   bash packaging/linux/make_appimage.sh dist/full/VizcachaIDE dist/release/VizcachaIDE-...AppImage 0.2.0
#
# AppDir layout:
#   AppRun                      -> exec usr/lib/vizcacha/VizcachaIDE
#   vizcacha.desktop, vizcacha.png, .DirIcon
#   usr/lib/vizcacha/           PyInstaller folder (+ toolchain/ in the full variant)
#   usr/share/applications/, usr/share/icons/hicolor/256x256/apps/
#
# appimagetool: taken from $APPIMAGETOOL if set, else downloaded once to packaging/cache/.
# APPIMAGE_EXTRACT_AND_RUN=1 lets it run without FUSE (CI containers).
# Glibc baseline = the build machine (ubuntu-22.04 -> glibc 2.35).
set -euo pipefail

SRC="${1:?usage: make_appimage.sh <dist/VizcachaIDE> <output.AppImage> [version]}"
OUT="${2:?usage: make_appimage.sh <dist/VizcachaIDE> <output.AppImage> [version]}"
VERSION="${3:-dev}"
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"
ARCH="$(uname -m)"
TOOL_URL="https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-${ARCH}.AppImage"

TOOL="${APPIMAGETOOL:-$REPO/packaging/cache/appimagetool-${ARCH}.AppImage}"
if [[ ! -x "$TOOL" ]]; then
  mkdir -p "$(dirname "$TOOL")"
  echo "Downloading appimagetool: $TOOL_URL"
  curl -fsSL -o "$TOOL" "$TOOL_URL"
  chmod +x "$TOOL"
fi

APPDIR="$(mktemp -d)/VizcachaIDE.AppDir"
trap 'rm -rf "$(dirname "$APPDIR")"' EXIT
mkdir -p "$APPDIR/usr/lib" "$APPDIR/usr/share/applications" \
         "$APPDIR/usr/share/icons/hicolor/256x256/apps"
cp -a "$SRC" "$APPDIR/usr/lib/vizcacha"

cat > "$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/lib/vizcacha/VizcachaIDE" "$@"
EOF
chmod +x "$APPDIR/AppRun"

cp "$HERE/vizcacha.desktop" "$APPDIR/vizcacha.desktop"
cp "$HERE/vizcacha.desktop" "$APPDIR/usr/share/applications/vizcacha.desktop"
cp "$REPO/logo.png" "$APPDIR/vizcacha.png"
cp "$REPO/logo.png" "$APPDIR/usr/share/icons/hicolor/256x256/apps/vizcacha.png"
ln -s vizcacha.png "$APPDIR/.DirIcon"

mkdir -p "$(dirname "$OUT")"
rm -f "$OUT"
ARCH="$ARCH" VERSION="$VERSION" APPIMAGE_EXTRACT_AND_RUN=1 "$TOOL" -n "$APPDIR" "$OUT"
echo "AppImage: $OUT"
