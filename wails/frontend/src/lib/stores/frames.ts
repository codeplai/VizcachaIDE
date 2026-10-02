import { derived, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { FrameVariables, SourceLocation, StackFrame } from '../domain'
import { debugState } from './debug'

/** The Calls view draws this many of the innermost calls; the rest become a "+N more" row. */
export const MAX_VISIBLE_CALLS = 8

/** One call drawn as a box. `depth` 0 is the outermost visible call. */
export interface CallBox {
  frameId: number
  /** "factorial", not "main.factorial". */
  name: string
  location: SourceLocation | null
  depth: number
  /** The innermost call: where the program is paused. */
  current: boolean
}

export interface CallBoxes {
  /** Outermost first, innermost last. */
  boxes: CallBox[]
  /** Calls left out because there are more than MAX_VISIBLE_CALLS. */
  hidden: number
}

const RUNTIME_PREFIXES = ['runtime.', 'testing.', 'reflect.']

const isUserFrame = (frame: StackFrame): boolean =>
  !RUNTIME_PREFIXES.some((prefix) => frame.function.startsWith(prefix))

/** "main.factorial" becomes "factorial"; "main.(*Cuenta).Depositar" becomes "(*Cuenta).Depositar". */
export const shortFunctionName = (name: string): string =>
  name.startsWith('main.') ? name.slice('main.'.length) : name

/**
 * Turns the stack (innermost first, as the debugger sends it) into nested boxes:
 * user code only, outermost first, keeping the `limit` innermost calls.
 */
export const callBoxes = (frames: StackFrame[], limit = MAX_VISIBLE_CALLS): CallBoxes => {
  const user = frames.filter(isUserFrame)
  const kept = user.slice(0, limit).reverse()
  const boxes = kept.map((frame, depth) => ({
    frameId: frame.frameId,
    name: shortFunctionName(frame.function),
    location: frame.location,
    depth,
    current: frame.frameId === user[0]?.frameId
  }))
  return { boxes, hidden: Math.max(0, user.length - limit) }
}

export const calls = derived(debugState, (state) => callBoxes(state?.frames ?? []))

/** Arguments and locals fetched so far, by frame id; emptied whenever the program moves. */
export const frameDetails = writable<Record<number, FrameVariables>>({})
const requested = new Set<number>()
let generation = 0

const forgetDetails = (): void => {
  generation++
  requested.clear()
  frameDetails.set({})
}

/** Asks the debugger for one frame, once per stop. */
export const loadFrameDetails = async (bridge: Bridge, frameId: number): Promise<void> => {
  if (requested.has(frameId)) return
  requested.add(frameId)
  const stop = generation
  const details = await bridge.debug.frameVariables(frameId)
  if (stop !== generation) return // the program moved while we waited
  frameDetails.update((all) => ({ ...all, [frameId]: details }))
}

/** "n=3, a=5": the arguments a call received, as shown in the box title. */
export const argumentsText = (details: FrameVariables | undefined): string =>
  (details?.arguments ?? []).map((item) => `${item.name}=${item.value}`).join(', ')

export const connectFrames = (bridge: Bridge): Unsubscribe => {
  const offs = [
    bridge.on('debug:stopped', forgetDetails),
    bridge.on('debug:terminated', forgetDetails)
  ]
  return () => offs.forEach((off) => off())
}
