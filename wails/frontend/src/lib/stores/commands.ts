// User actions. They call the bridge and update stores; components only call these.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { Breakpoint } from '../domain'
import { breakpoints, debugActive, debuggedPath, debugOutput, debugStarting } from './debug'
import { codeLanguageOf } from './codeLanguages'
import { activePath, buffers } from './files'
import { assistantOpen, cursor } from './layout'
import { showNotice } from './notice'
import { programArguments } from './programArguments'
import { lastRunConfiguration, pushRunText, resetRun, stoppedByUser } from './run'
import { withToolErrors } from './toolErrors'
import { isUntitled } from './untitled'

const breakpointsOf = (file: string): Breakpoint[] =>
  (get(breakpoints)[file] ?? []).map((line) => ({
    location: { file, line, column: 1 },
    condition: ''
  }))

/** The "Program arguments" text split like a shell, or null (after saying why) if it is invalid. */
const splitProgramArguments = async (bridge: Bridge, text: string): Promise<string[] | null> => {
  if (text.trim() === '') return []
  try {
    return await bridge.run.splitArguments(text)
  } catch {
    showNotice({ messageKey: 'errors.invalidArgs', values: {}, actions: [] })
    return null
  }
}

export const runActiveFile = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  const args = await splitProgramArguments(bridge, get(programArguments))
  if (!path || !args) return
  resetRun()
  await withToolErrors(bridge, codeLanguageOf(path), () =>
    isUntitled(path)
      ? bridge.run.runUntitled(path, get(buffers)[path] ?? '', args)
      : bridge.run.run(path, args)
  )
}

/** Compiles the open saved file without running it (the More menu's Build); the output is a run's. */
export const buildActiveFile = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  const args = await splitProgramArguments(bridge, get(programArguments))
  if (!path || !args || isUntitled(path)) return
  resetRun()
  await withToolErrors(bridge, codeLanguageOf(path), () => bridge.run.build(path, args))
}

export const stopProgram = (bridge: Bridge): Promise<void> => {
  stoppedByUser.set(true)
  return bridge.run.stop()
}

export const startDebugging = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (!path || get(debugActive) || isUntitled(path)) return
  const argsText = get(programArguments)
  if (!(await splitProgramArguments(bridge, argsText))) return
  debuggedPath.set(path)
  debugOutput.set([])
  debugStarting.set(true)
  assistantOpen.set(true) // the variables and the call stack live in the Assistant
  await withToolErrors(bridge, codeLanguageOf(path), () =>
    bridge.debug.start(path, breakpointsOf(path), argsText)
  )
}

export const stopDebugging = (bridge: Bridge): Promise<void> => bridge.debug.stop()
export const stepOver = (bridge: Bridge): Promise<void> => bridge.debug.stepOver()
export const stepInto = (bridge: Bridge): Promise<void> => bridge.debug.stepInto()
export const stepOut = (bridge: Bridge): Promise<void> => bridge.debug.stepOut()
export const resumeDebugging = (bridge: Bridge): Promise<void> => bridge.debug.resume()

/**
 * Sends one typed line to the program; shows it in Output unless the program echoes it (a PTY
 * does). While debugging, the program runs in a terminal and always echoes.
 */
export const sendProgramInput = async (bridge: Bridge, text: string): Promise<void> => {
  const echoed = get(debugActive) || get(debugStarting) || get(lastRunConfiguration)?.echo
  if (!echoed) pushRunText('stdout', `${text}\n`)
  await bridge.run.writeInput(text)
}

/** Runs the paused program until the line with the cursor (Ctrl+F10). */
export const runToCursor = async (bridge: Bridge): Promise<void> => {
  const file = get(activePath)
  if (!file || !get(debugActive)) return
  const { line, column } = get(cursor)
  await bridge.debug.runTo({ file, line, column })
}
