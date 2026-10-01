import { derived, get, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { DebugState, Variable } from '../domain'
import type { DebugOutputPayload } from '../events'

export const debugState = writable<DebugState | null>(null)
export const debugActive = writable(false)
export const debugStarting = writable(false)
export const debugOutput = writable<DebugOutputPayload[]>([])
/** Breakpoint lines per file (1-based). */
export const breakpoints = writable<Record<string, number[]>>({})

export const currentFrame = derived(debugState, (state) => state?.frames[0] ?? null)
export const currentLine = derived(currentFrame, (frame) => frame?.location?.line ?? null)

/** Puts `children` under the variable whose `reference` matches, at any depth. */
export const mergeChildren = (
  variables: Variable[],
  reference: number,
  children: Variable[]
): Variable[] =>
  variables.map((variable) => {
    if (variable.reference === reference) return { ...variable, children }
    if (variable.children.length === 0) return variable
    return { ...variable, children: mergeChildren(variable.children, reference, children) }
  })

export const toggleBreakpointLine = (
  all: Record<string, number[]>,
  file: string,
  line: number
): Record<string, number[]> => {
  const lines = all[file] ?? []
  const next = lines.includes(line) ? lines.filter((l) => l !== line) : [...lines, line].sort()
  return { ...all, [file]: next }
}

export const toggleBreakpoint = async (
  bridge: Bridge,
  file: string,
  line: number
): Promise<void> => {
  breakpoints.update((all) => toggleBreakpointLine(all, file, line))
  await bridge.debug.setBreakpoints(file, get(breakpoints)[file] ?? [])
}

export const connectDebug = (bridge: Bridge): Unsubscribe => {
  const offs = [
    bridge.on('debug:stopped', (state) => {
      debugStarting.set(false)
      debugActive.set(true)
      debugState.set(state)
    }),
    bridge.on('debug:variables', ({ reference, variables }) => {
      debugState.update((state) =>
        state
          ? { ...state, variables: mergeChildren(state.variables, reference, variables) }
          : state
      )
    }),
    bridge.on('debug:output', (line) => debugOutput.update((lines) => [...lines, line])),
    bridge.on('debug:terminated', () => {
      debugStarting.set(false)
      debugActive.set(false)
      debugState.set(null)
      debugOutput.set([])
    })
  ]
  return () => offs.forEach((off) => off())
}
