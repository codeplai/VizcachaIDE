import { writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { RunConfiguration } from '../domain'
import type { RunFinishedPayload } from '../events'
import { readPreference, writePreference } from './preferences'

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
/** True while Go is slow to answer on the very first build (shows the "first time" message). */
export const preparingGo = writable(false)

/** How long a run may stay silent before we say that Go is preparing itself. */
export const PREPARING_DELAY_MS = 3000
const FIRST_BUILD_KEY = 'firstBuildSeen'

const fileName = (path: string): string => path.split(/[\\/]/).pop() ?? path

/** Splits a chunk into lines, dropping the empty piece after a final newline. */
export const splitChunk = (text: string): string[] => {
  const lines = text.split(/\r?\n/)
  if (lines[lines.length - 1] === '') lines.pop()
  return lines
}

export const connectRun = (bridge: Bridge): Unsubscribe => {
  let timer: ReturnType<typeof setTimeout> | undefined
  const stopWaiting = (): void => {
    clearTimeout(timer)
    preparingGo.set(false)
  }
  const waitForFirstBuild = (): void => {
    if (readPreference(FIRST_BUILD_KEY, false)) return
    timer = setTimeout(() => preparingGo.set(true), PREPARING_DELAY_MS)
  }
  const offs = [
    bridge.on('run:started', (config) => {
      runResult.set(null)
      running.set(true)
      lastRunConfiguration.set(config)
      runLines.set([{ kind: 'start', text: fileName(config.target) }])
      waitForFirstBuild()
    }),
    bridge.on('run:output', ({ stream, text }) => {
      stopWaiting()
      const added = splitChunk(text).map((line) => ({ kind: stream, text: line }))
      runLines.update((lines) => [...lines, ...added])
    }),
    bridge.on('run:finished', (result) => {
      stopWaiting()
      writePreference(FIRST_BUILD_KEY, true)
      running.set(false)
      runResult.set(result)
    })
  ]
  return () => {
    stopWaiting()
    offs.forEach((off) => off())
  }
}

export const resetRun = (): void => {
  runLines.set([])
  runResult.set(null)
  running.set(false)
  stoppedByUser.set(false)
  preparingGo.set(false)
}
