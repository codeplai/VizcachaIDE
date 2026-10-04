package repl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/process"
)

// session is one repl.py process.
type session struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	answers *bufio.Reader
	stderr  *lockedBuffer
}

// lockedBuffer collects what the process prints on stderr (an interpreter crash).
type lockedBuffer struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.TrimSpace(b.data.String())
}

// request and answer are the lines of the protocol.
type request struct {
	Code string `json:"code"`
}

type answer struct {
	Result string `json:"result"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

// startSession runs `<python> -X utf8 -u <script>` with the environment of the adapter.
func startSession(interpreter python.Interpreter, baseEnvironment []string, scriptPath string) (*session, error) {
	cmd := exec.Command(interpreter.Path, "-X", "utf8", "-u", scriptPath)
	cmd.Env = python.EnvironmentList(python.Environment(baseEnvironment, interpreter))
	process.HideConsole(cmd)
	stderr := &lockedBuffer{}
	cmd.Stderr = stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("opening the console input: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opening the console output: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting %s: %w", interpreter.Path, err)
	}
	return &session{cmd: cmd, stdin: stdin, answers: bufio.NewReader(stdout), stderr: stderr}, nil
}

// ask sends one snippet and waits for its answer. It blocks; callers bound it with kill.
func (s *session) ask(code string) (domain.ConsoleResult, error) {
	line, err := json.Marshal(request{Code: code})
	if err != nil {
		return domain.ConsoleResult{}, fmt.Errorf("encoding the snippet: %w", err)
	}
	if _, err := s.stdin.Write(append(line, '\n')); err != nil {
		return domain.ConsoleResult{}, s.ended(err)
	}
	reply, err := s.answers.ReadBytes('\n')
	if err != nil {
		return domain.ConsoleResult{}, s.ended(err)
	}
	var decoded answer
	if err := json.Unmarshal(reply, &decoded); err != nil {
		return domain.ConsoleResult{}, fmt.Errorf("the console answered something unexpected: %w", err)
	}
	return domain.ConsoleResult{Result: decoded.Result, Output: decoded.Output, Error: decoded.Error}, nil
}

func (s *session) ended(cause error) error {
	if text := s.stderr.String(); text != "" {
		return fmt.Errorf("the Python console ended: %w\n%s", cause, text)
	}
	return fmt.Errorf("the Python console ended: %w", cause)
}

// kill stops the process and reaps it.
func (s *session) kill() {
	_ = s.stdin.Close()
	killTree(s.cmd.Process.Pid)
	_ = s.cmd.Process.Kill() // the process may already be gone
	go func() { _ = s.cmd.Wait() }()
}
