package python

import (
	"strings"
	"testing"
)

func TestScaffoldIsOneMainFile(t *testing.T) {
	for _, name := range []string{"hola", "Mi Programa", "Ñandú"} {
		files, main := Scaffold{}.Scaffold(name)
		if main != "main.py" || len(files) != 1 {
			t.Fatalf("%q: main = %q, files = %d", name, main, len(files))
		}
		text := files["main.py"]
		if !strings.Contains(text, `input("¿Cómo te llamas? ")`) || !strings.Contains(text, `print(f"Hola, {nombre}")`) {
			t.Errorf("%q: main.py = %q", name, text)
		}
	}
}
