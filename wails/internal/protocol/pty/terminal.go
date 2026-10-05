// Package pty runs a program inside a pseudoterminal (ConPTY on Windows, a Unix PTY elsewhere),
// so it behaves as in a real terminal: its output is line buffered and survives a crash, and
// what the user types is echoed back by the terminal itself.
//
// Spike result (M0, track N0): github.com/aymanbagabas/go-pty passes on Windows except for
// Ctrl+C, which ConPTY ignores when written as ETX. Interrupt therefore only asks politely;
// the caller (protocol/process) kills the process tree after its grace period, as it does today.
package pty

import (
	"os"

	gopty "github.com/aymanbagabas/go-pty"
)

// Columns is wide on purpose: ConPTY hard-wraps long lines, which would split the file paths
// that the error parsers read.
const Columns = 4096

const rows = 30

// etx is the byte a terminal sends for Ctrl+C.
const etx = 0x03

// plainEnvironment asks programs not to colour their output (Python 3.13+ colours tracebacks
// when it sees a terminal); Clean removes whatever escape sequences remain.
var plainEnvironment = []string{"NO_COLOR=1", "PYTHON_COLORS=0", "TERM=dumb"}

// Program is what to start in the terminal.
type Program struct {
	Command string
	Args    []string
	Dir     string
	// Env is the full environment; nil means the IDE's own environment.
	Env []string
	// Interactive starts a terminal a person types in: the given size (Columns and Rows, 80x24
	// when zero) and an environment left as it is, so the program colours its output and the
	// caller keeps the raw bytes for a terminal emulator.
	Interactive   bool
	Columns, Rows int
}

// size is the size of the pseudoterminal: wide for the Output panel, the caller's for a terminal.
func (p Program) size() (columns, lines int) {
	if !p.Interactive {
		return Columns, rows
	}
	columns, lines = p.Columns, p.Rows
	if columns <= 0 {
		columns = 80
	}
	if lines <= 0 {
		lines = 24
	}
	return columns, lines
}

func (p Program) environment() []string {
	if p.Interactive {
		return p.Env
	}
	return withPlainOutput(p.Env)
}

// Terminal is a program running in a pseudoterminal. Read returns its output with terminal
// escape sequences (pass it through Clean); Write sends keyboard input.
type Terminal struct {
	device gopty.Pty
	cmd    *gopty.Cmd
}

// Start opens a pseudoterminal and starts the program in it.
func Start(program Program) (*Terminal, error) {
	device, err := gopty.New()
	if err != nil {
		return nil, err
	}
	width, height := program.size()
	if err := device.Resize(width, height); err != nil {
		_ = device.Close()
		return nil, err
	}
	cmd := device.Command(program.Command, program.Args...)
	cmd.Dir = program.Dir
	cmd.Env = program.environment()
	if err := cmd.Start(); err != nil {
		_ = device.Close()
		return nil, err
	}
	return &Terminal{device: device, cmd: cmd}, nil
}

func withPlainOutput(env []string) []string {
	if env == nil {
		env = os.Environ()
	}
	return append(append([]string{}, env...), plainEnvironment...)
}

// Read reads program output, escape sequences included.
func (t *Terminal) Read(p []byte) (int, error) { return t.device.Read(p) }

// Write sends keyboard input. A line must end with "\r", like the Enter key.
func (t *Terminal) Write(p []byte) (int, error) { return t.device.Write(p) }

// Resize tells the program the new size of the terminal.
func (t *Terminal) Resize(columns, lines int) error { return t.device.Resize(columns, lines) }

// Interrupt sends Ctrl+C. Unix terminals deliver SIGINT; ConPTY may ignore it (see the
// package comment), so the caller must still kill the tree if the program stays alive.
func (t *Terminal) Interrupt() error {
	_, err := t.device.Write([]byte{etx})
	return err
}

// Pid is the process id of the program.
func (t *Terminal) Pid() int { return t.cmd.Process.Pid }

// Wait blocks until the program ends and returns its exit error, as exec.Cmd.Wait does.
func (t *Terminal) Wait() error { return t.cmd.Wait() }

// Close releases the pseudoterminal; pending reads return an error.
func (t *Terminal) Close() error { return t.device.Close() }
