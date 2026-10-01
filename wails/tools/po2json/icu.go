package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	placeholderPattern = regexp.MustCompile(`\{(\w+)\}`)
	pluralPattern      = regexp.MustCompile(`\{(\w+), plural, one \{((?:[^{}]|\{\w+\})*)\} other \{((?:[^{}]|\{\w+\})*)\}\}`)
)

// icuPlural builds an ICU plural message from the two forms of a gettext plural.
// The forms keep their own {count} placeholder.
func icuPlural(one, other string) string {
	return fmt.Sprintf("{count, plural, one {%s} other {%s}}", one, other)
}

// icuEscape makes plain gettext text safe for ICU MessageFormat: {name} placeholders
// stay, any other brace is quoted, and an apostrophe before a brace is doubled.
func icuEscape(text string) string {
	masked := placeholderPattern.ReplaceAllString(text, "\x00$1\x01")
	masked = strings.NewReplacer("'{", "''{", "'}", "''}", "'\x00", "''\x00").Replace(masked)
	masked = strings.NewReplacer("{", "'{'", "}", "'}'").Replace(masked)
	return strings.NewReplacer("\x00", "{", "\x01", "}").Replace(masked)
}

// icuUnescape undoes icuEscape, for go-i18n templates.
func icuUnescape(text string) string {
	return strings.NewReplacer("'{'", "{", "'}'", "}", "''{", "'{", "''}", "'}").Replace(text)
}

// goTemplate converts an ICU message into go-i18n output: a plain string
// ("{{.name}}" placeholders) or a {"one", "other"} map when it has a plural.
func goTemplate(icu string) any {
	match := pluralPattern.FindStringSubmatchIndex(icu)
	if match == nil {
		return goPlaceholders(icu)
	}
	name := icu[match[2]:match[3]]
	oneForm, otherForm := icu[match[4]:match[5]], icu[match[6]:match[7]]
	before, after := icu[:match[0]], icu[match[1]:]
	build := func(form string) string {
		form = strings.ReplaceAll(form, "#", "{"+name+"}")
		return goPlaceholders(before + form + after)
	}
	return map[string]string{"one": build(oneForm), "other": build(otherForm)}
}

func goPlaceholders(text string) string {
	return icuUnescape(placeholderPattern.ReplaceAllString(text, "{{.$1}}"))
}
