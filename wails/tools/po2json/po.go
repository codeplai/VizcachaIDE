package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// poEntry is one message of a gettext catalog.
type poEntry struct {
	Context string
	ID      string
	Plural  string
	Strs    []string
}

type poParser struct {
	entries []poEntry
	current *poEntry
	field   *string
}

// parsePO reads a gettext .po catalog. The header entry (empty msgid) is skipped.
func parsePO(r io.Reader) ([]poEntry, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	p := &poParser{}
	for lineNo := 1; scanner.Scan(); lineNo++ {
		if err := p.feed(strings.TrimSpace(scanner.Text())); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read po: %w", err)
	}
	p.flush()
	return p.entries, nil
}

func (p *poParser) flush() {
	if p.current != nil && p.current.ID != "" {
		p.entries = append(p.entries, *p.current)
	}
	p.current, p.field = nil, nil
}

func (p *poParser) feed(line string) error {
	switch {
	case line == "":
		p.flush()
	case strings.HasPrefix(line, "#"):
		return nil
	case strings.HasPrefix(line, `"`):
		return p.appendString(line)
	default:
		return p.startField(line)
	}
	return nil
}

func (p *poParser) appendString(quoted string) error {
	if p.field == nil {
		return fmt.Errorf("string without a field: %s", quoted)
	}
	text, err := strconv.Unquote(quoted)
	if err != nil {
		return fmt.Errorf("bad string %s: %w", quoted, err)
	}
	*p.field += text
	return nil
}

func (p *poParser) startField(line string) error {
	keyword, rest, found := strings.Cut(line, " ")
	if !found {
		return fmt.Errorf("unexpected line: %s", line)
	}
	if keyword == "msgctxt" || (keyword == "msgid" && p.current != nil && len(p.current.Strs) > 0) {
		p.flush()
	}
	if p.current == nil {
		p.current = &poEntry{}
	}
	switch {
	case keyword == "msgctxt":
		p.field = &p.current.Context
	case keyword == "msgid":
		p.field = &p.current.ID
	case keyword == "msgid_plural":
		p.field = &p.current.Plural
	case strings.HasPrefix(keyword, "msgstr"):
		p.current.Strs = append(p.current.Strs, "")
		p.field = &p.current.Strs[len(p.current.Strs)-1]
	default:
		return fmt.Errorf("unknown keyword %q", keyword)
	}
	return p.appendString(rest)
}
