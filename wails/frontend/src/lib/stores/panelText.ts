// Copy, paste and clear for the text panels (Output, Problems, Console), used by their
// right-click menus. The clipboard goes through the bridge: WebView2 does not let a page read
// the clipboard by itself.
import { get, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import { consoleEntries } from './console'
import { debugOutput } from './debug'
import { dismissNotice, notice, showNotice } from './notice'
import { runLines, runResult, stoppedByUser } from './run'

const COPIED_NOTICE_MS = 2500

/** What the user is typing in the program's input line (Output panel, while it runs). */
export const stdinDraft = writable('')

/** The text selected in the page right now ("" when nothing is selected). */
export const selectedText = (): string => window.getSelection()?.toString() ?? ''

const flashCopied = (): void => {
  showNotice({ messageKey: 'panels.copied', values: {}, actions: [], tone: 'info' })
  const shown = get(notice)
  setTimeout(() => get(notice) === shown && dismissNotice(), COPIED_NOTICE_MS)
}

export const copyText = async (bridge: Bridge, text: string): Promise<void> => {
  if (!text) return
  try {
    await bridge.system.writeClipboard(text)
  } catch {
    showNotice({ messageKey: 'panels.copyFailed', values: {}, actions: [] })
    return
  }
  flashCopied()
}

export const readClipboard = async (bridge: Bridge): Promise<string> => {
  try {
    return await bridge.system.readClipboard()
  } catch {
    return ''
  }
}

/** Selects all the text inside `element`, as Ctrl+A would do for that panel only. */
export const selectAllIn = (element: HTMLElement | undefined): void => {
  if (!element) return
  const range = document.createRange()
  range.selectNodeContents(element)
  const selection = window.getSelection()
  selection?.removeAllRanges()
  selection?.addRange(range)
}

/** Empties the Output panel (the program, if it is still running, keeps going). */
export const clearOutput = (): void => {
  runLines.set([])
  runResult.set(null)
  stoppedByUser.set(false)
  debugOutput.set([])
}

/** Empties the console's screen; the variables and functions defined so far are kept. */
export const clearConsoleScreen = (): void => consoleEntries.set([])
