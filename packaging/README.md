# Packaging VizcachaIDE

Everything needed to turn the source tree into Windows, macOS and Linux downloads.
*(Resumen en español al final.)*

## Variants

| Variant | Contents | Typical size (Windows x64) |
|---|---|---|
| **lite** | The IDE only. Uses the Go found on the user's system (`PATH` or Settings). | ~69 MB folder, ~29 MB portable zip |
| **full** | The IDE + Go, Delve (`dlv`) and `gopls`, pinned in [`versions.toml`](versions.toml). Nothing else to install. | ~307 MB folder, ~112 MB portable zip |

In the **full** variant the tools live next to the executable (convention agreed with track E):

```
VizcachaIDE/                     (macOS: VizcachaIDE.app/Contents/MacOS/)
├── VizcachaIDE[.exe]
├── _internal/                   PyInstaller runtime, logo, examples/, locale/*.mo, licenses
└── toolchain/
    ├── go/bin/go[.exe], gofmt[.exe]   GOROOT
    ├── bin/dlv[.exe], gopls[.exe]
    ├── licenses/                      Delve and gopls licenses (Go's is in go/LICENSE)
    └── VERSIONS.txt
```

On macOS the real folder is `Contents/Resources/toolchain` and `Contents/MacOS/toolchain` is a
symlink to it (keeps `codesign` happy). The app is re-signed ad hoc after the copy.

## Requirements

* Python **3.11+** for the packaging scripts (they use `tomllib`); the app itself runs on 3.10+.
* `pip install -r packaging/requirements-build.txt` (PyQt5, PyInstaller; Pillow on macOS).
* **full** only: a host `go` (any 1.21+) to cross-build `dlv`/`gopls`. `GOTOOLCHAIN` is set to
  the pinned Go version, so the host go downloads that exact compiler once from proxy.golang.org.
  Network access to `dl.google.com`, `proxy.golang.org` and `sum.golang.org` is required.
* Windows installer: [Inno Setup 6.3+](https://jrsoftware.org/isdl.php) (`choco install innosetup`).
* macOS dmg: `brew install create-dmg`.
* Linux AppImage: `curl` (appimagetool is downloaded to `packaging/cache/`, or set `APPIMAGETOOL`).

## Building

```bash
python packaging/build.py --variant lite              # -> dist/lite/VizcachaIDE[.app]
python packaging/build.py --variant full              # -> dist/full/VizcachaIDE[.app]
python packaging/build.py --variant full --package    # + dist/release/<installers>
python packaging/build.py --variant full --skip-fetch # reuse packaging/cache/stage/<os-arch>
python packaging/smoke_test.py dist/full/VizcachaIDE/VizcachaIDE.exe --seconds 6
```

PyInstaller cannot cross-compile: each OS/architecture is built on a matching machine
(that is what the release workflow does). The toolchain itself *can* be prepared for any target:

```bash
python packaging/fetch_toolchain.py --os darwin --arch arm64 --dest build/stage-mac
```

Outputs in `dist/release/` are named `VizcachaIDE-<version>-<platform>-<variant>…`:

| Platform | Artifacts | Script |
|---|---|---|
| Windows x64 | `…-setup.exe` (Inno Setup) and `…-portable.zip` | [`windows/vizcacha.iss`](windows/vizcacha.iss) |
| macOS arm64 / x86_64 | `….dmg` | [`macos/make_dmg.sh`](macos/make_dmg.sh) |
| Linux x86_64 | `….AppImage` | [`linux/make_appimage.sh`](linux/make_appimage.sh), [`linux/vizcacha.desktop`](linux/vizcacha.desktop) |

### Windows installer

Per-user installation by default (no administrator rights, `%LOCALAPPDATA%\Programs`); the first
page offers *Install for all users*. Setup languages: English and Spanish (auto-detected).
Optional tasks: desktop shortcut and "Open .go files with VizcachaIDE". The license page shows
[`NOTICE.md`](NOTICE.md) followed by the MIT license. Full and lite share the same AppId, so one
replaces the other. Silent install: `…-setup.exe /VERYSILENT /CURRENTUSER` (or `/ALLUSERS`).

### macOS

The architecture of the `.app` is that of the Python running PyInstaller. PyPI ships PyQt5 5.15.11
wheels for **both** `macosx_11_0_arm64` and `macosx_11_0_x86_64` (PyQt5-Qt5 5.15.19 and PyQt5-sip
as `universal2`), so we build two native apps: arm64 on `macos-14` and x86_64 on an Intel runner.
To build x86_64 on Apple Silicon, use an x86_64 Python under Rosetta 2:
`arch -x86_64 /usr/local/bin/python3 packaging/build.py --variant full --package`.
The app is ad-hoc signed only; Developer ID signing and notarization are future work (users must
right-click → Open on first launch).

### Linux

The AppImage contains the onedir folder in `usr/lib/vizcacha`. It needs the usual X11/Wayland Qt
libraries on the host (`libxkbcommon-x11-0`, `libxcb-*`, `libegl1`, `libfontconfig1`). The glibc
baseline is that of the build machine (Ubuntu 22.04 → glibc 2.35).

## Updating pinned versions

1. Edit [`versions.toml`](versions.toml) (`go.version`, `tools.dlv.version`, `tools.gopls.version`).
2. Regenerate the Go hashes: `python packaging/print_go_hashes.py 1.25.x` and paste the output.
   Hashes come from <https://go.dev/dl/?mode=json&include=all>.
3. gopls ≥ v0.22 needs Go 1.26 to build; keep gopls at v0.21.x while Go is pinned to 1.25.
4. `rm -rf packaging/cache/tools` (or bump versions: the cache is keyed by version).

Downloads are verified against the sha256 in `versions.toml`; dlv/gopls go through the Go checksum
database. `packaging/cache/` is git-ignored (the Go module cache inside is read-only: delete it with
`go clean -modcache` using `GOPATH=packaging/cache/gopath`, or `chmod -R u+w` first).

## Release workflow

[`.github/workflows/release.yml`](../.github/workflows/release.yml) runs on tags `v*`: matrix
`windows-latest`, `macos-14` (arm64), `macos-13` (x86_64), `ubuntu-22.04`; builds lite and full,
smoke-tests both headless, uploads `dist/release/*` and creates a **draft** GitHub Release with
EN/ES notes and `SHA256SUMS.txt`. `workflow_dispatch` produces the artifacts without a release.

## ⚠ License notice

The VizcachaIDE **source code is MIT**. The **distributed binaries** also contain:

| Component | License |
|---|---|
| PyQt5 | **GPLv3** |
| Qt 5 | LGPLv3 |
| Python | PSF |
| Go (full) | BSD-3-Clause |
| Delve (full) | MIT |
| gopls (full) | BSD-3-Clause |

Because PyQt5 is GPLv3, **the distributed package as a whole is licensed under the GPLv3**.
This is shown on the installer's license page, inside the dmg (`LICENSE-NOTICE.txt`), and shipped
as `NOTICE.md` + `licenses/GPL-3.0.txt` in every package. It should also appear in *Help → About*
(requested from the integrator as a contract change). Migrating `ui/` to PySide6 (LGPL) would lift
this constraint.

---

## Resumen en español

* **Variantes:** `lite` (sólo el IDE; usa el Go del sistema) y `full` (incluye Go, Delve y gopls
  fijados en `versions.toml`, en `toolchain/` junto al ejecutable).
* **Construir:** `python packaging/build.py --variant lite|full [--package]`. Python 3.11+ para los
  scripts. PyInstaller no hace compilación cruzada: cada SO se construye en su propia máquina.
* **Instaladores:** Windows con Inno Setup (por usuario, opción para todos los usuarios, idiomas
  inglés/español, accesos directos y asociación `.go` opcionales) + zip portable; macOS `.dmg`
  con create-dmg (arm64 y x86_64 nativos: PyQt5 tiene wheels arm64 en PyPI); Linux AppImage.
* **Versiones:** se actualizan en `versions.toml`; los sha256 de Go se regeneran con
  `print_go_hashes.py`.
* **Publicación:** al hacer push de un tag `v*`, `release.yml` construye todo y crea un Release en
  borrador con notas en inglés y español.
* **Licencia:** el código es MIT, pero el binario distribuido incluye PyQt5 (GPLv3), así que el
  **paquete distribuido queda bajo la GPLv3**. Se muestra en el instalador y se incluye
  `NOTICE.md` y `licenses/GPL-3.0.txt`.
