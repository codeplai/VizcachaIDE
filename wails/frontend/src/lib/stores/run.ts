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
      runLines.set([{ kind: 'start', text: fileName(config.target) }])
    }),
    bridge.on('run:output', ({ stream, text }) => {
      const added = splitChunk(text).map((line) => ({ kind: stream, text: line }))
      runLines.update((lines) => [...lines, ...added])
    }),
    bridge.on('run:finished', (result) => {
      running.set(false)
      runResult.set(result)
    })
  ]
  return () => offs.forEach((off) => off())
}

export const resetRun = (): void => {
  runLines.set([])
  runResult.set(null)
  running.set(false)
  stoppedByUser.set(false)
}
