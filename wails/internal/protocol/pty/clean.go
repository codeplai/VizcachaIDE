package pty

import "github.com/charmbracelet/x/ansi"

// Clean removes terminal escape sequences (cursor moves, erases, colours, window titles) so
// the text can go to the Output panel and to the error parsers.
//
// It works on whole strings: a sequence split between two reads survives in pieces, so callers
// should clean complete lines or buffer until a sequence ends.
func Clean(text string) string { return ansi.Strip(text) }
