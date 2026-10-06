import { get } from 'svelte/store'
import type { Bridge } from '../bridge'
import { isTerminalToggle } from '../panels/terminalKeys'
import {
  closeActive,
  debugActive,
  newFile,
  openFileFromDialog,
  openQuickOpen,
  showSearch,
  startNewProject,
  resumeDebugging,
  runActiveFile,
  showTerminal,
  runToCursor,
  saveActiveAs,
  saveActiveFile,
  startDebugging,
  stepInto,
  stepOut,
  stepOver,
  stopDebugging,
  stopProgram,
  toggleAssistant,
  toggleSidebar
} from '../stores'

type Action = (bridge: Bridge) => Promise<void>

const FILE_SHORTCUTS: Record<string, Action> = {
  n: newFile,
  o: openFileFromDialog,
  p: openQuickOpen,
  w: closeActive,
  s: saveActiveFile
}

/** Ctrl (Cmd on macOS) + N, O, P (Quick open), S, W, and Ctrl+Shift+S, N and F (Search). */
const fileShortcut = (event: KeyboardEvent): Action | null => {
  if (!(event.ctrlKey || event.metaKey) || event.altKey) return null
  const key = event.key.toLowerCase()
  if (key === 's' && event.shiftKey) return saveActiveAs
  if (key === 'n' && event.shiftKey) return startNewProject
  if (key === 'f' && event.shiftKey) return async () => showSearch()
  if (event.shiftKey) return null
  return FILE_SHORTCUTS[key] ?? null
}

/** Ctrl+B hides or shows the side panel, Ctrl+Alt+B the Assistant (as in VS Code). */
const panelShortcut = (event: KeyboardEvent): Action | null => {
  if (!(event.ctrlKey || event.metaKey) || event.shiftKey || event.key.toLowerCase() !== 'b')
    return null
  const toggle = event.altKey ? toggleAssistant : toggleSidebar
  return async () => toggle()
}

/** Ctrl+` shows and focuses the terminal. */
const terminalShortcut = (event: KeyboardEvent): Action | null =>
  isTerminalToggle(event) ? showTerminal : null

/** Inside the terminal the keys belong to the shell (Ctrl+W deletes a word, Ctrl+S stops output). */
const inTerminal = (event: KeyboardEvent): boolean =>
  event.target instanceof Element && event.target.closest('.xterm') !== null

const actionFor = (event: KeyboardEvent): Action | null => {
  const toggle = terminalShortcut(event)
  if (toggle) return toggle
  if (inTerminal(event)) return null
  const debugging = get(debugActive)
  const panelAction = panelShortcut(event)
  if (panelAction) return panelAction
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

/** Registers Ctrl+`, F5, F6, F7, F8, F9, Shift+F5, Shift+F6, Ctrl+N, Ctrl+O, Ctrl+S, Ctrl+Shift+S, Ctrl+Shift+N and Ctrl+W. Returns a function that removes them. */
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
