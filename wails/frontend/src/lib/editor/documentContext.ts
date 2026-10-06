// What the language extensions (completion, hover, signature, definition) need to ask gopls.
import type { EditorState } from '@codemirror/state'
import type { Bridge } from '../bridge'
import type { SourceLocation } from '../domain'
import { runesOfUnits, unitsOfRunes } from '../textEdits'

export type LanguageApi = Bridge['language']

export interface DocumentContext {
  /** File shown in the editor right now, or null when nothing is open. */
  path: () => string | null
  /** Sends pending edits to gopls, so answers match what is on screen. */
  flush: () => Promise<void>
}

/** Editor offset to a 1-based line and column (counted in characters, as the backend does). */
export const locationAt = (state: EditorState, pos: number, file: string): SourceLocation => {
  const line = state.doc.lineAt(pos)
  return { file, line: line.number, column: runesOfUnits(line.text, pos - line.from) + 1 }
}

/** 1-based line and column to an editor offset, clamped to the document. */
export const offsetOf = (state: EditorState, line: number, column: number): number => {
  const clamped = state.doc.line(Math.min(Math.max(line, 1), state.doc.lines))
  return Math.min(clamped.from + unitsOfRunes(clamped.text, Math.max(column, 1) - 1), clamped.to)
}
