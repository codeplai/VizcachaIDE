package toolchain

import (
	"errors"
	"fmt"
	"go/format"
	"go/scanner"
	"runtime"

	"github.com/codeplai/VizcachaIDE/wails/internal/app"
)

func isWindows() bool { return runtime.GOOS == "windows" }

// FormatSource implements app.Toolchain with the standard go/format package.
// A syntax error wraps app.ErrFormat and names the line.
func (t *Toolchain) FormatSource(text string) (string, error) {
	formatted, err := format.Source([]byte(text))
	if err == nil {
		return string(formatted), nil
	}
	var list scanner.ErrorList
	if errors.As(err, &list) && len(list) > 0 {
		return "", fmt.Errorf("%w: line %d: %s", app.ErrFormat, list[0].Pos.Line, list[0].Msg)
	}
	return "", fmt.Errorf("%w: %w", app.ErrFormat, err)
}
