package pty

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// cursorForward is "move the cursor N columns right" (CSI N C). ConPTY writes it instead of
// trailing spaces: the prompt `input("Nombre: ")` arrives as "Nombre:" + CSI 1 C.
var cursorForward = regexp.MustCompile(`\x1b\[(\d*)C`)

// maxForward caps the spaces one cursor move may become.
const maxForward = 256

// Clean removes terminal escape sequences (cursor moves, erases, colours, window titles) so
// the text can go to the Output panel and to the error parsers. A cursor move to the right becomes
// the spaces it stands for, so a prompt keeps its trailing space.
//
// It works on whole strings: a sequence split between two reads survives in pieces, so callers
// should clean complete lines or buffer until a sequence ends.
func Clean(text string) string {
	return ansi.Strip(cursorForward.ReplaceAllStringFunc(text, forwardAsSpaces))
}

func forwardAsSpaces(sequence string) string {
	count, err := strconv.Atoi(cursorForward.FindStringSubmatch(sequence)[1])
	if err != nil || count < 1 {
		count = 1 // CSI C without a number moves one column
	}
	return strings.Repeat(" ", min(count, maxForward))
}
