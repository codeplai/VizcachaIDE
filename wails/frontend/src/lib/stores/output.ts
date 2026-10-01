import { derived } from 'svelte/store'
import { debugActive, debugOutput, currentLine } from './debug'
import { activeFileName } from './files'
import { problemCount } from './diagnostics'
import { runLines, runResult } from './run'

/** A line of the Output panel: either program text or an i18n key with its values. */
export interface OutputLine {
  tone: 'system' | 'plain' | 'error' | 'success'
  text?: string
  key?: string
  values?: Record<string, string | number>
  /** Seconds, formatted for the language at render time. */
  seconds?: number
}

const programLines = derived([runLines, runResult, problemCount], ([lines, result, problems]) => {
  const view: OutputLine[] = lines.map((line) => {
    if (line.kind === 'start') {
      return { tone: 'system', key: 'run.starting', values: { file: line.text } }
    }
    return { tone: line.kind === 'stderr' ? 'error' : 'plain', text: line.text }
  })
  if (!result) return view
  if (result.exitCode === 0) {
    view.push({ tone: 'success', key: 'run.finished', seconds: result.durationMs / 1000 })
  } else if (problems > 0) {
    view.push({ tone: 'error', key: 'run.compileFailed', values: { count: problems } })
  } else {
    view.push({ tone: 'error', key: 'run.exitCode', values: { code: result.exitCode } })
  }
  return view
})

const debuggingLines = derived(
  [debugOutput, activeFileName, currentLine],
  ([lines, file, line]) => {
    const view: OutputLine[] = [
      { tone: 'system', key: 'run.debugging', values: { file, line: line ?? 0 } },
      { tone: 'system', key: 'run.debugStdin' }
    ]
    const extra = lines
      .filter((entry) => entry.category !== 'console')
      .map((entry): OutputLine => ({ tone: 'plain', text: entry.text }))
    return [...view, ...extra]
  }
)

/** What the Output panel shows: the debug session while one is active, the last run otherwise. */
export const outputLines = derived(
  [debugActive, debuggingLines, programLines],
  ([debugging, debug, program]) => (debugging ? debug : program)
)
