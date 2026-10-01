package delve

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-dap"
)

const fakeGoroot = "C:/Program Files/Go"

// loadTranscript reads a transcript recorded from dlv dap 1.27.2 (the same files the
// 1.0 tests use), replacing {SRC} with sourceDir and {GOROOT} with a fake Go folder.
func loadTranscript(t *testing.T, name, sourceDir string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ReplaceAll(string(raw), "{SRC}", filepath.ToSlash(sourceDir))
	return []byte(strings.ReplaceAll(text, "{GOROOT}", fakeGoroot))
}

// decodeMessages turns a list of recorded JSON messages into go-dap messages.
func decodeMessages(t *testing.T, raw []json.RawMessage) []dap.Message {
	t.Helper()
	messages := make([]dap.Message, 0, len(raw))
	for _, item := range raw {
		message, err := dap.DecodeProtocolMessage(item)
		if err != nil {
			t.Fatalf("decoding %s: %v", item, err)
		}
		messages = append(messages, message)
	}
	return messages
}

func functionsSession(t *testing.T, sourceDir string) []dap.Message {
	t.Helper()
	var transcript struct {
		Messages []json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(loadTranscript(t, "functions_session.json", sourceDir), &transcript); err != nil {
		t.Fatal(err)
	}
	return decodeMessages(t, transcript.Messages)
}

// panicSession is the "panic" part of panic_and_build_error.json, by name.
func panicPart(t *testing.T, sourceDir, name string) dap.Message {
	t.Helper()
	var transcript struct {
		Panic map[string]json.RawMessage `json:"panic"`
	}
	if err := json.Unmarshal(loadTranscript(t, "panic_and_build_error.json", sourceDir), &transcript); err != nil {
		t.Fatal(err)
	}
	return decodeMessages(t, []json.RawMessage{transcript.Panic[name]})[0]
}

func responseFor[T dap.Message](t *testing.T, messages []dap.Message) T {
	t.Helper()
	for _, message := range messages {
		if typed, ok := message.(T); ok {
			return typed
		}
	}
	var zero T
	t.Fatalf("no %T in the transcript", zero)
	return zero
}
