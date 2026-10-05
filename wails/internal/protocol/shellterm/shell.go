// Package shellterm runs the shells of the integrated terminal: each session is a login or
// interactive shell in a pseudoterminal, with the IDE's toolchains first in PATH. Its output is
// raw (escape sequences included) because the terminal emulator of the frontend draws it.
package shellterm

// Command is the shell to start and its arguments.
type Command struct {
	Path string
	Args []string
}

// ChooseShell picks the user's shell. On Windows it is pwsh.exe when it is on PATH, else
// powershell.exe. Elsewhere it is $SHELL, else /bin/zsh on macOS or /bin/bash. lookPath and
// getenv are exec.LookPath and os.Getenv in production.
func ChooseShell(goos string, lookPath func(string) (string, error), getenv func(string) string) Command {
	if goos == "windows" {
		return windowsShell(lookPath)
	}
	shell := getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
		if goos == "darwin" {
			shell = "/bin/zsh"
		}
	}
	return Command{Path: shell, Args: unixArguments(goos)}
}

// windowsShell prefers PowerShell 7 and falls back to the Windows PowerShell every Windows has.
func windowsShell(lookPath func(string) (string, error)) Command {
	for _, name := range []string{"pwsh.exe", "powershell.exe"} {
		if path, err := lookPath(name); err == nil {
			return Command{Path: path, Args: []string{"-NoLogo"}}
		}
	}
	return Command{Path: "powershell.exe", Args: []string{"-NoLogo"}}
}

// unixArguments starts a login shell on macOS, where an application launched from the Dock
// has a minimal PATH; Terminal.app does the same. Elsewhere the pseudoterminal already makes
// the shell interactive.
func unixArguments(goos string) []string {
	if goos == "darwin" {
		return []string{"-l"}
	}
	return nil
}
