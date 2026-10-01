# VizcachaIDE - Beginner-Friendly Go IDE

A simple, beginner-friendly IDE for Go programming, inspired by Thonny. VizcachaIDE provides an intuitive interface for learning and writing Go code with integrated debugging capabilities.

> **Project status (v0.2 in development).** The code base is being restructured for a
> public, bilingual (English / Español) 1.0 release on Windows, macOS and Linux. Editing,
> running programs and the console work today. **The debugger is being rebuilt on top of
> Delve**: in 0.1 it was only a simulation, so it is disabled until the real one lands.
> See [docs/PLAN_DESARROLLO.md](docs/PLAN_DESARROLLO.md) (Spanish) for the roadmap.

![VizcachaIDE](screenshot.png)

## Features

### 🎨 Simple and Clean Interface
- Qt5-based GUI with minimal, uncluttered design
- Single-window interface with integrated components
- Beginner-friendly layout

### 📝 Code Editor
- **Syntax highlighting** for Go keywords, types, and built-in functions
- **Line numbers** for easy reference
- **Auto-indentation** for clean code formatting
- **Bracket matching** to track code blocks
- **Code folding** for better organization
- **Breakpoint support** - click on line numbers to set/remove breakpoints
- **Tabbed interface** - work with multiple files simultaneously
- **Intelligent autocomplete** (Ctrl+Space) - suggestions for:
  - Go keywords (for, func, if, etc.)
  - Built-in types (int, string, bool, etc.)
  - Built-in functions (append, len, make, etc.)
  - Standard library packages (fmt, strings, os, etc.)
  - Package methods with documentation (fmt.Println, strings.Contains, etc.)
- **Documentation tooltips** - see brief descriptions for each suggestion

### ▶️ Execution Features
- **Run button** (F5) to execute Go code instantly
- **Stop button** (Shift+F5) to terminate running programs
- Real-time output display in integrated console
- Color-coded output (normal output, errors, success messages)

### 🐛 Debugging Capabilities
- **Step-by-step execution** powered by Delve debugger
- **Step Over** (F7) - Execute current line and move to next
- **Step Into** (F8) - Step into function calls
- **Step Out** (F9) - Step out of current function
- **Breakpoints** (F10) - Pause execution at specific lines
- **Current line highlighting** during debugging
- Integration with Delve (Go's official debugger)

### 🔍 Variable Inspector
- **Real-time variable viewer** during execution
- Display variable names, types, and values
- **Expandable view** for complex data structures:
  - Structs
  - Slices
  - Maps
  - Arrays
- Updates automatically as code executes

### 📊 Call Stack Viewer
- View the current call stack during debugging
- Track function call hierarchy
- Navigate through stack frames

### 💻 Output Console
- Integrated output panel for program output
- **Separate colors** for stdout and stderr
- Dark theme for comfortable reading
- Clear button to reset console

### 📁 File Operations
- **New** (Ctrl+N) - Create a new Go file
- **Open** (Ctrl+O) - Open existing .go files
- **Save** (Ctrl+S) - Save current file
- **Save As** (Ctrl+Shift+S) - Save with new name
- **Recent files** list
- Auto-restore last opened file

### ⚙️ Configuration & Settings
- **Configure dialog** (Run → Configure) with three tabs:
  - **Environment Tab**:
    - Custom Go executable path
    - GOPATH configuration
    - GOROOT configuration
    - Delve debugger path
    - Additional environment variables
  - **Editor Tab**:
    - Font family selection (Consolas, Courier New, Monaco, etc.)
    - Font size customization (8-24pt)
    - Tab size configuration
    - Auto-indentation toggle
    - Line numbers toggle
    - Word wrap toggle
  - **Appearance Tab**:
    - Editor theme (Light, Dark, Solarized Light/Dark)
    - Console theme (Light/Dark)
    - Custom background and text colors
    - Toolbar visibility
    - Status bar visibility
- Settings are automatically saved and restored

### 🎯 Beginner-Friendly Features
- Clear error messages with line numbers
- Highlight lines where errors occur
- Simple project structure (single files or simple projects)
- Easy configuration through GUI
- Intuitive keyboard shortcuts

## Requirements

### System Requirements
- Python 3.7 or higher
- Go 1.16 or higher
- PyQt5
- Delve debugger (for debugging features)

### Installation

1. **Clone or download this repository:**
   ```bash
   cd GoIDE
   ```

2. **Install Python dependencies:**
   ```bash
   pip install -r requirements.txt
   ```

3. **Install Go** (if not already installed):
   - Download from [https://golang.org/dl/](https://golang.org/dl/)
   - Follow installation instructions for your platform

4. **Install Delve debugger** (for debugging features):
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

   Make sure `dlv` is in your PATH:
   - On Windows: Add `%USERPROFILE%\go\bin` to PATH
   - On Linux/Mac: Add `~/go/bin` to PATH

## Usage

### Starting VizcachaIDE

Run the application:
```bash
python main.py
```

### Writing Your First Program

1. Click **File → New** or press **Ctrl+N**
2. Write your Go code:
   ```go
   package main

   import "fmt"

   func main() {
       fmt.Println("Hello, VizcachaIDE!")
   }
   ```
3. Click **File → Save** or press **Ctrl+S**
4. Save with a `.go` extension (e.g., `hello.go`)

### Running Your Program

1. Click the **▶ Run** button or press **F5**
2. View output in the console panel at the bottom
3. Click **⬛ Stop** or press **Shift+F5** to terminate if needed

### Using Autocomplete

1. Start typing any Go keyword, function, or package name
2. Press **Ctrl+Space** to trigger autocomplete
3. A popup will show:
   - Matching suggestions with icons
   - Function signatures
   - Brief documentation for each item
4. Navigate with **Up/Down** arrows
5. Press **Enter** to insert the selected completion
6. Press **Esc** to cancel

**Examples:**
- Type `fm` and press Ctrl+Space → see `fmt` package
- Type `fmt.` and press Ctrl+Space → see all fmt methods like `Println`, `Printf`, etc.
- Type `fo` and press Ctrl+Space → see `for` keyword
- Type `ap` and press Ctrl+Space → see `append` function with its signature

### Configuring VizcachaIDE

Access settings via **Run → Configure** to customize your IDE:

#### Environment Tab
- **Go Path**: Specify custom Go executable location
  - Leave empty to use system PATH
  - Example: `C:\Go\bin\go.exe` (Windows) or `/usr/local/go/bin/go` (Linux/Mac)
- **GOPATH**: Set your Go workspace directory
- **GOROOT**: Set Go installation directory (optional)
- **Delve Path**: Specify debugger executable location
- **Environment Variables**: Add custom variables (one per line)
  ```
  GOOS=linux
  GOARCH=amd64
  CGO_ENABLED=0
  ```

#### Editor Tab
- **Font Family**: Choose from monospace fonts (Consolas, Monaco, etc.)
- **Font Size**: Adjust editor text size (8-24pt)
- **Tab Size**: Set indentation width (2-8 spaces)
- **Auto-indent**: Toggle automatic indentation
- **Line Numbers**: Show/hide line numbers
- **Word Wrap**: Enable text wrapping

#### Appearance Tab
- **Editor Theme**: Light, Dark, Solarized Light/Dark
- **Console Theme**: Light or Dark
- **Custom Colors**: Pick custom background and text colors
- **Toolbar**: Show/hide toolbar
- **Status Bar**: Show/hide status bar

Click **Apply** to save settings without closing, or **OK** to save and close.

### Debugging Your Program

1. **Set breakpoints**: Click on line numbers where you want to pause execution
2. Click **🐛 Debug** button or press **F6** to start debugging
3. Use debug controls:
   - **Step Over** (F7): Execute current line and move to next
   - **Step Into** (F8): Enter into function calls
   - **Step Out** (F9): Exit current function
4. Watch variables update in real-time in the **Variables** panel
5. Monitor the call stack in the **Call Stack** panel
6. The current execution line is highlighted in yellow

### Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| New File | Ctrl+N |
| Open File | Ctrl+O |
| Save | Ctrl+S |
| Save As | Ctrl+Shift+S |
| **Autocomplete** | **Ctrl+Space** |
| Run | F5 |
| Stop | Shift+F5 |
| Start Debugging | F6 |
| Step Over | F7 |
| Step Into | F8 |
| Step Out | F9 |
| Toggle Breakpoint | F10 |
| Undo | Ctrl+Z |
| Redo | Ctrl+Y |
| Cut | Ctrl+X |
| Copy | Ctrl+C |
| Paste | Ctrl+V |

## Project Structure

VizcachaIDE follows Clean Architecture. Details and diagrams: [docs/ARQUITECTURA.md](docs/ARQUITECTURA.md)
and [docs/PLAN_DESARROLLO.md](docs/PLAN_DESARROLLO.md).

```
vizcacha/
├── domain/          # Pure Python model (no Qt): debugging, diagnostics, project...
├── application/     # Ports (interfaces) and use cases
├── infrastructure/  # Adapters: Go toolchain, settings, static completion...
├── i18n/            # English / Spanish catalogs (gettext + Babel)
└── ui/              # Qt widgets; features plug in through the Workbench
tests/               # pytest + pytest-qt
```

## Development

```bash
python -m venv .venv
.venv/Scripts/activate            # Windows  (Linux/macOS: source .venv/bin/activate)
pip install -r requirements-dev.txt

python -m vizcacha                 # run the IDE (python main.py also works)
python -m pytest                   # tests (Go integration tests run if `go` is on PATH)
ruff check . && lint-imports       # style + architecture contracts
```

Translations: write UI text in English inside `_()`, then

```bash
pybabel extract -F babel.cfg -o vizcacha/i18n/locale/vizcacha.pot .
pybabel update -i vizcacha/i18n/locale/vizcacha.pot -d vizcacha/i18n/locale -D vizcacha
pybabel compile -d vizcacha/i18n/locale -D vizcacha
```

## How It Works

### Code Execution
VizcachaIDE uses `go run` to compile and execute your Go programs. The output is captured and displayed in the integrated console with real-time updates.

### Debugging
The debugger integrates with Delve, Go's official debugger. When you start a debug session:
1. Delve compiles your program with debug symbols
2. Breakpoints are set at specified lines
3. You can step through code line by line
4. Variables are inspected at each step
5. The call stack is updated as you navigate

### Syntax Highlighting
The editor uses Pygments with custom highlighting rules for Go:
- **Keywords** (blue, bold): `func`, `var`, `if`, `for`, etc.
- **Types** (teal, bold): `int`, `string`, `bool`, etc.
- **Built-ins** (purple): `append`, `len`, `make`, etc.
- **Strings** (green): String literals
- **Numbers** (orange): Numeric literals
- **Comments** (gray, italic): Single and multi-line comments

## Customization

### Changing Editor Font
Edit `gui/editor.py`, line ~155:
```python
font = QFont("YourPreferredFont", 11)
```

### Changing Console Colors
Edit `gui/console.py` to customize the color scheme:
```python
palette.setColor(QPalette.Base, QColor("#YourBackgroundColor"))
palette.setColor(QPalette.Text, QColor("#YourTextColor"))
```

### Adding More Syntax Highlighting
Edit `gui/editor.py` in the `GoSyntaxHighlighter` class to add new highlighting rules.

## Troubleshooting

### "Delve debugger not found"
- Install Delve: `go install github.com/go-delve/delve/cmd/dlv@latest`
- Add Go's bin directory to your PATH
- Restart your terminal/IDE

### "Failed to start Go"
- Make sure Go is installed: `go version`
- Verify Go is in your PATH
- On Windows, you may need to restart after installing Go

### Syntax highlighting not working
- Make sure Pygments is installed: `pip install Pygments`
- Check the requirements.txt file

### Program won't run
- Save your file first (File → Save)
- Make sure the file has a `.go` extension
- Verify your code has a `main` package and `main()` function

## Future Enhancements

Potential features for future versions:
- [ ] Full Delve JSON-RPC API integration for advanced debugging
- [ ] Code completion using Go's language server protocol
- [ ] Go modules support
- [ ] Package management interface
- [ ] Integrated terminal
- [ ] Multiple file/project support
- [ ] Themes and color scheme options
- [ ] Code formatting with `gofmt`
- [ ] Linting with `golint` or `staticcheck`
- [ ] Testing interface for Go tests
- [ ] Git integration

## Contributing

Contributions are welcome! This project is designed to be beginner-friendly and extensible.

### Areas for Contribution
- Full Delve API integration for real debugging (currently simplified)
- Code completion and IntelliSense
- Additional themes and color schemes
- Better error highlighting and messages
- Multi-file project support
- Testing framework integration

## License

This project is open source. Feel free to use, modify, and distribute as needed.

## Credits

- Inspired by [Thonny](https://thonny.org/), the beginner-friendly Python IDE
- Built with [PyQt5](https://www.riverbankcomputing.com/software/pyqt/)
- Syntax highlighting powered by [Pygments](https://pygments.org/)
- Debugging powered by [Delve](https://github.com/go-delve/delve)

## Author

Created for beginners learning Go programming.

---

**Happy Coding! 🚀**
