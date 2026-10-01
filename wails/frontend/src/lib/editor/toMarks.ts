import type { ExplainedDiagnostic } from '../domain'
import { problemsOfFile, sameFile } from '../stores/diagnostics'
import type { EditorMarks, ProblemMark } from './marks'

const toProblemMark = (item: ExplainedDiagnostic): ProblemMark | null => {
  const { location, end, message } = item.diagnostic
  if (!location) return null
  const sameLine = end && end.line === location.line
  return {
    line: location.line,
    from: location.column,
    to: sameLine ? end.column : location.column,
    severity: item.diagnostic.severity,
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
  const marks = problemsOfFile(problems, path)
    .map(toProblemMark)
    .filter((mark): mark is ProblemMark => mark !== null)
  return {
    debugLine: path && debugFile && sameFile(debugFile, path) ? debugLine : null,
    breakpoints,
    problems: marks
  }
}
