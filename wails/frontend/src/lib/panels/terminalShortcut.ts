/**
 * The terminal shortcut as the learner's own keyboard shows it. isTerminalToggle takes the key
 * left of 1 whatever it prints (` on a US keyboard, º on a Spanish one, | on a Latin American
 * one) and Ctrl+Ñ only where a Ñ key exists, so the hint names the keys this keyboard has.
 */
import { readable } from 'svelte/store'
import { detectPlatform } from '../platform'

/** What navigator.keyboard.getLayoutMap gives: physical key code → what it prints. */
export type KeyboardLayout = Pick<Map<string, string>, 'get' | 'values'>

/** The label for a layout; without one (the browser does not say), the US key. */
export const terminalShortcutLabel = (layout: KeyboardLayout | null, mac: boolean): string => {
  const modifier = mac ? 'Cmd' : 'Ctrl'
  const leftOfOne = layout?.get('Backquote') ?? '`'
  const keys = [`${modifier}+${leftOfOne}`]
  const hasEnye = layout ? [...layout.values()].some((key) => key.toLowerCase() === 'ñ') : false
  if (hasEnye) keys.push(`${modifier}+Ñ`)
  return keys.join(' / ')
}

interface KeyboardApi {
  getLayoutMap?: () => Promise<KeyboardLayout>
}

const readLayout = async (): Promise<KeyboardLayout | null> => {
  const keyboard = (globalThis.navigator as { keyboard?: KeyboardApi } | undefined)?.keyboard
  if (!keyboard?.getLayoutMap) return null
  return keyboard.getLayoutMap().catch(() => null)
}

/** The shortcut label, read once from the keyboard when first shown. */
export const terminalShortcut = readable(terminalShortcutLabel(null, false), (set) => {
  const mac = detectPlatform() === 'macos'
  set(terminalShortcutLabel(null, mac))
  void readLayout().then((layout) => set(terminalShortcutLabel(layout, mac)))
})
