/** What the terminal asks of the clipboard. */
export interface TerminalClipboard {
  hasSelection: () => boolean
  copy: () => void
  paste: () => void
}

/**
 * Ctrl+` (or Cmd+`) shows the terminal: the global shortcut handles it, xterm must not eat it.
 * The key left of 1 counts whatever it prints (º on a Spanish keyboard, where ` is a dead key),
 * and Ctrl+Ñ too, the shortcut VS Code uses with Spanish keyboards.
 */
export const isTerminalToggle = (event: KeyboardEvent): boolean =>
  (event.ctrlKey || event.metaKey) &&
  !event.shiftKey &&
  !event.altKey &&
  (event.key === '`' || event.code === 'Backquote' || event.key.toLowerCase() === 'ñ')

/**
 * The xterm key handler: returns false for the keys the terminal must not send to the shell.
 * Ctrl+C goes to the shell (interrupt) unless text is selected, which it copies instead;
 * Ctrl+Shift+C and Ctrl+Shift+V copy and paste, as in Windows Terminal and VS Code.
 */
export const terminalKeyHandler =
  (clipboard: TerminalClipboard) =>
  (event: KeyboardEvent): boolean => {
    if (event.type !== 'keydown') return true
    if (isTerminalToggle(event)) return false
    const key = event.key.toLowerCase()
    const modifier = event.ctrlKey || event.metaKey
    if (!modifier || event.altKey) return true
    const copying = key === 'c' && (event.shiftKey || clipboard.hasSelection())
    const pasting = key === 'v' && (event.shiftKey || event.metaKey)
    if (!copying && !pasting) return true
    event.preventDefault()
    if (copying) clipboard.copy()
    else clipboard.paste()
    return false
  }
