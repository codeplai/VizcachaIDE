package console

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

// DefaultTimeout is how long a snippet may run before the session is abandoned.
const DefaultTimeout = 5 * time.Second

// Console is a yaegi-backed app.Console. It is safe for concurrent use.
type Console struct {
	timeout time.Duration
	mu      sync.Mutex
	current *session
}

// New creates a console whose snippets time out after timeout (DefaultTimeout if zero).
func New(timeout time.Duration) *Console {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	return &Console{timeout: timeout}
}

// Reset forgets every variable and function of the session.
func (c *Console) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = nil
}

type evalOutcome struct {
	value reflect.Value
	err   error
}

// Eval runs one snippet in the session, creating it on first use.
func (c *Console) Eval(code string) domain.ConsoleResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current == nil {
		fresh, err := newSession()
		if err != nil {
			return domain.ConsoleResult{Error: err.Error()}
		}
		c.current = fresh
	}
	running := c.current
	done := make(chan evalOutcome, 1)
	go func() { done <- evaluate(running, code) }()

	select {
	case outcome := <-done:
		return answer(running, outcome)
	case <-time.After(c.timeout):
		c.current = nil // the stuck goroutine keeps its old interpreter
		return domain.ConsoleResult{
			Output: running.output.take(),
			Error:  fmt.Sprintf("timeout: the code took more than %s and the console was reset", c.timeout),
		}
	}
}

// evaluate runs the snippet and turns a panic of the interpreter into an error.
func evaluate(s *session, code string) (outcome evalOutcome) {
	defer func() {
		if r := recover(); r != nil {
			outcome = evalOutcome{err: fmt.Errorf("%v", r)}
		}
	}()
	for _, path := range declaredImports(code) {
		s.imported[path] = true
	}
	for _, path := range missingImports(code) {
		if s.imported[path] {
			continue
		}
		s.imported[path] = true
		if _, err := s.interpreter.Eval(fmt.Sprintf("import %q", path)); err != nil {
			return evalOutcome{err: err}
		}
	}
	value, err := s.interpreter.Eval(code)
	if !isValueExpression(code) {
		value = reflect.Value{}
	}
	return evalOutcome{value: value, err: err}
}

func answer(s *session, outcome evalOutcome) domain.ConsoleResult {
	result := domain.ConsoleResult{Output: s.output.take()}
	if outcome.err != nil {
		result.Error = outcome.err.Error()
		return result
	}
	result.Result = format(outcome.value)
	return result
}
