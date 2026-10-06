package cmake

import (
	"strings"
	"testing"
	"time"
)

const buildOutput = `[0/2] Re-checking globbed directories...
[1/3] Building CXX object CMakeFiles/nandu.dir/main.cpp.obj
FAILED: [code=1] CMakeFiles/nandu.dir/main.cpp.obj
C:\llvm\bin\c++.exe -g -std=c++17 -o CMakeFiles/nandu.dir/main.cpp.obj -c main.cpp
D:/proyecto/main.cpp:4:12: error: use of undeclared identifier 'totl'
    4 |     return totl;
      |            ^
1 error generated.
ninja: build stopped: subcommand failed.
`

func TestTextKeepsTheCompilersDiagnosticsOnly(t *testing.T) {
	got := (&Filter{}).Text(buildOutput)
	want := "D:/proyecto/main.cpp:4:12: error: use of undeclared identifier 'totl'\n    4 |     return totl;\n      |            ^\n1 error generated.\n"
	if got != want {
		t.Fatalf("Text = %q, want %q", got, want)
	}
}

func TestTextHidesCMakeStatusButKeepsErrorsAndVcpkg(t *testing.T) {
	configure := "-- The CXX compiler identification is Clang 23.1.2\r\n-- Running vcpkg install\r\nInstalling 1/1 fmt:x64-mingw-static...\r\nCMake Error at CMakeLists.txt:7 (add_executable):\r\n  No SOURCES given to target: nandu\r\n\r\n-- Configuring incomplete, errors occurred!\r\n"
	got := (&Filter{}).Text(configure)
	for _, want := range []string{"-- Running vcpkg install", "Installing 1/1", "CMake Error at CMakeLists.txt:7", "No SOURCES given"} {
		if !strings.Contains(got, want) {
			t.Errorf("Text lacks %q: %q", want, got)
		}
	}
	if strings.Contains(got, "compiler identification") || strings.Contains(got, "Configuring incomplete") {
		t.Errorf("Text keeps CMake's status lines: %q", got)
	}
}

func TestBuildLineSendsDiagnosticsToStderrAndHidesProgress(t *testing.T) {
	filter := &Filter{}
	if _, _, keep := filter.BuildLine("stdout", "[2/3] Linking CXX executable nandu.exe"); keep {
		t.Error("progress must be hidden")
	}
	stream, text, keep := filter.BuildLine("stdout", "main.cpp:1:1: warning: x")
	if !keep || stream != "stderr" || text != "main.cpp:1:1: warning: x" {
		t.Errorf("BuildLine = %q %q %v", stream, text, keep)
	}
	if stream, _, _ := filter.Line("stdout", "Installing 1/1 fmt"); stream != "stdout" {
		t.Errorf("configure output stays on its stream, got %q", stream)
	}
}

func TestNoticeAppearsOnceAfterTheDelay(t *testing.T) {
	now := time.Now()
	filter := &Filter{Notice: func() string { return "Compiling..." }, Delay: 2 * time.Second, now: func() time.Time { return now }}
	if _, _, keep := filter.BuildLine("stdout", "[1/3] Building CXX object a.obj"); keep {
		t.Fatal("no notice before the delay")
	}
	now = now.Add(time.Second)
	if _, _, keep := filter.BuildLine("stdout", "[2/3] Building CXX object b.obj"); keep {
		t.Fatal("no notice before the delay")
	}
	now = now.Add(2 * time.Second)
	stream, text, keep := filter.BuildLine("stdout", "[3/3] Linking CXX executable x.exe")
	if !keep || stream != "stdout" || text != "Compiling...\n" {
		t.Fatalf("notice = %q %q %v", stream, text, keep)
	}
	if _, _, keep := filter.BuildLine("stdout", "[4/4] Linking CXX executable y.exe"); keep {
		t.Fatal("the notice shows once")
	}
}
