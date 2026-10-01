// Squiggles and hover messages come from @codemirror/lint; the app pushes the diagnostics in.
import { linter, setDiagnostics, type Diagnostic } from '@codemirror/lint'
import type { EditorState, Extension, StateEffect } from '@codemirror/state'
import type { ProblemMark } from './marks'

const lintSeverity = (mark: ProblemMark): Diagnostic['severity'] => {
  if (mark.severity === 'error' || mark.severity === 'warning') return mark.severity
  return 'info'
}

/** Gives a zero-width range something to underline: the rest of the word, or the next character. */
const visibleEnd = (state: EditorState, from: number, to: number, lineEnd: number): number => {
  if (to > from) return to
  const word = state.wordAt(from)
  if (word && word.to > from) return word.to
  return Math.min(from + 1, lineEnd)
}

export const toLintDiagnostics = (state: EditorState, problems: ProblemMark[]): Diagnostic[] =>
  problems.flatMap((problem) => {
    if (problem.line < 1 || problem.line > state.doc.lines) return []
    const line = state.doc.line(problem.line)
    const from = Math.min(line.from + problem.from - 1, line.to)
    const to = Math.min(line.from + problem.to - 1, line.to)
    const message =
      problem.title === problem.detail ? problem.title : `${problem.title}\n${problem.detail}`
    return [
      {
        from,
        to: visibleEnd(state, from, to, line.to),
        severity: lintSeverity(problem),
        message
      }
    ]
  })

export const pushDiagnostics = (
  state: EditorState,
  problems: ProblemMark[]
): StateEffect<unknown>[] => {
  const { effects } = setDiagnostics(state, toLintDiagnostics(state, problems))
  if (!effects) return []
  return 'length' in effects ? [...effects] : [effects]
}

/** No source of its own: diagnostics are pushed in with `pushDiagnostics`. */
export const problemLint: Extension = linter(null)
