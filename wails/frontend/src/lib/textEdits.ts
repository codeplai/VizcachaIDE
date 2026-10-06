// Text edits of the language server (a rename touches many places and files). Positions of the
// backend are 1-based lines and columns counted in runes (Unicode characters); JavaScript strings
// count UTF-16 units, so an emoji before the symbol must be counted once there and twice here.
import type { TextEdit } from './domain'
import { minimalChange } from './editor/textChange'

export interface Change {
  from: number
  to: number
  insert: string
}

/** UTF-16 units taken by the first `runes` characters of `line`. */
export const unitsOfRunes = (line: string, runes: number): number => {
  let units = 0
  for (let count = 0; count < runes && units < line.length; count++) {
    units += (line.codePointAt(units) ?? 0) > 0xffff ? 2 : 1
  }
  return units
}

/** Characters (runes) in the first `units` UTF-16 units of `line`. */
export const runesOfUnits = (line: string, units: number): number => {
  let runes = 0
  for (let at = 0; at < units && at < line.length; runes++) {
    at += (line.codePointAt(at) ?? 0) > 0xffff ? 2 : 1
  }
  return runes
}

const lineStarts = (text: string): number[] => {
  const starts = [0]
  for (let at = text.indexOf('\n'); at >= 0; at = text.indexOf('\n', at + 1)) starts.push(at + 1)
  return starts
}

const lineEnd = (text: string, starts: number[], index: number): number =>
  index + 1 < starts.length ? (starts[index + 1] ?? text.length) - 1 : text.length

/** Offset in `text` of a 1-based line and rune column; throws when it is outside the text. */
const offsetIn = (text: string, starts: number[], line: number, column: number): number => {
  const start = starts[line - 1]
  if (start === undefined || column < 1) throw new RangeError(`no line ${line}`)
  const end = lineEnd(text, starts, line - 1)
  const content = text.slice(start, end)
  const units = unitsOfRunes(content, column - 1)
  if (runesOfUnits(content, units) < column - 1) {
    throw new RangeError(`no column ${column} in line ${line}`)
  }
  return start + units
}

/**
 * The smallest changes that make `before` into `after`: one per line when both have the same
 * number of lines (a server that answers "replace the whole file" still changes only the names).
 */
const smallChanges = (before: string, after: string): Change[] => {
  const was = before.split('\n')
  const now = after.split('\n')
  if (was.length === 1 || was.length !== now.length) {
    const change = minimalChange(before, after)
    return change ? [change] : []
  }
  const changes: Change[] = []
  let offset = 0
  was.forEach((line, index) => {
    const change = minimalChange(line, now[index] ?? '')
    if (change) {
      changes.push({ from: offset + change.from, to: offset + change.to, insert: change.insert })
    }
    offset += line.length + 1
  })
  return changes
}

/**
 * Turns the edits of one file into changes of its text (offsets of the original text, in order).
 * Edits that change nothing are dropped. Throws a RangeError when an edit is outside the text
 * or two edits overlap.
 */
export const editsToChanges = (text: string, edits: TextEdit[]): Change[] => {
  const starts = lineStarts(text)
  const spans = edits.map((edit) => ({
    from: offsetIn(text, starts, edit.range.start.line, edit.range.start.column),
    to: offsetIn(text, starts, edit.range.end.line, edit.range.end.column),
    insert: edit.newText
  }))
  spans.sort((a, b) => a.from - b.from || a.to - b.to)
  const changes: Change[] = []
  let cursor = 0
  for (const span of spans) {
    if (span.to < span.from || span.from < cursor) throw new RangeError('overlapping edits')
    cursor = span.to
    for (const change of smallChanges(text.slice(span.from, span.to), span.insert)) {
      changes.push({
        from: span.from + change.from,
        to: span.from + change.to,
        insert: change.insert
      })
    }
  }
  return changes
}

/** Applies changes (offsets of the original text) to a string. */
export const applyChanges = (text: string, changes: Change[]): string => {
  let result = ''
  let cursor = 0
  for (const change of changes) {
    result += text.slice(cursor, change.from) + change.insert
    cursor = change.to
  }
  return result + text.slice(cursor)
}
