import type { SourceLocation } from '../domain'
import { parseAnsi } from './ansi'

/** A piece of an output line: plain text, or a "file.go:line:column" place the user can click. */
export interface OutputSegment {
  text: string
  location?: SourceLocation
  /** ANSI style of the piece (see ansi.ts). */
  color?: string
  background?: string
  decorations?: string[]
}

const PLACE = /((?:[A-Za-z]:)?[^\s:()'"]+\.go):(\d+)(?::(\d+))?/g

export const splitOutputLinks = (text: string): OutputSegment[] => {
  const segments: OutputSegment[] = []
  let last = 0
  for (const match of text.matchAll(PLACE)) {
    const [whole, file = '', line = '1', column = '1'] = match
    if (match.index > last) segments.push({ text: text.slice(last, match.index) })
    segments.push({ text: whole, location: { file, line: Number(line), column: Number(column) } })
    last = match.index + whole.length
  }
  if (last < text.length) segments.push({ text: text.slice(last) })
  return segments
}

/**
 * Splits one line of program output into pieces: ANSI colors and bold become style fields,
 * "file.go:line:column" places become clickable. The pieces are plain text, never HTML.
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
