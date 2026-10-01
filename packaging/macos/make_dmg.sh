#!/usr/bin/env bash
# Wrap VizcachaIDE.app in a drag-to-Applications .dmg with create-dmg.
#
#   bash packaging/macos/make_dmg.sh dist/full/VizcachaIDE.app dist/release/VizcachaIDE-...dmg
#
# Requires: create-dmg (`brew install create-dmg`). The architecture of the .dmg is the one
# of the .app, i.e. of the Python that ran PyInstaller (arm64 on Apple Silicon, x86_64 on
# Intel or under Rosetta: `arch -x86_64 python3 packaging/build.py ...`).
# Called by `python packaging/build.py --variant full|lite --package`.
set -euo pipefail

APP="${1:?usage: make_dmg.sh <VizcachaIDE.app> <output.dmg>}"
DMG="${2:?usage: make_dmg.sh <VizcachaIDE.app> <output.dmg>}"
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../.." && pwd)"

if ! command -v create-dmg >/dev/null 2>&1; then
  echo "create-dmg not found: brew install create-dmg" >&2
  exit 1
fi

STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT
# ditto keeps symlinks, extended attributes and the ad-hoc code signature.
ditto "$APP" "$STAGE/VizcachaIDE.app"
# The GPLv3 distribution notice is shown next to the app in the disk image.
{ cat "$REPO/packaging/NOTICE.md"; printf '\n\n'; cat "$REPO/LICENSE"; } > "$STAGE/LICENSE-NOTICE.txt"

VOLICON="$(find "$APP/Contents/Resources" -maxdepth 1 -name '*.icns' | head -n 1)"
EXTRA=()
if [[ -n "$VOLICON" ]]; then EXTRA+=(--volicon "$VOLICON"); fi
# On headless CI there is no Finder session to arrange the window: skip that step.
if [[ -n "${CI:-}" ]]; then EXTRA+=(--skip-jenkins); fi

mkdir -p "$(dirname "$DMG")"
rm -f "$DMG"
create-dmg \
  --volname "VizcachaIDE" \
  --window-pos 200 120 \
  --window-size 640 400 \
  --icon-size 96 \
  --icon "VizcachaIDE.app" 160 190 \
  --hide-extension "VizcachaIDE.app" \
  --icon "LICENSE-NOTICE.txt" 320 320 \
  --app-drop-link 480 190 \
  ${EXTRA[@]+"${EXTRA[@]}"} \
  "$DMG" "$STAGE"

echo "dmg: $DMG ($(uname -m) host, app arch: $(lipo -archs "$APP/Contents/MacOS/VizcachaIDE" 2>/dev/null || echo unknown))"
