#!/usr/bin/env bash
# Wrap VizcachaIDE.app in a drag-to-Applications .dmg.
#
#   bash wails/packaging/macos/make_dmg.sh <VizcachaIDE.app> <output.dmg>
#
# Uses create-dmg (`brew install create-dmg`) when present, otherwise plain hdiutil
# (no fancy window layout, but still app + Applications link + license notice).
# The architecture of the dmg is that of the .app (arm64 or x86_64).
set -euo pipefail

APP="${1:?usage: make_dmg.sh <VizcachaIDE.app> <output.dmg>}"
DMG="${2:?usage: make_dmg.sh <VizcachaIDE.app> <output.dmg>}"
HERE="$(cd "$(dirname "$0")" && pwd)"
REPO="$(cd "$HERE/../../.." && pwd)"

STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT
# ditto keeps symlinks (toolchain/), extended attributes and the ad-hoc code signature.
ditto "$APP" "$STAGE/VizcachaIDE.app"
{ cat "$REPO/wails/packaging/NOTICE.md"; printf '\n\n'; cat "$REPO/LICENSE"; } > "$STAGE/LICENSE-NOTICE.txt"

mkdir -p "$(dirname "$DMG")"
rm -f "$DMG"

if command -v create-dmg >/dev/null 2>&1; then
  EXTRA=()
  # On headless CI there is no Finder session to arrange the window: skip that step.
  if [[ -n "${CI:-}" ]]; then EXTRA+=(--skip-jenkins); fi
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
else
  echo "create-dmg not found: falling back to hdiutil"
  ln -s /Applications "$STAGE/Applications"
  hdiutil create -volname "VizcachaIDE" -srcfolder "$STAGE" -ov -format UDZO "$DMG"
fi

echo "dmg: $DMG (app arch: $(lipo -archs "$APP/Contents/MacOS/"* 2>/dev/null | head -n 1 || echo unknown))"
