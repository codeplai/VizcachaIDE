package delve

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

const listenAddress = "127.0.0.1:0"

var listeningPattern = regexp.MustCompile(`DAP server listening at:\s*([^\s:]+):(\d+)`)

var errDelveExited = errors.New("delve exited before listening")

// adapterProcess is a running "dlv dap" server.
type adapterProcess struct {
	cmd     *exec.Cmd
	address chan string
	exited  chan struct{}
}

// startAdapter launches "dlv dap --listen=127.0.0.1:0" in workingDir.
// missing is the message of the ErrToolNotFound returned when it cannot start.
func startAdapter(executable, workingDir string, environment map[string]string, missing string) (*adapterProcess, error) {
	cmd := exec.Command(executable, "dap", "--listen="+listenAddress)
	cmd.Dir = workingDir
	cmd.Env = mergedEnvironment(environment)
	hideConsoleWindow(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opening delve output: %w", err)
	}
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w: %s", app.ErrToolNotFound, missing)
	}
	process := &adapterProcess{cmd: cmd, address: make(chan string, 1), exited: make(chan struct{})}
	go process.scan(bufio.NewScanner(stdout))
	return process, nil
}

// scan reads Delve's console until it closes, publishing the listening address.
func (p *adapterProcess) scan(scanner *bufio.Scanner) {
	for scanner.Scan() {
		address := parseListeningAddress(scanner.Text())
		if address == "" {
			continue
		}
		select {
		case p.address <- address:
		default:
		}
	}
	_ = p.cmd.Wait()
	close(p.exited)
}

// waitAddress blocks until Delve says where it listens, or it dies, or ctx ends.
func (p *adapterProcess) waitAddress(ctx context.Context) (string, error) {
	select {
	case address := <-p.address:
		return address, nil
	case <-p.exited:
		return "", errDelveExited
	case <-ctx.Done():
		return "", fmt.Errorf("waiting for delve: %w", ctx.Err())
	}
}

// kill ends Delve (and the debuggee with it) by its own PID, never by name.
func (p *adapterProcess) kill() {
	if p != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
}

func parseListeningAddress(line string) string {
	match := listeningPattern.FindStringSubmatch(line)
	if match == nil {
		return ""
	}
	return match[1] + ":" + match[2]
}

func mergedEnvironment(extra map[string]string) []string {
	env := os.Environ()
	for name, value := range extra {
		env = append(env, name+"="+value)
	}
	return env
}
