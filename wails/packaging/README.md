# Packaging the Wails variant

Variants (what goes under `toolchain/`, next to the executable):

| Variant | Bundles | Notes |
|---|---|---|
| `lite` | nothing | uses the Go / Python installed on the system |
| `full-go` | Go, Delve, gopls | the old Go-only `full` (renamed) |
| `full-python` | CPython 3.12, debugpy, python-lsp-server (+ pyflakes), ruff | Python only |
| `full` | all of the above | nothing else to install |

All pins live in `packaging/versions.toml`. The Go part is the PyQt one: `build_release.py` imports
`packaging/fetch_toolchain.py` and `packaging/go_tools.py`; the Python part is
`packaging/fetch_python.py` (see "How Python is bundled"). The rename changes the file names of the
Go-only packages: `full` now means Go + Python. Stages live in `wails/dist/stage/<os>-<arch>/{go,python}`
and `full` is merged into `.../full` (licenses and `VERSIONS.txt` merged).

```bash
python wails/packaging/build_release.py --variant all                      # the four variants
python wails/packaging/build_release.py --variant both                     # lite + full-go (default)
python wails/packaging/build_release.py --variant full-python --target windows/amd64
python wails/packaging/build_release.py --variant lite,full --target darwin/arm64
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

## How Python is bundled

`packaging/fetch_python.py` (also usable alone: `python packaging/fetch_python.py --os windows --arch amd64 --dest <stage>`):

1. Downloads the `install_only` archive of **python-build-standalone** (astral-sh) pinned in
   `[python]` of `versions.toml` (CPython 3.12.15, release 20261003), verifies its sha256 (values copied
   from the release's `SHA256SUMS`, never invented) and extracts it to `toolchain/python/`. It reuses
   `download()` and `extract_archive()` of `fetch_toolchain.py`.
2. Prunes `include`, `Lib/test`, `idlelib`, `turtledemo`, `lib2to3`, `*.pdb` and every `__pycache__`
   (list in `[python].prune`). **tkinter (and turtle) is kept**: it is cheap (about 8 MB on disk,
   less than 3 MB compressed) and turtle graphics are a classic beginner exercise. `pip` and `venv`
   stay too (Packages, `.venv`).
3. `pip download --only-binary=:all: --platform <tag> --python-version 3.12` for the target, then
   `pip install --no-index --find-links <cache>/wheels/<target> --target <site-packages>`, so a
   Windows host can stage the macOS and Linux bundles. `[python.wheels]` pins debugpy,
   python-lsp-server (extra `pyflakes`, also pinned explicitly: without it pylsp publishes no diagnostics) and ruff; their dependencies come along (jedi, black, ...).
   Platform tags per target: `[python.pip_platforms]`.
4. Copies the licenses of CPython and of every installed wheel to `toolchain/licenses/python-*` and
   merges `python = ...` and the wheel versions into `toolchain/VERSIONS.txt`.

Layout (what `wails/internal/adapters/python/locator.go` expects): `toolchain/python/python.exe` on
Windows, `toolchain/python/bin/python3` elsewhere. NSIS and the dmg did not change: everything is
under `toolchain/`, and on macOS the existing `Contents/MacOS/toolchain` symlink covers Python too.

Smoke test of the bundled interpreter (no app needed; CI also passes it together with the exe):

```bash
python wails/packaging/smoke_test.py --python <stage>/toolchain/python/python.exe
# imports debugpy, pylsp and ruff (prints ok) and runs "python -m ruff --version"
```

Measured sizes, Windows amd64 (2026-10-03): staged `toolchain/python` about 200 MB on disk; the `full-python` portable zip (IDE + Python) is 68.2 MB, of which the toolchain alone is about 60 MB at zip level 9 (target 50 to 70 MB). The NSIS `-setup.exe` of `full-python` reached the LZMA stage (223 MB of install data) but its final size was not measured (solid LZMA is very slow on a loaded machine).

macOS and Linux bundles use the same code but were not run on those systems (CI only).

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
