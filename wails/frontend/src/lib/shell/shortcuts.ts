import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import {
  closeActive,
  debugActive,
  newFile,
  openFileFromDialog,
  resumeDebugging,
  runActiveFile,
  runToCursor,
  saveActiveAs,
  saveActiveFile,
  startDebugging,
  stepInto,
  stepOut,
  stepOver,
  stopDebugging,
  stopProgram
} from '../stores'

type Action = (bridge: Bridge) => Promise<void>

const FILE_SHORTCUTS: Record<string, Action> = {
  n: newFile,
  o: openFileFromDialog,
  w: closeActive,
  s: saveActiveFile
}

/** Ctrl (Cmd on macOS) + N, O, S, W, and Ctrl+Shift+S. */
const fileShortcut = (event: KeyboardEvent): Action | null => {
  if (!(event.ctrlKey || event.metaKey) || event.altKey) return null
  const key = event.key.toLowerCase()
  if (key === 's' && event.shiftKey) return saveActiveAs
  if (event.shiftKey) return null
  return FILE_SHORTCUTS[key] ?? null
}

const actionFor = (event: KeyboardEvent): Action | null => {
  const debugging = get(debugActive)
  const fileAction = fileShortcut(event)
  if (fileAction) return fileAction
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

/** Registers F5, F6, F7, F8, F9, Shift+F5, Shift+F6, Ctrl+N, Ctrl+O, Ctrl+S, Ctrl+Shift+S and Ctrl+W. Returns a function that removes them. */
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
