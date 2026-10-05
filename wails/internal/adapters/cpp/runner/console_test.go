package runner

import (
	"context"
	"runtime"
	"strings"
	"testing"
)

const accents = "#include <iostream>\nint main() {\n    std::cout << \"¿Cómo te llamas? Ñandú\" << std::endl;\n    return 0;\n}\n"

// std::cout writes UTF-8: in a Windows console it must still read "¿Cómo", not "┐C├│mo".
func TestAccentsReachTheTerminalIntact(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only the Windows console has its own code page")
	}
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		if err := r.Run(context.Background(), r.Configure(writeSource(t, accents), nil)); err != nil {
			t.Fatal(err)
		}
		if code := sink.waitFinished(t); code != 0 {
			t.Fatalf("exit code %d, stderr %q", code, errText(sink))
		}
		if out, _ := sink.text(); !strings.Contains(out, "¿Cómo te llamas? Ñandú") {
			t.Fatalf("stdout = %q", out)
		}
	})
}

const echoLine = "#include <iostream>\n#include <string>\nint main() {\n    std::string nombre;\n    std::cout << \"Nombre: \";\n    std::getline(std::cin, nombre);\n    std::cout << \"Hola, \" << nombre << \" (\" << nombre.size() << \" bytes)\" << std::endl;\n    return 0;\n}\n"

// What the student types with accents reaches std::cin as UTF-8 too.
func TestAccentsTypedReachTheProgram(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("only the Windows console has its own code page")
	}
	eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
		if err := r.Run(context.Background(), r.Configure(writeSource(t, echoLine), nil)); err != nil {
			t.Fatal(err)
		}
		sink.waitOutput(t, "Nombre:")
		if err := r.WriteInput("Ñandú"); err != nil {
			t.Fatal(err)
		}
		if code := sink.waitFinished(t); code != 0 {
			t.Fatalf("exit code %d, stderr %q", code, errText(sink))
		}
		if out, _ := sink.text(); !strings.Contains(out, "Hola, Ñandú (7 bytes)") {
			t.Fatalf("stdout = %q", out)
		}
	})
}
