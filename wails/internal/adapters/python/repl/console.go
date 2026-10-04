package repl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/python"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// DefaultTimeout is how long a snippet may run before the session is abandoned (the same as
// the console of Go).
const DefaultTimeout = 5 * time.Second

// Backend i18n keys of the messages of the console.
const (
	keyNoInput        = "errors.consoleNoInput"
	keyPythonNotFound = "errors.pythonNotFound"
)

// noInputMarker is the name of the exception repl.py raises when the code calls input().
const noInputMarker = "VizcachaConsoleNoInput"

// Finder locates the interpreter. *python.Locator is one.
type Finder interface {
	Find(ctx context.Context, folder string) (python.Interpreter, error)
}

// Options configures a Console. Finder and Translate are required.
type Options struct {
	Finder Finder
	// Translate returns the backend text of an i18n key in the language of the IDE.
	Translate func(key string) string
	// Folder is the folder whose .venv the interpreter search may use ("" for none).
	Folder string
	// Timeout is the limit of one snippet (DefaultTimeout if zero).
	Timeout time.Duration
	// CacheDir is where repl.py is copied (the user cache folder if empty).
	CacheDir string
	// BaseEnvironment is the "NAME=value" list Python starts from (os.Environ() if nil).
	BaseEnvironment []string
}

// Console is the app.Console of Python. It is safe for concurrent use.
type Console struct {
	options Options
	mu      sync.Mutex
	current *session
}

var _ app.Console = (*Console)(nil)

// New creates the console. No process starts until the first Eval.
func New(options Options) *Console {
	if options.Timeout <= 0 {
		options.Timeout = DefaultTimeout
	}
	if options.BaseEnvironment == nil {
		options.BaseEnvironment = os.Environ()
	}
	return &Console{options: options}
}

// Reset kills the process, so the next Eval starts a clean session.
func (c *Console) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.forget()
}

// forget kills and drops the process. The caller holds the lock.
func (c *Console) forget() {
	if c.current != nil {
		c.current.kill()
		c.current = nil
	}
}

type reply struct {
	result domain.ConsoleResult
	err    error
}

// Eval runs one snippet in the session, starting it on first use. A snippet that does not
// finish in time kills the process: its variables are lost and the next Eval starts again.
func (c *Console) Eval(code string) domain.ConsoleResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current == nil {
		started, err := c.start()
		if err != nil {
			return c.failure(err)
		}
		c.current = started
	}
	running := c.current
	done := make(chan reply, 1)
	go func() {
		result, err := running.ask(code)
		done <- reply{result: result, err: err}
	}()

	select {
	case answer := <-done:
		if answer.err != nil {
			c.forget()
			return domain.ConsoleResult{Error: answer.err.Error()}
		}
		return c.translated(answer.result)
	case <-time.After(c.options.Timeout):
		c.forget()
		return domain.ConsoleResult{
			Error: fmt.Sprintf("timeout: the code took more than %s and the console was reset", c.options.Timeout),
		}
	}
}

func (c *Console) start() (*session, error) {
	interpreter, err := c.options.Finder.Find(context.Background(), c.options.Folder)
	if err != nil {
		return nil, err
	}
	scriptPath, err := installScript(c.options.CacheDir)
	if err != nil {
		return nil, err
	}
	return startSession(interpreter, c.options.BaseEnvironment, scriptPath)
}

// failure turns a start error into a result: a missing Python gets its translated message.
func (c *Console) failure(err error) domain.ConsoleResult {
	if errors.Is(err, app.ErrToolNotFound) {
		return domain.ConsoleResult{Error: c.options.Translate(keyPythonNotFound)}
	}
	return domain.ConsoleResult{Error: err.Error()}
}

// translated replaces the error of input() by the translated explanation.
func (c *Console) translated(result domain.ConsoleResult) domain.ConsoleResult {
	if strings.Contains(result.Error, noInputMarker) {
		result.Error = c.options.Translate(keyNoInput)
	}
	return result
}
