package shellterm

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func found(names ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		if slices.Contains(names, name) {
			return `C:\bin\` + name, nil
		}
		return "", errors.New("not found")
	}
}

func environment(values map[string]string) func(string) string {
	return func(name string) string { return values[name] }
}

func TestWindowsPrefersPwshThenPowerShell(t *testing.T) {
	both := ChooseShell("windows", found("pwsh.exe", "powershell.exe"), environment(nil))
	if both.Path != `C:\bin\pwsh.exe` || !slices.Equal(both.Args, []string{"-NoLogo"}) {
		t.Fatalf("with pwsh: %+v", both)
	}
	only := ChooseShell("windows", found("powershell.exe"), environment(nil))
	if only.Path != `C:\bin\powershell.exe` {
		t.Fatalf("without pwsh: %+v", only)
	}
	neither := ChooseShell("windows", found(), environment(nil))
	if neither.Path != "powershell.exe" || !slices.Equal(neither.Args, []string{"-NoLogo"}) {
		t.Fatalf("nothing on PATH: %+v", neither)
	}
}

func TestUnixUsesShellVariableThenDefaults(t *testing.T) {
	cases := []struct {
		goos, shell, want string
		args              []string
	}{
		{"linux", "/usr/bin/fish", "/usr/bin/fish", nil},
		{"linux", "", "/bin/bash", nil},
		{"darwin", "", "/bin/zsh", []string{"-l"}},
		{"darwin", "/bin/bash", "/bin/bash", []string{"-l"}},
	}
	for _, c := range cases {
		got := ChooseShell(c.goos, found(), environment(map[string]string{"SHELL": c.shell}))
		if got.Path != c.want || !slices.Equal(got.Args, c.args) {
			t.Errorf("%s SHELL=%q: got %+v", c.goos, c.shell, got)
		}
	}
}

func TestEnvironmentPutsIDEFoldersBeforeTheUserPath(t *testing.T) {
	sep := string(os.PathListSeparator)
	base := []string{"=C:=C:\\dir", "Path=user1" + sep + "user2", "TERM=dumb", "HOME=/h"}
	env := Environment(base, []string{"ide1", "ide2"})
	want := "Path=ide1" + sep + "ide2" + sep + "user1" + sep + "user2"
	for _, entry := range []string{want, "TERM=xterm-256color", "COLORTERM=truecolor", "HOME=/h", "=C:=C:\\dir"} {
		if !slices.Contains(env, entry) {
			t.Errorf("missing %q in %v", entry, env)
		}
	}
	if slices.Contains(env, "TERM=dumb") {
		t.Error("TERM=dumb must be replaced")
	}
	if strings.Count(strings.Join(env, "\n"), "Path=") != 1 {
		t.Errorf("PATH must appear once: %v", env)
	}
}

func TestEnvironmentWithoutUserPathOrFolders(t *testing.T) {
	env := Environment([]string{"HOME=/h"}, nil)
	if !slices.Contains(env, "PATH=") {
		t.Fatalf("got %v", env)
	}
	env = Environment(nil, []string{"ide"})
	if !slices.Contains(env, "PATH=ide") {
		t.Fatalf("got %v", env)
	}
}

func TestCompleteLengthNeverSplitsARune(t *testing.T) {
	text := []byte("aé€😀")
	for cut := 0; cut <= len(text); cut++ {
		head := text[:cut]
		kept := completeLength(head)
		if !strings.EqualFold(string(head[:kept]), string(head[:kept])) || !validUTF8(head[:kept]) {
			t.Fatalf("cut %d keeps a split rune", cut)
		}
		if cut-kept > 3 {
			t.Fatalf("cut %d holds back %d bytes", cut, cut-kept)
		}
	}
	if completeLength([]byte{0xff, 0xfe}) != 2 {
		t.Error("bytes that are not UTF-8 must pass through")
	}
}

func validUTF8(b []byte) bool { return strings.ToValidUTF8(string(b), "") == string(b) }
