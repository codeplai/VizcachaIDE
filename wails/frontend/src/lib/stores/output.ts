import { derived } from 'svelte/store'
import { TERMINATED_BY_USER, type DebugOutputPayload } from '../events'
import { debugInputEnabled } from './codeLanguages'
import { debugActive, debugOutput, debugStarting, currentLine } from './debug'
import { activeFileName } from './files'
import { compileProblemCount } from './assistant'
import { styledSegments, type OutputSegment } from './outputLinks'
import { addChunk, running, runLines, runResult, stoppedByUser, type RawRunLine } from './run'

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
  segments: styledSegments(text)
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
  [runLines, runResult, compileProblemCount, stoppedByUser],
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

/** Program text of the debuggee (`debug:output`), joined like a run's output so a prompt stays one line. */
const debuggeeLines = (entries: DebugOutputPayload[]): OutputLine[] => {
  let lines: RawRunLine[] = []
  let tail: 'stdout' | 'stderr' | null = null
  for (const { category, text } of entries) {
    if (category === 'console') continue
    const next = addChunk(lines, tail, category, text)
    lines = next.lines
    tail = next.tail
  }
  return lines.map((line) => textLine('plain', line.text))
}

const debuggingLines = derived(
  [debugOutput, activeFileName, currentLine, debugStarting, debugInputEnabled],
  ([lines, file, line, starting, canType]) => {
    if (starting) return [{ tone: 'system', key: 'run.debugStarting' } satisfies OutputLine]
    const view: OutputLine[] = [
      line === null
        ? { tone: 'system', key: 'run.debugRunning', values: { file } }
        : { tone: 'system', key: 'run.debugging', values: { file, line } }
    ]
    // Delve gives a Go program no keyboard; a Python debuggee runs in a terminal and reads it.
    if (!canType) view.push({ tone: 'system', key: 'run.debugStdin' })
    return [...view, ...debuggeeLines(lines)]
  }
)

/** True when the Output input box is enabled: a program runs, or the debuggee can read the keyboard. */
export const programInputOpen = derived(
  [running, debugActive, debugStarting, debugInputEnabled],
  ([isRunning, debugging, starting, canType]) => isRunning || ((debugging || starting) && canType)
)

/** What the Output panel shows: the debug session while one is active, the last run otherwise. */
export const outputLines = derived(
  [debugActive, debugStarting, debuggingLines, programLines],
  ([debugging, starting, debug, program]) => (debugging || starting ? debug : program)
)
