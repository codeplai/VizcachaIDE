# Release QA — VizcachaIDE 1.0

## Resumen (español)

QA de release ejecutado en **Windows 10 Pro 22H2 (x64)** el **2026-10-01** sobre `main` @ `bd5b9e3`
(versión del paquete `0.2.0.dev0`).

- **Construido:** variantes `full` (Go 1.25.14, Delve 1.27.2 y gopls 0.21.1 incluidos) y `lite`, con sus
  zip portables. Tamaños y SHA256 más abajo.
- **Instalador `.exe` (Inno Setup): NO generado ni probado.** No se pudo instalar Inno Setup sin
  GitHub: winget y la web oficial descargan desde `github.com`, que está bloqueado en esta máquina.
  Se descartó usar un paquete no oficial (NuGet `Tools.InnoSetup`). Para cubrirlo hay
  `packaging/qa/install_qa.py --setup …`, listo para ejecutarse en GitHub Actions o en otra máquina.
- **QA funcional headless** con el workbench real (`build_workbench`, `QT_QPA_PLATFORM=offscreen`)
  usando la toolchain **incluida** en el bundle full y sin Go en el `PATH`:
  - **Inglés: 11/11.** Ejecución de hello y variables, 5 errores con el id correcto en el Assistant,
    depuración real con dlv (breakpoint en la línea 13, `result = 35`) y gofmt al guardar.
  - **Español: la funcionalidad pasa, pero el Assistant sigue en inglés.** Sus textos están solo en el
    borrador `docs/i18n/assistant.es.po` y no en `vizcacha.po` (pendiente del track G).
    **Esto bloquea un release bilingüe.**
- **Zip portable (full y lite):** arranca sin traceback. En full, el `go version` incluido y
  `go run hello.go` funcionan.
- **Firma de código:** no hay certificado. Windows mostrará SmartScreen ("Windows protegió su PC");
  en macOS, Gatekeeper bloquea la app hasta abrirla con clic derecho → Abrir. Ver §6.
- **Bugs y riesgos encontrados:** ver §5 (traducción del Assistant, ejemplos que aún dicen "GoIDE",
  primera ejecución lenta con la toolchain incluida y binarios de depuración que quedan en `%TEMP%`).

---

## 1. Environment

| Item | Value |
|---|---|
| Date | 2026-10-01 |
| OS | Windows 10 Pro 10.0.19045 x64, 16 GB RAM (heavily loaded: ~1 GB free during the run) |
| Commit | `bd5b9e3` (main) + this QA commit (packaging/ and docs/ only) |
| Python (build + QA) | 3.14.3 (`.venv`), PyInstaller 6.22.3, PyQt5 5.15 |
| Host Go (used only to build dlv/gopls) | 1.25.5 → `GOTOOLCHAIN=go1.25.14` |
| Network | PyPI, go.dev, dl.google.com and proxy.golang.org reachable; **github.com blocked** |

## 2. Artifacts built

Command: `python packaging/build.py --variant full --package`, then the same with `--variant lite`.
Wall-clock time: full 231 s (toolchain extraction + dlv/gopls compiled from proxy.golang.org +
PyInstaller), lite 85 s.

| Artifact | Size | SHA256 |
|---|---|---|
| `dist/full/VizcachaIDE/` (folder) | 313.8 MB | — |
| `dist/lite/VizcachaIDE/` (folder) | 75.7 MB | — |
| `VizcachaIDE-0.2.0.dev0-windows-x64-full-portable.zip` | 115,209,672 B (115.2 MB) | `3a0c1bd3f3cd785c27328989c21e8d47270c3be4b43d5e2ab3efb041a27fd514` |
| `VizcachaIDE-0.2.0.dev0-windows-x64-lite-portable.zip` | 32,416,009 B (32.4 MB) | `27f4bb35207e1049ddbe3fb6aac62d5f3dcae9458c09bc7655748cbf59c6a3fb` |
| `…-full-setup.exe` / `…-lite-setup.exe` | **not built** | — (no ISCC, see §4) |

These hashes belong to this local build; it is not reproducible bit for bit (PyInstaller embeds
timestamps), so the release must publish the `SHA256SUMS.txt` produced by `release.yml`.

Bundled toolchain (`toolchain/VERSIONS.txt`, checked by running the binaries): `go version go1.25.14
windows/amd64`, Delve `1.27.2`, `golang.org/x/tools/gopls v0.21.1`.

> Build note: the first attempt failed with `fatal error: runtime: cannot allocate memory` while
> compiling gopls, because other agents were using the machine's RAM. The rebuild worked with
> `GOFLAGS=-p=1 GOMAXPROCS=2 GOGC=50`. This is a property of the build machine, not of the app.

## 3. Results

### 3.1 Functional QA, headless (`packaging/qa/functional_qa.py`)

The script builds the real workbench (`vizcacha.ui.app.build_workbench`) offscreen and drives it
like a user would, through the menu actions. Each run works on copies in a temporary folder that
is deleted afterwards. With `--bundle`, `GoEnvironment` points to `dist/full/VizcachaIDE` and
**every PATH entry that contains a `go` is removed**, so passing proves that the bundled
toolchain is the one being used.

| Check | EN, bundled toolchain | ES, bundled toolchain | EN, system Go 1.25.5 |
|---|---|---|---|
| All 4 tools resolved as `bundled` inside the bundle; `go version` = go1.25.14 | PASS | PASS | n/a (PATH) |
| UI language (Run action text, empty Assistant text) | PASS | **FAIL**: Run is "▶ Ejecutar", but the empty Assistant text is still English | not run |
| Run `examples/hello.go` → exit 0 + expected output | PASS (2.7 s; **103.9 s on the very first run**, see B3) | PASS | PASS |
| Run `examples/variables.go` → `Count: 42` | PASS | PASS | PASS |
| Assistant id for `E-UNUSED-VAR` | PASS | PASS (id); **title not translated** | PASS |
| Assistant id for `E-UNDEFINED` | PASS | PASS (id); **title not translated** | PASS |
| Assistant id for `E-TYPE-MISMATCH` | PASS | PASS (id); **title not translated** | PASS |
| Assistant id for `P-INDEX-RANGE` (runtime panic) | PASS | PASS (id); **title not translated** | PASS |
| Assistant id for `P-DIVIDE-ZERO` (runtime panic) | PASS | PASS (id); **title not translated** | PASS |
| Debug `examples/functions.go` with real dlv: breakpoint at line 13, current line = 13, `result = 35`, Continue → exit 0 and `5 * 7 = 35` in the console; no debug binary left in `%TEMP%` | PASS (5.5 s) | PASS | PASS |
| gofmt on save (messy → canonical, on disk and in the editor) | PASS | PASS | PASS |
| **Total** | **11/11** | **10/16** (6 failures = Spanish translation) | 10/10 (earlier revision without the language checks) |

Startup time: `build_workbench` + `window.show()` takes 0.05 to 0.17 s (offscreen, from source).
The frozen executable stays up without a traceback in the 6 s smoke test. Measuring the time to
the first painted window of the frozen app needs a real display and is still pending (§7).

### 3.2 Portable zip (`packaging/qa/install_qa.py --zip`)

Each zip is extracted to a temporary folder, then checked as below.

| Check | full zip | lite zip |
|---|---|---|
| `VizcachaIDE.exe` alive after 6 s offscreen, no traceback, no error dialog | PASS (6.5 s) | PASS (6.4 s) |
| `toolchain/go/bin/go.exe version` → go1.25.14 (Go removed from PATH) | PASS | n/a |
| gofmt.exe, dlv.exe and gopls.exe present | PASS | n/a |
| `go run hello.go` with the bundled go + `GOROOT`/`GOTOOLCHAIN=local` (the app's environment) | PASS | n/a |
| No `toolchain/` folder | n/a | PASS |

### 3.3 Inno Setup installer: NOT TESTED

`install_qa.py --setup <setup.exe>` automates everything the task asks for, but it **has not been
run** because the installer could not be built:

1. Silent per-user install into a temporary folder:
   `/VERYSILENT /SUPPRESSMSGBOXES /CURRENTUSER /NORESTART /DIR=… /LOG=…`.
2. Start-menu shortcut present. No desktop shortcut and no `HKCU\Software\Classes\VizcachaIDE.go`
   by default, so the `.go` association is opt-in.
3. Launch smoke test, bundled `go version` and `go run hello.go` (same checks as the zip).
4. Reinstall with `/TASKS=associatego`: the ProgId key appears.
5. `unins000.exe /VERYSILENT`: the install folder is empty, and the shortcut and association are
   gone.

User settings are **kept by design** after uninstalling (QSettings in
`HKCU\Software\VizcachaIDE\VizcachaIDE`), like most Windows apps. The `.iss` file has no
`[UninstallDelete]` for them.

## 4. Why there is no installer, and how to get one

- `winget install --id JRSoftware.InnoSetup --scope user` fails with `0x80072ee2`: the package
  points to `github.com/jrsoftware/issrc/releases/...`.
- Since March 2026, jrsoftware.org only links to GitHub releases (`files.jrsoftware.org/is/6/`
  only has `.issig` signatures).
- A copy of ISCC exists on NuGet (`Tools.InnoSetup` 6.7.3, unofficial). Installing it was refused:
  it is not an official distribution channel.
- `packaging/installers.py` now also accepts **`ISCC=<path to ISCC.exe>`**, besides `PATH`,
  `Program Files (x86)\Inno Setup 6`, `Program Files\Inno Setup 6` and
  `%LOCALAPPDATA%\Programs\Inno Setup 6` (the per-user default of
  `innosetup-6.x.exe /CURRENTUSER /VERYSILENT`).

**To do (on a machine with GitHub, or in Actions on `windows-latest`, where Inno Setup 6 is
preinstalled):**

```powershell
innosetup-6.7.3.exe /CURRENTUSER /VERYSILENT /SUPPRESSMSGBOXES   # if not installed
python packaging/build.py --variant full --package
python packaging/build.py --variant lite --package
python packaging/qa/install_qa.py --setup dist/release/VizcachaIDE-<v>-windows-x64-full-setup.exe `
                                  --setup dist/release/VizcachaIDE-<v>-windows-x64-lite-setup.exe
```

Also check by hand once: the wizard in Spanish on a Spanish Windows, the license page (GPLv3
NOTICE + MIT) with correct accents, and "Install for all users" (needs admin).

## 5. Bugs and release risks found

| # | Severity | Area (owner) | Description |
|---|---|---|---|
| B1 | **Blocker for 1.0 (bilingual)** | i18n (track G) | With Spanish, the **Assistant is in English**: card titles, bodies and hints, and its empty text "Nothing to explain…". The translations are only in `docs/i18n/assistant.es.po` and not in `vizcacha/i18n/locale/es/LC_MESSAGES/vizcacha.po` (114 msgids). Menus are translated ("▶ Ejecutar"). |
| B2 | Low | examples | `examples/hello.go` prints `Hello, GoIDE!` and `examples/variables.go` uses `"GoIDE"`: old product name. `hello.go` also has a line with only trailing spaces, which gofmt removes when the user saves. |
| B3 | Medium (UX) | toolchain (track E) | The **first Run after installing `full` took 104 s** (later runs: about 2.5 s). The bundled GOROOT starts with a cold build cache, so `go run` compiles `fmt` and the runtime first. The machine was very loaded, so a normal PC will be faster, but beginners will think it hangs. The same applies to the first Debug session (`-gcflags all=-N -l`). Ideas: show "Preparing Go for the first time…" in the console, or warm the cache in the background on first start (`go build std`). |
| B4 | Low | debugger (track A) | `%TEMP%\vizcacha_debug_bin_<pid>.exe` files from earlier sessions (not this QA) were found in `%TEMP%`. A normal debug session cleans up (QA checks it: `leftover_debug_binary=[]`), but a crashed or killed IDE leaves its 2.3 MB binary behind. Idea: remove stale `vizcacha_debug_bin_*` at startup. |

B1 to B4 are not fixed here: QA does not touch `vizcacha/`.

**How to reproduce:**
- B1: `python packaging/qa/functional_qa.py --language es`, or by hand: Tools → Options → language
  Spanish, restart, then run `examples/errors/E-UNDEFINED/E-UNDEFINED.go`. The Assistant card
  says "Unknown name: total".
- B2: open `examples/hello.go` and press F5.
- B3: on a clean user profile (or after `go clean -cache` with the bundled go), install `full`,
  open `hello.go` and press F5. Time it until "Hello" appears.
- B4: start a debug session, kill VizcachaIDE from Task Manager while it is stopped at a
  breakpoint, then `dir %TEMP%\vizcacha_debug_bin_*`.

## 6. Code signing status

**There is no code-signing certificate.** Nothing is signed: not the PyInstaller exe, not the
Inno Setup installer and uninstaller, and not the macOS app (ad-hoc signed only). Go, dlv and
gopls are Google/Delve binaries compiled by us, so they are unsigned too.

| Platform | What users will see | What it takes |
|---|---|---|
| Windows | SmartScreen "Windows protected your PC" (More info → Run anyway) on the setup.exe and on the first launch. Some antivirus heuristics flag unsigned PyInstaller bootloaders. The warning fades only as download reputation builds up. | An Authenticode certificate. OV is about 200–400 USD/year, but it needs reputation. EV removes the SmartScreen warning at once. Azure Trusted Signing is about 10 USD/month for eligible orgs/individuals. Sign `VizcachaIDE.exe`, `toolchain/**/*.exe` and the installer: Inno `SignTool=` directive + `signtool sign /fd sha256 /tr <timestamp> /td sha256`. Run it in `release.yml` with the cert as a secret. |
| macOS | Gatekeeper: "VizcachaIDE can't be opened because Apple cannot check it for malicious software". The user must right-click → Open, or go to System Settings → Privacy & Security → Open Anyway (macOS 15 removed the right-click bypass for some cases). Quarantined downloads of the dmg are affected. | Apple Developer Program (99 USD/year). Developer ID Application certificate. `codesign --options runtime --timestamp`, deep-signing the bundled go/dlv/gopls. dlv needs the `com.apple.security.cs.debugger` entitlement and may need `disable-library-validation` for PyQt. Then `xcrun notarytool submit --wait` and `xcrun stapler staple` on the dmg. |
| Linux | Nothing blocks AppImages. Optionally publish a GPG signature / SHA256SUMS. | `gpg --detach-sign` of the AppImage and SHA256SUMS (optional). |

Until there is a certificate, the release notes must explain the SmartScreen/Gatekeeper steps in
English and Spanish, and publish SHA256 sums.

## 7. Release checklist (3 OS)

Legend: ✅ verified here, on Windows · ⏳ pending · 🤖 can be done by GitHub Actions
(`release.yml`, not run: GitHub is unreachable from this machine) · 👤 manual.

### Windows x64
- ✅ `build.py --variant full|lite --package` builds; zips created; sizes/hashes recorded (§2)
- ✅ zip full/lite: headless launch, no traceback; bundled go 1.25.14 + `go run` (§3.2)
- ✅ functional QA EN with bundled toolchain: run, Assistant ids, dlv debug, gofmt (§3.1)
- ⏳ **ES translation of the Assistant (B1)**: blocker
- ⏳ 🤖 Inno Setup installer built (full + lite), then `install_qa.py --setup` (§3.3)
- ⏳ 👤 installer wizard in ES/EN, license page, all-users mode (admin), upgrade lite→full (same AppId; `[InstallDelete]` removes the stale `toolchain/`)
- ⏳ 👤 real-display smoke test: first paint time, HiDPI 150 %, the dock layout of the Assistant
- ⏳ 👤 clean VM without Go (lite: Go missing → message/Settings; full: works offline)
- ⏳ Authenticode signing (§6)

### macOS (arm64 and x86_64)
- ⏳ 🤖 `macos-14` / `macos-13` runners: build full+lite, `make_dmg.sh`, smoke test
- ⏳ 👤 open the dmg on a clean Mac: Gatekeeper flow, app runs from /Applications
- ⏳ 👤 `functional_qa.py --bundle dist/full/VizcachaIDE.app/Contents/MacOS` (dlv needs the developer-tools permission / `DevToolsSecurity`)
- ⏳ `.go` double-click (needs `QFileOpenEvent`: deferred contract change)
- ⏳ Developer ID signing + notarization (§6)

### Linux x86_64
- ⏳ 🤖 `ubuntu-22.04`: build full+lite, AppImage, smoke test
- ⏳ 👤 run the AppImage on Ubuntu 24.04 and Fedora (Wayland + X11; `libxcb-*`, FUSE2)
- ⏳ 👤 `functional_qa.py --bundle <extracted AppImage>/usr/lib/vizcacha` (ptrace_scope for dlv)

### Release
- ⏳ 🤖 tag `v1.0.0` → draft GitHub Release with `SHA256SUMS.txt`, EN/ES notes (+ SmartScreen/Gatekeeper instructions), screenshots in both languages
- ⏳ 👤 check GPLv3 notice in installer/dmg and in Help → About

## 8. Commands to repeat this QA (Windows, PowerShell or Git Bash)

```bash
PY=.venv/Scripts/python          # Python 3.11+ with requirements-build.txt + pytest deps
# 1. Build (add GOFLAGS=-p=1 GOMAXPROCS=2 on machines short on RAM)
$PY packaging/build.py --variant full --package
$PY packaging/build.py --variant lite --package
sha256sum dist/release/*

# 2. Downloads: installer (needs ISCC) and portable zips
$PY packaging/qa/install_qa.py --setup dist/release/VizcachaIDE-<v>-windows-x64-full-setup.exe
$PY packaging/qa/install_qa.py --zip dist/release/VizcachaIDE-<v>-windows-x64-full-portable.zip \
                               --zip dist/release/VizcachaIDE-<v>-windows-x64-lite-portable.zip
$PY packaging/smoke_test.py dist/full/VizcachaIDE/VizcachaIDE.exe --seconds 6

# 3. Functional QA through the real workbench (offscreen), with the bundled toolchain
QT_QPA_PLATFORM=offscreen $PY packaging/qa/functional_qa.py --language en --bundle dist/full/VizcachaIDE
QT_QPA_PLATFORM=offscreen $PY packaging/qa/functional_qa.py --language es --bundle dist/full/VizcachaIDE
# ... or with the Go/dlv/gopls found on PATH
QT_QPA_PLATFORM=offscreen $PY packaging/qa/functional_qa.py --language en
```

Exit code 0 means every check passed. Every line is `[PASS]`/`[FAIL] <check> (<seconds>) <detail>`.
