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

# The plain image: app + Applications link + license, made from the folder without mounting it.
plain_dmg() {
  ln -sfn /Applications "$STAGE/Applications"
  hdiutil create -volname "VizcachaIDE" -srcfolder "$STAGE" -ov -format UDZO "$DMG"
}

# create-dmg mounts a read-write image to lay it out, and on CI its detach can time out
# ("timeout for DiskArbitration expired"): detach what it left mounted and try again.
fancy_dmg() {
  local extra=()
  # On headless CI there is no Finder session to arrange the window: skip that step.
  if [[ -n "${CI:-}" ]]; then extra+=(--skip-jenkins); fi
  create-dmg \
    --volname "VizcachaIDE" \
    --window-pos 200 120 \
    --window-size 640 400 \
    --icon-size 96 \
    --icon "VizcachaIDE.app" 160 190 \
    --hide-extension "VizcachaIDE.app" \
    --icon "LICENSE-NOTICE.txt" 320 320 \
    --app-drop-link 480 190 \
    ${extra[@]+"${extra[@]}"} \
    "$DMG" "$STAGE"
}

cleanup_mounts() {
  local volume
  for volume in /Volumes/dmg.* /Volumes/VizcachaIDE*; do
    if [[ -d "$volume" ]]; then hdiutil detach -force "$volume" >/dev/null 2>&1 || true; fi
  done
  rm -f "$(dirname "$DMG")"/rw.*."$(basename "$DMG")" "$DMG"
}

if command -v create-dmg >/dev/null 2>&1; then
  made=""
  for attempt in 1 2 3; do
    if fancy_dmg; then made=1; break; fi
    echo "create-dmg failed (attempt $attempt of 3): detaching and retrying"
    cleanup_mounts
    sleep $((attempt * 10))
  done
  if [[ -z "$made" ]]; then
    echo "create-dmg kept failing: making the plain image with hdiutil"
    plain_dmg
  fi
else
  echo "create-dmg not found: falling back to hdiutil"
  plain_dmg
fi

echo "dmg: $DMG (app arch: $(lipo -archs "$APP/Contents/MacOS/"* 2>/dev/null | head -n 1 || echo unknown))"
