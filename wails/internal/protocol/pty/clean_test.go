package pty_test

import (
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/pty"
)

// ConPTY writes a prompt's trailing space as a cursor move (CSI 1 C); it must stay a space.
func TestCleanKeepsSpacesWrittenAsCursorMoves(t *testing.T) {
	cases := map[string]string{
		"Nombre:\x1b[1C\x1b[?25hAna\r\n": "Nombre: Ana\r\n",
		"a\x1b[Cb":                       "a b",
		"x\x1b[3Cy":                      "x   y",
	}
	for raw, want := range cases {
		if got := pty.Clean(raw); got != want {
			t.Errorf("Clean(%q) = %q, want %q", raw, got, want)
		}
	}
}
