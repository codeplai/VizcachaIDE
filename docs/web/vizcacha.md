---
title: VizcachaIDE
description: The Go IDE for people just getting started. Write, run, understand your errors and debug step by step, in English or Spanish.
image: https://raw.githubusercontent.com/codeplai/VizcachaIDE/main/vizcachaidelogo.png
---

![VizcachaIDE by codeplai](https://raw.githubusercontent.com/codeplai/VizcachaIDE/main/vizcachaidelogo.png)

# VizcachaIDE

**The Go IDE for people just getting started.** Write your program, run it with one button, understand why it fails and watch it work step by step. All in a single window, in English or Spanish.

VizcachaIDE is inspired by [Thonny](https://thonny.org), the IDE that thousands of people use to learn Python, and brings that same idea to Go: fewer buttons, more clarity and explanations designed for learning.

## Who it is for

- **Students** taking their first steps in programming with Go.
- **Teachers** who need a simple tool for the classroom that works in Spanish and needs no complicated setup.
- **Anyone** who wants to learn Go without first fighting with terminals, environment variables and extensions.

## What you can do

![The Assistant explains an error in Spanish](https://raw.githubusercontent.com/codeplai/VizcachaIDE/main/docs/images/wails/assistant.es.png)

### Write and run
- Press **Run (F5)** and see your program's output instantly.
- Type in the console when your program asks for keyboard input.
- Pass arguments to your program and work with projects that use `go.mod`.
- Your code tidies itself on save, using Go's standard format (gofmt).

### Understand your errors
- When Go finds a problem, the **Assistant** explains **what happened and how to fix it**, in your language.
- It recognizes the 25 most common beginner errors: unused variables, extra imports, mismatched types, indexes out of range, uninitialized maps, deadlocks between goroutines and more.
- Go's original message is always in view, with a button to search for it online. That way you learn to read real errors.
- Errors are underlined **as you type**, before you run.

### Watch your program step by step

![Debugging step by step: variables, the one that just changed and how you got here](https://raw.githubusercontent.com/codeplai/VizcachaIDE/main/docs/images/wails/debugger.es.png)

- Click next to a line number to set a **breakpoint** and press **Debug (F6)**.
- Move forward with plainly worded buttons: **Next line**, **Go into function**, **Leave function**.
- See the value of your variables at every step. The one that just changed is highlighted, so you can see what the last line did.
- Find out **how you got there** (the call stack) and what each goroutine is doing.

### Write faster
- Smart Go autocompletion, with the documentation for each function.
- Parameter help while you type a call.
- Ctrl+click to jump to where a function is defined.
- Find and replace, go to line, zoom, light or dark theme.

## Designed for learning

- **English and Spanish** across the whole interface and in the error explanations. It detects your system language and you can change it whenever you like.
- **One main button.** Run is the most visible thing in the window; everything else appears when you need it.
- **Very readable type.** It uses Atkinson Hyperlegible, a typeface designed so that characters like 0 and O, or 1, l and I, are not confused.
- **Everything included.** The full version ships with Go, the Delve debugger and gopls: install VizcachaIDE and you can start programming.

## Download

VizcachaIDE is **free and open source**.

| Version | What it includes | Who it is for |
|---|---|---|
| **Full** | VizcachaIDE + Go + Delve + gopls | If you do not have Go installed (recommended to get started) |
| **Lite** | VizcachaIDE only | If you already have Go installed |

**Downloads and source code:** [github.com/codeplai/VizcachaIDE](https://github.com/codeplai/VizcachaIDE)

**Systems:**
- **Windows 10 and 11:** installer that needs no administrator permissions, and a portable version.
- **macOS and Linux:** preview version.

> **First time on Windows:** because the installer is not digitally signed yet, Windows may show "Windows protected your PC". Click **More info → Run anyway**.

## Project status

VizcachaIDE is a **release candidate**: it is already usable and we are polishing details before the final version.

- **New 2.0 edition:** redesigned interface, lighter and faster, with a debugger that explains every step.
- **Classic 1.x edition:** remains available while 2.0 reaches its final version.

## Made in Peru

VizcachaIDE is a project of **[Codeplai Games](https://codeplai.pe)**, created by Marks Calderon. Its name comes from the **vizcacha**, the Andean rodent that lives among the rocks of the highlands: small, curious and always alert.

Do you have ideas, did you find a bug, or do you want to use it in your class? Write to us at **hola@codeplai.pe** or open an *issue* on GitHub.

---

<small>VizcachaIDE is distributed under the MIT license and includes Go, Delve and gopls, each with its own open-source license.</small>
