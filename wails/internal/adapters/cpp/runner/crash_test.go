package runner

import (
	"context"
	"strings"
	"testing"
)

var crashPrograms = map[string]struct{ source, line string }{
	"segfault": {
		"#include <iostream>\nint main() {\n    std::cout << \"antes de caer\" << std::endl;\n    int* p = nullptr;\n    *p = 42;\n    return 0;\n}\n",
		"Segmentation fault",
	},
	"stack_overflow": {
		"int f(int n) {\n    volatile int a[1000];\n    a[0] = n;\n    return f(n + 1) + a[0];\n}\nint main() {\n    return f(0);\n}\n",
		"Stack overflow",
	},
	"divide_zero": {
		"#include <iostream>\nint main() {\n    int a = 10;\n    volatile int b = 0;\n    std::cout << a / b;\n    return 0;\n}\n",
		"Floating point exception",
	},
	"terminate": {
		"#include <stdexcept>\nint main() {\n    throw std::runtime_error(\"algo salio mal\");\n}\n",
		"Aborted",
	},
}

func TestRealCrashesPrintTheTerminalLine(t *testing.T) {
	for name, crash := range crashPrograms {
		t.Run(name, func(t *testing.T) {
			eachCompiler(t, func(t *testing.T, r *Runner, sink *testSink) {
				path := writeSource(t, crash.source)
				if err := r.Run(context.Background(), r.Configure(path, nil)); err != nil {
					t.Fatal(err)
				}
				code := sink.waitFinished(t)
				_, stderr := sink.text()
				if code == 0 || !strings.HasSuffix(strings.TrimSpace(stderr), crash.line) {
					t.Fatalf("exit %d, stderr %q, want it to end with %q", code, stderr, crash.line)
				}
			})
		})
	}
}

func TestCrashLineMapping(t *testing.T) {
	cases := []struct {
		goos string
		code int
		want string
	}{
		{"windows", 0, ""}, {"windows", 1, ""}, {"windows", 0xC0000005, "Segmentation fault"},
		{"windows", 0xC00000FD, "Stack overflow"}, {"windows", 0xC0000094, "Floating point exception"},
		{"windows", 0xC0000409, "Aborted"}, {"windows", 3, "Aborted"}, {"windows", -1073741819, "Segmentation fault"},
		{"linux", 0, ""}, {"linux", 3, ""}, {"linux", 139, "Segmentation fault"}, {"linux", 135, "Segmentation fault"},
		{"linux", 136, "Floating point exception"}, {"linux", 134, "Aborted"},
		{"linux", -11, "Segmentation fault"}, {"linux", -8, "Floating point exception"}, {"linux", -6, "Aborted"},
		{"linux", -1, ""}, {"linux", 1, ""},
	}
	for _, c := range cases {
		if got := crashLineFor(c.goos, c.code); got != c.want {
			t.Errorf("crashLineFor(%s, %d) = %q, want %q", c.goos, c.code, got, c.want)
		}
	}
}
