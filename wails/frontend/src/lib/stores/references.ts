// Find all references: the result list of the References tab.
import { derived, writable } from 'svelte/store'
import type { Bridge } from '../bridge'
import type { Reference, SourceLocation } from '../domain'
import { outputTab } from './layout'
import { showNotice } from './notice'

export interface ReferencesState {
  /** The name that was searched, for the title and the empty message. */
  symbol: string
  items: Reference[]
  loading: boolean
}

/** Null until the first search. */
export const referencesState = writable<ReferencesState | null>(null)

export interface ReferenceGroup {
  file: string
  items: Reference[]
}

/** Groups results by file, keeping the order in which files first appear. */
export const groupByFile = (items: Reference[]): ReferenceGroup[] => {
  const groups = new Map<string, ReferenceGroup>()
  for (const item of items) {
    const file = item.range.start.file
    const group = groups.get(file) ?? { file, items: [] }
    group.items.push(item)
    groups.set(file, group)
  }
  return [...groups.values()]
}

export const referenceGroups = derived(referencesState, (state) => groupByFile(state?.items ?? []))
export const referenceCount = derived(referencesState, (state) => state?.items.length ?? 0)

/** Looks for every use of the symbol at a position and shows them in the References tab. */
export const findReferences = async (
  bridge: Bridge,
  at: SourceLocation,
  symbol: string
): Promise<void> => {
  outputTab.set('references')
  referencesState.set({ symbol, items: [], loading: true })
  try {
    const items = await bridge.language.references(at)
    referencesState.set({ symbol, items, loading: false })
  } catch (error) {
    referencesState.set({ symbol, items: [], loading: false })
    const reason = error instanceof Error ? error.message : String(error)
    showNotice({ messageKey: 'refactor.referencesFailed', values: { reason }, actions: [] })
  }
}
