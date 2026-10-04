package errorcatalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
)

// ErrInvalidCatalog means a language's catalog JSON is inconsistent (a bug caught by its tests).
var ErrInvalidCatalog = errors.New("invalid error catalog")

var placeholderUse = regexp.MustCompile(`\{(\w+)\}`)

// texts are the three messages of one entry in one language. Placeholders are written {name}.
type texts struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Fix   string `json:"fix"`
}

type rawEntry struct {
	ID       string   `json:"id"`
	Patterns []string `json:"patterns"`
	// Codes are stable error codes of the compiler (rustc's "E0382", a lint name): a diagnostic
	// whose Code is one of them is recognised whatever its wording.
	Codes []string `json:"codes"`
	EN       texts    `json:"en"`
	ES       texts    `json:"es"`
}

// entry recognises one kind of message. The first entry whose code or pattern matches wins.
type entry struct {
	id       string
	codes    []string
	patterns []*regexp.Regexp
	texts    map[string]texts
}

// catalog is the ordered list of entries.
type catalog struct {
	entries []entry
}

// match is the result of recognising a message: the entry and its named groups.
type match struct {
	entry        *entry
	placeholders map[string]string
}

func loadCatalog(catalogJSON []byte) (*catalog, error) {
	var raw []rawEntry
	if err := json.Unmarshal(catalogJSON, &raw); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}
	result := &catalog{}
	seen := map[string]bool{}
	for _, item := range raw {
		built, err := buildEntry(item)
		if err != nil {
			return nil, err
		}
		if seen[built.id] {
			return nil, fmt.Errorf("%w: duplicate id %s", ErrInvalidCatalog, built.id)
		}
		seen[built.id] = true
		result.entries = append(result.entries, built)
	}
	return result, nil
}

func buildEntry(item rawEntry) (entry, error) {
	built := entry{id: item.ID, codes: item.Codes, texts: map[string]texts{"en": item.EN, "es": item.ES}}
	for _, pattern := range item.Patterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return entry{}, fmt.Errorf("%w: %s: %w", ErrInvalidCatalog, item.ID, err)
		}
		built.patterns = append(built.patterns, compiled)
	}
	if len(built.patterns) == 0 && len(built.codes) == 0 {
		return entry{}, fmt.Errorf("%w: %s has no patterns and no codes", ErrInvalidCatalog, item.ID)
	}
	return built, built.checkPlaceholders()
}

// checkPlaceholders requires every {name} of every text to be a named group of every pattern.
func (e entry) checkPlaceholders() error {
	for lang, t := range e.texts {
		for _, text := range []string{t.Title, t.Body, t.Fix} {
			if text == "" {
				return fmt.Errorf("%w: %s has an empty %s text", ErrInvalidCatalog, e.id, lang)
			}
			for _, used := range placeholderUse.FindAllStringSubmatch(text, -1) {
				if err := e.requireGroup(used[1]); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e entry) requireGroup(name string) error {
	if len(e.patterns) == 0 { // recognised by code only: nothing can fill {name}
		return fmt.Errorf("%w: %s uses {%s} but has no pattern", ErrInvalidCatalog, e.id, name)
	}
	for _, pattern := range e.patterns {
		if !slices.Contains(pattern.SubexpNames(), name) {
			return fmt.Errorf("%w: %s uses {%s} but a pattern lacks the group", ErrInvalidCatalog, e.id, name)
		}
	}
	return nil
}

// find returns the first entry that recognises the diagnostic, by its code (when the
// compiler gives a stable one, "" otherwise) or by its message, or nil. The patterns of an entry
// recognised by code still fill its placeholders when they match.
func (c *catalog) find(code, message string) *match {
	for i := range c.entries {
		placeholders, ok := c.entries[i].search(message)
		if ok || (code != "" && slices.Contains(c.entries[i].codes, code)) {
			if placeholders == nil {
				placeholders = map[string]string{}
			}
			return &match{entry: &c.entries[i], placeholders: placeholders}
		}
	}
	return nil
}

func (e *entry) search(message string) (map[string]string, bool) {
	for _, pattern := range e.patterns {
		indexes := pattern.FindStringSubmatchIndex(message)
		if indexes == nil {
			continue
		}
		placeholders := map[string]string{}
		for group, name := range pattern.SubexpNames() {
			if name != "" && indexes[2*group] >= 0 {
				placeholders[name] = message[indexes[2*group]:indexes[2*group+1]]
			}
		}
		return placeholders, true
	}
	return nil, false
}
