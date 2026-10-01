import type { ExplainedDiagnostic } from '../domain'
import type { EditorMarks, ProblemMark } from './marks'

const sameFile = (a: string, b: string): boolean =>
  a === b || a.endsWith(`/${b}`) || b.endsWith(`/${a}`)

const toProblemMark = (item: ExplainedDiagnostic): ProblemMark | null => {
  const { location, end, message } = item.diagnostic
  if (!location) return null
  const sameLine = end && end.line === location.line
  return {
    line: location.line,
    from: location.column,
    to: sameLine ? end.column : location.column,
    title: item.explanation?.title ?? message,
    detail: message
  }
}

/** Converts store data into what the editor draws for one file. */
export const toMarks = (
  path: string | null,
  problems: ExplainedDiagnostic[],
  breakpoints: number[],
  debugFile: string | null,
  debugLine: number | null
): EditorMarks => {
  const here = (file: string | undefined): boolean => !!path && !!file && sameFile(file, path)
  const marks = problems
    .filter((item) => here(item.diagnostic.location?.file))
    .map(toProblemMark)
    .filter((mark): mark is ProblemMark => mark !== null)
  return {
    debugLine: here(debugFile ?? undefined) ? debugLine : null,
    breakpoints,
    problems: marks
  }
}
