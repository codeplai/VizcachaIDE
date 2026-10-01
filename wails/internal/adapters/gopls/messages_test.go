package gopls

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// The change event must not carry a range: with {0:0 - 0:0} gopls inserts the text at the top
// of the file instead of replacing it, so every edit duplicated the document (found by the E2E QA).
func TestDidChangeSendsTheWholeTextWithoutARange(t *testing.T) {
	raw, err := json.Marshal(didChangeParams(document{path: "/tmp/main.go", text: "package main\n", version: 2}))
	if err != nil {
		t.Fatal(err)
	}
	body := string(raw)
	if strings.Contains(body, "range") {
		t.Errorf("didChange has a range: %s", body)
	}
	if !strings.Contains(body, `"text":"package main\n"`) || !strings.Contains(body, `"version":2`) {
		t.Errorf("didChange = %s", body)
	}
}

func TestRealGoplsReplacesTheDocumentWhenItChanges(t *testing.T) {
	session := startSession(t)
	fixed := "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(1)\n}\n"
	if err := session.server.ChangeDocument(context.Background(), session.file, fixed, 1); err != nil {
		t.Fatal(err)
	}
	waitUntil(t, queryTimeout, func() bool { return len(session.sink.diagnosticsOf(session.file)) == 0 })
}
