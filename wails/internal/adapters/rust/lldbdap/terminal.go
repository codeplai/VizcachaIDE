package lldbdap

import (
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// terminalWorks reports whether a program can run in a pseudoterminal on this machine, probing
// with the adapter itself.
func terminalWorks(adapter string) bool {
	terminal, err := pty.Start(pty.Program{Command: adapter, Args: []string{"--version"}})
	if err != nil {
		return false
	}
	_ = terminal.Wait()
	_ = terminal.Close()
	return true
}

// environmentMap turns "NAME=value" entries into a map.
func environmentMap(list []string) map[string]string {
	env := make(map[string]string, len(list))
	for _, entry := range list {
		for index := 1; index < len(entry); index++ { // index 1: Windows has "=C:=..." entries
			if entry[index] == '=' {
				env[entry[:index]] = entry[index+1:]
				break
			}
		}
	}
	return env
}
