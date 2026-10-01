// What the app tells the editor to draw for the open file (the editor never reads stores itself).
import { StateEffect, StateField } from '@codemirror/state'
import type { Severity } from '../domain'

export interface ProblemMark {
  line: number
  /** 1-based columns, `to` exclusive. */
  from: number
  to: number
  severity: Severity
  /** Translated title of the explanation, or the compiler message when there is none. */
  title: string
  detail: string
}

export interface EditorMarks {
  debugLine: number | null
  breakpoints: number[]
  problems: ProblemMark[]
}

export const emptyMarks: EditorMarks = { debugLine: null, breakpoints: [], problems: [] }

export const setMarks = StateEffect.define<EditorMarks>()

export const marksField = StateField.define<EditorMarks>({
  create: () => emptyMarks,
  update: (value, transaction) => {
    for (const effect of transaction.effects) if (effect.is(setMarks)) return effect.value
    return value
  }
})
