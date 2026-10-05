import { get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import { fileTree } from './files'
import { outputTab } from './layout'

/** One shell of the integrated terminal. `exitCode` is null while the shell is alive. */
export interface TerminalSession {
  id: string
  /** The number shown in the tab ("1: PowerShell"); never reused while the IDE runs. */
  number: number
  exitCode: number | null
}

export const terminalSessions = writable<TerminalSession[]>([])
export const activeTerminal = writable<string | null>(null)
/** Grows each time something asks the terminal to take the keyboard focus. */
export const terminalFocusRequests = writable(0)

let counter = 0
let starting = false
/** Output of a session whose view is not drawn yet (it is shown when the view appears). */
const backlog = new Map<string, string[]>()
const writers = new Map<string, (data: string) => void>()

const deliver = (id: string, data: string): void => {
  const write = writers.get(id)
  if (write) write(data)
  else backlog.set(id, [...(backlog.get(id) ?? []), data])
}

/** The view of a session registers how to draw its output and gets what arrived before. */
export const attachTerminalOutput = (id: string, write: (data: string) => void): Unsubscribe => {
  writers.set(id, write)
  backlog.get(id)?.forEach(write)
  backlog.delete(id)
  return () => writers.delete(id)
}

/** Opens a shell in the open folder (the home folder when none is open) and shows it. */
export const startTerminal = async (bridge: Bridge): Promise<void> => {
  if (starting) return
  starting = true
  try {
    const id = await bridge.terminal.start(get(fileTree)?.path ?? '', 80, 24)
    terminalSessions.update((all) => [...all, { id, number: ++counter, exitCode: null }])
    activeTerminal.set(id)
  } finally {
    starting = false
  }
}

/** Ends a session, with every program it started, and removes its tab. */
export const closeTerminal = async (bridge: Bridge, id: string): Promise<void> => {
  const remaining = get(terminalSessions).filter((session) => session.id !== id)
  terminalSessions.set(remaining)
  backlog.delete(id)
  if (get(activeTerminal) === id) activeTerminal.set(remaining.at(-1)?.id ?? null)
  await bridge.terminal.close(id)
}

/** Shows the Terminal tab, starts the first shell if there is none, and gives it the focus. */
export const showTerminal = async (bridge: Bridge): Promise<void> => {
  outputTab.set('terminal')
  if (get(terminalSessions).length === 0) await startTerminal(bridge)
  terminalFocusRequests.update((count) => count + 1)
}

export const connectTerminal = (bridge: Bridge): Unsubscribe => {
  const offs = [
    bridge.on('terminal:output', ({ id, data }) => deliver(id, data)),
    bridge.on('terminal:exit', ({ id, exitCode }) =>
      terminalSessions.update((all) =>
        all.map((session) => (session.id === id ? { ...session, exitCode } : session))
      )
    )
  ]
  return () => offs.forEach((off) => off())
}

/** Forgets every session (tests). */
export const resetTerminals = (): void => {
  terminalSessions.set([])
  activeTerminal.set(null)
  backlog.clear()
  writers.clear()
  counter = 0
  starting = false
}
