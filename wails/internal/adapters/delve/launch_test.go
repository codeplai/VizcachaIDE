package delve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
)

func TestStaleDebugBinariesAreRemovedButNotOurOwn(t *testing.T) {
	dir := t.TempDir()
	own := filepath.Join(dir, debugBinaryName+"_42.exe")
	stale := filepath.Join(dir, debugBinaryName+"_7.exe")
	unrelated := filepath.Join(dir, "other.exe")
	for _, path := range []string{own, stale, unrelated} {
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	removed := removeStaleDebugBinaries(42, dir)

	if len(removed) != 1 || removed[0] != stale {
		t.Errorf("removed = %v, want [%s]", removed, stale)
	}
	for _, kept := range []string{own, unrelated} {
		if _, err := os.Stat(kept); err != nil {
			t.Errorf("%s must stay: %v", kept, err)
		}
	}
}

func TestLaunchArgumentsRunTheProgramRemotelyFromItsFolder(t *testing.T) {
	dir := t.TempDir()
	config := domain.NewFileRunConfiguration(domain.CodeLanguageGo, filepath.Join(dir, "functions.go"), []string{"-n", "3"})

	request, err := launchRequest(config, map[string]string{"GOFLAGS": "-mod=mod"}, "out.exe")
	if err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(request.Arguments, &got); err != nil {
		t.Fatal(err)
	}
	if got["mode"] != "debug" || got["outputMode"] != "remote" || got["cwd"] != dir || got["output"] != "out.exe" {
		t.Errorf("arguments = %v", got)
	}
	if got["program"] != filepath.Join(dir, "functions.go") {
		t.Errorf("program = %v", got["program"])
	}
}

func TestListeningAddressIsParsedFromDelveOutput(t *testing.T) {
	if got := parseListeningAddress("DAP server listening at: 127.0.0.1:52341"); got != "127.0.0.1:52341" {
		t.Errorf("address = %q", got)
	}
	if got := parseListeningAddress("Type 'dlv help'"); got != "" {
		t.Errorf("address = %q, want empty", got)
	}
}
