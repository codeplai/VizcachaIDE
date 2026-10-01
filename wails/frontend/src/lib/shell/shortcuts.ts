import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import {
  debugActive,
  resumeDebugging,
  runActiveFile,
  runToCursor,
  saveActiveFile,
  startDebugging,
  stepInto,
  stepOut,
  stepOver,
  stopDebugging,
  stopProgram
} from '../stores'

type Action = (bridge: Bridge) => Promise<void>

const actionFor = (event: KeyboardEvent): Action | null => {
  const debugging = get(debugActive)
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') return saveActiveFile
  switch (event.key) {
    case 'F5':
      return event.shiftKey ? (debugging ? stopDebugging : stopProgram) : runActiveFile
    case 'F6':
      return event.shiftKey ? resumeDebugging : startDebugging
    case 'F10':
      return event.ctrlKey || event.metaKey ? runToCursor : null
    case 'F7':
      return stepOver
    case 'F8':
      return stepInto
    case 'F9':
      return stepOut
    default:
      return null
  }
}

/** Registers F5, F6, F7, F8, F9, Shift+F5, Shift+F6 and Ctrl+S. Returns a function that removes them. */
export const registerShortcuts = (bridge: Bridge): (() => void) => {
  const handler = (event: KeyboardEvent): void => {
    const action = actionFor(event)
    if (!action) return
    event.preventDefault()
    void action(bridge)
  }
  window.addEventListener('keydown', handler)
  return () => window.removeEventListener('keydown', handler)
}
