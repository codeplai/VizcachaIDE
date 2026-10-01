import { derived } from 'svelte/store'
import { TERMINATED_BY_USER } from '../events'
import { debugActive, debugOutput, debugStarting, currentLine } from './debug'
import { activeFileName } from './files'
import { problemCount } from './diagnostics'
import { splitOutputLinks, type OutputSegment } from './outputLinks'
import { runLines, runResult, stoppedByUser } from './run'

/** A line of the Output panel: either program text or an i18n key with its values. */
export interface OutputLine {
  tone: 'system' | 'plain' | 'error' | 'success'
  text?: string
  /** `text` split into plain pieces and clickable "file.go:line:column" places. */
  segments?: OutputSegment[]
  key?: string
  values?: Record<string, string | number>
  /** Seconds, formatted for the language at render time. */
  seconds?: number
}

const textLine = (tone: 'plain' | 'error', text: string): OutputLine => ({
  tone,
  text,
  segments: splitOutputLinks(text)
})

const verdictOf = (
  exitCode: number,
  durationMs: number,
  problems: number,
  stopped: boolean
): OutputLine => {
  if (stopped || exitCode === TERMINATED_BY_USER) return { tone: 'system', key: 'run.stopped' }
  if (exitCode === 0) return { tone: 'success', key: 'run.finished', seconds: durationMs / 1000 }
  if (problems > 0) return { tone: 'error', key: 'run.compileFailed', values: { count: problems } }
  return { tone: 'error', key: 'run.exitCode', values: { code: exitCode } }
}

const programLines = derived(
  [runLines, runResult, problemCount, stoppedByUser],
  ([lines, result, problems, stopped]) => {
    const view: OutputLine[] = lines.map((line) => {
      if (line.kind === 'start') {
        return { tone: 'system', key: 'run.starting', values: { file: line.text } }
      }
      return textLine(line.kind === 'stderr' ? 'error' : 'plain', line.text)
    })
    if (result) view.push(verdictOf(result.exitCode, result.durationMs, problems, stopped))
    return view
  }
)

const debuggingLines = derived(
  [debugOutput, activeFileName, currentLine, debugStarting],
  ([lines, file, line, starting]) => {
    if (starting) return [{ tone: 'system', key: 'run.debugStarting' } satisfies OutputLine]
    const view: OutputLine[] = [
      { tone: 'system', key: 'run.debugging', values: { file, line: line ?? 0 } },
      { tone: 'system', key: 'run.debugStdin' }
    ]
    const extra = lines
      .filter((entry) => entry.category !== 'console')
      .map((entry) => textLine('plain', entry.text))
    return [...view, ...extra]
  }
)

/** What the Output panel shows: the debug session while one is active, the last run otherwise. */
export const outputLines = derived(
  [debugActive, debugStarting, debuggingLines, programLines],
  ([debugging, starting, debug, program]) => (debugging || starting ? debug : program)
)
