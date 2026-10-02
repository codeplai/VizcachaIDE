import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'

/** One snippet of the console scrollback with what came back from the interpreter. */
export interface ConsoleEntry {
  id: number
  code: string
  result: string
  output: string
  error: string
}

export const consoleEntries = writable<ConsoleEntry[]>([])
/** Snippets already run, oldest first, for the up/down arrows. */
export const consoleHistory = writable<string[]>([])
/** Where the arrows are in the history; `null` while the user types something new. */
export const consoleCursor = writable<number | null>(null)

let nextId = 1

/** Runs a snippet and appends it to the scrollback. Blank snippets are ignored. */
export const evalConsole = async (bridge: Bridge, code: string): Promise<void> => {
  if (code.trim() === '') return
  consoleHistory.update((all) => (all[all.length - 1] === code ? all : [...all, code]))
  consoleCursor.set(null)
  const answer = await bridge.console.eval(code).catch((reason: unknown) => ({
    result: '',
    output: '',
    error: String(reason)
  }))
  consoleEntries.update((all) => [...all, { id: nextId++, code, ...answer }])
}

/** Clears the scrollback and the interpreter session; the history is kept. */
export const resetConsole = async (bridge: Bridge): Promise<void> => {
  await bridge.console.reset()
  consoleEntries.set([])
  consoleCursor.set(null)
}

/**
 * Moves through the history. `direction` -1 goes to an older snippet, +1 to a newer one;
 * returns the text for the input, or `null` when there is nothing to change.
 */
export const stepConsoleHistory = (direction: -1 | 1): string | null => {
  const history = get(consoleHistory)
  const current = get(consoleCursor)
  if (history.length === 0) return null
  if (current === null) {
    if (direction === 1) return null
    consoleCursor.set(history.length - 1)
    return history[history.length - 1] ?? ''
  }
  const next = current + direction
  if (next >= history.length) {
    consoleCursor.set(null)
    return ''
  }
  const clamped = Math.max(0, next)
  consoleCursor.set(clamped)
  return history[clamped] ?? ''
}
