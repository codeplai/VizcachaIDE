import type { ITheme } from '@xterm/xterm'

const ANSI_NAMES = [
  'black',
  'red',
  'green',
  'yellow',
  'blue',
  'magenta',
  'cyan',
  'white',
  'bright-black',
  'bright-red',
  'bright-green',
  'bright-yellow',
  'bright-blue',
  'bright-magenta',
  'bright-cyan',
  'bright-white'
] as const

/** The xterm colors, read from the app tokens (the window colors and the 16 ANSI names). */
export const terminalTheme = (): ITheme => {
  const style = getComputedStyle(document.documentElement)
  const token = (name: string): string => style.getPropertyValue(name).trim()
  const [black, red, green, yellow, blue, magenta, cyan, white, ...bright] = ANSI_NAMES.map(
    (name) => token(`--ansi-${name}`)
  )
  return {
    background: token('--win'),
    foreground: token('--ink'),
    cursor: token('--go'),
    cursorAccent: token('--win'),
    selectionBackground: token('--go-soft'),
    selectionInactiveBackground: token('--go-soft'),
    black,
    red,
    green,
    yellow,
    blue,
    magenta,
    cyan,
    white,
    brightBlack: bright[0],
    brightRed: bright[1],
    brightGreen: bright[2],
    brightYellow: bright[3],
    brightBlue: bright[4],
    brightMagenta: bright[5],
    brightCyan: bright[6],
    brightWhite: bright[7]
  }
}

/** The terminal font: the editor's monospace font. */
export const terminalFont = (): string =>
  getComputedStyle(document.documentElement).getPropertyValue('--mono').trim() || 'monospace'

/** Calls `changed` when the theme switches (Settings, or the system when it is "system"). */
export const watchTheme = (changed: () => void): (() => void) => {
  const observer = new MutationObserver(changed)
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  const system = window.matchMedia?.('(prefers-color-scheme: dark)')
  system?.addEventListener('change', changed)
  return () => {
    observer.disconnect()
    system?.removeEventListener('change', changed)
  }
}
