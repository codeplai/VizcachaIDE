# Packaging the Wails variant

Variants: **lite** (IDE only, uses the Go on the user's system) and **full** (adds Go, Delve and
gopls, pinned in `packaging/versions.toml`). The toolchain code is the PyQt one:
`build_release.py` imports `packaging/fetch_toolchain.py` and `packaging/go_tools.py`.

```bash
python wails/packaging/build_release.py --variant both                     # host platform
python wails/packaging/build_release.py --variant full --target darwin/arm64
python wails/packaging/build_release.py --cache-dir /path/to/packaging/cache   # reuse downloads
```

Output: `wails/dist/release/VizcachaIDE-<version>-<os>-<arch>-<variant>…` + `SHA256SUMS-*.txt`.
Version = `info.productVersion` in `wails/wails.json` (`2.1.0`). Wails cannot cross-compile
with CGO targets, so each OS/arch is built on its own machine (see `.github/workflows/wails-release.yml`,
tags `wails-v*`).

| Platform | Artifacts |
|---|---|
| Windows amd64 | NSIS `-setup.exe` (per user, EN/ES, WebView2 bootstrapper, license page) and `-portable.zip` |
| macOS arm64 / amd64 | `.dmg` containing `VizcachaIDE.app` (ad-hoc signed, not notarized) |
| Linux amd64 | `.AppImage` and `.tar.gz` |

## Where the locator finds the toolchain

The Go runner (`internal/adapters/golang/runner`, through `internal/protocol/toollocator`) looks in
`filepath.Dir(os.Executable())/toolchain/{go/bin,bin}`:

* Windows / Linux tar.gz: next to the executable.
* AppImage: `os.Executable()` is `<mount>/usr/bin/VizcachaIDE`, so the toolchain is in `usr/bin/toolchain`.
* macOS: `os.Executable()` is `VizcachaIDE.app/Contents/MacOS/vizcacha`, so the locator looks in
  `Contents/MacOS/toolchain`. The real files live in `Contents/Resources/toolchain` and
  `Contents/MacOS/toolchain` is a relative symlink to them (as in the PyQt packaging; `codesign`
  seals Resources). **No Go change is needed.**

## Windows installer (NSIS)

* Template: `wails/build/windows/installer/project.nsi` (per user, `$LOCALAPPDATA\Programs\VizcachaIDE`,
  Start-menu shortcut, English + Spanish chosen from the Windows language, license page, WebView2
  bootstrapper through Wails' own macro). The full variant is the same script run again with
  `-DARG_TOOLCHAIN_DIR=<stage>` (adds `toolchain\`) and `-DARG_OUTFILE`.
* Silent install / uninstall: `setup.exe /S /D=C:\Some\Dir` (`/D` last, unquoted) and `uninstall.exe /S`.
  The NSIS uninstaller relaunches itself from `%TEMP%` and returns immediately: poll for the folder.
* NSIS without admin: `winget install NSIS.NSIS --scope user` fails (the package has no user scope).
  Use the official zip (`nsis-3.10.zip` from SourceForge), unzip it anywhere and put it on `PATH` or
  set `NSIS_HOME`. CI uses `choco install nsis`.
* The WebView2 bootstrapper (`tmp\MicrosoftEdgeWebview2Setup.exe`) is downloaded by `wails build -nsis`.
  Installing the runtime itself may need elevation on PCs that lack it (Windows 10 21H2+/11 have it).
* QA: `python wails/packaging/qa/install_qa.py --setup … --zip …`.

## Linux: why also a tar.gz

WebKitGTK cannot be bundled in an AppImage (its helper processes and GIO/GStreamer modules use
absolute paths), so both formats need `libwebkit2gtk-4.1-0` from the host (Ubuntu 22.04+,
Debian 12+, Fedora 38+). The AppImage checks for it and explains how to install it. The tar.gz
is the fallback for hosts without FUSE.

## Not signed

No Authenticode, Developer ID or notarization yet (pending certificates). Hence SmartScreen /
Gatekeeper warnings; the release notes explain how to proceed.
