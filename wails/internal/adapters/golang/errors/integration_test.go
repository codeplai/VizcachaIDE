package errors

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var realGoSamples = []string{"E-UNUSED-VAR", "E-UNDEFINED", "E-TYPE-MISMATCH", "E-MISSING-BRACE", "P-INDEX-RANGE", "P-NIL-MAP", "P-DEADLOCK", "V-PRINTF-ARGS"}

// Compiles real examples with the installed Go and checks that Go's current output
// still maps to the expected id. Skipped when go is not installed.
func TestRealGoOutputProducesTheExpectedID(t *testing.T) {
	goPath, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go is not installed")
	}
	explainer := newExplainer(t)
	for _, id := range realGoSamples {
		t.Run(id, func(t *testing.T) {
			dir := t.TempDir()
			source, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "examples", "errors", id, id+".go"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, id+".go"), source, 0o600); err != nil {
				t.Fatal(err)
			}
			verb := "run"
			if id[0] == 'V' {
				verb = "vet"
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			command := exec.CommandContext(ctx, goPath, verb, id+".go")
			command.Dir = dir
			output, runErr := command.CombinedOutput()
			if runErr == nil {
				t.Fatalf("go %s succeeded, expected a failure:\n%s", verb, output)
			}
			diagnostics := explainer.Parse(string(output), dir)
			if len(diagnostics) != 1 || diagnostics[0].Code != id {
				t.Fatalf("diagnostics = %+v\noutput:\n%s", diagnostics, output)
			}
			if location := diagnostics[0].Location; location != nil && filepath.Base(location.File) != id+".go" {
				t.Errorf("location = %+v", location)
			}
		})
	}
}
