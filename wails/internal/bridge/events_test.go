package bridge

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
)

// The constants of events.go must match the ones of frontend/src/lib/events.ts.
func TestEventNamesMatchFrontend(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "frontend", "src", "lib", "events.ts"))
	if err != nil {
		t.Fatalf("read events.ts: %v", err)
	}
	pattern := regexp.MustCompile(`(?m)^\s+\w+:\s*'([a-z]+:[a-z]+)',?\s*$`)
	var fromTS []string
	for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
		fromTS = append(fromTS, match[1])
	}
	if len(fromTS) == 0 {
		t.Fatal("no event names found in events.ts")
	}
	want := slices.Clone(AllEvents)
	slices.Sort(want)
	slices.Sort(fromTS)
	if !slices.Equal(want, fromTS) {
		t.Errorf("events.go and events.ts differ\n go: %v\n ts: %v", want, fromTS)
	}
}

func TestEventNamesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range AllEvents {
		if seen[name] {
			t.Errorf("duplicate event %q", name)
		}
		seen[name] = true
	}
}

func TestAllEventsCoversEveryConstant(t *testing.T) {
	source, err := os.ReadFile("events.go")
	if err != nil {
		t.Fatal(err)
	}
	declared := regexp.MustCompile(`(?m)^\s+Event\w+\s*=\s*"([a-z]+:[a-z]+)"`).FindAllStringSubmatch(string(source), -1)
	if len(declared) != len(AllEvents) {
		t.Fatalf("%d constants but %d entries in AllEvents", len(declared), len(AllEvents))
	}
	for _, match := range declared {
		if !slices.Contains(AllEvents, match[1]) {
			t.Errorf("%q is missing from AllEvents", match[1])
		}
	}
}
