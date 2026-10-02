package bridge

import "testing"

func TestWithGoExtension(t *testing.T) {
	cases := map[string]string{"": "", "hola": "hola.go", "hola.go": "hola.go", "notas.txt": "notas.txt"}
	for in, want := range cases {
		if got := withGoExtension(in); got != want {
			t.Errorf("withGoExtension(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDefaultDialogFolder(t *testing.T) {
	if got := defaultDialogFolder("a", "b"); got != "a" {
		t.Errorf("got %q", got)
	}
	if got := defaultDialogFolder("", "b"); got != "b" {
		t.Errorf("got %q", got)
	}
}
