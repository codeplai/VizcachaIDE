## VizcachaIDE 2.1.0 — ready for more languages

VizcachaIDE is a beginner-friendly IDE in English and Spanish, inspired by Thonny. Version 2.1
rebuilds the inside of the IDE around **language profiles**: Go works exactly as before, and
Python and C++ can now be added without changing the rest of the IDE. *(Notas en español:
[notes-2.1.0.es.md](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.1.0/docs/release/notes-2.1.0.es.md).)*

![Settings, Tools tab, with the Go tools found](https://raw.githubusercontent.com/codeplai/VizcachaIDE/wails-v2.1.0/docs/wails/qa-2.1/persist-02-tools.png)

### What's new
- **New file of…** in the File menu: Go, Python or C++, each with its own starter template.
  Settings lets you choose the language of new files.
- **Tools by language.** Settings → Tools shows each language's tools, where they were found and
  their version.
- **Clearer "tool missing" notices.** The notice names the missing tool and offers *Install* or
  *Copy the command*, plus *Choose in Settings*.
- **Lighter in the background.** The code helper (gopls) stops after 5 minutes without open files
  and starts again by itself.

### Same as before for Go
Run with keyboard input and arguments, Stop, errors explained in your language, suggestions and
live problems, gofmt on save, Go modules and the real debugger all behave as in 2.0. We checked it
with 34 automatic steps in English and Spanish.

### Good to know
- **Your settings are kept.** The tool paths of a 2.0 `settings.json` are converted the first time
  2.1 starts.
- **Python and C++ are not runnable yet.** You can create and edit their files, but running one
  says it isn't available yet. Their support is the next step.
- The "Go modules" dialog is now called **Packages**. For Go it has the same actions.

### Downloads

| File | Contents |
|---|---|
| `VizcachaIDE-2.1.0-windows-amd64-full-setup.exe` | Windows installer: IDE + Go, Delve and gopls |
| `VizcachaIDE-2.1.0-windows-amd64-lite-setup.exe` | Windows installer: IDE only (uses the Go on your system) |
| `VizcachaIDE-2.1.0-windows-amd64-{full,lite}-portable.zip` | Windows, no installation needed |
| `VizcachaIDE-2.1.0-darwin-{arm64,amd64}-{full,lite}.dmg` | macOS (**preview**, not notarized) |
| `VizcachaIDE-2.1.0-linux-amd64-{full,lite}.{AppImage,tar.gz}` | Linux (**preview**) |
| `SHA256SUMS-*.txt` | Checksums of every file |

**Which one?** If you are new to Go, take **full**. If you already have Go, **lite** is much
smaller. This release was tested on Windows 10; macOS and Linux builds come from CI and have not
been tried by hand.

See the [CHANGELOG](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.1.0/CHANGELOG.md) for the
complete list.
