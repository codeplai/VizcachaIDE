// Package console implements app.Console with yaegi, a Go interpreter written in Go.
//
// One interpreter session lives as long as the console is not reset. Each snippet runs
// in its own goroutine with a timeout: yaegi cannot preempt running code, so a snippet
// that does not finish is abandoned together with its interpreter and a fresh session
// takes its place.
package console
