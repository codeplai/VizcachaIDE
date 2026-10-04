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
	EN       texts    `json:"en"`
	ES       texts    `json:"es"`
}

// entry recognises one kind of message. The first entry whose pattern matches wins.
type entry struct {
	id       string
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
	built := entry{id: item.ID, texts: map[string]texts{"en": item.EN, "es": item.ES}}
	for _, pattern := range item.Patterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return entry{}, fmt.Errorf("%w: %s: %w", ErrInvalidCatalog, item.ID, err)
		}
		built.patterns = append(built.patterns, compiled)
	}
	if len(built.patterns) == 0 {
		return entry{}, fmt.Errorf("%w: %s has no patterns", ErrInvalidCatalog, item.ID)
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
	for _, pattern := range e.patterns {
		if !slices.Contains(pattern.SubexpNames(), name) {
			return fmt.Errorf("%w: %s uses {%s} but a pattern lacks the group", ErrInvalidCatalog, e.id, name)
		}
	}
	return nil
}

// find returns the first entry that recognises the message, or nil.
func (c *catalog) find(message string) *match {
	for i := range c.entries {
		if placeholders, ok := c.entries[i].search(message); ok {
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
