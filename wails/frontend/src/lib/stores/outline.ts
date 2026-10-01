import { writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { DocumentSymbol } from '../domain'

export const outline = writable<DocumentSymbol[]>([])

/** Asks the language server for the declarations of a file. */
export const refreshOutline = async (bridge: Bridge, path: string | null): Promise<void> => {
  outline.set(path ? await bridge.language.documentSymbols(path) : [])
}
