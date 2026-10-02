//go:build !windows

package filesystem

import (
	"fmt"
	"os/exec"
)

// run starts a command. With wait it returns when the command ends; without it the
// command is left running (file managers may outlive us).
func run(wait bool, name string, args ...string) error {
	command := exec.Command(name, args...)
	if !wait {
		if err := command.Start(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		go func() { _ = command.Wait() }()
		return nil
	}
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, output)
	}
	return nil
}
