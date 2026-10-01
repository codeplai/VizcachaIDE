// User actions. They call the bridge and update stores; components only call these.
import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { Breakpoint } from '../domain'
import { breakpoints, debugActive, debugStarting } from './debug'
import { activePath } from './files'
import { resetRun } from './run'

const breakpointsOf = (file: string): Breakpoint[] =>
  (get(breakpoints)[file] ?? []).map((line) => ({
    location: { file, line, column: 1 },
    condition: ''
  }))

export const runActiveFile = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (!path) return
  resetRun()
  await bridge.run.run(path, [])
}

export const stopProgram = (bridge: Bridge): Promise<void> => bridge.run.stop()

export const startDebugging = async (bridge: Bridge): Promise<void> => {
  const path = get(activePath)
  if (!path || get(debugActive)) return
  debugStarting.set(true)
  await bridge.debug.start(path, breakpointsOf(path))
}

export const stopDebugging = (bridge: Bridge): Promise<void> => bridge.debug.stop()
export const stepOver = (bridge: Bridge): Promise<void> => bridge.debug.stepOver()
export const stepInto = (bridge: Bridge): Promise<void> => bridge.debug.stepInto()
export const stepOut = (bridge: Bridge): Promise<void> => bridge.debug.stepOut()
export const resumeDebugging = (bridge: Bridge): Promise<void> => bridge.debug.resume()
