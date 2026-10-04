## VizcachaIDE 2.2.0 — Python arrives

VizcachaIDE is a beginner-friendly IDE in English and Spanish, inspired by Thonny. Version 2.2 adds
**Python** as a second language, with the same experience Go already had. *(Notas en español:
[notes-2.2.0.es.md](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.2.0/docs/release/notes-2.2.0.es.md).)*

![Debugging a Python program that reads the keyboard](https://raw.githubusercontent.com/codeplai/VizcachaIDE/wails-v2.2.0/docs/wails/qa-2.2/es-py-05-debug-input.png)

### What's new
- **Run Python with F5.** Your program runs in a real terminal: `input()` reads what you type in
  Output, and what it prints before crashing is never lost.
- **Errors explained in your language.** About 30 common Python errors (a name that does not exist,
  wrong indentation, a missing colon, adding text and numbers, dividing by zero…) are explained with
  what happened, why, and how to fix it. The original message stays as Python wrote it.
- **A real debugger.** Breakpoints, step by step, variables that light up when they change, the
  call stack, a stop on the line of an uncaught error, and you can **type answers while debugging**.
- **Help while you write.** Problems underlined as you type, suggestions, documentation on hover and
  go to definition.
- **Format on save**, a **Python console** that remembers your variables, and a **Packages** dialog
  to install libraries with pip.
- **Choose your languages.** The first-run wizard asks which programming languages you will use;
  menus and Settings show only those.

### Downloads

| File | Contents |
|---|---|
| `VizcachaIDE-2.2.0-windows-amd64-full-setup.exe` | Windows installer: IDE + Go and Python, with their debuggers and code helpers |
| `VizcachaIDE-2.2.0-windows-amd64-full-python-setup.exe` | Windows installer: IDE + Python |
| `VizcachaIDE-2.2.0-windows-amd64-full-go-setup.exe` | Windows installer: IDE + Go (the old "full") |
| `VizcachaIDE-2.2.0-windows-amd64-lite-setup.exe` | Windows installer: IDE only (uses the Go or Python on your system) |
| `VizcachaIDE-2.2.0-windows-amd64-<variant>-portable.zip` | Windows, no installation needed |
| `VizcachaIDE-2.2.0-darwin-*`, `VizcachaIDE-2.2.0-linux-*` | macOS and Linux (**preview**) |
| `SHA256SUMS-*.txt` | Checksums of every file |

**Which one?** For a Python course take **full-python** (about 70 MB zipped on Windows); for Go and
Python, **full**. With **lite**, install Python 3.10 or newer and run
`python -m pip install debugpy "python-lsp-server[pyflakes]" ruff`. This release was tested on
Windows 10; macOS and Linux builds come from CI and have not been tried by hand.

See the [CHANGELOG](https://github.com/codeplai/VizcachaIDE/blob/wails-v2.2.0/CHANGELOG.md) for the
complete list.
