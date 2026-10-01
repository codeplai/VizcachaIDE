package main

import (
	"bytes"
	"crypto/sha1" //nolint:gosec // not security related: only a short stable suffix for keys
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// message is one text in every language, written as ICU MessageFormat.
type message struct {
	EN string
	ES string
}

type catalog map[string]message

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// loadSemantic reads the hand-written semantic catalog ({"key": {"en": "...", "es": "..."}}).
func loadSemantic(path string) (catalog, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read semantic catalog: %w", err)
	}
	var parsed map[string]map[string]string
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse semantic catalog: %w", err)
	}
	result := catalog{}
	for key, texts := range parsed {
		if texts["en"] == "" || texts["es"] == "" {
			return nil, fmt.Errorf("key %q needs both en and es", key)
		}
		result[key] = message{EN: texts["en"], ES: texts["es"]}
	}
	return result, nil
}

// loadLegacy merges the 1.0 catalogs. Keys are "legacy.<slug>_<hash>".
func loadLegacy(enPath, esPath string) (catalog, error) {
	english, err := readPO(enPath)
	if err != nil {
		return nil, err
	}
	spanish, err := readPO(esPath)
	if err != nil {
		return nil, err
	}
	result := catalog{}
	for _, entry := range english {
		key := legacyKey(entry)
		result[key] = message{EN: poText(entry), ES: poText(entry)}
	}
	for _, entry := range spanish {
		key := legacyKey(entry)
		current, known := result[key]
		if !known {
			return nil, fmt.Errorf("%q is in es but not in en", entry.ID)
		}
		current.ES = poText(entry)
		result[key] = current
	}
	return result, nil
}

func readPO(path string) ([]poEntry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read po: %w", err)
	}
	entries, err := parsePO(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return entries, nil
}

// poText returns the ICU text of an entry, falling back to the msgid when untranslated.
func poText(entry poEntry) string {
	if entry.Plural == "" {
		return icuEscape(firstNonEmpty(entry.Strs, entry.ID))
	}
	one, other := entry.ID, entry.Plural
	if len(entry.Strs) >= 2 && entry.Strs[0] != "" {
		one, other = entry.Strs[0], entry.Strs[1]
	}
	return icuPlural(icuEscape(one), icuEscape(other))
}

func firstNonEmpty(values []string, fallback string) string {
	if len(values) > 0 && values[0] != "" {
		return values[0]
	}
	return fallback
}

func legacyKey(entry poEntry) string {
	words := strings.Fields(nonAlphanumeric.ReplaceAllString(strings.ToLower(entry.ID), " "))
	if len(words) > 5 {
		words = words[:5]
	}
	slug := strings.Join(words, "_")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	sum := sha1.Sum([]byte(entry.Context + "\x00" + entry.ID)) //nolint:gosec // see import
	return "legacy." + slug + "_" + hex.EncodeToString(sum[:])[:6]
}
