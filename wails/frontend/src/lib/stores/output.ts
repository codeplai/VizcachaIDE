import { derived } from 'svelte/store'
import { TERMINATED_BY_USER, type DebugOutputPayload } from '../events'
import { debugInputEnabled } from './codeLanguages'
import { debugActive, debugOutput, debugStarting, currentLine } from './debug'
import { activeFileName } from './files'
import { compileProblemCount } from './assistant'
import { styledSegments, type OutputSegment } from './outputLinks'
import {
  addChunk,
  lastRunConfiguration,
  running,
  runLines,
  runResult,
  stoppedByUser,
  type RawRunLine
} from './run'

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

/** Exit codes of a native program killed by the system: Unix -signal or 128 + signal, Windows 0xC000xxxx. */
const UNIX_CRASH_CODES = [134, 135, 136, 139]
const WINDOWS_CRASH_FLOOR = 0xc0000000
const isCrash = (exitCode: number): boolean =>
  UNIX_CRASH_CODES.includes(exitCode) || exitCode >>> 0 >= WINDOWS_CRASH_FLOOR

const verdictOf = (
  exitCode: number,
  durationMs: number,
  problems: number,
  stopped: boolean,
  native: boolean
): OutputLine => {
  if (stopped || exitCode === TERMINATED_BY_USER) return { tone: 'system', key: 'run.stopped' }
  if (exitCode === 0) return { tone: 'success', key: 'run.finished', seconds: durationMs / 1000 }
  // A crash comes first: its runtime diagnostic also counts as a problem.
  if (native && isCrash(exitCode)) return { tone: 'error', key: 'run.crashed' }
  if (problems > 0) return { tone: 'error', key: 'run.compileFailed', values: { count: problems } }
  return { tone: 'error', key: 'run.exitCode', values: { code: exitCode } }
}

const programLines = derived(
  [runLines, runResult, compileProblemCount, stoppedByUser, lastRunConfiguration],
  ([lines, result, problems, stopped, configuration]) => {
    const view: OutputLine[] = lines.map((line) => {
      if (line.kind === 'start') {
        return { tone: 'system', key: 'run.starting', values: { file: line.text } }
      }
      return textLine(line.kind === 'stderr' ? 'error' : 'plain', line.text)
    })
    const native = configuration?.codeLanguage === 'cpp'
    if (result) view.push(verdictOf(result.exitCode, result.durationMs, problems, stopped, native))
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

/** A finished debug session: what its program printed, until the next run or session. */
const endedDebugLines = derived(debugOutput, (lines): OutputLine[] => {
  const printed = debuggeeLines(lines)
  return printed.length === 0 ? [] : [{ tone: 'system', key: 'run.debugEnded' }, ...printed]
})

/** What the Output panel shows: the debug session while one is active (or the one that just
 * ended, until something else runs), the last run otherwise. */
export const outputLines = derived(
  [debugActive, debugStarting, debuggingLines, endedDebugLines, programLines],
  ([debugging, starting, debug, ended, program]) => {
    if (debugging || starting) return debug
    return ended.length > 0 ? ended : program
  }
)
