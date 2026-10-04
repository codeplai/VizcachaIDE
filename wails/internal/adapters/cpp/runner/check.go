package runner

import (
	"context"
	"errors"
	"os"
	"runtime"
	"time"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/cpp"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

const checkTimeout = 60 * time.Second

// Check implements app.CodeChecker: "<cxx> -fsyntax-only <flags> <sources>" with the warnings of
// a compilation (-Wall -Wextra), without events and outside the run slot. It returns the
// compiler's text, "" when the code is clean. Findings are output, not an error.
func (r *Runner) Check(ctx context.Context, config domain.RunConfiguration) (string, error) {
	if info, err := os.Stat(folderOf(config)); err != nil || !info.IsDir() {
		return "", nil // for example an untitled run, whose folder is already deleted
	}
	sources, err := sourcesOf(config)
	if err != nil {
		return "", err
	}
	compiler, err := r.locator.Compiler(ctx)
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	args := append([]string{"-fsyntax-only"}, checkFlags(compiler.Family)...)
	output, err := capture(ctx, compiler.Path, folderOf(config), r.base, append(args, sources...))
	if err != nil && !errors.Is(err, ErrCompileFailed) {
		return "", err
	}
	return output, nil
}

// checkFlags are the compilation's flags without -static, which only matters when linking.
func checkFlags(family cpp.Family) []string {
	var flags []string
	for _, flag := range cpp.CompileFlags(family, runtime.GOOS) {
		if flag != "-static" {
			flags = append(flags, flag)
		}
	}
	return flags
}
