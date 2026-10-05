package vcpkg

import (
	"strings"
	"unicode"
)

// platform is the system the IDE builds for, as vcpkg's "supports" expressions see it.
type platform struct{ triplet, goos, goarch string }

func (p platform) key() string { return p.triplet + p.goos + p.goarch }

// value says whether a platform identifier is true here. Identifiers it does not know are true:
// a port is hidden only when its "supports" clearly excludes this system.
func (p platform) value(name string) bool {
	switch name {
	case "windows", "mingw":
		return p.goos == "windows"
	case "linux":
		return p.goos == "linux"
	case "osx":
		return p.goos == "darwin"
	case "x64", "x86_64":
		return p.goarch == "amd64"
	case "arm64":
		return p.goarch == "arm64"
	case "x86", "arm", "uwp", "xbox", "android", "ios", "emscripten", "wasm32", "freebsd", "openbsd", "qnx":
		return false
	case "static":
		return true
	case "staticcrt":
		return p.goos == "windows"
	}
	return true
}

// supports evaluates a vcpkg "supports" expression: identifiers, "!", "&", "|" and parentheses
// ("windows & !arm"). An expression it cannot read counts as supported.
func (p platform) supports(expression string) bool {
	parser := &expressionParser{tokens: tokenize(expression), platform: p}
	result := parser.or()
	if parser.failed || parser.position != len(parser.tokens) {
		return true
	}
	return result
}

func tokenize(expression string) []string {
	var tokens []string
	var word strings.Builder
	flush := func() {
		if word.Len() > 0 {
			tokens = append(tokens, word.String())
			word.Reset()
		}
	}
	for _, character := range expression {
		switch {
		case strings.ContainsRune("!&|()", character):
			flush()
			tokens = append(tokens, string(character))
		case unicode.IsSpace(character):
			flush()
		default:
			word.WriteRune(character)
		}
	}
	flush()
	return tokens
}

type expressionParser struct {
	tokens   []string
	position int
	platform platform
	failed   bool
}

func (e *expressionParser) peek() string {
	if e.position < len(e.tokens) {
		return e.tokens[e.position]
	}
	return ""
}

func (e *expressionParser) or() bool {
	result := e.and()
	for e.peek() == "|" {
		e.position++
		result = e.and() || result
	}
	return result
}

func (e *expressionParser) and() bool {
	result := e.not()
	for e.peek() == "&" {
		e.position++
		result = e.not() && result
	}
	return result
}

func (e *expressionParser) not() bool {
	switch token := e.peek(); token {
	case "!":
		e.position++
		return !e.not()
	case "(":
		e.position++
		result := e.or()
		if e.peek() != ")" {
			e.failed = true
		}
		e.position++
		return result
	case "", ")", "&", "|":
		e.failed = true
		return true
	default:
		e.position++
		return e.platform.value(token)
	}
}
