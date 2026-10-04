package console

import (
	"bytes"
	"strings"
	"sync"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// session is one interpreter plus the buffer that catches what its snippets print.
type session struct {
	interpreter *interp.Interpreter
	imported    map[string]bool
	output      *safeBuffer
}

// safeBuffer is written by the interpreter goroutine and read by Eval.
type safeBuffer struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (b *safeBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data.Write(p)
}

// take returns what was written so far and empties the buffer.
func (b *safeBuffer) take() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	text := b.data.String()
	b.data.Reset()
	return text
}

func newSession() (*session, error) {
	output := &safeBuffer{}
	interpreter := interp.New(interp.Options{
		Stdin:  strings.NewReader(""),
		Stdout: output,
		Stderr: output,
	})
	if err := interpreter.Use(stdlib.Symbols); err != nil {
		return nil, err
	}
	return &session{interpreter: interpreter, output: output, imported: map[string]bool{}}, nil
}
