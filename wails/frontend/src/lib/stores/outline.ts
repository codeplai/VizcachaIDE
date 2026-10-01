import { derived, writable } from 'svelte/store'
import type { Bridge, Unsubscribe } from '../bridge'
import type { DocumentSymbol, SymbolKind } from '../domain'
import { activePath, activeText } from './files'

export const outline = writable<DocumentSymbol[]>([])

const GLYPHS: Record<SymbolKind, string> = {
  function: 'ƒ',
  method: 'ƒ',
  struct: '▢',
  interface: '◇',
  type: 'T',
  variable: 'v',
  constant: 'c',
  field: '·',
  package: '▤',
  other: '•'
}

/** One character that hints what kind of declaration a row is. */
export const symbolGlyph = (kind: SymbolKind): string => GLYPHS[kind]

const REFRESH_DELAY_MS = 400

/** Asks the language server for the declarations of a file. */
export const refreshOutline = async (bridge: Bridge, path: string | null): Promise<void> => {
  outline.set(path ? await bridge.language.documentSymbols(path) : [])
}

/** Keeps the outline up to date when the file changes, waiting for the user to stop typing. */
export const connectOutline = (bridge: Bridge): Unsubscribe => {
  let timer: ReturnType<typeof setTimeout> | undefined
  const current = derived([activePath, activeText], ([path, text]) => ({ path, text }))
  const off = current.subscribe(({ path }) => {
    clearTimeout(timer)
    timer = setTimeout(() => void refreshOutline(bridge, path), REFRESH_DELAY_MS)
  })
  return () => {
    clearTimeout(timer)
    off()
  }
}
