package runner

import (
	"fmt"
	"path/filepath"

	"github.com/codeplai/VizcachaIDE/wails/internal/adapters/rust"
	"github.com/codeplai/VizcachaIDE/wails/internal/app"
	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

var _ app.MemberRunner = (*Runner)(nil)

// ConfigureMember implements app.MemberRunner: the member of the workspace around path that the
// student chose after ErrChooseMember, run through its main program (docs/PLAN_RUST.md section
// 3.2 point 8).
func (r *Runner) ConfigureMember(path, member string, programArgs []string) (domain.RunConfiguration, error) {
	workspace, ok, err := rust.FindProject(path)
	if err != nil || !ok {
		return domain.RunConfiguration{}, fmt.Errorf("%w: no Cargo workspace around %s", ErrNoBinary, path)
	}
	for _, folder := range workspace.Members {
		crate, found, err := rust.FindProject(filepath.Join(folder, "Cargo.toml"))
		if err != nil || !found || crate.Name != member {
			continue
		}
		binary, hasProgram := crate.BinaryFor("")
		if !hasProgram {
			return domain.RunConfiguration{}, ErrNoBinary
		}
		return r.Configure(binary.Path, programArgs), nil
	}
	return domain.RunConfiguration{}, fmt.Errorf("%w: no member %q", ErrChooseMember, member)
}
