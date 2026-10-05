package shellterm

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestCloseKillsTheWholeProcessTree(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("the tree check uses PowerShell child processes")
	}
	sink := newRecordingSink()
	dir := t.TempDir()
	host := newTestHost(t, sink)
	id, err := host.Start(dir, 80, 24)
	if err != nil {
		t.Skipf("no pseudoterminal: %v", err)
	}
	if err := host.Write(id, "$p = Start-Process -PassThru -WindowStyle Hidden ping -ArgumentList '-n','60','127.0.0.1'; 'child-pid=' + $p.Id\r"); err != nil {
		t.Fatal(err)
	}
	var pid string
	eventually(t, "the child pid", func() bool {
		text := sink.text(id)
		i := strings.LastIndex(text, "child-pid=")
		if i < 0 {
			return false
		}
		digits := strings.TrimLeft(text[i+len("child-pid="):], " ")
		end := strings.IndexFunc(digits, func(r rune) bool { return r < '0' || r > '9' })
		if end <= 0 {
			return false
		}
		pid = digits[:end]
		return true
	})
	if err := host.Close(id); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the child to end", func() bool {
		out, _ := exec.Command("tasklist", "/FI", "PID eq "+pid, "/NH").Output()
		return !strings.Contains(string(out), "ping.exe")
	})
}
