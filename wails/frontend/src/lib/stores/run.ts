import { writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { RunConfiguration } from '../domain'
import type { RunFinishedPayload } from '../events'

/** One raw line of program output, before it is turned into display text. */
export interface RawRunLine {
  kind: 'start' | 'stdout' | 'stderr'
  /** File name for `start`, program text otherwise. */
  text: string
}

export const runLines = writable<RawRunLine[]>([])
export const runResult = writable<RunFinishedPayload | null>(null)
export const running = writable(false)
/** The configuration of the last run (its `module` feeds the "Go modules" dialog). */
export const lastRunConfiguration = writable<RunConfiguration | null>(null)
/** True when the user pressed Stop for the current run. */
export const stoppedByUser = writable(false)

const fileName = (path: string): string => path.split(/[\\/]/).pop() ?? path

/** Splits a chunk into lines, dropping the empty piece after a final newline. */
export const splitChunk = (text: string): string[] => {
  const lines = text.split(/\r?\n/)
  if (lines[lines.length - 1] === '') lines.pop()
  return lines
}

type Stream = 'stdout' | 'stderr'

/**
 * Adds a chunk of program output. A chunk is whatever one read of the pipe returned, so it can end in
 * the middle of a line (Go prints "panic: " and the message in two writes, a prompt has no newline):
 * `tail` is the stream whose last line is still open, and the next chunk of that stream continues it.
 */
export const addChunk = (
  lines: RawRunLine[],
  tail: Stream | null,
  stream: Stream,
  text: string
): { lines: RawRunLine[]; tail: Stream | null } => {
  const pieces = text.split(/\r?\n/)
  const open = pieces[pieces.length - 1] !== ''
  const complete = open ? pieces : pieces.slice(0, -1)
  const next = [...lines]
  complete.forEach((piece, index) => {
    const last = next[next.length - 1]
    if (index === 0 && tail === stream && last && last.kind === stream) {
      next[next.length - 1] = { ...last, text: last.text + piece }
    } else {
      next.push({ kind: stream, text: piece })
    }
  })
  return { lines: next, tail: open ? stream : null }
}

let openTail: Stream | null = null

/** Appends program output (or what the user typed for it) to the run's lines. */
export const pushRunText = (stream: Stream, text: string): void => {
  runLines.update((lines) => {
    const result = addChunk(lines, openTail, stream, text)
    openTail = result.tail
    return result.lines
  })
}

/**
 * The "preparing Go for the first time" message is not made here: the backend prints it as program
 * output (`run:output`) when a run stays silent, so it arrives like any other line.
 */
export const connectRun = (bridge: Bridge): Unsubscribe => {
  const offs = [
    bridge.on('run:started', (config) => {
      runResult.set(null)
      running.set(true)
      lastRunConfiguration.set(config)
      openTail = null
      runLines.set([{ kind: 'start', text: fileName(config.target) }])
    }),
    bridge.on('run:output', ({ stream, text }) => pushRunText(stream, text)),
    bridge.on('run:finished', (result) => {
      running.set(false)
      runResult.set(result)
    })
  ]
  return () => offs.forEach((off) => off())
}

export const resetRun = (): void => {
  openTail = null
  runLines.set([])
  runResult.set(null)
  running.set(false)
  stoppedByUser.set(false)
}
