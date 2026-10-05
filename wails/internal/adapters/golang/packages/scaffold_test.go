package packages

import (
	"strings"
	"testing"
)

func TestModuleName(t *testing.T) {
	cases := map[string]string{
		"hola":             "hola",
		"Mi Programa":      "mi-programa",
		"Ñandú Número 1":   "nandu-numero-1",
		"2048":             "2048",
		"  espacios  ":     "espacios",
		"juego_v1.2":       "juego_v1.2",
		"Cálculo (final)!": "calculo-final",
		"$$$":              "proyecto",
		"-.-":              "proyecto",
	}
	for name, want := range cases {
		if got := ModuleName(name); got != want {
			t.Errorf("ModuleName(%q) = %q, want %q", name, got, want)
		}
	}
}

func TestScaffoldFiles(t *testing.T) {
	files, main := Scaffold{}.Scaffold("Mi Programa")
	if main != "main.go" {
		t.Fatalf("main file = %q", main)
	}
	if !strings.Contains(files["go.mod"], "module mi-programa\n") || !strings.Contains(files["go.mod"], "go 1.21") {
		t.Errorf("go.mod = %q", files["go.mod"])
	}
	if !strings.Contains(files["main.go"], "bufio") || !strings.Contains(files["main.go"], "Hola, %s") {
		t.Errorf("main.go does not greet: %q", files["main.go"])
	}
	if len(files) != 2 {
		t.Errorf("files = %d, want 2", len(files))
	}
}
