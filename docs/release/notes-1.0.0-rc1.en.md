## VizcachaIDE 1.0.0-rc1 — first release candidate

VizcachaIDE is a beginner-friendly IDE for Go, inspired by Thonny, in English and Spanish.
This is the first **release candidate** of 1.0: please try it and report problems before the
final release. *(Notas en español: [notes-1.0.0-rc1.es.md](https://github.com/codeplai/VizcachaIDE/blob/v1.0.0-rc1/docs/release/notes-1.0.0-rc1.es.md).)*

![VizcachaIDE stopped at a breakpoint](https://raw.githubusercontent.com/codeplai/VizcachaIDE/v1.0.0-rc1/docs/images/debugger.en.png)

### Highlights
- **A real debugger based on Delve.** Breakpoints, step over / into / out, run to cursor,
  variables you can expand, call stack and goroutines. The simulated debugger of 0.1 is gone.
- **An Assistant that explains Go errors** in plain English or Spanish: 25 kinds of common compiler
  errors, runtime panics and `go vet` warnings, with a suggested fix, the original message and
  a link to its line.
- **Code intelligence with gopls:** completion (Ctrl+Space), errors underlined while you type,
  documentation on hover, Ctrl+click to go to a definition, parameter hints and an Outline panel.
- **A better editor:** themes, gofmt on save, find and replace, go to line, comment toggling,
  zoom and recent files.
- **Go modules, program arguments and keyboard input** for your programs, a Files panel, and a
  *Go Modules* dialog for `go mod init`, `go get` and `go mod tidy`.
- **"Full" packages with Go included**, so beginners do not have to install anything else.

See the [CHANGELOG](https://github.com/codeplai/VizcachaIDE/blob/v1.0.0-rc1/CHANGELOG.md) for
the complete list.

### Downloads

| File | Contents |
|---|---|
| `VizcachaIDE-<version>-windows-x64-full-setup.exe` | Windows installer: IDE + Go 1.25.14, Delve 1.27.2 and gopls 0.21.1 |
| `VizcachaIDE-<version>-windows-x64-lite-setup.exe` | Windows installer: IDE only (uses the Go on your system) |
| `VizcachaIDE-<version>-windows-x64-{full,lite}-portable.zip` | Windows, no installation needed |
| `VizcachaIDE-<version>-macos-arm64-{full,lite}.dmg` | macOS on Apple Silicon (**preview**) |
| `VizcachaIDE-<version>-macos-x86_64-{full,lite}.dmg` | macOS on Intel (**preview**) |
| `VizcachaIDE-<version>-linux-x86_64-{full,lite}.AppImage` | Linux x86_64 (**preview**) |
| `SHA256SUMS.txt` | Checksums of every file |
| `NOTICE.md` | License notice of the distributed program |

**Which one?** If you are new to Go, take **full**. If you already have Go installed, **lite** is
much smaller (about 30 MB instead of about 110 MB zipped on Windows). With lite, install Delve
and gopls for debugging and code intelligence:
`go install github.com/go-delve/delve/cmd/dlv@latest` and
`go install golang.org/x/tools/gopls@latest`.

### System requirements
- **Windows:** Windows 10 or 11, 64-bit (x64).
- **macOS:** macOS 11 Big Sur or newer, Apple Silicon or Intel. The app is not notarized: the
  first time, right-click it and choose **Open**.
- **Linux:** x86_64 with glibc 2.35 or newer (for example Ubuntu 22.04+) and the usual Qt
  libraries (`libxkbcommon-x11-0`, `libxcb-*`, `libegl1`, `libfontconfig1`). Make the file
  executable with `chmod +x` and run it.
- **lite only:** Go installed (1.21 or newer recommended; tested with 1.25).
- Disk space: about 70 MB (lite) or 310 MB (full) once installed on Windows.

### License
The VizcachaIDE source code is MIT. These binaries also bundle PyQt5, which is GPLv3, so **the
distributed program as a whole is licensed under the GPLv3**. The complete list of bundled
components and their licenses is in `NOTICE.md`. The GPL applies to the IDE, not to the Go
programs you write with it.

### Known limitations
- **macOS and Linux packages are a preview**: they are built by the release workflow but have
  not been tested on real machines yet. Please report anything that does not work.
- The Windows installer had not been tested before this release candidate (the application
  itself and the portable zip were).
- No code signing: Windows SmartScreen may warn about an unknown publisher.
- The Spanish translation is not complete yet, and the language follows the operating system
  (there is no selector in Options yet).
- A program being debugged cannot read keyboard input (stdin).
- The default layout of the panels is still being tuned; the Assistant panel can start small
  (drag its border or undock it).

### Thanks
To the projects that make VizcachaIDE possible: Thonny (the inspiration), Delve, gopls, Go,
PyQt5 and Qt.
