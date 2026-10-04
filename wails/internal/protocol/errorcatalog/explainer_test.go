package errorcatalog_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/codeplai/VizcachaIDE/wails/internal/domain"
	"github.com/codeplai/VizcachaIDE/wails/internal/protocol/errorcatalog"
)

const sampleCatalog = `[
 {"id": "E-UNUSED", "patterns": ["unused (?P<name>\\w+)"],
  "en": {"title": "{name} is unused", "body": "b {name}", "fix": "use { braces } of {name}"},
  "es": {"title": "{name} sin uso", "body": "c {name}", "fix": "usa {name}"}}
]`

var sampleLine = regexp.MustCompile(`^(?P<path>[^:]+):(?P<line>\d+):(?P<message>.*)$`)

// sampleParser reads "file:line:message" lines, the way a language parser would.
func sampleParser(raw, workingDir string, identify func(string) string) []domain.Diagnostic {
	var diagnostics []domain.Diagnostic
	for _, line := range errorcatalog.SplitLines(raw) {
		groups := errorcatalog.NamedGroups(sampleLine, line)
		if groups == nil {
			continue
		}
		location := errorcatalog.LocationFrom(groups, workingDir)
		diagnostics = append(diagnostics, domain.Diagnostic{
			Location: &location, Message: groups["message"], Code: identify(groups["message"]),
		})
	}
	return diagnostics
}

func newSample(t *testing.T) *errorcatalog.Explainer {
	t.Helper()
	explainer, err := errorcatalog.NewExplainer([]byte(sampleCatalog), sampleParser)
	if err != nil {
		t.Fatal(err)
	}
	return explainer
}

func TestParserReceivesTheCatalogIdentifier(t *testing.T) {
	got := newSample(t).Parse("a.x:3:unused n\r\nb.x:4:other\r\n", "/w")
	if len(got) != 2 || got[0].Code != "E-UNUSED" || got[1].Code != "" || got[0].Location.Line != 3 {
		t.Errorf("diagnostics = %+v", got)
	}
}

func TestExplainFillsPlaceholdersInBothLanguages(t *testing.T) {
	explainer := newSample(t)
	diagnostic := domain.Diagnostic{Message: "unused n"}

	en := explainer.Explain(diagnostic, domain.LanguageEN)
	if en.Title != "n is unused" || en.FixHint != "use { braces } of n" || en.Placeholders["name"] != "n" {
		t.Errorf("en = %+v", en)
	}
	if es := explainer.Explain(diagnostic, domain.LanguageES); es.Title != "n sin uso" {
		t.Errorf("es = %+v", es)
	}
	if fallback := explainer.Explain(diagnostic, "fr"); fallback.Title != "n is unused" {
		t.Errorf("fr = %+v, want English", fallback)
	}
	if explainer.Explain(domain.Diagnostic{Message: "nothing"}, "en") != nil {
		t.Error("an unknown message must not be explained")
	}
}

func TestInconsistentCatalogsAreRejected(t *testing.T) {
	cases := map[string]string{
		"bad json":       `[`,
		"no patterns":    `[{"id":"E-X","patterns":[],"en":{"title":"t","body":"b","fix":"f"},"es":{"title":"t","body":"b","fix":"f"}}]`,
		"missing group":  `[{"id":"E-X","patterns":["x"],"en":{"title":"{a}","body":"b","fix":"f"},"es":{"title":"t","body":"b","fix":"f"}}]`,
		"empty text":     `[{"id":"E-X","patterns":["x"],"en":{"title":"","body":"b","fix":"f"},"es":{"title":"t","body":"b","fix":"f"}}]`,
		"duplicate id":   strings.Replace(sampleCatalog, "]", ","+sampleCatalog[1:], 1),
		"invalid regexp": `[{"id":"E-X","patterns":["("],"en":{"title":"t","body":"b","fix":"f"},"es":{"title":"t","body":"b","fix":"f"}}]`,
	}
	for name, catalog := range cases {
		if _, err := errorcatalog.NewExplainer([]byte(catalog), sampleParser); !errors.Is(err, errorcatalog.ErrInvalidCatalog) {
			t.Errorf("%s: error = %v, want ErrInvalidCatalog", name, err)
		}
	}
}

func TestResolvePathKeepsAbsoluteAndJoinsRelative(t *testing.T) {
	if got := errorcatalog.ResolvePath(`C:\a\b.go`, "/w"); got != `C:\a\b.go` {
		t.Errorf("absolute = %q", got)
	}
	if got := errorcatalog.ResolvePath("/a/b.go", "/w"); got != "/a/b.go" {
		t.Errorf("posix = %q", got)
	}
	if got := errorcatalog.ResolvePath(`.\x\y.go`, ""); !strings.HasSuffix(got, "y.go") || strings.HasPrefix(got, ".") {
		t.Errorf("relative without directory = %q", got)
	}
}
