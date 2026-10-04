import type { SourceLocation } from '../domain'
import { parseAnsi } from './ansi'

/** A piece of an output line: plain text, or a "file:line:column" place the user can click. */
export interface OutputSegment {
  text: string
  location?: SourceLocation
  /** ANSI style of the piece (see ansi.ts). */
  color?: string
  background?: string
  decorations?: string[]
}

// Source files of the languages that print "file:line:column" (Go, C++ with .h and .hpp, Rust).
const SOURCE_EXTENSION = String.raw`go|cpp|cxx|cc|c\+\+|hpp|hh|h|rs`
// A Windows path may hold spaces, accents and backslashes ("C:\Users\Ana\mis programas\main.cpp").
const WINDOWS_FILE = String.raw`(?<![A-Za-z0-9_])[A-Za-z]:[\\/][^:*?"<>|]*?\.(?:${SOURCE_EXTENSION})`
const PLAIN_FILE = String.raw`[^\s:()'"]+\.(?:${SOURCE_EXTENSION})`
// The line is required: GCC's linker form "main.cpp:(.text+0x1a)" is not a place to go to.
const PLACE = new RegExp(String.raw`(${WINDOWS_FILE}|${PLAIN_FILE}):(\d+)(?::(\d+))?`, 'gi')

// Frames inside Rust's own library are not the student's code: "/rustc/<hash>/library/core/..."
// and the rust-src copy in the toolchain (".../lib/rustlib/src/rust/library/...").
const RUST_LIBRARY = /(^|[\\/])rustc[\\/][0-9a-f]{7,40}[\\/]|rustlib[\\/]src[\\/]rust[\\/]/i

export const splitOutputLinks = (text: string): OutputSegment[] => {
  const segments: OutputSegment[] = []
  let last = 0
  for (const match of text.matchAll(PLACE)) {
    const [whole, file = '', line = '1', column = '1'] = match
    if (RUST_LIBRARY.test(file)) continue
    if (match.index > last) segments.push({ text: text.slice(last, match.index) })
    segments.push({ text: whole, location: { file, line: Number(line), column: Number(column) } })
    last = match.index + whole.length
  }
  if (last < text.length) segments.push({ text: text.slice(last) })
  return segments
}

/**
 * Splits one line of program output into pieces: ANSI colors and bold become style fields,
 * "file:line:column" places become clickable. The pieces are plain text, never HTML.
 */
export const styledSegments = (text: string): OutputSegment[] =>
  parseAnsi(text).flatMap((piece) =>
    splitOutputLinks(piece.text).map((segment) => ({
      ...segment,
      ...(piece.color ? { color: piece.color } : {}),
      ...(piece.background ? { background: piece.background } : {}),
      ...(piece.decorations.length > 0 ? { decorations: piece.decorations } : {})
    }))
  )
