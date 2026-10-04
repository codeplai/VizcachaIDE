package process

import "os/exec"

// PrepareTree makes cmd the head of its own process tree (a process group on Unix, a new
// process group with a hidden console on Windows), so KillTree can end it with its children.
// Call it before cmd.Start.
func PrepareTree(cmd *exec.Cmd) { prepareTree(cmd) }

// KillTree ends pid and all its descendants at once. pid must belong to a command started after
// PrepareTree. A Windows venv's python.exe is a launcher that runs the real interpreter as a child,
// so killing only the launcher would leave the student's code running.
func KillTree(pid int) { killTree(pid) }
